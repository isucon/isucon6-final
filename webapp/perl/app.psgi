#!/usr/bin/env plackup
use 5.014;
use warnings;

use FindBin;
use lib "$FindBin::Bin/lib";
use File::Spec;
use Plack::Builder;

use Isuketch::Web;

my $root_dir = $FindBin::Bin;
my $app = Isuketch::Web->new(root_dir => $root_dir);
my $psgi_app = $app->build_app;

sub {
    my ($env) = @_;
    if ($env->{PATH_INFO} =~ m<\A/api/stream/rooms/([^/]+)\z>) {
        return $app->get_api_stream_room($env, $1);
    } else {
        return $psgi_app->($env);
    }
};
