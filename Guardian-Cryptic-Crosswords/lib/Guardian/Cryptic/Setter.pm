package Guardian::Cryptic::Setter;

use strict;
use warnings;

use Carp;
use JSON;
use File::Glob ':glob';
use POSIX 'strftime';
use Data::Dumper;

use constant CROSSWORD_PATH => "./crosswords/cryptic/setter/";

=head1 new

Standard constructor from the child class 'Crosswords'.
Passes in the Setter's name as an argument,

=cut

sub new
{
	my ($class, %args) = @_;

	if (!defined $args{'setter'}) {
		confess "Setter is not defined";
	}

	my $self = {
		'_setter' => $args{'setter'},
		'_loaded' => 0,
	};

	bless $self, $class;

	$self->_get_setter_data();
	return $self;
}

=head1 name

Returns the name of the setter this object refers to.

=cut

sub name
{
	my ($self) = @_;

	return $self->{'_setter'};
}

=head1 ids

Returns all the crossword IDs the setter for this object has, as an array
reference.

=cut

sub ids
{
	my ($self) = @_;

	my @ids = map {
		/(\d+)\.JSON/ and $1;
	} @{ $self->{'_files'} };

	return \@ids;
}

=head1 date

Returns the year every crossword that setter has produced, as an array
reference.

=cut

sub date
{
	my ($self) = @_;
	my $name = $self->name();

	$self->load() unless $self->{'_loaded'};

	my @dates = map {
		my $ts = $self->{'_data'}->{$name}->{$_}->{'date'};
		my $year = strftime '%Y', localtime($ts / 1000);

		$year;

	} keys %{ $self->{'_data'}->{$name} };

	return \@dates;
}

=head1 load

Loads all the JSON structures for each crossword for the setter, into RAM.

=cut

sub load
{
	my ($self, %args) = @_;
	my %data_block;

	foreach my $file (@{ $self->{'_files'} }) {
		open(my $content, "<", $file) or die "Couldn't open $file: $!";
		my $j_content = do { local $/ = undef; <$content> };
		my $data = from_json($j_content);

		$data_block{$data->{'number'}} = $data;
	}

	my $name = $self->name();

	$self->{'_data'} = { $name => \%data_block };
	$self->{'_loaded'} = 1;
}

=head1 entry

Returns the entries (an array ref of hashes) for each set of answers/clues.

=cut

sub entry
{
	my ($self, %args) = @_;

	my $name = $self->name();

	$self->load() unless $self->{'_loaded'};

	if (!defined $args{'id'} or $args{'id'} !~ /\d+/) {
		confess "Bad id given";
	}

	if (!exists $self->{'_data'}->{$name}->{$args{'id'}}) {
		confess "Couldn't find $name/$args{'id'}";
	}

	return $self->{'_data'}->{$name}->{$args{'id'}}->{'entries'};
}

=head1 total

Returns the total number of crosswords a given setter has produced.

=cut

sub total
{
	my ($self) = @_;
	my $name = $self->name();

	$self->load() unless $self->{'_loaded'};

	return scalar keys %{ $self->{'_data'}->{$name} };
}

=head1 answers

Returns a hashref contaning the key of the clue, the answer, and whether that
answer was grouped with another.  In such circumstances, one may pass in

   join => 1

to join such grouped answers togther, referened via the key

   joined_solution

=cut

sub answers
{
	my ($self, %args) = @_;

	my $name = $self->name();

	$self->load() unless $self->{'_loaded'};

	if (!exists $self->{'_data'}->{$name}->{$args{'id'}}) {
		confess "Couldn't find $name/$args{'id'}";
	}

	my $entry = $self->entry(id => $args{'id'});
	my %answers = map {
		my $group = $_->{'group'} if scalar @{$_->{'group'}} > 1;
		$_->{'id'} => {
			'solution' => $_->{'solution'},
			'group' => $group,
		};
	} @$entry;

	if (!defined $args{'join'}) {
		return \%answers;
	}

	# Fall-through to scanning the answerset, joining any words which
	# cross boundaries.
	foreach my $id (keys %answers) {
		if (defined $answers{$id}->{'group'} and
		    @{$answers{$id}->{'group'}} > 1) {
			# Check the first item in the group.  This is an
			# ordered list of how the clues fit together for the
			# constituent wordplay.
			#
			# If the first item is the same as the solution we're
			# currently looking at, then we need to assume this is
			# the starting word to append from -- taking into
			# account all other clues.  As those clues are looked
			# up, the answers are marked as joined.
			my $group = $answers{$id}->{'group'};
			my $start_clue = $group->[0];
			my $s = 0;
			if ($start_clue ne $id) {
				$s = 1;
			}

			my $e = scalar @$group - 1;

			foreach my $group_clue_id (($s..$e)) {
				my $group_clue = $group->[$group_clue_id];
				if ($group_clue eq $start_clue) {
					next;
				}

				if (exists $answers{$group_clue}->{'joined'} and
				    $start_clue ne $group_clue) {
					next;
				}

				$answers{$id}->{'joined_solution'} =
					$answers{$start_clue}->{'solution'};

				if (defined
				    $answers{$group_clue}->{'solution'})
				{
					$answers{$id}->{'joined_solution'} .=
					$answers{$group_clue}->{'solution'};

					$answers{$group_clue}->{'joined'} = 1;
				}
			}
		}

	}

	return \%answers;
}

=head1 _get_setter_data

A private function to return setter information from JSON files.

=cut

sub _get_setter_data
{
	my ($self) = @_;

	$self->{'_files'} = [
		bsd_glob(CROSSWORD_PATH . $self->name() . "/*.JSON")
	];
}

1;
