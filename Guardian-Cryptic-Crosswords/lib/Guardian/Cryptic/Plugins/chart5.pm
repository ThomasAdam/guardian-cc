package Guardian::Cryptic::Plugins::chart5;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";

use parent 'Guardian::Cryptic::ChartRenderer';
use JSON;

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
						'solution' => '$entries.solution',
					},
					'count' => {
						'$sum' => 1
					},
					'clues' => {
						'$push' => '$entries.clue'
					},
					'cnumbers' => {
						'$push' => '$number'
					},
					'type' => {
						'$push' => '$crosswordType'
					},
					'urls' => {
						'$push' => {
							'$concat' => [
								'<a href="https://www.theguardian.com/',
								'$id',
								'">',
								'$number',
								'</a>'
							],
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
			},
	], {'allowDiskUse' => 1});

	my @res = map { $_->{id} = delete $_->{_id}; $_ } $agg->all;

	return \@res;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();

	my $interim_js = [];

	foreach my $j (@$interdata) {
		my $clues = join '<br />', @{$j->{'clues'}};
		my $types = join '<br />', @{$j->{'type'}};
		my $nums  = join '<br />', @{$j->{'urls'}};

		push @{$interim_js},
			[
				$j->{'id'}->{'name'},
				$j->{'id'}->{'solution'},
				$clues,
				$types,
				$nums,
			]
	};

	my $ajax_data = {
		data => $interim_js,
	};

	open (my $fh, ">", "./ds_ajax.txt") or die;
	print {$fh} to_json($ajax_data, {utf8 => 1, pretty => 1});
	close ($fh);

	my $data = {
		'title' => "Number of duplicate answers and their questions",
		'preamble' => "This table shows the number of times a given " .
		              "clue has been used and the different questions " .
						"which have been used to make up that clue.",
		'order' => 5,
		'js' => {
			columns => [
				{ 'title' => 'Setter' },
				{ 'title' => 'Answer' },
				{ 'title' => 'Clues'  },
				{ 'title' => 'Type'   },
				{ 'title' => 'Crossword' },
			],
		}
	};

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
