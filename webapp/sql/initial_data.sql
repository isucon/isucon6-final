INSERT INTO `room` (`id`, `name`, `created_at`, `canvas_width`, `canvas_height`) VALUES
(1, 'ひたすら椅子を描くスレ', '2016-01-01 00:00:00', 1028, 768);

INSERT INTO `stroke` (`id`, `room_id`, `created_at`, `stroke_width`, `red`, `green`, `blue`, `alpha`) VALUES
(1, 1, '2016-01-01 00:00:01', 20, 128, 128, 128, 128);

INSERT INTO `path` (`id`, `stroke_id`, `x`, `y`) VALUES
(1, 1, 100, 100),
(2, 1, 100, 200),
(3, 1, 200, 200),
(4, 1, 150, 150);
