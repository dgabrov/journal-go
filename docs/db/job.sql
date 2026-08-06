-- auto-generated definition
create table job
(
    job_id          varchar(64)                              not null
        primary key,
    name            varchar(255) default ''                  not null,
    status          varchar(128) default 'pending'           not null,
    journal_item_id varchar(64)                              not null,
    create_dt       datetime     default current_timestamp() not null,
    constraint fk_job_journal_item
        foreign key (journal_item_id) references journal_item (journal_item_id)
);

