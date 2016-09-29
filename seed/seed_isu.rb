require 'json'

data_str = ""
File.open("data/main001.json") do |file|
  data_str += file.read
end

data = JSON.parse(data_str)

output = []

data.each do |s|
  o = {
    width: rand(40) + 10,
    red: rand(100) + 100,
    green: rand(100) + 100,
    blue: rand(100) + 100,
    alpha: (rand(5) + 5) / 10.0,
    points: [],
  }
  s['points'].each do |p|
    o[:points].push({x: p["x"] + rand(3) - 1.5, y: p["y"] + rand(3) - 1.5})
  end
  output.push(o)
end

puts JSON.dump(output)
