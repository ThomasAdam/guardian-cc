#!/usr/bin/perl

use strict;
use warnings;

use JSON;
use File::Basename;
use File::Path qw/make_path/;
use File::Glob ':glob';
use File::Copy;
use POSIX 'strftime';
use List::MoreUtils qw/mesh pairwise/;
use Data::Dumper;

use constant CROSSWORD_PATH => "./crosswords/cryptic/setter/";

our $setter_cache;
BEGIN {
	my @names_local = map {
		basename($_)
	} grep { -d } bsd_glob(CROSSWORD_PATH . "*");

	$setter_cache->{'names'} = \@names_local;
};

sub move_by_setter
{
	foreach my $file (glob(CROSSWORD_PATH . "*.JSON")) {
		my $content;
		open($content, "<", $file) or die "Couldn't open $file: $!";
		my $j_content = do { local $/ = undef; <$content> };
		my $data = from_json($j_content);
		my $setter = $data->{'creator'}->{'name'};

		if ($setter eq '' or !defined $setter) {
			$setter = "unknown";
		}

		if (! -d CROSSWORD_PATH/"$setter") {
			make_path(CROSSWORD_PATH . "$setter");
		}
		move("./$file", CROSSWORD_PATH . "$file");
	}
}

sub by_setter
{
	my @names = @{ $setter_cache->{'names'} };
	my %n_and_n = map {
		my @dates;

		# Use 'bsd_glob' to handle spaces in directory names.
		my @f = bsd_glob(CROSSWORD_PATH . "$_/*.JSON");

		foreach my $file (@f) {
			open(my $content, "<", $file) or die "Couldn't open $file: $!";
			my $j_content = do { local $/ = undef; <$content> };
			my $data = from_json($j_content);

			# The date is in milliseconds.
			push @dates, strftime '%Y-%m-%d',
				localtime($data->{'date'} / 1000);
		}
		$_ => {
			total => scalar @f,
			dates => \@dates,
		};

	} @names;
	
	return \%n_and_n;
}

sub setter_per_year
{
	my @names = @{$setter_cache->{'names'}};
	my %n_and_n = map {
		my $dates;

		# Use 'bsd_glob' to handle spaces in directory names.
		my @f = bsd_glob(CROSSWORD_PATH . "$_/*.JSON");

		foreach my $file (@f) {
			open(my $content, "<", $file) or die "Couldn't open $file: $!";
			my $j_content = do { local $/ = undef; <$content> };
			my $data = from_json($j_content);

			# The date is in milliseconds.
			my $year = strftime '%Y', localtime($data->{'date'} / 1000);
			$dates->{$year}++;
		}
		$_ => {
			dates => $dates,
		};

	} @names;

	return \%n_and_n;
}

sub setter_by_word_frequency
{
	my @names = @{$setter_cache->{'names'}};
	my %n_and_n = map {
		my $words;

		# Use 'bsd_glob' to handle spaces in directory names.
		my @f = bsd_glob(CROSSWORD_PATH . "$_/*.JSON");

		foreach my $file (@f) {
			open(my $content, "<", $file) or die "Couldn't open $file: $!";
			my $j_content = do { local $/ = undef; <$content> };
			my $data = from_json($j_content);

			foreach my $e (@{ $data->{'entries'} }) {
				$words->{$e->{'solution'}}++;
			}
		}
		$_ => {
			words => $words,
		};
	} @names;

	return \%n_and_n;
}

