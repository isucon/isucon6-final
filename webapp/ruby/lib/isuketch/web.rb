require 'json'
require 'time'

require 'mysql2'
require 'mysql2-cs-bind'
require 'sinatra/base'

module Isuketch
  class Web < ::Sinatra::Base
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
        |, room_id)
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

      def to_rfc_3339(str)
        dt = Time.parse(str)
        Time.iso8601(dt)
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
        |, csrf_token)
      end

      def get_stroke_points(stroke_id)
        db.xquery(%|
          SELECT `id`, `stroke_id`, `x`, `y`
          FROM `points`
          WHERE `stroke_id` = ?
          ORDER BY `id` ASC
        |, stroke_id)
      end

      def get_watcher_count(room_id, greater_than_id)
        db.xquery(%|
          SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at`
          FROM `strokes`
          WHERE `room_id` = ?
            AND `id` > ?
          ORDER BY `id` ASC;
        |, room_id, greater_than_id)
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
      |, id)

      content_type :json
      JSON.generate(
        token: token[:csrf_token],
      )
    end

    get '/api/room' do
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

      posted_room = JSON.parse(request.body)
      if (posted_room['name'] || '').empty? || (posted_room['canvas_width'] || '').empty? || (posted_room['canvas_height'] || '').empty?
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
  end
end
