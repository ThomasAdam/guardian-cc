package Guardian::Cryptic::Plugins::chart1;

use lib "$ENV{'HOME'}/guardian-cc-import/Guardian-Cryptic-Crosswords/lib";
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
				'$group' => {
					'_id'   => '$creator.name',
					'count' => {'$sum' => 1},
				}
			},
			{
				'$sort' => {'count' => -1}
			}
	]);

	my %data = map {
		$_->{'_id'} => $_->{'count'}
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
		'title' => "Total number of crosswords, set by author",
		'preamble' => "This chart shows the number of crosswords set " .
			      "per setter.  No real surprises here as to the " .
			      "most prolific setters.",
		'order' => 1,
		'div_id' => 'mychart1',
		'js_var' => 'chart1',
		'default_chart' => "bar",
		'chart' => {
			'bindto' => '#myChart1',
			'size' => {
				'height' => 800
			},
			'data' => {
				'columns' => \@ordered_axis,
				'type' => 'bar',
				'empty' => {
					'label' => {
						'text' => 'Unknown'
					}
				},
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
					'label' => 'No. of crosswords set',
					'max' => 800,
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
