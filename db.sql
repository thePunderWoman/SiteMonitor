Create TABLE IF NOT EXISTS Setting (
	`id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	`server` varchar(100) NULL,
	`email` varchar(100) NULL,
	`requireSSL` int NOT NULL DEFAULT 0,
	`username` varchar(100) NULL,
	`password` varchar(100) NULL,
	`port` int NOT NULL
);

CREATE TABLE IF NOT EXISTS Website (
	`id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	`name` varchar(250) NULL,
	`URL` varchar(250) NULL,
	`checkInterval` int NOT NULL,
	`monitoring` int NOT NULL DEFAULT 0,
	`public` int NOT NULL DEFAULT 0,
	`emailInterval` int NOT NULL,
	`logDays` int NOT NULL
);

CREATE TABLE IF NOT EXISTS History (
	`id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	`siteID` INT NOT NULL REFERENCES Website(id),
	`checked` TIMESTAMP NOT NULL,
	`status` varchar(30) NOT NULL,
	`emailed` int NOT NULL DEFAULT 0,
	`code` int NULL,
	`responseTime` double(16,4) NULL
);

CREATE TABLE IF NOT EXISTS Notify (
	`id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	`siteID` INT NOT NULL REFERENCES Website(id),
	`name` varchar(100) NOT NULL,
	`email` varchar(250) NOT NULL
);