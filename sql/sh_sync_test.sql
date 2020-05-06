-- cleanup first
delete from enrollments where uuid like 'origin:%';
delete from identities where uuid like 'origin:%';
delete from profiles where uuid like 'origin:%';
delete from uidentities where uuid like 'origin:%';
delete from domains_organizations where domain like 'origin.%';
delete from organizations where name like 'origin:%';
delete from sync_orgs;
delete from sync_uuids;

-- now tests
-- organizations
set @origin = 'orgs0';
insert into organizations(name) select 'origin:orgs0:org';
set @origin = 'orgs1';
insert into organizations(name) select 'origin:orgs1:org';
update organizations set name = 'origin:orgs1:rorg' where name = 'origin:orgs1:org';
set @origin = 'orgs2';
insert into organizations(name) select 'origin:orgs2:org';
set @origin = 'orgs3';
update organizations set name = 'origin:orgs3:rorg' where name = 'origin:orgs2:org';

-- domains_organizations
set @origin = 'doms0';
insert into organizations(name) select 'origin:doms0:org';
insert into domains_organizations(domain, is_top_domain, organization_id) select 'origin.doms0.dom', 0, id from organizations where name = 'origin:doms0:org';
set @origin = 'doms1';
insert into organizations(name) select 'origin:doms1:org';
insert into domains_organizations(domain, is_top_domain, organization_id) select 'origin.doms1.dom', 0, id from organizations where name = 'origin:doms1:org';
update domains_organizations set domain = 'origin.doms1.rdom' where domain = 'origin.doms1.dom';
set @origin = 'doms2';
insert into domains_organizations(domain, is_top_domain, organization_id) select 'origin.doms2.dom', 0, id from organizations where name = 'origin:doms1:org';
set @origin = 'doms3';
update domains_organizations set domain = 'origin.doms2.rdom' where domain = 'origin.doms2.dom';

-- profiles
-- uidentities
set @origin = 'uids0';
insert into uidentities(uuid) select 'origin:uids0:uuid';
set @origin = 'uids1';
insert into uidentities(uuid) select 'origin:uids1:uuid';
update uidentities set last_modified = now() where uuid = 'origin:uids1:uuid';
set @origin = 'uids2';
insert into uidentities(uuid) select 'origin:uids2:uuid';
set @origin = 'uids3';
update uidentities set last_modified = now() where uuid = 'origin:uids2:uuid';

-- profiles
set @origin = 'prof0';
insert into profiles(uuid) select 'origin:uids0:uuid';
set @origin = 'prof1';
insert into profiles(uuid) select 'origin:uids1:uuid';
update profiles set country_code = 'PL' where uuid = 'origin:uids1:uuid';
set @origin = 'prof2';
insert into profiles(uuid) select 'origin:uids2:uuid';
set @origin = 'prof3';
update profiles set country_code = 'PL' where uuid = 'origin:uids2:uuid';

-- identities
set @origin = 'ids0';
insert into uidentities(uuid) select 'origin:ids0:uuid';
insert into identities(uuid, id, source) select 'origin:ids0:uuid', 'origin:ids0:id', 's';
set @origin = 'ids1';
insert into uidentities(uuid) select 'origin:ids1:uuid';
insert into identities(uuid, id, source) select 'origin:ids1:uuid', 'origin:ids1:id', 's';
update identities set source = 't' where id = 'origin:ids1:id';
set @origin = 'ids2';
insert into identities(uuid, id, source) select 'origin:ids1:uuid', 'origin:ids2:id', 's';
set @origin = 'ids3';
update identities set source = 't' where id = 'origin:ids2:id';

-- enrollments
set @origin = 'rol0';
insert into organizations(name) select 'origin:rol0:org';
insert into uidentities(uuid) select 'origin:rol0:uuid';
insert into enrollments(uuid, start, end, organization_id) select 'origin:rol0:uuid', '1900-01-01', '2100-01-01', id from organizations where name = 'origin:rol0:org';
set @origin = 'rol1';
insert into organizations(name) select 'origin:rol1:org';
insert into uidentities(uuid) select 'origin:rol1:uuid';
insert into enrollments(uuid, start, end, organization_id) select 'origin:rol1:uuid', '1900-01-01', '2020-01-01', id from organizations where name = 'origin:rol1:org';
update enrollments set start = '2020-01-01', end = '2021-01-01' where uuid = 'origin:rol1:uuid' and start = '1900-01-01';
set @origin = 'rol3';
insert into enrollments(uuid, start, end, organization_id) select 'origin:rol1:uuid', '2020-01-01', '2100-01-01', id from organizations where name = 'origin:rol1:org';
set @origin = 'rol4';
update enrollments set start = '2021-01-01', end = '2022-01-01' where uuid = 'origin:rol1:uuid' and end = '2100-01-01';


-- final output
select count(*) as domains_orgs from domains_organizations where domain like 'origin.%';
select * from domains_organizations where domain like 'origin.%';
select count(*) as orgs from organizations where name like 'origin:%';
select * from organizations where name like 'origin:%';
select count(*) as sync_orgs from sync_orgs;
select * from sync_orgs order by last_modified;

select '===================================================';
select count(*) as enrollments from enrollments where uuid like 'origin:%';
select * from enrollments where uuid like 'origin:%';
select count(*) as identities from identities where uuid like 'origin:%';
select * from identities where uuid like 'origin:%';
select count(*) as profiles from profiles where uuid like 'origin:%';
select * from profiles where uuid like 'origin:%';
select count(*) as uidentities from uidentities where uuid like 'origin:%';
select * from uidentities where uuid like 'origin:%';
select count(*) as sync_uuids from sync_uuids;
select * from sync_uuids order by last_modified;

-- cleanup at the end
delete from enrollments where uuid like 'origin:%';
delete from identities where uuid like 'origin:%';
delete from profiles where uuid like 'origin:%';
delete from uidentities where uuid like 'origin:%';
delete from domains_organizations where domain like 'origin.%';
delete from organizations where name like 'origin:%';
delete from sync_orgs;
delete from sync_uuids;
