BEGIN;
CREATE TABLE IF NOT EXISTS task (
	id int(11) NOT NULL AUTO_INCREMENT,
  	summary TEXT NOT NULL,
	created_at timestamp NOT NULL,
	created_by varchar(255) NOT NULL,
	PRIMARY KEY (`id`)
);
CREATE TABLE IF NOT EXISTS user (
	id int(11) NOT NULL AUTO_INCREMENT,
	email varchar(255) NOT NULL,
	role varchar(255) NOT NULL,
	PRIMARY KEY (`id`)
);
insert into user values(1, "admin@domain.com", "manager");
insert into user values(2, "tech@domain.com", "technician");
COMMIT;
