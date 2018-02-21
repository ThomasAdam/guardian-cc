package Guardian::Cryptic::ChartRenderer;

use Module::Pluggable 
    search_path => ["Guardian::Cryptic::Plugins"],
    instantiate => 'new';

use strict;
use warnings;

sub new
{
	my ($class) = @_;

	return bless {}, $class;
}

1;
