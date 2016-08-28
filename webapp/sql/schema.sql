CREATE TABLE `rooms` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `canvas_width` int(10) NOT NULL,
  `canvas_height` int(10) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `stroke` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `room_id` int(10) NOT NULL,
  `created_at` datetime NOT NULL,
  `stroke_width` tinyint(3) NOT NULL,
  `red` tinyint(3) NOT NULL,
  `blue` tinyint(3) NOT NULL,
  `green` tinyint(3) NOT NULL,
  `alpha` tinyint(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `room_id` (`room_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `path` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `stroke_id` int(10) NOT NULL,
  `x` float NOT NULL,
  `y` float NOT NULL,
  PRIMARY KEY (`id`),
  KEY `room_id` (`stroke_id`)
);
