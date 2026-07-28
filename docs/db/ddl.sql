create table journal
(
    journal_id varchar(64)                              not null
        primary key,
    title      varchar(255) default ''                  not null,
    created_dt datetime     default current_timestamp() not null
);

create table journal_item
(
    journal_item_id varchar(64)                          not null
        primary key,
    journal_id      varchar(64)                          not null,
    created_dt      datetime default current_timestamp() not null,
    updated_dt      datetime default current_timestamp() not null,
    comments        mediumtext                           null,
    constraint fk_journal_item_journal
        foreign key (journal_id) references journal (journal_id)
);

create table attachment
(
    attachment_id   varchar(64)                              not null
        primary key,
    journal_item_id varchar(64)                              not null,
    title           varchar(255) default ''                  not null,
    content_type    varchar(128)                             null,
    width           int          default 0                   not null,
    height          int          default 0                   not null,
    created_dt      datetime     default current_timestamp() not null,
    updated_dt      datetime     default current_timestamp() not null,
    constraint fk_picture_item
        foreign key (journal_item_id) references journal_item (journal_item_id)
);

create table relation
(
    relation_cd varchar(64)  not null
        primary key,
    name        varchar(128) not null
);

create table user
(
    user_id     varchar(64)  not null
        primary key,
    provided_id varchar(128) not null,
    name        varchar(128) not null,
    login       varchar(128) not null,
    created_dt  datetime     not null,
    constraint uq_provided_id
        unique (provided_id) comment 'one entry for provided id'
);

create table session
(
    session_id  varchar(64)            not null
        primary key,
    user_id     varchar(64)            not null,
    expired_ind varchar(1) default 'N' not null,
    expire_dt   datetime               not null,
    token       varchar(64)            not null,
    constraint fk_session_user
        foreign key (user_id) references user (user_id)
);

create table user_journal
(
    user_journal_id varchar(64)                          not null
        primary key,
    relation_cd     varchar(64)                          not null,
    user_id         varchar(64)                          not null,
    journal_id      varchar(64)                          not null,
    created_dt      datetime default current_timestamp() not null,
    constraint ix_uj_unique
        unique (user_id, journal_id) comment 'only one relation between user and its journal, either owns or reads not both',
    constraint fk_uj_relation
        foreign key (relation_cd) references relation (relation_cd),
    constraint fk_user_journal_journal
        foreign key (journal_id) references journal (journal_id),
    constraint fk_user_journal_user
        foreign key (user_id) references user (user_id)
);



INSERT INTO relation (relation_cd, name) VALUES ('owner', 'Journal owner');
INSERT INTO relation (relation_cd, name) VALUES ('read', 'Read access');

