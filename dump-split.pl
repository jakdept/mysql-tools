#!/bin/env perl

use strict;
use warnings;
use diagnostics;
use Getopt::Long;

# set default options
my %opt = (
		"debug"            => 0,
		"help"             => 0,
		"usage"            => 0,
		"per-table"        => 0,
		"create-info"      => 1,
		"create-db-info"   => 0,
		"wipe-table"       => 1,
		"wipe-database"    => 0,
		"force-overwrite"  => 0
);

# Gather the options from the command line
GetOptions(\%opt,
		'debug+',
		'help',
		'usage',
		'per-table!',
		'create-info!',
		'create-db-info!',
		'wipe-table!',
		'wipe-database!',
		'force-overwrite!'
);

sub usage {
	# Shown with --help option passed
	print "\n".
		"   dump-split - a script to pull apart MySQL dumps\n".
		"   Internal bug reports and feature requests go to the author\n".
		"   This tool is for internal use only\n".
		"\n".
		"   General usage:\n".
		"     dump-split.pl /path/to/mysql/dump/file.sql /empty/output/directory/\n".
		"\n".
		"   Important Usage Guidelines:\n".
		"      * Your input file should be a single file - no symlinks\n".
		"      * Your output directory should be empty\n".
		"\n".
		"   Flags:\n".
		"      --per-table          Dumps are created per table instead of just per database (default disabled)\n".
		"      --create-info        CREATE and DROP TABLE statements are written (default enabled)\n".
		"      --create-db-info     CREATE and DROP DATABASE statements are written (default disabled)\n".
		"      --wipe-table         Write statements for tables to wipe out tables (default enabled) (depends on create-info)\n".
		"      --wipe-database      Write statements for databases to wipe out databases (default enabled) (depends on create-db-info)\n".
		"      --force-overwrite    Write INSERT statements that overwrite conflicting data (default enabled)\n".
		"      --help OR --usage    Print out this information\n".
		"\n".
		"   For more information please see the Readme\n".
		"\n";
	exit;
}

sub is_dir_empty {
	my $target_dir = shift;

	if (! -d $target_dir) {
		# the directory doesn't exist, so we're going to have to create it
		mkdir $target_dir or return 0;
	} else {
		opendir(DIR, $target_dir) or die("failed to check the directory");
		while(my $entry = readdir DIR) {
			next if($entry ~=/^\.\.?$/);

			# if we got to this point, $entry is pointing to something in this directory
			closedir DIR;
			return 0;
		}
	}
	# if we got to this point, the directory either was just created or had nothing in it
	# of note - return code 1 is true
	return 1;
}

sub open_output {
	my $output_dir = shift;
	my $current_database = shift;
	my $current_table = "";
	if (exists $_[0]) {$current_table = shift;};
	my $target_file = "";
	if($output_dir ~= !/^.*\/$/) {$output_dir = $output_dir."/";} # add a trailing slash if it's not there

	if(tell(OUTFILE) =! -1) {close (OUTFILE);}; # if the filehandle is already open, close it

	if($current_database ~= /^$/) {
		$target_file = $output_dir.$current_table.".sql";
	} elsif ($current_table ~= /^$/) {
		$target_file = $output_dir.$current_database.".sql";
	} else {
		$target_file = $output_dir.$current_database."."$current_table.".sql";
	}

	open(OUTFILE, ">>", $target_file);
}


sub process_file {
	my $output_dir = shift;
	my $read_file = shift;
	my $current_database = "";
	my $current_table = "";

	if(exists $_[0]) {
		my $current_database = shift;
	}

	open(READFILE, "<", $read_file);

	while (<READFILE>) {
		$line = $_;
		if ($line ~= /^\s*DROP\s+DATABASE\s+(IF\s+EXISTS\s+)?`?(\w+)`?.*;$/i) {
			if (! $current_database eq $2) {
				$current_database = $2;
				if ($opt{'create-db-info'} == 1 && $opt{'wipe-database'} == 1) {
					open_output($output_dir, $current_database);
					print OUTFILE "DROP DATABASE IF EXISTS `$2`;\n";
				}
			}
		} elsif ($line ~= /^\s*CREATE\s+DATABASE\s+(IF\s+EXISTS\s+)?`?(\w+)`?(.*);$/i) {
			if (! $current_database eq $2) {
				$current_database = $2;
			}
			if ($opt{'create-db-info'} == 1) {
				open_output($output_dir, $current_database);
				if ($opt{'wipe-database'} == 1) {
					print OUTFILE "CREATE DATABASE `$2`$3;\n";
				} else {
					print OUTFILE "CREATE DATABASE IF NOT EXISTS `$2`$3;\n";
				}
			}
		} elsif($line ~= /^\s*DROP\s+TABLE\s+(IF\s+EXISTS\s+)?`?(\w+)`?;$/i ) {
			# find any drop table statements - $2 is the table name
			if($current_table != $2 && $opt{'per-table'} == 1) {
				$current_table = $2;
				open_output($output_dir, $current_database, $current_table);
			}
			if ($opt{'create-info'} == 1 && $opt{'wipe-table'} == 1) {
				print OUTFILE "DROP TABLE IF EXISTS `$2`;\n";
			}
		} elsif($line ~= /^\s*CREATE\s+TABLE\s+(IF\s+NOT\s+EXISTS\s+)?`?(\w+)`?(.*)$/i ) {
			# find any create table statements  - note no semi-colon on the end
			if($current_table != $2) {
				$current_table = $2;
				open_output($output_dir, $current_database, $current_table);
			}
			if ($opt{'create-info'} == 1) {
				if ($opt{'wipe-table'} == 1) {
					print OUTFILE "CREATE TABLE `$2`$3";
				} else {
					print OUTFILE "CREATE TABLE IF NOT EXISTS `$2`$3";
				}
			}
		} elsif ($line ~= /^\s*INSERT\s+(IGNORE\s+)?(INTO\s+)?`?(\w+)`?\s+(\(.*\)\s+VALUES)?\s+(.*)(ON\s+DUPLICATE\s+KEY\s+UPDATE)?$/i) {
			# we should be looking at INSERT statements here
			# $1 IGNORE portion is optional and we won't use
			# $2 INTO is optional and we won't use from original
			# $3 is the name of the table to insert into
			# $4 is the column names to insert into, along with the keyword "VALUES" - this will need to be dumped in
			# $5 is the actual values that are being inserted
			# $6 is duplicate key update and will be discarded
			# we actually use $3, $4, and $5 below
			if($current_table != $3) {
				$current_table = $3;
				open_output($output_dir, $current_database, $current_table);
			} 
			if($opt{'force-overwrite'} == 1) {
				print OUTFILE "INSERT INTO `$3` $4 $5 ON DUPLICATE KEY UPDATE;";
			} else {
				print OUTFILE "INSERT IGNORE INTO `$3` $4 $5;";
			}
		} else {  # fallthrough for any lines that have not matched yet - just write them to the current file as is
			print OUTFILE $line;
		}
	}
}

##### main section #####

if ($opt{'help'} == 1 || $opt{'usage'} == 1) {usage();};

if (! -e $ARGV[0]) {
	print "You should provide a real input file\n";
	usage();
} elsif (! -f $ARGV[0]) {
	print "You should provide a real input file?\n";
	usage();
}

if (! -e $ARGV[1]) {
	print "You should provide a folder for output\n";
	usage();
} elsif (! -d $ARGV[1]) {
	print "You should probably provide a folder for output\n";
	usage();
} elsif (is_dir_empty($ARGV[1]) {
	print "You should provide an empty directory\n";
	usage();
}

process_file($ARGV[0], $ARGV[1]);