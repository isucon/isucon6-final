<?php

require __DIR__ . '/../vendor/autoload.php';

function getPDO() {
    $host = getenv('MYSQL_HOST') ?: 'localhost';
    $port = getenv('MYSQL_PORT') ?: 3306;
    $user = getenv('MYSQL_USER') ?: 'root';
    $pass = getenv('MYSQL_PASS') ?: '';
    $dbname = 'isuchannel';
    $dbh = new PDO("mysql:host={$host};port={$port};dbname={$dbname};charset=utf8", $user, $pass);
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

function typeCastPointData($data) {
    return [
        'id' => (int)$data['id'],
        'stroke_id' => (int)$data['stroke_id'],
        'x' => (float)$data['x'],
        'y' => (float)$data['y'],
    ];
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
        'created_at' => isset($data['created_at']) ? date_create($data['created_at'])->format(DateTime::ISO8601) : '',
    ];
}

function typeCastRoomData($data) {
    return [
        'id' => (int)$data['id'],
        'name' => $data['name'],
        'canvas_width' => (int)$data['canvas_width'],
        'canvas_height' => (int)$data['canvas_height'],
        'created_at' => isset($data['created_at']) ? date_create($data['created_at'])->format(DateTime::ISO8601) : '',
        'strokes' => isset($data['strokes']) ? array_map('typeCastStrokeData', $data['strokes']) : [],
        'stroke_count' => (int)$data['stroke_count'],
    ];
}


class TokenException extends Exception {}

function checkToken($x_csrf_token) {
    $dbh = getPDO();
    $sql = 'SELECT * FROM `tokens` WHERE `token` = :token AND `created_at` > CURRENT_TIMESTAMP - INTERVAL 1 DAY';
    $token = selectOne($dbh, $sql, [':token' => $x_csrf_token]);
    if (is_null($token)) {
        throw new TokenException();
    }
    return $token;
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

    $sql = 'INSERT INTO `tokens` (`token`) VALUES';
    $sql .= ' (SHA2(RAND(), 512))';

    $id = execute($dbh, $sql);

    $sql = 'SELECT * FROM `tokens` WHERE id = :id';
    $token = selectOne($dbh, $sql, [':id' => $id]);

    return $response->withJson(['token' => $token['token']]);
});

$app->get('/api/rooms', function ($request, $response, $args) {
    $dbh = getPDO();
    $sql = 'SELECT `rooms`.* FROM `rooms` JOIN';
    $sql .= ' (SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`';
    $sql .= ' GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100) AS `t`';
    $sql .= ' ON `rooms`.`id` = `t`.`room_id`';
    $rooms = selectAll($dbh, $sql);

    foreach ($rooms as $i => $room) {
        $sql = 'SELECT COUNT(*) AS stroke_count FROM `strokes` WHERE `room_id` = :room_id';
        $result = selectOne($dbh, $sql, [':room_id' => $room['id']]);
        $rooms[$i]['stroke_count'] = (int)$result['stroke_count'];
    }

    //$this->logger->info(var_export($rooms, true));
    return $response->withJson(['rooms' => array_map('typeCastRoomData', $rooms)]);
});

