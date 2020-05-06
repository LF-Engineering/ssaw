-- organizations
drop trigger if exists organizations_before_insert_trigger;
drop trigger if exists organizations_after_insert_trigger;
drop trigger if exists organizations_before_update_trigger;
drop trigger if exists organizations_after_update_trigger;

drop trigger if exists domains_organizations_before_insert_trigger;
drop trigger if exists domains_organizations_after_insert_trigger;
drop trigger if exists domains_organizations_before_update_trigger;
drop trigger if exists domains_organizations_after_update_trigger;

-- users/profiles
drop trigger if exists uidentities_before_insert_trigger;
drop trigger if exists uidentities_after_insert_trigger;
drop trigger if exists uidentities_before_update_trigger;
drop trigger if exists uidentities_after_update_trigger;

drop trigger if exists profiles_before_insert_trigger;
drop trigger if exists profiles_after_insert_trigger;
drop trigger if exists profiles_before_update_trigger;
drop trigger if exists profiles_after_update_trigger;

drop trigger if exists identities_before_insert_trigger;
drop trigger if exists identities_after_insert_trigger;
drop trigger if exists identities_before_update_trigger;
drop trigger if exists identities_after_update_trigger;

drop trigger if exists enrollments_before_insert_trigger;
drop trigger if exists enrollments_after_insert_trigger;
drop trigger if exists enrollments_before_update_trigger;
drop trigger if exists enrollments_after_update_trigger;

delimiter $

-- organizations
-- organizations table
create trigger organizations_before_insert_trigger before insert on organizations
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger organizations_after_insert_trigger after insert on organizations
for each row begin
  insert into sync_orgs(name, src, op) values(new.name, new.src, 'i') on duplicate key update last_modified = now();
end$
create trigger organizations_before_update_trigger before update on organizations
for each row begin
  if old.name != new.name then
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger organizations_after_update_trigger after update on organizations
for each row begin
  if old.name != new.name then
    insert into sync_orgs(name, src, op) values(new.name, new.src, 'u') on duplicate key update last_modified = now();
  end if;
end$

-- domains_organizations table
create trigger domains_organizations_before_insert_trigger before insert on domains_organizations
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger domains_organizations_after_insert_trigger after insert on domains_organizations
for each row begin
  insert into sync_orgs(name, src, op) (
    select name, coalesce(@origin, 'unknown'), 'u' from organizations where id = new.organization_id
  ) on duplicate key update last_modified = now(), op = 'u';
end$
create trigger domains_organizations_before_update_trigger before update on domains_organizations
for each row begin
  if old.domain != new.domain or not(old.is_top_domain <=> new.is_top_domain) or old.organization_id != new.organization_id then
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger domains_organizations_after_update_trigger after update on domains_organizations
for each row begin
  if old.domain != new.domain or not(old.is_top_domain <=> new.is_top_domain) or old.organization_id != new.organization_id then
    set @origin = coalesce(@origin, 'unknown');
    insert into sync_orgs(name, src, op) (
      select name, @origin, 'u' from organizations where id = old.organization_id
    ) on duplicate key update last_modified = now(), op = 'u';
    insert into sync_orgs(name, src, op) (
      select name, @origin, 'u' from organizations where id = new.organization_id
    ) on duplicate key update last_modified = now(), op = 'u';
  end if;
end$

-- users/profiles
-- uidentities table
create trigger uidentities_before_insert_trigger before insert on uidentities
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger uidentities_after_insert_trigger after insert on uidentities
for each row begin
  insert into sync_uuids(uuid, src, op) values(new.uuid, new.src, 'i') on duplicate key update last_modified = now();
end$
create trigger uidentities_before_update_trigger before update on uidentities
for each row begin
  if not(old.last_modified <=> new.last_modified) then
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger uidentities_after_update_trigger after update on uidentities
for each row begin
  if not(old.last_modified <=> new.last_modified) then
    insert into sync_uuids(uuid, src, op) values(new.uuid, new.src, 'u') on duplicate key update last_modified = now(), op = 'u';
  end if;
end$

-- profiles table
create trigger profiles_before_insert_trigger before insert on profiles
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger profiles_after_insert_trigger after insert on profiles
for each row begin
  insert into sync_uuids(uuid, src, op) values(new.name, new.src, 'i') on duplicate key update last_modified = now();
end$
create trigger profiles_before_update_trigger before update on profiles
for each row begin
  if not(old.name <=> new.name) or not(old.email <=> new.email) or not(old.gender <=> new.gender) or not(old.gender_acc <=> new.gender_acc) or not(old.is_bot <=> new.is_bot) or not(old.country_code <=> new.country_code) then 
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger profiles_after_update_trigger after update on profiles
for each row begin
  if not(old.name <=> new.name) or not(old.email <=> new.email) or not(old.gender <=> new.gender) or not(old.gender_acc <=> new.gender_acc) or not(old.is_bot <=> new.is_bot) or not(old.country_code <=> new.country_code) then 
    insert into sync_uuids(uuid, src, op) values(new.uuid, new.src, 'u') on duplicate key update last_modified = now(), op = 'u';
  end if;
end$

-- identities table
create trigger identities_before_insert_trigger before insert on identities
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger identities_after_insert_trigger after insert on identities
for each row begin
  insert into sync_uuids(uuid, src, op) values(new.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
end$
create trigger identities_before_update_trigger before update on identities
for each row begin
  if old.source != new.source or not(old.name <=> new.name) or not(old.email <=> new.email) or not(old.username <=> new.username) or not(old.uuid <=> new.uuid) then
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger identities_after_update_trigger after update on identities
for each row begin
  if old.source != new.source or not(old.name <=> new.name) or not(old.email <=> new.email) or not(old.username <=> new.username) or not(old.uuid <=> new.uuid) then
    set @origin = coalesce(@origin, 'unknown');
    insert into sync_uuids(uuid, src, op) values(old.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
    if not(old.uuid <=> new.uuid) then
      insert into sync_uuids(uuid, src, op) values(new.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
    end if;
  end if;
end$

-- enrollments table
create trigger enrollments_before_insert_trigger before insert on enrollments
for each row begin
  set new.src = coalesce(@origin, 'unknown'), new.op = 'i';
end$
create trigger enrollments_after_insert_trigger after insert on enrollments
for each row begin
  insert into sync_uuids(uuid, src, op) values(new.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
end$
create trigger enrollments_before_update_trigger before update on enrollments
for each row begin
  if old.uuid != new.uuid or old.organization_id != new.organization_id or old.start != new.start or old.end != new.end then
    set new.src = coalesce(@origin, 'unknown'), new.op = 'u';
  end if;
end$
create trigger enrollments_after_update_trigger after update on enrollments
for each row begin
  if old.uuid != new.uuid or old.organization_id != new.organization_id or old.start != new.start or old.end != new.end then
    set @origin = coalesce(@origin, 'unknown');
    insert into sync_uuids(uuid, src, op) values(old.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
    if not(old.uuid <=> new.uuid) then
      insert into sync_uuids(uuid, src, op) values(new.uuid, coalesce(@origin, 'unknown'), 'u') on duplicate key update last_modified = now(), op = 'u';
    end if;
  end if;
end$

delimiter ;