sub write_html
{
	my $setter_data = by_setter();
	my $setter_year = setter_per_year();
	my $setter_freq = setter_by_word_frequency();
	
	my @names = @{$setter_cache->{'names'}};
	my @labels = sort keys %$setter_data;
	my @values = map { $setter_data->{$_}->{'total'} } sort keys %$setter_data; 

	#my @dps = pairwise { [($a, $b)] } @labels, @values;
	my @dps = (['Setters', @values]);
	my $chart_data = {
		'bindto' => '#myChart',
		'size' => {
			'height' => 800
		},
		'data' => {
			'columns' => \@dps,
			'type' => 'bar',
		},
		'axis' => {
			'x' => {
			'type' => 'category',
				'tick' => {
					'rotate' => '75',
					'multiline' => 0
				},
				'height' => 0,
				'categories' => \@labels,
			},
			'y' => {
				'label' => 'No. of crosswords set',
				'max' => 800,
				'tick' => {
					'steps' => 20,
				}
			}
		},

	};
	my $json_obj = JSON->new();
	my $to_go = $json_obj->pretty->encode($chart_data);

	my @year_range = (1998..2018);
	my @setters = sort keys %$setter_year;

	my @chart2_data = map {
		my @d;

		foreach my $y (@year_range) {
			if (!exists $setter_year->{$_}->{'dates'}->{$y}) {
				push @d, 0;
			} else {
				push @d, $setter_year->{$_}->{'dates'}->{$y};
			}
		}
		[$_, @d]
	} @setters;

	my $chart_data2 = {
		'bindto' => '#myChart2',
		'size' => {
			'height' => 800
		},
		'data' => {
			'columns' => \@chart2_data,
			'type' => 'area',
		},
		'axis' => {
			'x' => {
			'type' => 'category',
				'tick' => {
					'rotate' => '75',
					'multiline' => 0
				},
				'height' => 0,
				'categories' => \@year_range,
			},
			'y' => {
				'label' => 'Crosswords per year',
			}
		},

	};
	my $to_go2 = $json_obj->pretty->encode($chart_data2);

	my @chart3_data = map {
		my %words;

		foreach my $w (keys %{ $setter_freq->{$_}->{'words'} }) {
			if ($setter_freq->{$_}->{'words'}->{$w} > 1) {
				$words{$setter_freq->{$_}->{'words'}->{$w}} = 1;
			}
		}
		[scalar keys %words]
	} @names;



	my @dps3 = (['Setters', @chart3_data]);
	my $chart_data3 = {
		'bindto' => '#myChart3',
		'size' => {
			'height' => 800
		},
		'data' => {
			'columns' => \@dps3,
			'type' => 'bar',
		},
		'axis' => {
			'x' => {
			'type' => 'category',
				'tick' => {
					'rotate' => '75',
					'multiline' => 0
				},
				'categories' => \@names,
				'height' => 0,
			},
			'y' => {
				'label' => 'Frequency of duplicated answers',
			}
		},

	};
	my $to_go3 = $json_obj->pretty->encode($chart_data3);

	my $html = <<HTML;
<html>
<head>
<link href="./chart_files/c3-0.4.18/c3.css" rel="stylesheet">
<link href="./gcc.css" rel="stylesheet">
<script src="https://d3js.org/d3.v3.min.js"></script>
<script src="./chart_files/c3-0.4.18/c3.min.js"></script>
</head>
<body>
<h1>Guardian Cryptic Crossword Analysis</h2>
<p>This page renders some charts from data gathered via the Guardian newspaper's
cryptic crosswords.  The data shown here was scraped via the web.</p>

<p>The git repository containing this data <a href="https://github.com/ThomasAdam/guardian-cc">is here.</a></p>

<h2>1.  Number of crosswords set per setter</h2>
<p>This chart shows the number of crosswords set per setter.  No real surprises
here as to the most prolific setters.</p>
<div id="myChart"></div>
<script>
var chart = c3.generate(
    $to_go
);
setTimeout(function () {
    chart.transform('line');
}, 8000);
</script>
<br />
<hr />
<h2>2.  Crosswords per year, per setter</h2>
<p>This chart shows an area span for the number of crosswords set per setter,
per year.  Interesting to see when a setter started and stopped.</p>
<div id="myChart2"></div>
<script>
var chart2 = c3.generate(
    $to_go2	
);
</script>
<br />
<hr />
<h2>3.  Frequency of word duplications across all crosswords, per setter</h2>
<p>This chart shows the number of words a given setter has used more than once,
across all the crosswords that setter has set.</p>
<div id="myChart3"></div>
<script>
var chart2 = c3.generate(
    $to_go3	
);
</script>
</body>
</html>
HTML

open my $out, ">", "./gcc-analysis.html" or die;
print $out $html;
close $out;
}

write_html;
