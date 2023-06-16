DROP MATERIALIZED VIEW IF EXISTS mv_total_content_count;
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_total_content_count
AS
select count(total_content_count) from (select distinct(cid) as total_content_count from contents) as total_content_count;
CREATE UNIQUE INDEX ON mv_total_content_count(total_content_count);

DROP MATERIALIZED VIEW IF EXISTS mv_total_size;
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_total_size
AS
select sum(total_size) from (select distinct(cid),size as total_size from contents) as total_size;
CREATE UNIQUE INDEX ON mv_total_size(total_size);

DROP MATERIALIZED VIEW IF EXISTS mv_total_api_keys;
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_total_api_keys
AS
select count(total_api_keys) from (select count(*) as total_api_keys from contents group by requesting_api_key) as total_api_keys;
CREATE UNIQUE INDEX ON mv_total_api_keys(total_api_keys);