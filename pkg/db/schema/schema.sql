drop table if exists subscribers;
drop table if exists users;
drop table if exists comics;

CREATE EXTENSION unaccent;

create table comics (
    id serial UNIQUE not null,
    page VARCHAR(64) not null,
    "name" VARCHAR(128)not null,
    "url" VARCHAR(256) not null unique,
    "img_url" VARCHAR(128) not null,
    "cloud_img_url" VARCHAR(256) not null,
    "latest_chap" VARCHAR(128) not null,
    "chap_url" VARCHAR(128) not null,
    PRIMARY KEY (id)
);
create table users (
    "id" serial UNIQUE not null,
    "name" VARCHAR(64) not null,
    "psid" VARCHAR(64) UNIQUE,
    "appid" VARCHAR(64) UNIQUE,
    "profile_pic" VARCHAR(256),
    PRIMARY KEY (id)
);
create table subscribers (
    "id" serial UNIQUE not null,
    "user_id" INT REFERENCES users(id) not null,
    "comic_id" INT REFERENCES comics(id) not null,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);