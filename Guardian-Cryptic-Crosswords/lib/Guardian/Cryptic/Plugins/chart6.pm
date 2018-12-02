package Guardian::Cryptic::Plugins::chart6;

use lib "$ENV{'HOME'}/guardian-cc/Guardian-Cryptic-Crosswords/lib";

use parent 'Guardian::Cryptic::ChartRenderer';
use JSON;

my $tmpl_file = "chart6.tmpl";

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
				'$match' => {
					'pdf' => {
						'$exists' => 1,
						'$ne' =>  undef
					}
				}
			},
			{
				'$project' => {
					'_id' => {
						'name' => '$creator.name',
						'number' => '$number',
						'pdf' => '$pdf',
						'date' => {
							'$dateFromString' => {
								'dateString' => '$date'
							}
						}
					}
				}
			},
			{
				'$sort' => {
					'_id.name' => 1,
					'_id.number' => 1
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

	foreach my $j (@$interdata) {
		my $dt = $j->{'id'}->{'date'};
		push @{$interim_js},
			[
				$j->{'id'}->{'name'},
				'<a href="' . $j->{'id'}->{'pdf'} . '">' .  $j->{'id'}->{'number'} . '</a>',
				$dt->ymd,
			]
	};

	my $ajax_data = {
		data => $interim_js,
	};

	open (my $fh, ">", "./ds_ajax2.txt") or die;
	print {$fh} to_json($ajax_data, {utf8 => 1, pretty => 1});
	close ($fh);

	my $data = {
		'title' => "List of all crosswords by setter, which has a PDF version",
		'preamble' => "This table shows the crossword number and a link to the " .
		              "PDF crossword, if available.",
		'order' => 6,
		'js' => {
			columns => [
				{ 'title' => 'Setter' },
				{ 'title' => 'Crossword (PDF)' },
				{ 'title' => 'Date published' },
			],
		}
	};

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
