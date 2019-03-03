package Guardian::Cryptic::Plugins::chart1;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";
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
					'_id'   => {
						name => '$creator.name',
						type => '$crosswordType',
					},
					'count' => {'$sum' => 1},
				}
			},
			{
				'$sort' => {'count' => -1, '_id.type' => 1}
			}
	]);

	my @res = $res->all;
	return \@res;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();
	my (@cryptic, @prize, @labels, @columns);
	my %name_map = ();

	use Data::Dumper;

	my $position = 0;
	foreach my $k (@$interdata) {
		my $setter = $k->{'_id'}->{'name'};
		my $type   = $k->{'_id'}->{'type'};
		my $count  = $k->{'count'};

		$name_map{$setter}->{$type} = {
			pos => ++$position,
			count => $count
		};

		if (!exists $name_map{$setter}->{'prize'}) {
			$name_map{$setter}->{'prize'} = {
				'pos' => $name_map{$setter}->{'cryptic'}->{'pos'},
				'count' => 0,
			};
		}

		if (!exists $name_map{$setter}->{'cryptic'}) {
			$name_map{$setter}->{'cryptic'} = {
				'pos' => $name_map{$setter}->{'prize'}->{'pos'},
				'count' => 0,
			};
		}
	}

	@labels = ("x");

	foreach my $s (sort {
		$name_map{$a}->{'cryptic'}->{'pos'} <=>
		$name_map{$b}->{'cryptic'}->{'pos'} } keys %name_map) {

		push @labels, $s;
		push @cryptic, $name_map{$s}->{'cryptic'}->{'count'};
		push @prize,   $name_map{$s}->{'prize'}->{'count'};
	}
	unshift @cryptic, "Cryptic";
	unshift @prize, "Prize";
	push @columns, \@labels, \@cryptic, \@prize;

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
				'x' => 'x',
				'columns' => \@columns,
				'type' => 'bar',
				'empty' => {
					'label' => {
						'text' => 'Unknown'
					}
				},
				'groups' => [ ["Prize", "Cryptic"] ],
			},
			'axis' => {
				'x' => {
				'type' => 'category',
					'tick' => {
						'rotate' => '75',
						'multiline' => 0
					},
					'height' => 0,
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
