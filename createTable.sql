CREATE TABLE IF NOT EXISTS `usersnapshots` (
	`tid` INT(11) NOT NULL AUTO_INCREMENT,
	`sid` INT(11) NOT NULL,
	`id` VARCHAR(50) NOT NULL,
	`ask` INT(11) NOT NULL,
	`answer` INT(11) NOT NULL,
	`post` INT(11) NOT NULL,
	`edits` INT(11) NOT NULL,
	`agree` INT(11) NOT NULL,
	`thanks` INT(11) NOT NULL,
	`follower` INT(11) NOT NULL,
	`followee` INT(11) NOT NULL,
	`fav` INT(11) NOT NULL,
	`logs` INT(11) NOT NULL,
	`mostvote` INT(11) NOT NULL,
	`mostvote5` INT(11) NOT NULL,
	`mostvote10` INT(11) NOT NULL,
	`count10000` INT(11) NOT NULL,
	`count5000` INT(11) NOT NULL,
	`count2000` INT(11) NOT NULL,
	`count1000` INT(11) NOT NULL,
	`count500` INT(11) NOT NULL,
	`count200` INT(11) NOT NULL,
	`count100` INT(11) NOT NULL,
	`date` DATETIME NOT NULL,
	PRIMARY KEY (`tid`),
	UNIQUE INDEX `index_sid_id` (`sid`, id),
	INDEX `index_sid` (`sid`)
)
COLLATE='utf8_general_ci'
ENGINE=InnoDB
;

CREATE TABLE IF NOT EXISTS `users` (
	`tid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`hash` VARCHAR(100) NOT NULL,
	`id` VARCHAR(100) NOT NULL,
	`name` VARCHAR(100) NOT NULL,
	`sex` TINYINT(1) NULL DEFAULT NULL,
	`avatar` VARCHAR(100) NULL DEFAULT NULL,
	`signature` VARCHAR(500) NULL DEFAULT NULL,
	`description` VARCHAR(2000) NULL DEFAULT NULL,
	`cheat` TINYINT(1) NULL DEFAULT NULL,
	`stopped` TINYINT(1) NULL DEFAULT NULL,
	PRIMARY KEY (`tid`),
	UNIQUE INDEX `index_id` (`id`)
)
COLLATE='utf8_general_ci'
ENGINE=InnoDB
;

CREATE TABLE IF NOT EXISTS `snapshot_config` (
	`tid` INT(11) NOT NULL AUTO_INCREMENT,
	`sid` INT(11) NOT NULL,
	`startAt` DATETIME NOT NULL,
	`endAt` DATETIME NOT NULL,
	`finished` TINYINT(1) NULL DEFAULT NULL,
	PRIMARY KEY (`tid`)
)
COLLATE='utf8_general_ci'
ENGINE=InnoDB
;

CREATE TABLE IF NOT EXISTS `usertopanswers` (
	`tid` BIGINT(20) NOT NULL AUTO_INCREMENT,
	`id` VARCHAR(50) NOT NULL,
	`sid` INT(11) NOT NULL,
	`title` VARCHAR(50) NOT NULL,
	`agree` INT(11) NOT NULL,
	`date` DATETIME NOT NULL,
	`answerid` CHAR(50) NOT NULL,
	`link` VARCHAR(50) NOT NULL,
	`ispost` TINYINT(1) NOT NULL,
	`noshare` TINYINT(1) NOT NULL,
	`len` INT(11) NOT NULL,
	`summary` VARCHAR(1400) NOT NULL,
	PRIMARY KEY (`tid`),
	UNIQUE INDEX `index_sid_uid` (`sid`, `id`, `answerid`)
)
COLLATE='utf8_general_ci'
ENGINE=InnoDB
;
