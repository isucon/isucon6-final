CREATE TABLE `rooms` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `canvas_width` int(10) unsigned NOT NULL,
  `canvas_height` int(10) unsigned NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `strokes` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `room_id` bigint(20) NOT NULL,
  `width` tinyint(3) unsigned NOT NULL,
  `red` tinyint(3) unsigned NOT NULL,
  `green` tinyint(3) unsigned NOT NULL,
  `blue` tinyint(3) unsigned NOT NULL,
  `alpha` decimal(3,2) unsigned NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `room_id` (`room_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `points` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `stroke_id` bigint(20) NOT NULL,
  `x` decimal(10,4) NOT NULL,
  `y` decimal(10,4) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `room_id` (`stroke_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tokens` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `token` varbinary(128) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
