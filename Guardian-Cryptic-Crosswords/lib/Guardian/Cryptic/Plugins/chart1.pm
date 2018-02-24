package Guardian::Cryptic::Plugins::chart1;

use lib "$ENV{'HOME'}/projects/cc/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::Crosswords;

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

	my $setters = Guardian::Cryptic::Crosswords::setters();

	my %data = map {
		my ($name, $total) = ($_->name(), $_->total());
		$name => $total
	} @$setters;

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
		'chart' => {
			'bindto' => '#myChart1',
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
