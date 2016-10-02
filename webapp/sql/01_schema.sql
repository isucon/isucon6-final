use `isuketch`;

CREATE TABLE `rooms` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `canvas_width` int(10) unsigned NOT NULL,
  `canvas_height` int(10) unsigned NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
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
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
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
  `csrf_token` varchar(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY (`csrf_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `room_watchers` (
  `room_id` bigint(20) NOT NULL,
  `token_id` bigint(20) NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY (`room_id`, `token_id`),
  KEY (`room_id`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `room_owners` (
  `room_id` bigint(20) NOT NULL,
  `token_id` bigint(20) NOT NULL,
  UNIQUE KEY (`room_id`, `token_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
