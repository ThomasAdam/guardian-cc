#!/usr/bin/env perl

use lib "$ENV{'HOME'}/guardian-cc-import/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::ChartRenderer;

use strict;
use warnings;

my $cr = Guardian::Cryptic::ChartRenderer->new();
my @plugins = $cr->plugins();

my %collective_data;

foreach my $p (@plugins) {
	warn "Looking at: " . $p . "...\n";
	my $order = $p->render() if $p->can("render");
	$collective_data{$order} = ${ $p->tmpl_data() },
}

my $tmpl_file = "main.tmpl";
$cr->save(file => "main.tmpl", content => \%collective_data);

my $html = $cr->tmpl_data;
open my $out, ">", "./gcc-analysis.html" or die;
binmode($out, ':utf8');
print $out $$html;
close $out;
