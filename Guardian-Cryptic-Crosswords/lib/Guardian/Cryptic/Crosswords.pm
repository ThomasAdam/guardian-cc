package Guardian::Cryptic::Crosswords;

use strict;
use warnings;

use Carp;
use File::Basename;
use File::Glob ':glob';

use parent 'Guardian::Cryptic::Setter';

use Data::Dumper;

our %setters;
use constant CROSSWORD_PATH => "./crosswords/cryptic/setter/";

BEGIN {
	%setters = map {
		basename($_) => 1
	} grep { -d } bsd_glob(CROSSWORD_PATH . "*");
};

=head1 NAME

Guardian::Cryptic::Crosswords - An API for querying information about cryptic
crosswords.

=head1 VERSION

Version 0.01

=cut

our $VERSION = '0.01';

=head1 SYNOPSIS

Quick summary of what the module does.

Perhaps a little code snippet.

    use Guardian::Cryptic::Crosswords;

    my $cc = Guardian::Cryptic::Crosswords->new(setter => 'Vlad');
    my $total = $cc->total();

=head1 SUBROUTINES/METHODS

=head2 new

Instantiates a new crypitc crossword object.  The name of the setter must be
supplied.

=cut

sub new
{
	my ($class, %args) = @_;

	my $self = $class->SUPER::new(%args);

	bless $self, $class;

	return $self;
}

=head2 setters

Creates a Setter object for all setters.

Returns an array ref.

=cut

sub setters
{
	my ($self) = @_;

	my @setters = map {
		Guardian::Cryptic::Setter->new(setter => $_);
	} keys %setters;

	return \@setters;
}

=head1 AUTHOR

Thomas Adam, C<< <thomas at xteddy.org> >>

=head1 BUGS

Please report any bugs or feature requests to C<bug-guardian-cryptic-crosswords at rt.cpan.org>, or through
the web interface at L<http://rt.cpan.org/NoAuth/ReportBug.html?Queue=Guardian-Cryptic-Crosswords>.  I will be notified, and then you'll
automatically be notified of progress on your bug as I make changes.

=head1 SUPPORT

You can find documentation for this module with the perldoc command.

    perldoc Guardian::Cryptic::Crosswords

=head1 ACKNOWLEDGEMENTS

=head1 LICENSE AND COPYRIGHT

Copyright 2018 Thomas Adam.

=cut

1; # End of Guardian::Cryptic::Crosswords
