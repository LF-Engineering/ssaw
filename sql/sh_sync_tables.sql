-- organizations for sync
drop table if exists sync_orgs;
create table sync_orgs(
  id int(11) not null auto_increment,
  name varchar(192) collate utf8mb4_unicode_520_ci not null,
  src varchar(32) collate utf8mb4_unicode_520_ci not null,
  op char(1) collate utf8mb4_unicode_520_ci not null,
  last_modified datetime(6) not null default now(),
  primary key(id)
) engine=InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_520_ci;

-- users/profiles for sync
drop table if exists sync_uuids;
create table sync_uuids(
  id int(11) not null auto_increment,
  uuid varchar(128) collate utf8mb4_unicode_520_ci not null,
  src varchar(32) collate utf8mb4_unicode_520_ci not null,
  op varchar(8) collate utf8mb4_unicode_520_ci not null,
  last_modified datetime(6) not null default now(),
  primary key(id)
) engine=InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_520_ci;

alter table sync_orgs add index sync_orgs_name_idx(name);
alter table sync_uuids add index sync_uuids_uuid_idx(uuid);

-- source and operation for organizations
alter table organizations drop column if exists src;
alter table organizations drop column if exists op;
alter table organizations add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table organizations add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;

alter table domains_organizations drop column if exists src;
alter table domains_organizations drop column if exists op;
alter table domains_organizations add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table domains_organizations add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;

-- source and operation for profiles
alter table uidentities drop column if exists src;
alter table uidentities drop column if exists op;
alter table uidentities add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table uidentities add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;

alter table profiles drop column if exists src;
alter table profiles drop column if exists op;
alter table profiles add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table profiles add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;

alter table identities drop column if exists src;
alter table identities drop column if exists op;
alter table identities add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table identities add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;

alter table enrollments drop column if exists src;
alter table enrollments drop column if exists op;
alter table enrollments add column if not exists src varchar(32) collate utf8mb4_unicode_520_ci;
alter table enrollments add column if not exists op varchar(1) collate utf8mb4_unicode_520_ci;
