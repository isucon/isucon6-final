DROP DATABASE IF EXISTS `isuketch`;
CREATE DATABASE `isuketch`;

GRANT ALL PRIVILEGES ON `isuketch`.* TO 'isucon'@'localhost' IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON `isuketch`.* TO 'isucon'@'%' IDENTIFIED BY 'isucon';
