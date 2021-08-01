alter table nideshop_address add column province_name varchar(48);
alter table nideshop_address add column city_name varchar(48);
alter table nideshop_address add column district_name varchar(48);
ALTER TABLE `nideshop_user` AUTO_INCREMENT=10000000;
alter table nideshop_cart add column goods_brief varchar(255);

alter table nideshop_order add column province_name varchar(48);
alter table nideshop_order add column city_name varchar(48);
alter table nideshop_order add column district_name varchar(48);