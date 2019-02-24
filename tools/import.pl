#!/usr/bin/env perl

use strict;
use warnings;

use MongoDB;
use JSON;
use File::Find;
use Data::Dumper;
use DateTime;

my @file_path = ("./crosswords/cryptic/setter", "./crosswords/prize/setter");
my @files;
my $client = MongoDB->connect();
my $db = $client->get_database('guardian');
my $cc = $db->get_collection('cryptic');

sub wanted
{
	$_ =~ /\.JSON$/ && push @files, $File::Find::name;
}

if (!defined $ARGV[0] or $ARGV[0] eq "") {
	find({wanted => \&wanted}, @file_path);
} else {
	push @files, $ARGV[0];
}

my $json_obj = JSON->new();
my $j_text;
foreach my $f (@files) {
	my $j = do {
		open my $fh, "<", $f or die;
		local $/ = undef;
		<$fh>;
	};

	eval {
		$j_text = $json_obj->decode($j);
	};
	die if $@;

	$j_text->{'creator'}->{'name'} = "Unknown"
		unless exists $j_text->{'creator'}->{'name'};

	$j_text->{'creator'}->{'webUrl'} = "http://www.example.org"
		unless exists $j_text->{'creator'}->{'webUrl'};

	# There's one case where a setter's name has a space at the end of it.
	# So that this matches with the other setter, this should be corrected.
	$j_text->{'creator'}->{'name'} =~ s/\s+$//;

	# For the benefit of mongodb, we can't have nice things...
	foreach my $h (@{ $j_text->{'entries'} }) {
		my $sl = $h->{'separatorLocations'};
		if (exists($sl->{'.'})) {
			my $temp_h = $sl->{'.'};
			delete $h->{'separatorLocations'}->{'.'};
			$h->{'separatorLocations'}->{'-'} = $temp_h;
		}
	}

	# Fix up the date...
	my $dt = DateTime->from_epoch(epoch => $j_text->{'date'} / 1000);
	$j_text->{'date'} = "$dt";

	my $r = $cc->insert_one($j_text);
	print "Added: " . $r->inserted_id . "\n";
}
