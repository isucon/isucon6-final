#!/usr/bin/env ruby

# initial_data.sqlを生成するためのスクリプト
# data/main001.jsonが既にあると仮定して、
# generate_initial_data.rb > ../webapp/sql/initial_data.sql
# とやって、あとは ../webapp/sql/README.md の手順に従って初期データを用意する

require 'json'

json = File.read(File.dirname(__FILE__) + '/data/main001.json') # TODO: 今後mainXXXとsubXXXの複数ファイルになる
strokes = JSON.parse(json, {:symbolize_names => true})

stroke_id = 0
point_id = 0

room_sql = []
stroke_sql = []
point_sql = []

(1..1000).each do |room_id|
  room_sql << "(#{room_id}, 'ひたすら椅子を描くスレ【#{room_id}】', '2016-01-01 00:00:00' + INTERVAL #{room_id} SECOND, 1028, 768)"

  strokes.each do|stroke|
    stroke_id += 1
    stroke_sql << "(#{stroke_id}, #{room_id}, '2016-01-01 00:00:00' + INTERVAL #{stroke_id} SECOND, #{stroke[:width]}, #{stroke[:red]}, #{stroke[:green]}, #{stroke[:blue]}, #{stroke[:alpha]})"

    stroke[:points].each do |point|
      point_id += 1
      point_sql << "(#{point_id}, #{stroke_id}, #{point[:x]}, #{point[:y]})"
    end
  end
end

puts "SET GLOBAL max_allowed_packet=1073741824;"
puts "BEGIN;"

puts "INSERT INTO `rooms` (`id`, `name`, `created_at`, `canvas_width`, `canvas_height`) VALUES"
puts room_sql.join(",\n") + ";"

puts "INSERT INTO `strokes` (`id`, `room_id`, `created_at`, `width`, `red`, `green`, `blue`, `alpha`) VALUES"
puts stroke_sql.join(",\n") + ";"

puts "INSERT INTO `points` (`id`, `stroke_id`, `x`, `y`) VALUES"
puts point_sql.join(",\n") + ";"

puts "COMMIT;"
