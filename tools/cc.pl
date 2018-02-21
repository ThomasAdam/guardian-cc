#!/usr/bin/env perl

use lib "$ENV{'HOME'}/projects/cc/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::Crosswords;
use Guardian::Cryptic::ChartRenderer;
use Template;
use JSON;

use Data::Dumper;

use strict;
use warnings;

my $cr = Guardian::Cryptic::ChartRenderer->new();
my @plugins = $cr->plugins();

my %collective_data;

foreach my $p (@plugins) {
	my $data = $p->render() if $p->can("render");
	$collective_data{$data->{'order'}} = $data;
}

my $json_obj = JSON->new();
my $html = <<HTML;
<html>
<head>
<link href="./ui/chart_files/c3-0.4.18/c3.css" rel="stylesheet">
<link href="./gcc.css" rel="stylesheet">
<script src="https://d3js.org/d3.v3.min.js"></script>
<script src="./ui/chart_files/c3-0.4.18/c3.min.js"></script>
</head>
<body>
<h1>Guardian Cryptic Crossword Analysis</h2>
<p>This page renders some charts from data gathered via the Guardian newspaper's
cryptic crosswords.  The data shown here was scraped via the web.</p>

<p>The git repository containing this data <a href="https://github.com/ThomasAdam/guardian-cc">is here.</a></p>
HTML

foreach my $d (sort { $a <=> $b } keys %collective_data) {
    my $chart_json = $json_obj->pretty->encode($collective_data{$d}->{'chart'});
    $html .= <<HTML;
<h2>$d.  $collective_data{$d}->{'title'}</h2>
<p>$collective_data{$d}->{'preamble'}</p>
<div id="$collective_data{$d}->{'div_id'}"></div>
<script>
var chart = c3.generate(
    $chart_json
);
</script>
<br />
<hr />
</body>
</html>
HTML

open my $out, ">", "./gcc-analysis.html" or die;
print $out $html;
close $out;
}
