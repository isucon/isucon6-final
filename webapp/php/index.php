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
  $room = ['id' => (int)$m[1], 'name' => 'ひたすら椅子を描く部屋', 'strokes' => []];
  echo json_encode(['room' => $room]);
} elseif (preg_match('@^/api/rooms/(\d+)/strokes$@', $uri, $m)) {
  while(true) {
    echo 'data: ' . json_encode(['stroke' => 'black', 'stroke-width' => 1, 'fill' => 'none', 'd' => 'M 0 200 L 25 50 l 50 -25 V 175 H 100 v -50 h 25 m 25 0 l 50 0 L 175 175 z']) . "\n\n";
    ob_flush();
    flush();
    sleep(1);
  }
}
