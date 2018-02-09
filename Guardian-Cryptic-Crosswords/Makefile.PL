use 5.006;
use strict;
use warnings;
use ExtUtils::MakeMaker;

WriteMakefile(
    NAME             => 'Guardian::Cryptic::Crosswords',
    AUTHOR           => q{Thomas Adam <thomas@xteddy.org>},
    VERSION_FROM     => 'lib/Guardian/Cryptic/Crosswords.pm',
    ABSTRACT_FROM    => 'lib/Guardian/Cryptic/Crosswords.pm',
    LICENSE          => 'artistic_2',
    PL_FILES         => {},
    MIN_PERL_VERSION => 5.006,
    CONFIGURE_REQUIRES => {
        'ExtUtils::MakeMaker' => 0,
    },
    BUILD_REQUIRES => {
        'Test::More' => 0,
    },
    PREREQ_PM => {
        #'ABC'              => 1.6,
        #'Foo::Bar::Module' => 5.0401,
    },
    dist  => { COMPRESS => 'gzip -9f', SUFFIX => 'gz', },
    clean => { FILES => 'Guardian-Cryptic-Crosswords-*' },
);
