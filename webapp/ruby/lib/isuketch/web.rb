require 'json'
require 'time'

require 'mysql2'
require 'mysql2-cs-bind'
require 'sinatra/base'

module Isuketch
  class Web < ::Sinatra::Base
    JSON.load_default_options[:symbolize_names] = true

    helpers do
      def db
        Thread.current[:db] ||=
          begin
            host = ENV['MYSQL_HOST'] || 'localhost'
            port = ENV['MYSQL_PORT'] || '3306'
            user = ENV['MYSQL_USER'] || 'root'
            pass = ENV['MYSQL_PASS'] || ''
            name = 'isuketch'
            mysql = Mysql2::Client.new(
              username: user,
              password: pass,
              database: name,
              host: host,
              port: port,
              encoding: 'utf8mb4',
              init_command: %|
                SET TIME_ZONE = 'UTC'
              |,
            )
            mysql.query_options.update(symbolize_keys: true)
            mysql
          end
      end

      def get_room(room_id)
        db.xquery(%|
          SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at`
          FROM `rooms`
          WHERE `id` = ?
        |, room_id).first
      end

      def get_strokes(room_id, greater_than_id)
        db.xquery(%|
          SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at`
          FROM `strokes`
          WHERE `room_id` = ?
            AND `id` > ?
          ORDER BY `id` ASC;
        |, room_id, greater_than_id)
      end

      def to_room_json(room)
        {
          id: room[:id].to_i,
          name: room[:name],
          canvas_width: room[:canvas_width].to_i,
          canvas_height: room[:canvas_height].to_i,
          created_at: (room[:created_at] ? to_rfc_3339(room[:created_at]) : ''),
          strokes: (room[:strokes] ? room[:strokes].map {|stroke| to_stroke_json(stroke) } : []),
          stroke_count: room[:stroke_count] || 0,
          watcher_count: room[:watcher_count] || 0,
        }
      end

      def to_rfc_3339(dt)
        dt.strftime('%Y-%m-%dT%H:%M:%S.%6N%:z').
          sub(/\+00:00/, 'Z') # RFC3339では+00:00のときはZにするという仕様
      end

      def to_stroke_json(stroke)
        {
          id: stroke[:id].to_i,
          room_id: stroke[:room_id].to_i,
          width: stroke[:width].to_i,
          red: stroke[:red].to_i,
          green: stroke[:green].to_i,
          blue: stroke[:blue].to_i,
          alpha: stroke[:alpha].to_i,
          points: (stroke[:points] ? stroke[:points].map {|point| to_point_json(point) } : []),
          created_at: (stroke[:created_at] ? to_rfc_3339(stroke[:created_at]) : ''),
        }
      end

      def to_point_json(point)
        {
          id: point[:id].to_i,
          stroke_id: point[:stroke_id].to_i,
          x: point[:x].to_i,
          y: point[:y].to_i,
        }
      end

      def check_token(csrf_token)
        db.xquery(%|
          SELECT `id`, `csrf_token`, `created_at` FROM `tokens`
          WHERE `csrf_token` = ?
            AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY
        |, csrf_token).first
      end

      def get_stroke_points(stroke_id)
        db.xquery(%|
          SELECT `id`, `stroke_id`, `x`, `y`
          FROM `points`
          WHERE `stroke_id` = ?
          ORDER BY `id` ASC
        |, stroke_id)
      end

      def get_watcher_count(room_id)
        db.xquery(%|
          SELECT COUNT(*) AS `watcher_count`
          FROM `room_watchers`
          WHERE `room_id` = ?
            AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND
        |, room_id).first[:watcher_count].to_i
      end
    end

    post '/api/csrf_token' do
      db.xquery(%|
        INSERT INTO `tokens` (`csrf_token`)
        VALUES
        (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))
      |)

      id = db.last_id
      token = db.xquery(%|
        SELECT `id`, `csrf_token`, `created_at`
        FROM `tokens`
        WHERE `id` = ?
      |, id).first

      content_type :json
      JSON.generate(
        token: token[:csrf_token],
      )
    end

    get '/api/rooms' do
      results = db.xquery(%|
        SELECT `room_id`, MAX(`id`) AS `max_id`
        FROM `strokes`
        GROUP BY `room_id`
        ORDER BY `max_id` DESC
        LIMIT 100
      |)

      rooms = results.map {|res|
        room = get_room(res[:room_id])
        room[:stroke_count] = get_strokes(room[:id], 0).size
        room
      }

      content_type :json
      JSON.generate(
        rooms: rooms.map {|room| to_room_json(room) },
      )
    end

    post '/api/rooms' do
      token = check_token(request.env['HTTP_X_CSRF_TOKEN'])
      unless token
        halt(400, {'Content-Type' => 'application/json'}, JSON.generate(
          error: 'トークンエラー。ページを再読み込みしてください。'
        ))
      end

      posted_room = JSON.load(request.body)
      if (posted_room[:name] || '').empty? || !posted_room[:canvas_width] || !posted_room[:canvas_height]
        halt(400, {'Content-Type' => 'application/json'}, JSON.generate(
          error: 'リクエストが正しくありません。'
        ))
      end

      room_id = nil
      begin
        db.xquery(%|BEGIN|)

        db.xquery(%|
          INSERT INTO `rooms`
          (`name`, `canvas_width`, `canvas_height`)
          VALUES
          (?, ?, ?)
        |, posted_room['name'], posted_room['canvas_width'], posted_room['canvas_height'])
        room_id = db.last_id

        db.xquery(%|
          INSERT INTO `room_owners`
          (`room_id`, `token_id`)
          VALUES
          (?, ?)
        |, room_id, token[:id])
      rescue
        db.xquery(%|ROLLBACK|)
        halt(500, ['Content-Type'], JSON.generate(
          error: 'エラーが発生しました。'
        ))
      else
        db.xquery(%|COMMIT|)
      end

      room = get_room(room_id)
      content_type :json
      JSON.generate(
        room: to_room_json(room)
      )
    end

    get '/api/rooms/:id' do |id|
      room = get_room(id)
      unless room
        halt(404, ['Content-Type'], JSON.generate(
          error: 'この部屋は存在しません。'
        ))
      end

      strokes = get_strokes(room[:id], 0)
      strokes.each do |stroke|
        stroke[:points] = get_stroke_points(stroke[:id])
      end
      room[:strokes] = strokes
      room[:watcher_count] = get_watcher_count(room[:id])

      content_type :json
      JSON.generate(
        room: to_room_json(room)
      )
    end

    post '/api/strokes/rooms/:id' do |id|
      token = check_token(request.env['HTTP_X_CSRF_TOKEN'])
      unless token
        halt(400, {'Content-Type' => 'application/json'}, JSON.generate(
          error: 'トークンエラー。ページを再読み込みしてください。'
        ))
      end

      room = get_room(id)
      unless room
        halt(404, ['Content-Type'], JSON.generate(
          error: 'この部屋は存在しません。'
        ))
      end

      posted_stroke = JSON.load(request.body)
      if !posted_stroke[:width] || !posted_stroke[:points]
        halt(400, {'Content-Type' => 'application/json'}, JSON.generate(
          error: 'リクエストが正しくありません。'
        ))
      end

      stroke_count = get_strokes(room[:id], 0).count
      if stroke_count == 0
        count = db.xquery(%|
          SELECT COUNT(*) as cnt FROM `room_owners`
          WHERE `room_id` = ?
            AND `token_id` = ?
        |, room[:id], token[:id]).first[:cnt].to_i
        if count == 0
          halt(400, {'Content-Type' => 'application/json'}, JSON.generate(
            error: '他人の作成した部屋に1画目を描くことはできません'
          ))
        end
      end

      stroke_id = nil
      begin
        db.xquery(%| BEGIN |)

        db.xquery(%|
          INSERT INTO `strokes`
          (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)
          VALUES
          (?, ?, ?, ?, ?, ?)
        |, room[:id], posted_stroke[:width], posted_stroke[:red], posted_stroke[:green], posted_stroke[:blue], posted_stroke[:alpha])
        stroke_id = db.last_id

        posted_stroke[:points].each do |point|
          db.xquery(%|
            INSERT INTO `points`
            (`stroke_id`, `x`, `y`)
            VALUES
            (?, ?, ?)
          |, stroke_id, point[:x], point[:y])
        end
      rescue
        db.xquery(%| ROLLBACK |)
        halt(500, {'Content-Type' => 'application/json'}, JSON.generate(
          error: 'エラーが発生しました。'
        ))
      else
        db.xquery(%| COMMIT |)
      end

      stroke = db.xquery(%|
        SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at`
        FROM `strokes`
        WHERE `id`= ?
      |, stroke_id).first
      stroke[:points] = get_stroke_points(stroke_id)

      content_type :json
      JSON.generate(
        stroke: to_stroke_json(stroke)
      )
    end

  end
end
