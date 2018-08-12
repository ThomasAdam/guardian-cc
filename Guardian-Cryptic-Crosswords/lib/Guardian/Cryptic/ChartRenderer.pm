package Guardian::Cryptic::ChartRenderer;

use lib "$ENV{'HOME'}/guardian-cc-import/Guardian-Cryptic-Crosswords/lib";
use Module::Pluggable
	search_path => ["Guardian::Cryptic::Plugins"],
    instantiate => 'new';

use Template;
use Carp qw/confess/;
use MongoDB;

use strict;
use warnings;

sub new
{
	my ($class) = @_;

	my $tmpl = Template->new(
		INCLUDE_PATH => 'ui/chart_defs/',
		EVAL_PERL => 1,
		INTERPOLATE => 0,
	);

	my $mc = MongoDB->connect();
	my $db = $mc->get_database('guardian');
	my $col = $db->get_collection('cryptic');

	my $out_fh;

	return bless {
		'mongo' => {
			'client' => $mc,
			'db' => $db,
			'col' => $col,
		},
		'_tt' => $tmpl,
		'_output' => \$out_fh,
	}, $class;
}

sub tt
{
	my ($self) = @_;

	return $self->{'_tt'};
}

sub save
{
	my ($self, %args) = @_;
	my $file = $args{'file'};
	my $content = { content => $args{'content'} };


	$self->tt()->process($file, $content, $self->{'_output'}) ||
	   die $self->tt()->error(), "\n";
}

sub tmpl_data
{
	my ($self) = @_;

	return $self->{'_output'};
}

1;
