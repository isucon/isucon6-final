<?php

$stderr = fopen('php://stderr', 'w');
fwrite($stderr, var_export($_SERVER, true) . "\n");

$uri = $_SERVER['REQUEST_URI'];
if ($uri === '/api/rooms') {
  $rooms = [];
  for ($i = 1; $i <= 10; $i++) {
    $rooms[] = ['id' => $i, 'name' => 'ひたすら椅子を描く部屋'];
  }
  echo json_encode(['rooms' => $rooms]);
} elseif ($uri === '/api/csrf_token') {
  echo json_encode(['token' => md5(rand())]);
} elseif (preg_match('@^/api/rooms/(\d+)$@', $uri, $m)) {
  $room = ['id' => (int)$m[1], 'name' => 'ひたすら椅子を描く部屋', 'strokes' => [
    [
      'id' => microtime(true) * 1000000,
      'red' => 128,
      'green' => 128,
      'blue' => 128,
      'alpha' => 0.5,
      'width' => 5,
      'points' => [
        ['x' => 1, 'y' => 2],
        ['x' => 10, 'y' => 31],
        ['x' => 44, 'y' => 19],
        ['x' => 81, 'y' => 61],
        ['x' => 115, 'y' => 118],
        ['x' => 174, 'y' => 71],
        ['x' => 227, 'y' => 124],
        ['x' => 365, 'y' => 243],
      ]
    ]
  ]];
  echo json_encode(['room' => $room]);
} elseif (preg_match('@^/api/rooms/(\d+)/strokes$@', $uri, $m)) {
  while(true) {
    echo 'data: ' . json_encode([
      'id' => microtime(true) * 1000000,
      'red' => 128,
      'green' => 128,
      'blue' => 128,
      'alpha' => 0.5,
      'width' => 5,
      'points' => [
        ['x' => 1, 'y' => 2],
        ['x' => 10, 'y' => 31],
        ['x' => 44, 'y' => 19],
        ['x' => 81, 'y' => 61],
        ['x' => 115, 'y' => 118],
        ['x' => 174, 'y' => 71],
        ['x' => 227, 'y' => 124],
        ['x' => 365, 'y' => 243],
      ]
    ]) . "\n\n";
    ob_flush();
    flush();
    sleep(1);
  }
}
