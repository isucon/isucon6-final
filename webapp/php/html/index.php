<?php

require __DIR__ . '/../vendor/autoload.php';

function getPDO() {
    $host = getenv('MYSQL_HOST') ?: 'localhost';
    $port = getenv('MYSQL_PORT') ?: 3306;
    $user = getenv('MYSQL_USER') ?: 'root';
    $pass = getenv('MYSQL_PASS') ?: '';
    $dbname = 'isuketch';
    $dbh = new PDO("mysql:host={$host};port={$port};dbname={$dbname};charset=utf8mb4", $user, $pass);
    $dbh->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $dbh->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
    return $dbh;
}

function execute($dbh, $sql, array $params = []) {
    $stmt = $dbh->prepare($sql);
    $stmt->execute($params);
    return (int)$dbh->lastInsertId();
}

function selectOne($dbh, $sql, array $params = []) {
    $stmt = $dbh->prepare($sql);
    $stmt->execute($params);
    return array_pop($stmt->fetchAll());
}

function selectAll($dbh, $sql, array $params = []) {
    $stmt = $dbh->prepare($sql);
    $stmt->execute($params);
    return $stmt->fetchAll();
}

function printAndFlush($content) {
    print($content);
    ob_flush();
    flush();
}

function typeCastPointData($data) {
    return [
        'id' => (int)$data['id'],
        'stroke_id' => (int)$data['stroke_id'],
        'x' => (float)$data['x'],
        'y' => (float)$data['y'],
    ];
}

function toRFC3339Micro($date) {
    // RFC3339では+00:00のときはZにするという仕様だが、PHPの"P"は準拠していないため
    return str_replace('+00:00', 'Z', date_create($date)->format("Y-m-d\TH:i:s.uP"));
}

function typeCastStrokeData($data) {
    return [
        'id' => (int)$data['id'],
        'room_id' => (int)$data['room_id'],
        'width' => (int)$data['width'],
        'red' => (int)$data['red'],
        'green' => (int)$data['green'],
        'blue' => (int)$data['blue'],
        'alpha' => (float)$data['alpha'],
        'points' => isset($data['points']) ? array_map('typeCastPointData', $data['points']) : [],
        'created_at' => isset($data['created_at']) ? toRFC3339Micro($data['created_at']) : '',
    ];
}

function typeCastRoomData($data) {
    return [
        'id' => (int)$data['id'],
        'name' => $data['name'],
        'canvas_width' => (int)$data['canvas_width'],
        'canvas_height' => (int)$data['canvas_height'],
        'created_at' => isset($data['created_at']) ? toRFC3339Micro($data['created_at']) : '',
        'strokes' => isset($data['strokes']) ? array_map('typeCastStrokeData', $data['strokes']) : [],
        'stroke_count' => (int)$data['stroke_count'],
        'watcher_count' => (int)$data['watcher_count'],
    ];
}


class TokenException extends Exception {}

function checkToken($dbh, $csrf_token) {
    $sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens`';
    $sql .= ' WHERE `csrf_token` = :csrf_token AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY';
    $token = selectOne($dbh, $sql, [':csrf_token' => $csrf_token]);
    if (is_null($token)) {
        throw new TokenException();
    }
    return $token;
}

function getStrokePoints($dbh, $stroke_id) {
    $sql = 'SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = :stroke_id ORDER BY `id` ASC';
    return selectAll($dbh, $sql, [':stroke_id' => $stroke_id]);
}

function getStrokes($dbh, $room_id, $greater_than_id) {
    $sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`';
    $sql .= ' WHERE `room_id` = :room_id AND `id` > :greater_than_id ORDER BY `id` ASC';
    return selectAll($dbh, $sql, [':room_id' => $room_id, ':greater_than_id' => $greater_than_id]);
}

function getRoom($dbh, $room_id) {
    $sql = 'SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = :room_id';
    return selectOne($dbh, $sql, [':room_id' => $room_id]);
}

