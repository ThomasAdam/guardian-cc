package Guardian::Cryptic::Plugins::chart4;

use lib "$ENV{'HOME'}/projects/guardian-cc/Guardian-Cryptic-Crosswords/lib";
use List::Util qw/max/;
use DateTime;
use DateTime::Format::Duration;
use Sort::Key::DateTime qw/ dtkeysort dtsort /;

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
	my %res;

	my $names_agg = $self->{'mongo'}->{'col'}->aggregate([
			{
				'$group' => {
					'_id' => '$creator.name'
				}
			},
			{
				'$sort' => {
					'_id' => 1
				}
			}
	]);

	my @names = map { $_->{'_id'} } $names_agg->all;

	foreach my $name (@names) {
		my $res = $self->{'mongo'}->{'col'}->aggregate([
				{
					'$match' => {
						'creator.name' => "$name"
					}
				},
				{
					'$project' => {
						'creator.name' => '$creator.name',
						'date' => {
							'$dateFromString' => {
								'dateString' => '$date'
							}
						}
					}
				},
				{
					'$sort' => {
						"date" => 1,
					}
				},
				{
					'$group' => {
						"_id" => {
							'name' => '$creator.name',
							'date' => '$date',
						},
					}
				}

		]);

		my @dt_objs = map {
			$_->{'_id'}->{'date'}
		} $res->all;

		my @dt_sorted = dtkeysort { $_ } @dt_objs;

		$res{$name}->{'range'} = {
			'first' => $dt_sorted[0]->strftime("%Y-%m-%d"),
			'last'  => $dt_sorted[-1]->strftime("%Y-%m-%d"),
		};

		my $duration = $dt_sorted[-1] - $dt_sorted[0];
		my $fmt_duration = DateTime::Format::Duration->new(
			pattern => '%Y years, %m months, %e days',
			normalize => 1,
		);

		$res{$name}->{'range'}->{'duration'} =
			$fmt_duration->format_duration($duration);
	}

	my $self_agg = $self->{'mongo'}->{'col'}->aggregate([
			{
				'$unwind' => '$entries'
			},
			{
				'$project' => {
					'name' => '$creator.name',
					'clues' => '$entries.clue'
				}
			},
			{
				'$match' => {
					'$expr' => {
						'$ne' => [
							{
								'$indexOfCP' => ['$clues', '$name']
							},
							-1
						]
					}
				}
			},
			{
				'$group' => {
					'_id' => '$name',
					'count' => {
						'$sum' => 1
					}
				}
			},
			{
				'$sort' => {
					'_id' => 1
				}
			}
	]);

	my @sagg = $self_agg->all;

	foreach (@sagg) {
		$res{$_->{'_id'}}->{'self_word_count'} = $_->{'count'};
	}

	my $self_agg_graph = $self->{'mongo'}->{'col'}->aggregate([
			{
				'$project' => {
					'_id' => '$creator.name',
					'ndate' => {
						'$year' => {
							'$dateFromString' => {
								'dateString' => '$date'
							}
						}
					},
					'type' => '$crosswordType',
				}
			},
			{
				'$group' => {
					'_id' => {
						'name' => '$_id',
						'year' => '$ndate',
						'type' => '$type',
					},
					'count' => {
						'$sum' => 1
					}
				}
			},
			{
				'$project' => {
					'_id' => '$_id',
					'count' => '$count',
					'type'  => '$_id.type',
					'pmonth' => {
						'$ceil' => {
							'$divide' => ['$count', 12]
						}
					}
				}
			},
			{
				'$sort' => {
					'_id' => 1
				}
			}
	]);

	my @date_range = $self_agg_graph->all;

	foreach (@date_range) {
		push @{ $res{$_->{'_id'}->{'name'}}->{'graph'} },
		{
			'gdata' => {
				$_->{'_id'}->{'year'},
				$_->{'count'},
			},
			'avg' => $_->{'pmonth'},
			'type' => $_->{'type'},
		}
	}

	foreach my $name (keys %res) {
		my $total = 0;
		my $ptotal = 0;
		my $ctotal = 0;
		foreach my $d (@{ $res{$name}->{'graph'} }) {
			# Keep a running total of crosswords produced by this setter.
			# Note the list context for the hash values here -- this is in
			# order to get the actual count for the year.  Each hash entry
			# only ever has one key, hence why we only extract the first
			# entry.
			$total += (values %{ $d->{'gdata'} })[0];
			$ctotal +=(values %{ $d->{'gdata'} })[0] if $d->{'type'} eq 'cryptic';
			$ptotal +=(values %{ $d->{'gdata'} })[0] if $d->{'type'} eq 'prize';
		}
		$res{$name}->{'range'}->{'total'}->{'all'} = $total;
		$res{$name}->{'range'}->{'total'}->{'cryptic'} = $ctotal;
		$res{$name}->{'range'}->{'total'}->{'prize'} = $ptotal;
	}

	return \%res;
}

sub render
{
	my ($self) = @_;

	my $interdata = $self->interpolate();

	use Data::Dumper;
	warn Dumper "id", $interdata;

	my $data = {
		'title' => "Setter Biographies",
		'preamble' => "This shows information about each setter",
		'order' => 4,
		'default_chart' => 'area',
	};

	my $chart_count = -1;
	foreach my $k (sort keys %{$interdata}) {
		my (@labels, @values, @avg);
		$chart_count++;
		foreach my $h (@{ $interdata->{$k}->{'graph'} }) {
			push @labels, keys %{$h->{'gdata'}};
			push @values, values %{$h->{'gdata'}};
			push @avg, $h->{'avg'};
		}
		$interdata->{$k}->{'chart'}->{'clabels'} = [[$k, @values], ["Average per month", @avg]];
		$interdata->{$k}->{'chart'}->{'labels' } = \@labels;
		delete $interdata->{$k}->{'graph'};
		push @{ $data->{'charts'} },
			{
				info => {
					'div_id' => "mychart4$chart_count",
					'js_var' => "chart4$chart_count",
					'person' => $k,
					'range' => $interdata->{$k}->{'range'},
					'self_ref' => $interdata->{$k}->{'self_word_count'} // 0,
				},
				'chart' => {
					'bindto' => "#myChart4$chart_count",
					'size' => {
						'height' => 200,
						'width' => 600
					},
					'data' => {
						'columns' => $interdata->{$k}->{'chart'}->{'clabels'},
						'type' => 'area',
					},
					'axis' => {
						'x' => {
						'type' => 'category',
							'tick' => {
								'rotate' => '75',
								'multiline' => 0
							},
							'height' => 0,
							'categories' => $interdata->{$k}->{'chart'}->{'labels'},
						},
						'y' => {
							'label' => 'Number of crosswords',
							'tick' => {
								'steps' => 1,
							},
							'min' => 1
						}
					},
					'legend' => {
						'show' => 0,
					}
				},
		};
	}

	$self->save(file => $tmpl_file, content => $data);

	return $data->{'order'};
}

1;
