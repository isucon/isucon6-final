<?php

require __DIR__ . '/../vendor/autoload.php';

function getPDO() {
    $host = getenv('MYSQL_HOST') ?: 'localhost';
    $port = getenv('MYSQL_PORT') ?: 3306;
    $user = getenv('MYSQL_USER') ?: 'root';
    $pass = getenv('MYSQL_PASS') ?: '';
    $dbname = 'isuchannel';
    $dbh = new PDO("mysql:host={$host};port={$port};dbname={$dbname}", $user, $pass);
    $dbh->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $dbh->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
    $dbh->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); // キャストしなくてよくなる
    return $dbh;
}

function execute($dbh, $sql, array $params = []) {
    $stmt = $dbh->prepare($sql);
    $stmt->execute($params);
    return $dbh->lastInsertId();
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


class TokenException extends Exception {}

function checkToken($request) {
    if (!$request->hasHeader('x-csrf-token')) {
        throw new TokenException();
    }

    $dbh = getPDO();
    $sql = 'SELECT * FROM `csrf_token` WHERE `token` = :token';
    $token = selectOne($dbh, $sql, [':token' => $request->getHeaderLine('x-csrf-token')]);
    if (is_null($token)) {
        throw new TokenException();
    }
    if (time() - strtotime($token['created_at']) > 60 * 60 * 24 * 7) {
        throw new TokenException();
    }
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

    $sql = 'INSERT INTO `csrf_token` (`token`) VALUES';
    $sql .= ' (SHA2(RAND(), 512))';

    $id = execute($dbh, $sql);

    $sql = 'SELECT * FROM `csrf_token` WHERE id = :id';
    $token = selectOne($dbh, $sql, [':id' => $id]);

    return $response->withJson(['token' => $token['token']]);
});

$app->get('/api/rooms', function ($request, $response, $args) {
    $dbh = getPDO();
    $sql = 'SELECT `room`.* FROM `room` JOIN';
    $sql .= ' (SELECT `room_id`, MAX(`id`) AS `max_id` FROM `stroke`';
    $sql .= ' GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100) AS `t`';
    $sql .= ' ON `room`.`id` = `t`.`room_id`';
    $rooms = selectAll($dbh, $sql);

    foreach ($rooms as $i => $room) {
        $sql = 'SELECT COUNT(*) AS stroke_count FROM `stroke` WHERE `room_id` = :room_id';
        $result = selectOne($dbh, $sql, [':room_id' => $room['id']]);
        $rooms[$i]['stroke_count'] = (int)$result['stroke_count'];
    }

    //$this->logger->info(var_export($rooms, true));
    return $response->withJson(['rooms' => $rooms]);
});

$app->post('/api/rooms', function ($request, $response, $args) {
    try {
        checkToken($request);
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $dbh = getPDO();

    $room = $request->getParsedBody();
    if (empty($room['name']) || empty($room['canvas_width']) || empty($room['canvas_height'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $sql = 'INSERT INTO `room` (`name`, `canvas_width`, `canvas_height`)';
    $sql .= ' VALUES (:name, :canvas_width, :canvas_height)';
    $id = execute($dbh, $sql, [':name' => $room['name'], ':canvas_width' => $room['canvas_width'], ':canvas_height' => $room['canvas_height']]);

    $room['id'] = (int)$id;
    $room['strokes'] = [];
    return $response->withJson(['room' => $room]);
});

$app->get('/api/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'SELECT * FROM `room` WHERE `room`.`id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $args['id']]);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $sql = 'SELECT * FROM `stroke` WHERE `room_id` = :id ORDER BY `id` ASC';
    $strokes = selectAll($dbh, $sql, [':id' => $args['id']]);

    foreach ($strokes as $i => $stroke) {
        $sql = 'SELECT * FROM `point` WHERE `stroke_id` = :id ORDER BY `id` ASC';
        $strokes[$i]['points'] = selectAll($dbh, $sql, [':id' => $stroke['id']]);
    }

    $room['strokes'] = $strokes;

    //$this->logger->info(var_export($room, true));
    return $response->withJson(['room' => $room]);
});

$app->get('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {

    header('Content-Type: text/event-stream');
    echo "retry:500\n\n";
    ob_flush();
    flush();

    $dbh = getPDO();

    $lastId = 0;
    if ($request->hasHeader('Last-Event-ID')) {
        $lastId = (int)$request->getHeaderLine('Last-Event-ID');
    }

    $loop = 3;
    while ($loop > 0) {
        $loop--;
        $this->logger->info($loop);

        sleep(1);

        $sql = 'SELECT * FROM `stroke` WHERE `room_id` = :room_id AND `id` > :id ORDER BY `id` ASC';
        $strokes = selectAll($dbh, $sql, [':room_id' => $args['id'], ':id' => $lastId]);

        foreach ($strokes as $i => $stroke) {
            $stroke_id = $stroke['id']
            $sql = 'SELECT * FROM `point` WHERE `stroke_id` = :stroke_id ORDER BY `id` ASC';
            $strokes[$i]['points'] = selectAll($dbh, $sql, [':stroke_id' => $stroke_id]);

            echo 'id:' . $stroke_id . "\n\n";
            echo 'data:' . json_encode($strokes[$i]) . "\n\n";
            ob_flush();
            flush();

            $lastId = $stroke_id;
        }
    }

    return $response;
});

$app->post('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    try {
        checkToken($request);
    } catch (TokenException $e) {
        return $response->withStatus(400)->withJson(['error' => 'トークンエラー。ページを再読み込みしてください。']);
    }

    $dbh = getPDO();

    $sql = 'SELECT * FROM `room` WHERE `room`.`id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $args['id']]);

    if ($room === null) {
        return $response->withStatus(404)->withJson(['error' => 'この部屋は存在しません。']);
    }

    $stroke = $request->getParsedBody();
    if (empty($stroke['width']) || empty($stroke['points'])) {
        return $response->withStatus(400)->withJson(['error' => 'リクエストが正しくありません。']);
    }

    $sql = 'SELECT COUNT(*) AS stroke_count FROM `stroke` WHERE `room_id` = :room_id';
    $result = selectOne($dbh, $sql, [':room_id' => $room['id']]);
    if ($result['stroke_count'] > 1000) {
        return $response->withStatus(400)->withJson(['error' => '1000画を超えました。これ以上描くことはできません。']);
    }

    $dbh->beginTransaction();
    try {
        $sql = 'INSERT INTO `stroke` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)';
        $sql .= ' VALUES(:room_id, :width, :red, :green, :blue, :alpha)';
        $id = execute($dbh, $sql, [':room_id' => $args['id'], ':width' => $stroke['width'], ':red' => $stroke['red'], ':green' => $stroke['green'], ':blue' => $stroke['blue'], ':alpha' => $stroke['alpha']]);

        $stroke['id'] = (int)$id;

        $sql = 'INSERT INTO `point` (`stroke_id`, `x`, `y`) VALUES (:stroke_id, :x, :y)';
        foreach ($stroke['points'] as $coord) {
            execute($dbh, $sql, ['stroke_id' => $id, 'x' => $coord['x'], 'y' => $coord['y']]);
        }

        $dbh->commit();
    } catch (Exception $e) {
        $dbh->rollback();
        $this->logger->error($e->getMessage());
        return $response->withStatus(500)->withJson(['error' => 'エラーが発生しました。']);
    }

    //$this->logger->info(var_export($stroke, true));
    return $response->withJson(['stroke' => $stroke]);
});

// Run app
$app->run();
