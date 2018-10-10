package Guardian::Cryptic::Plugins::chart2;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";
use parent 'Guardian::Cryptic::ChartRenderer';

use POSIX qw/strftime/;

my $tmpl_file = "chart.tmpl";

sub new
{
	my ($class) = @_;

	my $self = $class->SUPER::new();
	bless $self, $class;
}

sub interpolate
{
	my ($self) = @_;

	my $res = $self->{'mongo'}->{'col'}->aggregate([
		{
			'$project' => {
				'_id' => '$creator.name',
				'ndate' => {
					'$year' => {
						'$dateFromString' => {
							'dateString' => '$date'
						}
					}
				}
			}
		},
		{
			'$group' => {
				'_id' => {
					'name' => '$_id',
					'year' =>  '$ndate'
				},
				'count' => {
					'$sum' => 1
				}
			}
		},
		{
			'$sort' => {'_id' => 1}
		}
	]);

	my @all = $res->all();
	my %data = ();

	foreach my $r (@all) {
		$data{$r->{'_id'}->{'name'}}->{ $r->{'_id'}->{'year'} } = $r->{'count'};
	} @all;

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
		'js_var' => 'chart2',
		'default_chart' => 'area',
		'chart' => {
			'bindto' => '#myChart2',
			'size' => {
				'height' => 800
			},
			'data' => {
				'columns' => \@chart2_data,
				'type' => 'area',
			},
			'tooltip' => {
				'show' => 0,
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

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