$app->post('/api/rooms', function ($request, $response, $args) {
    try {
        checkToken($request->getHeaderLine('x-csrf-token'));
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $dbh = getPDO();

    $postedRoom = $request->getParsedBody();
    if (empty($postedRoom['name']) || empty($postedRoom['canvas_width']) || empty($postedRoom['canvas_height'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $sql = 'INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)';
    $sql .= ' VALUES (:name, :canvas_width, :canvas_height)';
    $id = execute($dbh, $sql, [
        ':name' => $postedRoom['name'],
        ':canvas_width' => $postedRoom['canvas_width'],
        ':canvas_height' => $postedRoom['canvas_height']
    ]);

    $sql = 'SELECT * FROM `rooms` WHERE `id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $id]);

    return $response->withJson(['room' => typeCastRoomData($room)]);
});

$app->get('/api/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'SELECT * FROM `rooms` WHERE `id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $args['id']]);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $sql = 'SELECT * FROM `strokes` WHERE `room_id` = :id ORDER BY `id` ASC';
    $strokes = selectAll($dbh, $sql, [':id' => $args['id']]);

    foreach ($strokes as $i => $stroke) {
        $sql = 'SELECT * FROM `points` WHERE `stroke_id` = :id ORDER BY `id` ASC';
        $strokes[$i]['points'] = selectAll($dbh, $sql, [':id' => $stroke['id']]);
    }

    $room['strokes'] = $strokes;

    //$this->logger->info(var_export($room, true));
    return $response->withJson(['room' => typeCastRoomData($room)]);
});

$app->get('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {

    sleep(1);

    $dbh = getPDO();

    $last_stroke_id = 0;
    if ($request->hasHeader('Last-Event-ID')) {
        $last_stroke_id = (int)$request->getHeaderLine('Last-Event-ID');
    }
    $sql = 'SELECT * FROM `strokes` WHERE `room_id` = :room_id AND `id` > :id ORDER BY `id` ASC';
    $strokes = selectAll($dbh, $sql, [':room_id' => $args['id'], ':id' => $last_stroke_id]);

    $body = "retry:500\n\n";
    foreach ($strokes as $i => $stroke) {
        $last_stroke_id = $stroke['id'];
        $sql = 'SELECT * FROM `points` WHERE `stroke_id` = :stroke_id ORDER BY `id` ASC';
        $strokes[$i]['points'] = selectAll($dbh, $sql, [':stroke_id' => $last_stroke_id]);

        $body .= 'id:' . $last_stroke_id . "\n\n";
        $body .= 'data:' . json_encode(typeCastStrokeData($strokes[$i])) . "\n\n";
    }

    return $response
        ->withHeader('Content-type', 'text/event-stream')
        ->write($body);
});

$app->post('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    try {
        checkToken($request->getHeaderLine('x-csrf-token'));
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $dbh = getPDO();

    $room_id = $args['id'];

    $sql = 'SELECT * FROM `rooms` WHERE `id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $room_id]);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $postedStroke = $request->getParsedBody();
    if (empty($postedStroke['width']) || empty($postedStroke['points'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $sql = 'SELECT COUNT(*) AS stroke_count FROM `strokes` WHERE `room_id` = :room_id';
    $result = selectOne($dbh, $sql, [':room_id' => $room_id]);
    if ($result['stroke_count'] > 10000) {
        return $response->withStatus(400)->withJson(['error' => '10000画を超えました。これ以上描くことはできません。']);
    }

    $dbh->beginTransaction();
    try {
        $sql = 'INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)';
        $sql .= ' VALUES(:room_id, :width, :red, :green, :blue, :alpha)';
        $stroke_id = execute($dbh, $sql, [
            ':room_id' => $room_id,
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

    $sql = 'SELECT * FROM `strokes` WHERE `id` = :stroke_id';
    $stroke = selectOne($dbh, $sql, [':stroke_id' => $stroke_id]);

    $sql = 'SELECT * FROM `points` WHERE `stroke_id` = :stroke_id ORDER BY `id` ASC';
    $stroke['points'] = selectAll($dbh, $sql, [':stroke_id' => $stroke_id]);

    //$this->logger->info(var_export($stroke, true));
    return $response->withJson(['stroke' => typeCastStrokeData($stroke)]);
});

$app->get('/api/initialize', function ($request, $response, $args) {
    $dbh = getPDO();

    $sqls = [
        'DELETE FROM `points` WHERE `id` > 1443000',
        'DELETE FROM `strokes` WHERE `id` > 41000',
        'DELETE FROM `rooms` WHERE `id` > 1000',
        'DELETE FROM `tokens` WHERE `id` > 0',
    ];

    foreach ($sqls as $sql) {
        execute($dbh, $sql);
    }

    return $response->write('ok');
});

// Run app
$app->run();
