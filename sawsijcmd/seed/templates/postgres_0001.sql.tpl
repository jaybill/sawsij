CREATE TABLE "{{ .schema }}"."sawsij_db_version" (
    "version_id" int8 NOT NULL,
    "ran_on" timestamp NULL default now(),
    PRIMARY KEY("version_id")
);

INSERT INTO "{{ .schema }}"."sawsij_db_version" ("version_id") VALUES (1);

CREATE TABLE "{{ .schema }}"."user"  ( 
	"id"           	serial NOT NULL,
	"username"     	varchar(64) NOT NULL,
	"password_hash"	text NOT NULL,
	"full_name"    	text NOT NULL,
	"email"        	text NULL,
	"created_on"   	timestamp NULL,
	"role"         	int NULL,
	PRIMARY KEY("id")
);

ALTER TABLE "{{ .schema }}"."user"
	ADD CONSTRAINT "UNIQUE_user_1"
	UNIQUE ("username");

INSERT INTO  "{{ .schema }}"."user"(username, password_hash, full_name, email, created_on, role) 
	VALUES ('admin','{{ .password_hash }}', 'Administrator','{{ .admin_email }}' , now(), 3);