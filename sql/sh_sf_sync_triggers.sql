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

delimiter $

-- organizations
-- organizations table
create trigger organizations_insert_sf_sync_trigger after insert on organizations
for each row begin
  insert into orgs_for_sf_sync(name) values(new.name) on duplicate key update last_modified = now();
end$
create trigger organizations_update_sf_sync_trigger after update on organizations
for each row begin
  insert into orgs_for_sf_sync(name) values(old.name) on duplicate key update last_modified = now();
  insert into orgs_for_sf_sync(name) values(new.name) on duplicate key update last_modified = now();
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
  insert into orgs_for_sf_sync(name) (select name from organizations where id = old.organization_id) on duplicate key update last_modified = now();
  insert into orgs_for_sf_sync(name) (select name from organizations where id = new.organization_id) on duplicate key update last_modified = now();
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

delimiter ;
