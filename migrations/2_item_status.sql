-- +migrate Up

alter table items rename to item;
alter table movies rename to movie;
alter table tv_episodes rename to tv_episode;
alter table tv_seasons rename to tv_season;

create table item_status
(
    item_id     uuid      not null references item,
    received_at timestamp not null,
    status      text      not null,
    primary key (item_id, received_at)
);

-- +migrate Down

drop table item_status;

alter table tv_season rename to tv_seasons;
alter table tv_episode rename to tv_episodes;
alter table movie rename to movies;
alter table item rename to items;
