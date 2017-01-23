#!/usr/bin/perl

use warnings;
use strict;

my $file;
my @message_nums;

sub print_file {
    my ($file, $message_nums) = @_;
    print join(',', sort {$a <=> $b} @$message_nums) . " ";
    print "$file\n";
}

while (<>) {
    # Make this easy to use from a simple grep|sort of the log.
    my ($file_messageno) = split(' ');
    
    my ($nfile, $message_num) = split(':', $file_messageno);

    if (!$file) {
	$file = $nfile;
    }
    
    if ($nfile ne $file) {
	&print_file($file, \@message_nums);
	$file = $nfile;
	@message_nums = ();
    }
    push @message_nums, $message_num;
}
# Last entry
&print_file($file, \@message_nums);
