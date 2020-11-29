drop table if exists subscribers;
drop table if exists users;
drop table if exists images;
drop table if exists comics;
drop table if exists pages;


create table comics
(
    "id" serial UNIQUE,
    "page" VARCHAR(64),
    "name" VARCHAR(128),
    "url" VARCHAR(256) not null unique,
    "img_url" VARCHAR(128),
    "imgur_id" VARCHAR(32),
    "imgur_link" VARCHAR(128),
    "latest_chap" VARCHAR(128) not null,
    "chap_url" VARCHAR(128) not null,
    "date" VARCHAR(32),
    "date_format" VARCHAR(32),
    PRIMARY KEY (id, url, img_url)
);

create table users
(
    "name" VARCHAR(64),
    "psid" VARCHAR(64) UNIQUE,
    "appid" VARCHAR(64) UNIQUE,
    "profile_pic" VARCHAR(256),
    PRIMARY KEY (psid, appid)
);
create table subscribers
(
    "id" serial,
    "user_psid" VARCHAR(64) references users(psid) on DELETE CASCADE on UPDATE CASCADE,
    "comic_id" INT references comics(id) on DELETE CASCADE on UPDATE CASCADE,
    "created_at" timestamp with time zone DEFAULT now()
);