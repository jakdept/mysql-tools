## Reference for dump-split.pl ##

Typically run with:

```
$ dump-split.pl mysqldump_to_split.sql /directory/to/put/files/in/
```

## Flags ##
* `per-table` - default off
	* If enabled, per table dumps are created instead of per database dumps
* `create-info` - default on
	* If enabled, create table and drop table statements are written as applicable (see `wipe-table` flag)
	* Can be disabled with `no-create-info`
* `create-db-info` - default is off
	* If enabled, create databse and drop database statements are written as applicable (see `wipe-database` flag)
	* Can be disabled with `no-create-info`
* `wipe-table` - default is on
	* If enabled, drop table statements are written as `DROP TABLE ... IF EXISTS` and create table statements are written as  `CREATE TABLE`
	* If disabled, drop table statements are not written, and create table statements are written as `CREATE TABLE IF NOT EXISTS`
	* Can be disabled with `no-wipe-table`
* `wipe-database` - default is off
	* If enabled, drop database statements are written as `DROP DATABASE ... IF EXISTS` and create database statements are written as  `CREATE TABLE`
	* If disabled, drop database statements are not written, and create database statements are written as `CREATE TABLE IF NOT EXISTS`
	* Can be disabled with `no-wipe-database`
* `force-overwrite` - disabled by default
	* If enabled, INSERT statements are written as `INSERT INTO ... ON DUPLICATE KEY UPDATE` so it replaces the row on conflicts
	* If disabled, INSERT statements are written as `INSERT IGNORE INTO ...` so duplicates are not replaced, but do not cause the import to fail

## Credit ##

The fault for this script lies with Jack Hayhurst jhayhurst@liquidweb.com