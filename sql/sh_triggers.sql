drop trigger if exists uidentities_insert_sf_sync_trigger;
drop trigger if exists uidentities_update_sf_sync_trigger;
drop trigger if exists uidentities_delete_sf_sync_trigger;
delimiter $
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
