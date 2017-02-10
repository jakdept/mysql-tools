mysqlstat.pl
	used to determine MySQL's memory usage and more

## Example Output ##

```
root@host.jhayhurst.com [~]# ./mysqlstat.pl --query-cache --datadir files

Information for host.jhayhurst.com running MySQL Version 5.0.96-community
[YAS] Engines: +InnoDB +MyISAM +MEMORY +BLACKHOLE +ARCHIVE
[YAS] There are currently 70 files open - the limit is 1024
[YAS] SELECT queries are waiting on a lock 0.000 % of the time - there have been 883 requests

[YAS] 0.003 % of your queries are slow by the threshold 10 seconds

[YAS] This is 9 queries out of a total of 2955


########## Memory Usage per connections ##########
                             |       base |        now |       peak |      limit
            # of connections |          1 |          1 |          2 |        150
     innodb_buffer_pool_size |       8.0M |       8.0M |       8.0M |       8.0M
             key_buffer_size |      16.0M |      16.0M |      16.0M |      16.0M
            query_cache_size |       8.0M |       8.0M |       8.0M |       8.0M
          Total Memory Guess |      41.2M |      41.2M |      50.5M |       1.4G

########## Detailed Engine Total by Files (fast) ##########
[FAK] MyISAM total space usage (tables & indexes) 15.2M
[FAK] MyISAM key_buffer_size is 16.0M - 1.509% of the engine usage
[UHH] MyISAM contains 15.2M of data and 8.9M of indexes
[UHH] key_buffer_size has been 0.020 % full

[FAK] MyISAM total space usage (tables & indexes) 10.6G
[FAK] MyISAM innodb_buffer_pool_size is 8.0M - 679.625% of the engine usage
[UHH] When using the short method, the usage difference between data and indexes is not available
[UHH] innodb_buffer_pool_size has been 0.039 % full

########## Query Cache Statistics ##########
[FAK] The MySQL Query Cache hit rate is 2/1678 - 0.001 %
[FAK] You cannot touch 0.701 % of these queries or 1177 queries

########## Security Concerns ##########
[FAK] old_password discovered - localhost @ imscared
[FAK] old_password discovered - localhost @ mail
[FAK] blank password discovered - localhost @ badjak
```

## Flags ##
* `--table-locks` - prints out information about wait time for table locks
* `--security` - checks for users with old style passwords and no password set
* `--status-all-engines` - prints out status information for all engines
* `--check-slow_queries` - reports information on slow queries
* `--datadir-size` - used to specify a method to calculate the amount stored in the datadir
  * `files` - works based on file inodes (fastest)
  * `query` - works based on a long query method and gives InnoDB primary index vs secondary index (slower) 
  * both methods return the `key_buffer_size` and `innodb_buffer_pool_size` max fill %

* `--mem-display` - determines the level of memory usage output
  * `full` - usage for every setting is displayed
  * `auto` - critical values and anything set in the my.cnf are displayed
  * `summary` - only the critical values and summary are displayed
  * `none` - nothing is displayed

* `--scale-`* - used to adjust/scale the different values where applicable
  * `scale-innodb_buffer_pool_size` - determines the expeceted scale between `innodb_buffer_pool_size` and InnoDB tablespace
  * `scale-key_buffer_size` - determines the expected scale between `key_buffer_size` and MyISAM tablespace
  * `scale-read_buffer_size` - how many times to count `read_buffer_size` against each thread
  * `scale-read_rnd_buffer_size` - how many times to count `read_rnd_buffer_size` against each thread
  * `scale-sort_buffer_size` - how many times to count `sort_buffer_size` against each thread
  * `scale-myisam_sort_buffer_size` - how many times to count `myisam_sort_buffer_size` against each thread
    * `myisam_sort_buffer_size` *should* only apply to alters, repairs, and optimizes (in theory) 
  * `scale-max_allowed_packet` - the maximum size of a network buffer to allocate per thread
    * note, `net_buffer_length` is the actual size of the buffer, `max_allowed_packet` is an upper bound

* `--query-cache` - prints out information on the query_cache within MySQL
  * Useful informtion can be found at:
    * http://dev.mysql.com/doc/refman/5.1/en/query-cache.html
    * http://dev.mysql.com/doc/refman/5.1/en/query-cache-operation.html
    * http://dev.mysql.com/doc/refman/5.1/en/query-cache-configuration.html
    * http://dev.mysql.com/doc/refman/5.1/en/query-cache-status-and-maintenance.html
  * MySQL query_cache was designed for a dual core, 500Mhz Linux system - it's dated.
  * Different things to keep in mind on the `query-cache` system within MySQL
    * Only SELECT queries can be fleshed out
    * Query case matters - if both "SELECT * FROM `table`;" and "select * from `table`;" were run, it would be cached twice
    * Query cannot be cached if it contains:
      * `BENCHMARK()`
      * `CURDATE()`, `CURRENT_DATE()`, `CURRENT_TIMESTAMP()`, `NOW()`, `CURTIME()`, `SYSDATE()`, `CURRENT_TIME()`
      * `LAST_INSERT_ID()`, `CONNETION_ID()`, `MASTER_POS_WAIT()`
      * `SLEEP()`
      * `USER()`
      * `FOUND_ROWS()`
      * `RAND()`
      * `UUID()`
      * `GET_LOCK()`, `RELEASE_LOCK()`
      * `DATABASE()`
    * Queries also cannot be cached if they:
      * Deal with `INFILE` or `OUTFILE` in any way - loading or exporting data
      * If they use a tempoary table
      * Sub-queries are not cached
      * Executing user has a column livel privilege for any involved table
      * Query does not use any table
      * Query generates any warnings or errors
      * Query touches something in `mysql` or `INFORMATION_SCHEMA` or `PERFORMANCE_SCHEMA`

## Planned future features ##
- [x] `query-cache` - break down query_cache usage
- [x] `table-locks` - break apart table lock problems and indicators
- [x] `datadir-size` - more options
	- [x] `query` - gather datadir information based on the old (bad) query method
- [x] `security` - checks to look for old/no passwords for users, find alternate superusers, find remote root users
- [x] add engine check?
- [x] `slow-query` - flag to break apart slow query issues
- [x]  `open-files` - check for open files
- [x] `key_buffer_size` filled % and hit rate, same thing with `innodb_buffer_pool_size`

## Known Issues ##
* The `query_cache` statistics are off
  * maybe use Questions from show global stat; to answer this?
  * Just trying using `mysql -e 'show global status LIKE "Questions";'` for now

Major contributions by:
		Jack Hayhurst - jhayhurst@liquidweb.com

## external credit ##

* http://www.omh.cc/mycnf/
* http://mysqltuner.pl/
* http://www.percona.com/blog/2006/05/17/mysql-server-memory-usage/
* http://shooltz.net/stats/stats