function getWatcherCount($dbh, $room_id) {
    $sql = 'SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`';
    $sql .= ' WHERE `room_id` = :room_id AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND';
    $result = selectOne($dbh, $sql, [':room_id' => $room_id]);
    return $result['watcher_count'];
}

function updateRoomWatcher($dbh, $room_id, $token_id) {
    $sql = 'INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (:room_id, :token_id)';
    $sql .= ' ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)';
    execute($dbh, $sql, [':room_id' => $room_id, ':token_id' => $token_id]);
}

// Instantiate the app
$settings = [
    'displayErrorDetails' => getenv('ISUCON_ENV') !== 'production',
    'addContentLengthHeader' => false, // Allow the web server to send the content-length header

    // Monolog settings
    'logger' => [
        'name' => 'isucon6',
        'path' => 'php://stdout',
        'level' => \Monolog\Logger::DEBUG,
    ],
];
$app = new \Slim\App(['settings' => $settings]);

$container = $app->getContainer();

// monolog
$container['logger'] = function ($c) {
    $settings = $c->get('settings')['logger'];
    $logger = new Monolog\Logger($settings['name']);
    $logger->pushHandler(new Monolog\Handler\StreamHandler($settings['path'], $settings['level']));
    return $logger;
};

// Routes

$app->post('/api/csrf_token', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'INSERT INTO `tokens` (`csrf_token`) VALUES';
    $sql .= ' (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))';

    $id = execute($dbh, $sql);

    $sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = :id';
    $token = selectOne($dbh, $sql, [':id' => $id]);

    return $response->withJson(['token' => $token['csrf_token']]);
});

