-- +migrate Up

create table items
(
    id          uuid primary key,
    external_id text unique not null,
    kind        text        not null,
    title       text,
    description text,
    image_path  text,
    rating      float
);

create table movies
(
    item_id       uuid primary key references items,
    collection_id text
);

create table tv_series
(
    item_id      uuid primary key references items,
    season_count integer
);

create table tv_seasons
(
    item_id       uuid references items,
    number        integer,
    description   text,
    release_year  integer,
    episode_count integer,
    primary key (item_id, number)
);

create table tv_episodes
(
    item_id       uuid references items,
    title         text,
    description   text,
    season_number integer,
    number        integer,
    primary key (item_id, season_number, number)
);

-- +migrate Down

drop table items cascade;
