<?php

header('Content-Type: application/json');
$stderr = fopen('php://stderr', 'w');
fwrite($stderr, var_export($_SERVER, true) . "\n");

switch($_SERVER['REQUEST_URI']) {
case '/api/csrf_token':
    echo json_encode(['token' => md5(rand())]);
    break;
case '/api/rooms':
    $rooms = [];
    for ($i = 1; $i <= 10; $i++) {
        $rooms[] = ['id' => $i, 'name' => 'ひたすら椅子を描く部屋', 'watcherCount' => 10];
    }
    echo json_encode(['rooms' => $rooms]);
    break;
case '/api/rooms/1':
    $room = ['id' => 1, 'name' => 'ひたすら椅子を描く部屋', 'watcherCount' => 10];
    echo json_encode(['room' => $room]);
    break;
}
