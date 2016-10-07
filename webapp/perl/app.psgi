#!/usr/bin/env plackup
use 5.014;
use warnings;

use FindBin;
use lib "$FindBin::Bin/lib";
use File::Spec;
use Plack::Builder;

use Isuketch::Web;

my $root_dir = $FindBin::Bin;
my $app = Isuketch::Web->psgi($root_dir);
