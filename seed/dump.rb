#!/usr/bin/env ruby

require 'json'

# 1. docker exec webapp_mysql_1 sh -c 'export MYSQL_PWD=password; mysql -uroot < /sql/create_database.sql && mysql -uroot -Disuchannel < /sql/schema.sql'
# 2. 部屋を作って描く
# 3. mysql -h127.0.0.1 -uisucon -P33060 -Disuchannel -p -N -e 'SELECT strokes.id, width, red, green, blue, alpha, x, y FROM strokes JOIN points ON strokes.id = points.stroke_id ORDER BY strokes.id, points.id' | ruby dump.rb > data/main001.json
# という感じで、 mainXXX.json と subXXX.json をいっぱい作る
# mainはキャンバスの中央部に大きく描く、subはキャンバスの端のほうに描く
# 初期データやベンチマーカーが投稿するデータはそれらの色を変えたり座標を変えたりしながら作る

strokes = []
last_id = nil

while $_ = STDIN.gets
  stroke_id, width, red, green, blue, alpha, x, y = $_.strip.split("\t")

  if last_id != stroke_id
    last_id = stroke_id

    strokes << {
      width: width.to_i,
      red: red.to_i,
      green: green.to_i,
      blue: blue.to_i,
      alpha: alpha.to_f,
      points: [],
    }
  end

  strokes.last[:points] << {
    x: x.to_f,
    y: y.to_f,
  }
end

puts JSON.dump(strokes)
