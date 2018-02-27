package Guardian::Cryptic::Plugins::chart4;

use lib "$ENV{'HOME'}/projects/cc/Guardian-Cryptic-Crosswords/lib";
use Guardian::Cryptic::Crosswords;
use POSIX 'strftime';
use List::Util qw/max/;

use parent 'Guardian::Cryptic::ChartRenderer';

my $tmpl_file = "chart4.tmpl";

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
	my %data;
	my %seen;

	foreach (@$setters) {
		my $name = $_->name();
		my $ids = $_->ids();

		foreach my $id (@$ids) {
			my $answerset = $_->answers(id => $id, join => 1);
			foreach my $clue (keys %$answerset) {
				my $q = $answerset->{$clue}->{'clue'};
				my $answer = 
				    $answerset->{$clue}->{'joined_solution'} //
			            $answerset->{$clue}->{'solution'};
				$seen{$name}->{$answer}++;
				$data{$name}->{$answer}->{'seen'} = 
					$seen{$name}->{$answer};
				push @{$data{$name}->{$answer}->{'question'}},
					$q;
			}
		}
	}

	# Go back through the dataset and eliminate those entries less than
	# twice.  There's too much data to display.
	foreach my $k (keys %data) {
		foreach my $w (keys %{ $data{$k} }) {
			if ($data{$k}->{$w}->{'seen'} == 1) {
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
	my $data = {
		'title' => "Number of duplicate answers and their questions",
		'preamble' => "This table shows the number of times a given " .
		              "clue has been used and the different questions " .
			      "which have been used to make up that clue.",
		'order' => 4,
		'table' => $interdata,
	};

	$self->save(file => $tmpl_file, content => $data); 

	return $data->{'order'};
}

1;
