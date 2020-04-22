-- organizations
drop trigger if exists organizations_insert_sf_sync_trigger;
drop trigger if exists organizations_update_sf_sync_trigger;
drop trigger if exists organizations_delete_sf_sync_trigger;

drop trigger if exists domains_organizations_insert_sf_sync_trigger;
drop trigger if exists domains_organizations_update_sf_sync_trigger;
drop trigger if exists domains_organizations_delete_sf_sync_trigger;

-- users/profiles
drop trigger if exists uidentities_insert_sf_sync_trigger;
drop trigger if exists uidentities_update_sf_sync_trigger;
drop trigger if exists uidentities_delete_sf_sync_trigger;

drop trigger if exists profiles_insert_sf_sync_trigger;
drop trigger if exists profiles_update_sf_sync_trigger;
drop trigger if exists profiles_delete_sf_sync_trigger;

drop trigger if exists identities_insert_sf_sync_trigger;
drop trigger if exists identities_update_sf_sync_trigger;
drop trigger if exists identities_delete_sf_sync_trigger;

drop trigger if exists enrollments_insert_sf_sync_trigger;
drop trigger if exists enrollments_update_sf_sync_trigger;
drop trigger if exists enrollments_delete_sf_sync_trigger;

delimiter $

-- organizations
-- organizations table
create trigger organizations_insert_sf_sync_trigger after insert on organizations
for each row begin
  insert into orgs_for_sf_sync(name) values(new.name) on duplicate key update last_modified = now();
end$
create trigger organizations_update_sf_sync_trigger after update on organizations
for each row begin
  if old.name != new.name then
    insert into orgs_for_sf_sync(name) values(old.name) on duplicate key update last_modified = now();
    insert into orgs_for_sf_sync(name) values(new.name) on duplicate key update last_modified = now();
  end if;
end$
create trigger organizations_delete_sf_sync_trigger after delete on organizations
for each row begin
  insert into orgs_for_sf_sync(name) values(old.name) on duplicate key update last_modified = now();
end$

-- domains_organizations table
create trigger domains_organizations_insert_sf_sync_trigger after insert on domains_organizations
for each row begin
  insert into orgs_for_sf_sync(name) (select name from organizations where id = new.organization_id) on duplicate key update last_modified = now();
end$
create trigger domains_organizations_update_sf_sync_trigger after update on domains_organizations
for each row begin
  if old.domain != new.domain or NOT(old.is_top_domain <=> new.is_top_domain) or old.organization_id != new.organization_id then
    insert into orgs_for_sf_sync(name) (select name from organizations where id = old.organization_id) on duplicate key update last_modified = now();
    insert into orgs_for_sf_sync(name) (select name from organizations where id = new.organization_id) on duplicate key update last_modified = now();
  end if;
end$
create trigger domains_organizations_delete_sf_sync_trigger after delete on domains_organizations
for each row begin
  insert into orgs_for_sf_sync(name) (select name from organizations where id = old.organization_id) on duplicate key update last_modified = now();
end$

-- users/profiles
-- uidentities table
create trigger uidentities_insert_sf_sync_trigger after insert on uidentities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger uidentities_update_sf_sync_trigger after update on uidentities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger uidentities_delete_sf_sync_trigger after delete on uidentities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
end$

-- profiles table
create trigger profiles_insert_sf_sync_trigger after insert on profiles
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger profiles_update_sf_sync_trigger after update on profiles
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger profiles_delete_sf_sync_trigger after delete on profiles
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
end$

-- identities table
create trigger identities_insert_sf_sync_trigger after insert on identities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger identities_update_sf_sync_trigger after update on identities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger identities_delete_sf_sync_trigger after delete on identities
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
end$

-- enrollments table
create trigger enrollments_insert_sf_sync_trigger after insert on enrollments
for each row begin
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger enrollments_update_sf_sync_trigger after update on enrollments
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
  insert into uuids_for_sf_sync(uuid) values(new.uuid) on duplicate key update last_modified = now();
end$
create trigger enrollments_delete_sf_sync_trigger after delete on enrollments
for each row begin
  insert into uuids_for_sf_sync(uuid) values(old.uuid) on duplicate key update last_modified = now();
end$

delimiter ;
