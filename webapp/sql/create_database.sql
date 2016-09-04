DROP DATABASE IF EXISTS `isuchannel`;
CREATE DATABASE `isuchannel`;

GRANT ALL PRIVILEGES ON `isuchannel`.* TO 'isucon'@'localhost' IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON `isuchannel`.* TO 'isucon'@'%' IDENTIFIED BY 'isucon';
