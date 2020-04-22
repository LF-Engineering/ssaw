drop table if exists uuids_for_sf_sync;
create table uuids_for_sf_sync(
  uuid varchar(128) collate utf8mb4_unicode_520_ci not null,
  last_modified datetime(6) not null default now(),
  primary key(uuid)
) engine=InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_520_ci;

drop table if exists orgs_for_sf_sync;
create table orgs_for_sf_sync(
  name varchar(192) collate utf8mb4_unicode_520_ci not null,
  last_modified datetime(6) not null default now(),
  primary key(name)
) engine=InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_520_ci;

