package Guardian::Cryptic::Plugins::chart2;

use lib "$ENV{'HOME'}/projects/cc/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::Crosswords;
use POSIX 'strftime';

sub new
{
	my ($class) = @_;

	return bless {}, $class;
}

sub interpolate
{
	my ($self) = @_;

	my $setters = Guardian::Cryptic::Crosswords::setters();
	my %data;

	foreach (@$setters) {
		my $name = $_->name();
		my $dates = $_->date();

		foreach my $d (@$dates) {
			$data{$name}->{$d}++;
		}
	}

	return \%data;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();
	my @setters = sort keys %$interdata;

	my $current_year = strftime "%Y", localtime;
	my @year_range = (1998..$current_year);
	my @chart2_data = map {
		my @d;
		my $name = $_;

		foreach my $y (@year_range) {
			if (!exists $interdata->{$name}->{$y}) {
				push @d, 0;
			} else {
				push @d, $interdata->{$name}->{$y};
			}
		}
		[$_, @d],
	} @setters;

	my $data = {
		'title' => "Crosswords per year, per setter",
		'preamble' => "This chart shows an area span for the number of crosswords set per setter, per year.  Interesting to see when a setter started and stopped.",
		'order' => 2,
		'div_id' => 'mychart2',
		'chart' => {
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
		},
	};

	return $data;
}

1;
