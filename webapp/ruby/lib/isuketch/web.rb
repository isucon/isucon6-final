require 'json'

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
  end
end
