package Guardian::Cryptic::Plugins::chart3;

use lib "$ENV{'HOME'}/projects/guardian-cc/Guardian-Cryptic-Crosswords/lib";
use List::Util qw/max/;

use parent 'Guardian::Cryptic::ChartRenderer';

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
				'$unwind' => '$entries'
			},
			{
				'$group' => {
					'_id' => {
						'name' => '$creator.name',
						'solution' => '$entries.solution'
					},
					'count' => {
						'$sum' => 1
					}
				}
			},
			{
				'$match' => {
					'count' => {
						'$gt' => 1
					}
				}
			},
			{
				'$group' => {
					'_id' => '$_id.name',
					'max' => {
						'$max' => '$count'
					}
				}
			},
			{
				'$sort' => {
					'max' => -1
				}
			}
	]);

	my %data = map {
		$_->{'_id'} => $_->{'max'}
	} $res->all;

	return \%data;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();
	my @ordered_values;
	my @ordered_labels;

	foreach my $k (sort {$interdata->{$b} <=> $interdata->{$a}}
	    keys %$interdata)
	{
		push @ordered_labels, $k;
		push @ordered_values, $interdata->{$k};
	}

	my @ordered_axis = (["Setters", @ordered_values]);

	my $data = {
		'title' => "Frequency of word duplications across all " .
			   "crosswords, per setter",
		'preamble' => "This chart shows the number of words a given " .
			      "setter has used more than once, across all " .
			      "crosswords for that setter.",
		'order' => 3,
		'div_id' => 'mychart3',
		'js_var' => 'chart3',
		'default_chart' => 'bar',
		'chart' => {
			'bindto' => '#myChart3',
			'size' => {
				'height' => 800
			},
			'data' => {
				'columns' => \@ordered_axis,
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
					'categories' => \@ordered_labels,
				},
				'y' => {
					'label' => 'Frequency of duplicated ' .
						   'answers',
					'tick' => {
						'steps' => 20,
					}
				}
			},
		},

	};

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
