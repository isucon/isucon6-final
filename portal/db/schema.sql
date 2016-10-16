CREATE TABLE IF NOT EXISTS teams (
    id INT UNSIGNED NOT NULL,
    name VARCHAR(128) NOT NULL,
    password VARCHAR(128) NOT NULL,
    ip_address VARCHAR(32),
    instance_name VARCHAR(255),
    category ENUM('general', 'students', 'official') NOT NULL,
    azure_resource_group VARCHAR(32) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (name)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS results (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    team_id INT UNSIGNED NOT NULL, -- teams.id
    queue_id INT UNSIGNED NOT NULL, -- queues.id
    pass TINYINT UNSIGNED NOT NULL,
    score BIGINT NOT NULL,
    messages MEDIUMTEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY team_id (team_id)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS queues (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    team_id INT NOT NULL,
    status ENUM('waiting', 'running', 'done', 'aborted') NOT NULL DEFAULT 'waiting',
    bench_node VARCHAR(64) DEFAULT NULL,
    stderr MEDIUMTEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY queues_team_status_idx (team_id, status)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS proxies (
    ip_address VARCHAR(128) NOT NULL,
    PRIMARY KEY (ip_address)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS messages (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    message VARCHAR(255) NOT NULL,
    kind ENUM('danger', 'warning', 'info', 'success') NOT NULL,
    PRIMARY KEY (id)
) DEFAULT CHARSET=utf8mb4;
