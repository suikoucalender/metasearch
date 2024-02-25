#!/usr/bin/perl
use strict;
use warnings;
use Getopt::Long;

use constant {
    OPT_NO  => 0,
    OPT_YES => 1,
};

my $help = OPT_NO;
my $key_header;
my $key_header_bool;
my $message = << '&EOT&';
# Usage: *.pl [-k] <table 1> <table 2> [table 3] ...
#  -h --help       Print this message.
#  -k              Key header column exists (key must be unique, and the header must be in the first table).
&EOT&

my $result = GetOptions(
    'help|h'         => \$help,
    'k'            => \$key_header_bool,
			);
if($#ARGV<1){
    print STDERR "You must set table files.\n";
    $help = OPT_YES;
}
if($help == OPT_YES){
    print STDERR $message."\n";
    exit(0);
}

my %hash_table=();

#print Dumper \%hash_table;
my $line="";
my @ntab_list=();
for(my $i=0;$i<$#ARGV+1;$i++){
    open(IN, $ARGV[$i]);
    my $ntab_max=0;
    while($line=<IN>){
        #print $line;
#$line=~s/\n//;
#        my @tempn = split(/\t/,$line);
	my $tempn=$line=~tr/\t/\t/;
#print $line . ": " . $tempn . "\n";
        if($tempn > $ntab_max){
            $ntab_max = $tempn;
        }
    }
    close(IN);

    $ntab_list[$i]=$ntab_max;
    open(IN, $ARGV[$i]);
    while($line=<IN>){
        #print $line;
#        my @tempn = split(/\t/,$line);
	my $tempn=$line=~tr/\t/\t/;
        if($line=~m/^([^\t]+)\t(.*)/){
            $hash_table{$1}[$i]=$2;
            for(my $j=0;$j<$ntab_max-$tempn;$j++){
                $hash_table{$1}[$i]="$hash_table{$1}[$i]\t";
            }
        }
    }
    close(IN);
}

if(defined($key_header_bool)){
    open(IN, $ARGV[0]);
    my $header=<IN>;
    close(IN);
    my @header_items=split(/\t/,$header);
    $key_header=$header_items[0];
}

if(defined($key_header)){
    print $key_header;
    for(my $i=0;$i<$#ARGV+1;$i++){
        if(!defined($hash_table{$key_header}[$i])){
            $hash_table{$key_header}[$i]="";
#            if($ntab_list[$i]-2>0){
#                $hash_table{$key_header}[$i]="\t";
#            }
            for(my $j=1;$j<$ntab_list[$i];$j++){
                $hash_table{$key_header}[$i]=$hash_table{$key_header}[$i]."\t";
            }
        }
#else{
#    my $tempn=$hash_table{$key_header}[$i]=~tr/\t/\t/;
#    for(my $j=0;$j<$ntab_list[$i]-$tempn-1;$j++){
#$hash_table{$key_header}[$i]="$hash_table{$key_header}[$i]\t";
#    }
#}
	print "\t$hash_table{$key_header}[$i]";
    }
    print "\n";
}

foreach my $key (sort keys(%hash_table)){
    if(defined($key_header)){
        if($key ne $key_header){
            print "$key";
            for(my $i=0;$i<$#ARGV+1;$i++){
                if(!defined($hash_table{$key}[$i])){
                    $hash_table{$key}[$i]="";
#                    if($ntab_list[$i]-2>0){
#                        $hash_table{$key}[$i]="\t";
#                    }
                    for(my $j=1;$j<$ntab_list[$i];$j++){
                        $hash_table{$key}[$i]=$hash_table{$key}[$i]."\t";
                    }
                }
                print "\t$hash_table{$key}[$i]";
            }
            print "\n";
        }
    }
    else{
        print "$key";
        for(my $i=0;$i<$#ARGV+1;$i++){
            if(!defined($hash_table{$key}[$i])){
                $hash_table{$key}[$i]="";
#                if($ntab_list[$i]-2>0){
#                    $hash_table{$key}[$i]="\t";
#                }
                for(my $j=1;$j<$ntab_list[$i];$j++){
                    $hash_table{$key}[$i]=$hash_table{$key}[$i]."\t";
                }
            }
            print "\t$hash_table{$key}[$i]";
        }
        print "\n";
    }
}