$app->get('/api/rooms', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`';
    $sql .= ' GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100';
    $results = selectAll($dbh, $sql);

    $rooms = [];
    foreach ($results as $result) {
        $room = getRoom($dbh, $result['room_id']);
        $room['stroke_count'] = count(getStrokes($dbh, $room['id'], 0));
        $rooms[] = $room;
    }

    return $response->withJson(['rooms' => array_map('typeCastRoomData', $rooms)]);
});

$app->post('/api/rooms', function ($request, $response, $args) {
    $dbh = getPDO();

    try {
        $token = checkToken($dbh, $request->getHeaderLine('x-csrf-token'));
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $postedRoom = $request->getParsedBody();
    if (empty($postedRoom['name']) || empty($postedRoom['canvas_width']) || empty($postedRoom['canvas_height'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $dbh->beginTransaction();
    try {
        $sql = 'INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)';
        $sql .= ' VALUES (:name, :canvas_width, :canvas_height)';
        $room_id = execute($dbh, $sql, [
            ':name' => $postedRoom['name'],
            ':canvas_width' => $postedRoom['canvas_width'],
            ':canvas_height' => $postedRoom['canvas_height']
        ]);

        $sql = 'INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (:room_id, :token_id)';
        execute($dbh, $sql, [
            ':room_id' => $room_id,
            ':token_id' => $token['id'],
        ]);

        $dbh->commit();
    } catch (Exception $e) {
        $dbh->rollback();
        $this->logger->error($e->getMessage());
        return $response->withStatus(500)->withJson(['error' => 'エラーが発生しました。']);
    }

    $room = getRoom($dbh, $room_id);

    return $response->withJson(['room' => typeCastRoomData($room)]);
});

$app->get('/api/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    $room = getRoom($dbh, $args['id']);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $strokes = getStrokes($dbh, $room['id'], 0);

    foreach ($strokes as $i => $stroke) {
        $strokes[$i]['points'] = getStrokePoints($dbh, $stroke['id']);
    }

    $room['strokes'] = $strokes;
    $room['watcher_count'] = getWatcherCount($dbh, $room['id']);

    return $response->withJson(['room' => typeCastRoomData($room)]);
});

$app->get('/api/stream/rooms/[{id}]', function ($request, $response, $args) {
    header('Content-Type: text/event-stream');

    $dbh = getPDO();

    try {
        $token = checkToken($dbh, $request->getQueryParam('csrf_token'));
    } catch (TokenException $e) {
        printAndFlush(
            "event:bad_request\n" .
            "data:トークンエラー。ページを再読み込みしてください。\n\n"
        );
        return;
    }


    $room = getRoom($dbh, $args['id']);

    if ($room === null) {
        printAndFlush(
            "event:bad_request\n" .
            "data:この部屋は存在しません\n\n"
        );
        return;
    }

    updateRoomWatcher($dbh, $room['id'], $token['id']);
    $watcher_count = getWatcherCount($dbh, $room['id']);

    printAndFlush(
        "retry:500\n\n" .
        "event:watcher_count\n" .
        'data:' . $watcher_count . "\n\n"
    );

    $last_stroke_id = 0;
    if ($request->hasHeader('Last-Event-ID')) {
        $last_stroke_id = (int)$request->getHeaderLine('Last-Event-ID');
    }

    $loop = 6;
    while ($loop > 0) {
        $loop--;
        usleep(500 * 1000); // 500ms

        $strokes = getStrokes($dbh, $room['id'], $last_stroke_id);
        //$this->logger->info(var_export($strokes, true));

        foreach ($strokes as $stroke) {
            $stroke['points'] = getStrokePoints($dbh, $stroke['id']);
            printAndFlush(
                'id:' . $stroke['id'] . "\n\n" .
                "event:stroke\n" .
                'data:' . json_encode(typeCastStrokeData($stroke)) . "\n\n"
            );
            $last_stroke_id = $stroke['id'];
        }

        updateRoomWatcher($dbh, $room['id'], $token['id']);
        $new_watcher_count = getWatcherCount($dbh, $room['id']);
        if ($new_watcher_count !== $watcher_count) {
            $watcher_count = $new_watcher_count;
            printAndFlush(
                "event:watcher_count\n" .
                'data:' . $watcher_count . "\n\n"
            );
        }
    }
});

$app->post('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    try {
        $token = checkToken($dbh, $request->getHeaderLine('x-csrf-token'));
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $room = getRoom($dbh, $args['id']);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $postedStroke = $request->getParsedBody();
    if (empty($postedStroke['width']) || empty($postedStroke['points'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $stroke_count = count(getStrokes($dbh, $room['id'], 0));
    if ($stroke_count == 0) {
        $sql = 'SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = :room_id AND `token_id` = :token_id';
        $result = selectOne($dbh, $sql, [':room_id' => $room['id'], ':token_id' => $token['id']]);
        if ($result['cnt'] == 0) {
            return $response->withStatus(400)->withJson(['error' => '他人の作成した部屋に1画目を描くことはできません']);
        }
    }

    $dbh->beginTransaction();
    try {
        $sql = 'INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)';
        $sql .= ' VALUES(:room_id, :width, :red, :green, :blue, :alpha)';
        $stroke_id = execute($dbh, $sql, [
            ':room_id' => $room['id'],
            ':width' => $postedStroke['width'],
            ':red' => $postedStroke['red'],
            ':green' => $postedStroke['green'],
            ':blue' => $postedStroke['blue'],
            ':alpha' => $postedStroke['alpha']
        ]);

        $sql = 'INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (:stroke_id, :x, :y)';
        foreach ($postedStroke['points'] as $point) {
            execute($dbh, $sql, [
                ':stroke_id' => $stroke_id,
                ':x' => $point['x'],
                ':y' => $point['y']
            ]);
        }

        $dbh->commit();
    } catch (Exception $e) {
        $dbh->rollback();
        $this->logger->error($e->getMessage());
        return $response->withStatus(500)->withJson(['error' => 'エラーが発生しました。']);
    }

    $sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`';
    $sql .= ' WHERE `id` = :stroke_id';
    $stroke = selectOne($dbh, $sql, [':stroke_id' => $stroke_id]);

    $stroke['points'] = getStrokePoints($dbh, $stroke_id);

    return $response->withJson(['stroke' => typeCastStrokeData($stroke)]);
});

// Run app
$app->run();
