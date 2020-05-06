-- cleanup first
delete from sync_orgs;
delete from organizations where name like 'origin:%';
delete from sync_uuids;

-- now tests
-- organizations
set @origin = 'orgs1';
insert into organizations(name) select 'origin:orgs1:org';
update organizations set name = 'origin:orgs1:rorg' where name = 'origin:orgs1:org';
set @origin = 'orgs2';
insert into organizations(name) select 'origin:orgs2:org';
set @origin = 'orgs3';
update organizations set name = 'origin:orgs3:rorg' where name = 'origin:orgs2:org';

-- domains_organizations
set @origin = 'doms1';
insert into organizations(name) select 'origin:doms1:org';
insert into domains_organizations(domain, is_top_domain, organization_id) select 'origin.doms1.dom', 0, id from organizations where name = 'origin:doms1:org';
update domains_organizations set domain = 'origin.doms1.rdom' where domain = 'origin.doms1.dom';
set @origin = 'doms2';
insert into domains_organizations(domain, is_top_domain, organization_id) select 'origin.doms2.dom', 0, id from organizations where name = 'origin:doms1:org';
set @origin = 'doms3';
update domains_organizations set domain = 'origin.doms2.rdom' where domain = 'origin.doms2.dom';


-- final output
select count(*) as domains_orgs from domains_organizations where domain like 'origin.%';
select * from domains_organizations where domain like 'origin.%';
select count(*) as orgs from organizations where name like 'origin:%';
select * from organizations where name like 'origin:%';
select count(*) as sync_orgs from sync_orgs;
select * from sync_orgs order by last_modified;
