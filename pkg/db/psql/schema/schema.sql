drop table if exists subscribers;
drop table if exists users;
drop table if exists comics;

create table comics (
    "id" serial UNIQUE,
    "page" VARCHAR(64),
    "name" VARCHAR(128),
    "url" VARCHAR(256) not null unique,
    "img_url" VARCHAR(128),
    "cloud_img" VARCHAR(256),
    "latest_chap" VARCHAR(128) not null,
    "chap_url" VARCHAR(128) not null,
    PRIMARY KEY (id, url, img_url)
);
create table users (
    "id" serial UNIQUE,
    "name" VARCHAR(64),
    "psid" VARCHAR(64) UNIQUE,
    "appid" VARCHAR(64) UNIQUE,
    "profile_pic" VARCHAR(256),
    PRIMARY KEY (psid, appid)
);
create table subscribers (
    "id" serial,
    "user_psid" VARCHAR(64) references users(psid) on DELETE CASCADE on UPDATE CASCADE,
    "comic_id" INT references comics(id) on DELETE CASCADE on UPDATE CASCADE,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);