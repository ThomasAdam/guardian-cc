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
					'cnumbers' => {
						'$push' => '$number'
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

	my $interim_js = [];

	use Data::Dumper;
	foreach my $j (@$interdata) {
		my @new_clues = grep { !/^See\s*\d+|^\s*\(\d+\)/i } @{$j->{'clues'}};
		next unless @new_clues;
		next if $new_clues[0] eq '';
		my $clues = join '<br />', @new_clues;
		my $nums;
		foreach my $u (@{$j->{'cnumbers'}}) {
			$nums .= "<a href='https://theguardian.com/crosswords/cryptic/$u'" . ">$u</a><br />";
		}
		push @{$interim_js},
			[
				$j->{'id'}->{'name'},
				$clues,
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
				{ 'title' => 'Crossword' },
			],
		}
	};

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
