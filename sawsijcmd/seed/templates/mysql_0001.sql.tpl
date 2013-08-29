CREATE TABLE `{{ .schema }}_sawsij_db_version` (
	`version_id` BIGINT NOT NULL,
	`ran_on` DATETIME NULL,
	PRIMARY KEY (`version_id`)
);

INSERT INTO `{{ .schema }}_sawsij_db_version` (`version_id`)
VALUES
	(1);

CREATE TABLE `{{ .schema }}_user` (
	`id` BIGINT NOT NULL AUTO_INCREMENT,
	`username` VARCHAR (64) NOT NULL,
	`password_hash` text NOT NULL,
	`full_name` text NOT NULL,
	`email` text NULL,
	`created_on` DATETIME NULL,
	`role` INT NULL,
	PRIMARY KEY (`id`)
);

ALTER TABLE `{{ .schema }}_user` ADD CONSTRAINT `UNIQUE_user_1` UNIQUE (`username`);

INSERT INTO  `{{ .schema }}_user` (username, password_hash, full_name, email, created_on, role) 
	VALUES ('admin','{{ .password_hash }}', 'Administrator','{{ .admin_email }}' , now(), 3);