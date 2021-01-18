drop table if exists subscribers;
drop table if exists users;
drop table if exists comics;

create table comics (
    "id" serial UNIQUE not null,
    "page" VARCHAR(64) not null,
    "name" VARCHAR(128)not null,
    "url" VARCHAR(256) not null unique,
    "img_url" VARCHAR(128) not null,
    "cloud_img_url" VARCHAR(256) not null,
    "latest_chap" VARCHAR(128) not null,
    "chap_url" VARCHAR(128) not null,
    PRIMARY KEY (id, url, img_url)
);
create table users (
    "id" serial UNIQUE not null,
    "name" VARCHAR(64) not null,
    "psid" VARCHAR(64) UNIQUE,
    "appid" VARCHAR(64) UNIQUE,
    "profile_pic" VARCHAR(256),
    PRIMARY KEY (id,psid, appid)
);
create table subscribers (
    "id" serial UNIQUE not null,
    "user_psid" VARCHAR(64) references users(psid) not null,
    "user_appid" VARCHAR(64) references users(appid) not null,
    "comic_id" INT references comics(id) not null,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);