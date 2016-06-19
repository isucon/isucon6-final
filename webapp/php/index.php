<?php

$uri = $_SERVER['REQUEST_URI'];
if ($uri === '/api/rooms') {
  $rooms = [];
  for ($i = 1; $i <= 10; $i++) {
    $rooms[] = ['id' => $i, 'name' => 'ひたすら椅子を描く部屋'];
  }
  echo json_encode(['rooms' => $rooms]);
} elseif ($uri === '/api/rooms/1') {
  $room = ['id' => 1, 'name' => 'ひたすら椅子を描く部屋', 'strokes' => []];
  echo json_encode(['room' => $room]);
} elseif ($uri === '/api/rooms/1/strokes') {
  while(true) {
    echo 'data: ' . json_encode(['stroke' => 'black', 'stroke-width' => 1, 'fill' => 'none', 'd' => 'M 0 200 L 25 50 l 50 -25 V 175 H 100 v -50 h 25 m 25 0 l 50 0 L 175 175 z']) . "\n\n";
    ob_flush();
    flush();
    sleep(1);
  }
}
