<?php

require __DIR__ . '/../vendor/autoload.php';

//session_start();

// Instantiate the app
$settings = [
    'displayErrorDetails' => true, // set to false in production
    'addContentLengthHeader' => false, // Allow the web server to send the content-length header

    // Monolog settings
    'logger' => [
        'name' => 'isucon6',
        'path' => __DIR__ . '/../logs/app.log',
        'level' => \Monolog\Logger::DEBUG,
    ],
];
$app = new \Slim\App($settings);

$container = $app->getContainer();

// monolog
$container['logger'] = function ($c) {
    $settings = $c->get('settings')['logger'];
    $logger = new Monolog\Logger($settings['name']);
    $logger->pushProcessor(new Monolog\Processor\UidProcessor());
    $logger->pushHandler(new Monolog\Handler\StreamHandler($settings['path'], $settings['level']));
    return $logger;
};

// Routes

$app->get('/api/rooms', function ($request, $response, $args) {
    $rooms = [];
    for ($i = 1; $i <= 10; $i++) {
        $rooms[] = ['id' => $i, 'name' => 'ひたすら椅子を描く部屋'];
    }
    return $response->withJson(['rooms' => $rooms]);
});

$app->get('/api/csrf_token', function ($request, $response, $args) {
    return $response->withJson(['token' => md5(rand())]);
});

$app->get('/api/rooms/[{id}]', function ($request, $response, $args) {
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
    return $response->withJson(['room' => $room]);
});

$app->post('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    $stroke = $request->getParsedBody();

    $stroke['id'] = microtime(true) * 1000000;
    return $response->withJson(['stroke' => $stroke]);
});

$app->get('/api/strokes/rooms/[{id}]', function ($request, $response, $args) {
    sleep(1);
    return $response
        //->withHeader('Transfer-Encoding', 'chunked') // TODO: これを付けるとなぜかApacheがbodyを出力しない
        ->withHeader('Content-type', 'text/event-stream')
        ->write('data: ' . json_encode([
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
        ]) . "\n\n");
});

$app->get('/[{name}]', function ($request, $response, $args) {
    // Sample log message
    $this->logger->info("Slim-Skeleton '/' route");

    // Render index view
    return $this->renderer->render($response, 'index.phtml', $args);
});

// Run app
$app->run();
