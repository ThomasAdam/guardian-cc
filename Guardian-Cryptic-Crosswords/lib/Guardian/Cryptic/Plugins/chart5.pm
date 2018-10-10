package Guardian::Cryptic::Plugins::chart5;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";

use parent 'Guardian::Cryptic::ChartRenderer';

my $tmpl_file = "chart5.tmpl";

sub new
{
	my ($class) = @_;

	my $self = $class->SUPER::new();
	bless $self, $class;
}

sub interpolate
{
	my ($self) = @_;

	my $agg = $self->{'mongo'}->{'col'}->aggregate([
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
					},
					'clues' => {
						'$push' => '$entries.clue'
					},
					'cnumbers' => {
						'$push' => {
							'$concat' => ["https://www.theguardian.com/", '$id' ]
						}
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
				'$sort' => {
					'_id.name' => 1
				}
			}
	]);

	my @res = map { $_->{id} = delete $_->{_id}; $_ } $agg->all;

	return \@res;
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
		'order' => 5,
		'table' => $interdata,
	};

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
