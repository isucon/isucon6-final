#!/usr/bin/env ruby

# initial_data.sqlを生成するためのスクリプト
# data/foo.jsonが既にあると仮定して、
# generate_initial_data.rb > ../webapp/sql/initial_data.sql
# とやって、あとは ../webapp/sql/README.md の手順に従って初期データを用意する

require 'json'

ROOM_ID_MAX = 1000
TOKENS_ID_MAX = 5000

json = File.read(File.dirname(__FILE__) + '/data/foo.json')
strokes = JSON.parse(json, {:symbolize_names => true})

stroke_id = 0
point_id = 0

room_sql = []
stroke_sql = []
point_sql = []
room_owner_sql = []

(1..ROOM_ID_MAX).each do |room_id|
  room_sql.push("(#{room_id}, 'ひたすら椅子を描くスレ【#{room_id}】', '2016-01-01 00:00:00.#{sprintf('%06d', rand(1000000))}' + INTERVAL #{room_id} SECOND, 1028, 768)")

  strokes.each do |stroke|
    stroke_id += 1

    s = {
      width: rand(40) + 10,
      red: rand(100) + 100,
      green: rand(100) + 100,
      blue: rand(100) + 100,
      alpha: (rand(5) + 5) / 10.0,
    }

    stroke_sql.push("(#{stroke_id}, #{room_id}, '2016-01-01 00:00:00.#{sprintf('%06d', rand(1000000))}' + INTERVAL #{stroke_id} SECOND, #{s[:width]}, #{s[:red]}, #{s[:green]}, #{s[:blue]}, #{s[:alpha]})")

    stroke[:points].each do |p|
      point_id += 1
      point_sql.push("(#{point_id}, #{stroke_id}, #{p[:x] + rand(3) - 1.5}, #{p[:y] + rand(3) - 1.5})")
    end
  end

  # ownerはtokenの作成日時よりも後のユーザーでないとおかしい
  room_owner_sql.push("(#{room_id}, #{rand(room_id - 1) + 1})")
end

token_sql = []
room_watcher_sql = []

(1..TOKENS_ID_MAX).each do |token_id|
  token_sql.push("(SHA2(CONCAT(RAND(), UUID_SHORT()), 256), '2016-01-01 00:00:00.#{sprintf('%06d', rand(1000000))}' + INTERVAL #{token_id} SECOND)")
  room_id = rand(ROOM_ID_MAX) + 1

  # room/tokenの作成日時よりも前にwatchするのはおかしい
  interval = ((room_id > token_id) ? room_id : token_id) + rand(100) + 1
  date_at = "'2016-01-01 00:00:00.#{sprintf('%06d', rand(1000000))}' + INTERVAL #{interval} SECOND"
  room_watcher_sql.push("(#{room_id}, #{token_id}, #{date_at}, #{date_at})")
end

puts "SET NAMES utf8mb4;"

puts "use `isuketch`;"

puts "BEGIN;"

puts "INSERT INTO `rooms` (`id`, `name`, `created_at`, `canvas_width`, `canvas_height`) VALUES"
puts room_sql.join(",\n") + ";"

puts "INSERT INTO `strokes` (`id`, `room_id`, `created_at`, `width`, `red`, `green`, `blue`, `alpha`) VALUES"
puts stroke_sql.join(",\n") + ";"

puts "INSERT INTO `points` (`id`, `stroke_id`, `x`, `y`) VALUES"
puts point_sql.join(",\n") + ";"

puts "INSERT INTO `tokens` (`csrf_token`, `created_at`) VALUES"
puts token_sql.join(",\n") + ";"

puts "INSERT INTO `room_watchers` (`room_id`, `token_id`, `created_at`, `updated_at`) VALUES"
puts room_watcher_sql.join(",\n") + ";"

puts "INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES"
puts room_owner_sql.join(",\n") + ";"

puts "COMMIT;"
