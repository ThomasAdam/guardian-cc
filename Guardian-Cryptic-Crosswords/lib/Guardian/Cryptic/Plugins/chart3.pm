package Guardian::Cryptic::Plugins::chart3;

use lib "$ENV{'HOME'}/projects/cc/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::Crosswords;
use POSIX 'strftime';
use List::Util qw/max/;

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
		my $ids = $_->ids();

		foreach my $id (@$ids) {
			my $answerset = $_->answers(id => $id, join => 1);
			foreach my $clue (keys %$answerset) {
				my $answer = 
				    $answerset->{$clue}->{'joined_solution'} //
			            $answerset->{$clue}->{'solution'};
				$data{$name}->{$answer}++;
			}
		}
	}

	foreach my $k (keys %data) {
		foreach my $w (keys %{ $data{$k} }) {
			if ($data{$k}->{$w} == 1) {
				delete $data{$k}->{$w};
			}
		}
	}

	return \%data;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();
	my @labels = sort keys %$interdata;
	my %value_map = map {
		$_ => {
			words => $interdata->{$_},
			highest => max values %{ $interdata->{$_} },
		}
	} @labels;

	my @values = map { $value_map{$_}->{'highest'} } @labels;

	my @clabels = (['Setters', @values]);

	my $data = {
		'title' => "Frequency of word duplications across all " .
			   "crosswords, per setter",
		'preamble' => "This chart shows the number of words a given " .
			      "setter has used more than once, across all " .
			      "crosswords for that setter.",
		'order' => 3,
		'div_id' => 'mychart3',
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

	return $data;
}

1;
