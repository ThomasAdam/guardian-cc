package Guardian::Cryptic::Plugins::chart3;

use lib "$ENV{'HOME'}/projects/guardian-cc-import/Guardian-Cryptic-Crosswords/lib";
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
	my @labels = sort keys %$interdata;
	my @values = map { $interdata->{$_} } @labels;

	my @clabels = (['Setters', @values]);

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
				'columns' => \@clabels,
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
