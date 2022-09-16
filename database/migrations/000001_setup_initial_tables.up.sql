CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
create table if not exists account
(
    spotify_id     varchar(2048)                         not null
    constraint account_pk
    primary key,
    email          varchar(150)                          not null,
    access_token   varchar(2048)                         not null,
    expires_in     timestamp                             not null,
    picture        varchar,
    name           varchar default ''::character varying not null
    );

create unique index if not exists account_email_uindex
    on account (email);

create unique index if not exists account_uuid_uindex
    on account (spotify_id);

