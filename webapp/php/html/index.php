<?php

require __DIR__ . '/../vendor/autoload.php';

//session_start();

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
    return $stmt->fetch();
}

function selectAll($dbh, $sql, array $params = []) {
    $stmt = $dbh->prepare($sql);
    $stmt->execute($params);
    return $stmt->fetchAll();
}

// Instantiate the app
$settings = [
    'displayErrorDetails' => true, // set to false in production
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

$app->get('/api/rooms', function ($request, $response, $args) {
    $dbh = getPDO();
    // TODO: max_created_at も使うべきか？
    $sql = 'SELECT `room`.* FROM `room` JOIN';
    $sql .= ' (SELECT `room_id`, MAX(`created_at`) AS `max_created_at` FROM `stroke`';
    $sql .= ' GROUP BY `room_id` ORDER BY `max_created_at` DESC LIMIT 30) AS `t`';
    $sql .= ' ON `room`.`id` = `t`.`room_id`';
    // TODO: これだと一画も描かれてない部屋が取得できない
    $rooms = selectAll($dbh, $sql);
    $this->logger->info(var_export($rooms, true));
    return $response->withJson(['rooms' => $rooms]);
});

$app->get('/api/csrf_token', function ($request, $response, $args) {
    return $response->withJson(['token' => md5(rand())]);
});

$app->get('/api/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'SELECT * FROM `room` WHERE `room`.`id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $args['id']]);

    if ($room === false) {
        // TODO: 404
        return $response->withJson(['room' => null]);
    }

    $sql = 'SELECT * FROM `stroke` WHERE `room_id` = :id ORDER BY `id` ASC';
    $strokes = selectAll($dbh, $sql, [':id' => $args['id']]);

    foreach ($strokes as &$stroke) {
        $sql = 'SELECT * FROM `point` WHERE `stroke_id` = :id ORDER BY `id` ASC';
        $stroke['points'] = selectAll($dbh, $sql, [':id' => $stroke['id']]);
    }

    $room['strokes'] = $strokes;

    $this->logger->info(var_export($room, true));
    return $response->withJson(['room' => $room]);
});

// TODO: $app->post('/api/rooms', ...)

$app->post('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    $dbh = getPDO();

    $sql = 'SELECT * FROM `room` WHERE `room`.`id` = :id';
    $room = selectOne($dbh, $sql, [':id' => $args['id']]);

    if ($room === false) {
        // TODO: 404
        return $response->withJson(['room' => null]);
    }
    // TODO: bad request if strokes have reached a certain limit (1000?)

    $stroke = $request->getParsedBody();

    $dbh->query('BEGIN');
    try {
        $sql = 'INSERT INTO `stroke` (`room_id`, `created_at`, `stroke_width`, `red`, `green`, `blue`, `alpha`)';
        $sql .= ' VALUES(:room_id, :created_at, :stroke_width, :red, :green, :blue, :alpha)';
        $id = execute($dbh, $sql, [':room_id' => $args['id'], ':stroke_width' => $stroke['width'], ':red' => $stroke['red'], ':green' => $stroke['green'], ':blue' => $stroke['blue'], ':alpha' => $stroke['alpha']]);

        $stroke['id'] = $id;

        $sql = 'INSERT INTO `point` (`stroke_id`, `x`, `y`) VALUES (:stroke_id, :x, :y)';
        foreach ($stroke['points'] as $coord) {
            execute($dbh, $sql, ['stroke_id' => $id, 'x' => $coord['x'], 'y' => $coord['y']]);
        }

        $dbh->query('COMMIT');
    } catch (Exception $e) {
        $dbh->query('ROLLBACK');
        // TODO: 500
    }

    return $response->withJson(['stroke' => $stroke]);
});

$app->get('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {

    sleep(1);

    $dbh = getPDO();

    if ($request->hasHeader('Last-Event-ID')) {
        $id = $request->getHeaderLine('Last-Event-ID');
        $sql = 'SELECT * FROM `stroke` WHERE `room_id` = :room_id AND `id` > :id ORDER BY `id` ASC';
        $strokes = selectAll($dbh, $sql, [':room_id' => $args['id'], ':id' => $id]);
    } else {
        $sql = 'SELECT * FROM `stroke` WHERE `room_id` = :room_id ORDER BY `id` ASC';
        $strokes = selectAll($dbh, $sql, [':room_id' => $args['id']]);
    }

    $body = "retry:500\n\n";
    foreach ($strokes as &$stroke) {
        $sql = 'SELECT * FROM `point` WHERE `stroke_id` = :id ORDER BY `id` ASC';
        $stroke['points'] = selectAll($dbh, $sql, [':id' => $stroke['id']]);

        $body .= 'id:' . $stroke['id'] . "\n\n";
        $body .= 'data:' . json_encode($stroke) . "\n\n";
    }

    return $response
        //->withHeader('Transfer-Encoding', 'chunked') // TODO: これを付けるとなぜかApacheがbodyを出力しない
        ->withHeader('Content-type', 'text/event-stream')
        ->write($body);
});

// Run app
$app->run();
