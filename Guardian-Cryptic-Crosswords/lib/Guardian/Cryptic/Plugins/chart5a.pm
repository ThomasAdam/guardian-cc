package Guardian::Cryptic::Plugins::chart5a;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";

use parent 'Guardian::Cryptic::ChartRenderer';
use JSON;

my $tmpl_file = "chart5a.tmpl";

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
						'clue' => '$entries.clue',
					},
					'count' => {
						'$sum' => 1
					},
					'clues' => {
						'$push' => '$entries.clue'
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
			}
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
		my @new_clues = grep { !/
			^\s+\(\d+\)$|
			^See\s+(?:clues|special)\s+(?:page|instructions)\s+\(\d+\)$|
			^Follow\s+the\s+link\s+below\s+to\s+see\s+today\'s\s+clues.*$|
			^See\s+(?:\d+)??\s+\(\d+\)$|
			^See\s+\(\d+\)\s*$|
			^See\s+\d+\s*(?:across|down)??$
			/x } @{$j->{'clues'}};
		next unless @new_clues;
		next if $new_clues[0] eq '';

=head
		my $pos = -1;
		foreach my $type (@{$j->{'type'}}) {
			$pos++;
			my $class = "span-$type";
			my $c = $new_clues[$pos];
			my $u = $j->{'urls'}->[$pos];

			$new_clues[$pos] = "<span class='$class'>$c</span>";
			$j->{'urls'}->[$pos] = "<span class='$class'>$u</span>";
			$j->{'type'}->[$pos] = "<span class='$class'>$type</span>";
		}
=cut

		my $clues = join '<br />', @new_clues;
		my $types = join '<br />', @{$j->{'type'}};
		my $nums  = join '<br />', @{$j->{'urls'}};
		push @{$interim_js},
			[
				$j->{'id'}->{'name'},
				$clues,
				$types,
				$nums,
			]
	};

	my $ajax_data = {
		data => $interim_js,
	};

	open (my $fh, ">", "./ds_ajax5a.txt") or die;
	print {$fh} to_json($ajax_data, {utf8 => 1, pretty => 1});
	close ($fh);

	my $data = {
		'title' => "Number of duplicate clues per setter",
		'preamble' => "This table shows the number of times a given " .
		              "clue has been used per setter.",
		'order' => '5a',
		'js' => {
			columns => [
				{ 'title' => 'Setter' },
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
