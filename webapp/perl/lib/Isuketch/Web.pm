package Isuketch::Web;
use 5.014;
use strict;
use warnings;
use utf8;
use Kossy;
use DBIx::Sunny;
use Time::HiRes qw(usleep);
use JSON qw(encode_json decode_json);
use DateTime::Format::MySQL;
use DateTime::Format::RFC3339;

sub config {
    state $conf = {
        db_host => $ENV{MYSQL_HOST} // 'localhost',
        db_port => $ENV{MYSQL_PORT} // 3306,
        db_user => $ENV{MYSQL_USER} // 'root',
        db_pass => $ENV{MYSQL_PASS} // '',
        db_name => 'isuketch',
    };
    my $key = shift;
    my $v = $conf->{$key};
    unless (defined $v) {
        die "config value of $key undefined";
    }
    return $v;
}

sub dbh {
    my ($self) = @_;
    return $self->{dbh} //= DBIx::Sunny->connect(
        'dbi:mysql:host=' . config('db_host') . ';port=' . config('db_port') . ';dbname=' . config('db_name'),
        config('db_user'),
        config('db_pass'),
        {
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    $dbh->do('SET NAMES utf8mb4');
                    $dbh->do('SET TIME_ZONE = "UTC"');
                    return;
                },
            },
        }
    );
}

sub check_token {
    my ($dbh, $csrf_token) = @_;
    my $token = $dbh->select_row(q[
        SELECT `id`, `csrf_token`, `created_at` FROM `tokens`
        WHERE `csrf_token` = ?
          AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY
    ], $csrf_token);
    die 'token not found' unless $token;
    return $token;
}

sub to_point_json {
    my ($data) = @_;
    return {
        id        => int $data->{id},
        stroke_id => int $data->{stroke_id},
        x         => 0+$data->{x},
        y         => 0+$data->{y},
    };
}

sub to_rfc_3339 {
    my ($string) = @_;
    my $dt = DateTime::Format::MySQL->parse_datetime($string);
    $dt->set_time_zone('UTC');
    return DateTime::Format::RFC3339->format_datetime($dt);
}

sub to_stroke_json {
    my ($data) = @_;
    return {
        id         => int $data->{id},
        room_id    => int $data->{room_id},
        width      => int $data->{width},
        red        => int $data->{red},
        green      => int $data->{green},
        blue       => int $data->{blue},
        alpha      => 0+$data->{alpha},
        points     => $data->{points} ? [ map { to_point_json($_) } @{ $data->{points} } ] : [],
        created_at => $data->{created_at} ? to_rfc_3339($data->{created_at}) : '',
    };
}

sub to_room_json {
    my ($data) = @_;
    return {
        id            => int $data->{id},
        name          => $data->{name},
        canvas_width  => int $data->{canvas_width},
        canvas_height => int $data->{canvas_height},
        created_at    => $data->{created_at} ? to_rfc_3339($data->{created_at}) : '',
        strokes       => $data->{strokes} ? [ map { to_stroke_json($_) } @{ $data->{strokes} } ] : [],
        stroke_count  => int($data->{stroke_count} // 0),
        watcher_count => int($data->{watcher_count} // 0),
    };
}

sub get_stroke_points {
    my ($dbh, $stroke_id) = @_;
    return $dbh->select_all(q[
        SELECT `id`, `stroke_id`, `x`, `y`
        FROM `points`
        WHERE `stroke_id` = ?
        ORDER BY `id` ASC
    ], $stroke_id);
}

sub get_strokes {
    my ($dbh, $room_id, $greater_than_id) = @_;
    return $dbh->select_all(q[
        SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at`
        FROM `strokes`
        WHERE `room_id` = ?
          AND `id` > ?
        ORDER BY `id` ASC;
    ], $room_id, $greater_than_id);
}

sub get_room {
    my ($dbh, $room_id) = @_;
    return $dbh->select_row(q[
        SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at`
        FROM `rooms`
        WHERE `id` = ?
    ], $room_id);
}

sub get_watcher_count {
    my ($dbh, $room_id) = @_;
    return $dbh->select_one(q[
        SELECT COUNT(*) AS `watcher_count`
        FROM `room_watchers`
        WHERE `room_id` = ?
          AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND
    ], $room_id) // 0;
}

sub update_room_watcher {
    my ($dbh, $room_id, $token_id) = @_;
    $dbh->query(q[
        INSERT INTO `room_watchers` (`room_id`, `token_id`)
        VALUES (?, ?)
        ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)
    ], $room_id, $token_id);
}

post '/api/csrf_token' => sub {
    my ($self, $c) = @_;
    $self->dbh->query(q[
        INSERT INTO `tokens` (`csrf_token`)
        VALUES
        (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))
    ]);

    my $id = $self->dbh->last_insert_id;
    my $token = $self->dbh->select_row(q[
        SELECT `id`, `csrf_token`, `created_at`
        FROM `tokens`
        WHERE `id` = ?
    ], $id);

    return $c->render_json({
        token => $token->{csrf_token}
    });
};

get '/api/rooms' => sub {
    my ($self, $c) = @_;
    my $results = $self->dbh->select_all(q[
        SELECT `room_id`, MAX(`id`) AS `max_id`
        FROM `strokes`
        GROUP BY `room_id`
        ORDER BY `max_id` DESC
        LIMIT 100
    ]);

    my @rooms;
    foreach my $result (@$results) {
        my $room = get_room($self->dbh, $result->{room_id});
        $room->{stroke_count} = scalar @{ get_strokes($self->dbh, $room->{id}, 0) };
        push @rooms, $room;
    }

    return $c->render_json({
        rooms => [
            map { to_room_json($_) } @rooms
        ]
    });
};

post '/api/rooms' => sub {
    my ($self, $c) = @_;
    my $token = eval {
        check_token($self->dbh, $c->req->header('X-CSRF-Token'));
    };
    if ($@) {
        $c->render_json({
            error => 'トークンエラー。ページを再読み込みしてください。'
        });
        $c->res->code(400);
        return $c->res;
    }

    my $posted_room = decode_json $c->req->content;
    if (!length($posted_room->{name}) || !length($posted_room->{canvas_width}) || !length($posted_room->{canvas_height})) {
        $c->render_json({
            error => 'リクエストが正しくありません。'
        });
        $c->res->code(400);
        return $c->res;
    }

    my $txn = $self->dbh->txn_scope;
    my $room_id;
    eval {
        $self->dbh->query(q[
            INSERT INTO `rooms`
            (`name`, `canvas_width`, `canvas_height`)
            VALUES
            (?, ?, ?)
        ], $posted_room->{name}, $posted_room->{canvas_width}, $posted_room->{canvas_height});

        $room_id = $self->dbh->last_insert_id;

        $self->dbh->query(q[
            INSERT INTO `room_owners`
            (`room_id`, `token_id`)
            VALUES
            (?, ?)
        ], $room_id, $token->{id});
        $txn->commit;
    };
    if (my $e = $@) {
        $txn->rollback;
        warn $e;
        $c->render_json({
            error => 'エラーが発生しました。'
        });
        $c->res->code(500);
        return $c->res;
    }

    my $room = get_room($self->dbh, $room_id);
    return $c->render_json({
        room => to_room_json($room),
    });
};

get '/api/rooms/:id' => sub {
    my ($self, $c) = @_;
    my $room = get_room($self->dbh, $c->args->{id});
    unless ($room) {
        $c->render_json({
            error => 'この部屋は存在しません。'
        });
        $c->res->code(404);
        return $c->res;
    }

    my $strokes = get_strokes($self->dbh, $room->{id}, 0);
    foreach my $stroke (@$strokes) {
        $stroke->{points} = get_stroke_points($self->dbh, $stroke->{id});
    }
    $room->{strokes} = $strokes;
    $room->{watcher_count} = get_watcher_count($self->dbh, $room->{id});
    return $c->render_json({
        room => to_room_json($room),
    });
};

# get /api/stream/rooms/:id
sub get_api_stream_room {
  my ($self, $env, $room_id) = @_;
  my $req = Plack::Request->new($env);

  return sub {
      my ($respond) = @_;
      my $writer = $respond->([ 200, [ 'Content-Type' => 'text/event-stream' ] ]);

      my $token = eval {
          check_token($self->dbh, $req->parameters->{csrf_token});
      };
      if ($@) {
          $writer->write(
            "event:bad_request\n" .
            "data:トークンエラー。ページを再読み込みしてください。\n\n"
          );
          $writer->close();
          return;
      }

      my $room = get_room($self->dbh, $room_id);
      unless ($room) {
          $writer->write(
            "event:bad_request\n" .
            "data:この部屋は存在しません\n\n"
          );
          $writer->close();
          return;
      }

      update_room_watcher($self->dbh, $room->{id}, $token->{id});
      my $watcher_count = get_watcher_count($self->dbh, $room->{id});

      $writer->write(
        "retry:500\n\n" .
        "event:watcher_count\n" .
        "data:$watcher_count\n\n"
      );

      my $last_stroke_id = 0;
      if ($req->header('Last-Event-ID')) {
          $last_stroke_id = int $req->header('Last-Event-ID');
      }

      my $loop = 6;
      while ($loop > 0) {
          $loop--;
          usleep(500 * 1000); # 500ms

          my $strokes = get_strokes($self->dbh, $room->{id}, $last_stroke_id);
          foreach my $stroke (@$strokes) {
              $stroke->{points} = get_stroke_points($self->dbh, $stroke->{id});
              $writer->write(
                  "id:$stroke->{id}\n\n" .
                  "event:stroke\n" .
                  'data:' . encode_json(to_stroke_json($stroke)) . "\n\n"
              );
              $last_stroke_id = $stroke->{id};
          }

          update_room_watcher($self->dbh, $room->{id}, $token->{id});
          my $new_watcher_count = get_watcher_count($self->dbh, $room->{id});
          if ($new_watcher_count != $watcher_count) {
              $watcher_count = $new_watcher_count;
              $writer->write(
                  "event:watcher_count\n" .
                  "data:$watcher_count\n\n"
              );
          }
      }
      $writer->close;
  };
}

post '/api/strokes/rooms/:id' => sub {
    my ($self, $c) = @_;
    my $token = eval {
        check_token($self->dbh, $c->req->header('X-CSRF-Token'));
    };
    if ($@) {
        $c->render_json({
            error => 'トークンエラー。ページを再読み込みしてください。'
        });
        $c->res->code(400);
        return $c->res;
    }

    my $room = get_room($self->dbh, $c->args->{id});
    unless ($room) {
        $c->render_json({
            error => 'この部屋は存在しません。'
        });
        $c->res->code(404);
        return $c->res;
    }

    my $posted_stroke = decode_json $c->req->content;
    if (!length($posted_stroke->{width}) || !length($posted_stroke->{points})) {
        $c->render_json({
            error => 'リクエストが正しくありません。'
        });
        $c->res->code(400);
        return $c->res;
    }

    my $stroke_count = scalar @{ get_strokes($self->dbh, $room->{id}, 0) };
    if ($stroke_count == 0) {
        my $count = $self->dbh->select_one(q[
            SELECT COUNT(*) as cnt FROM `room_owners`
            WHERE `room_id` = ?
              AND `token_id` = ?
        ], $room->{id}, $token->{id});
        if ($count == 0) {
            $c->render_json({
                error => '他人の作成した部屋に1画目を描くことはできません'
            });
            $c->res->code(400);
            return $c->res;
        }
    }

    my $txn = $self->dbh->txn_scope;
    my $stroke_id;
    eval {
        $self->dbh->query(q[
            INSERT INTO `strokes`
            (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)
            VALUES
            (?, ?, ?, ?, ?, ?)
        ], $room->{id}, $posted_stroke->{width}, $posted_stroke->{red}, $posted_stroke->{green}, $posted_stroke->{blue}, $posted_stroke->{alpha});
        $stroke_id = $self->dbh->last_insert_id;
        foreach my $point (@{ $posted_stroke->{points} }) {
            $self->dbh->query(q[
                INSERT INTO `points`
                (`stroke_id`, `x`, `y`)
                VALUES
                (?, ?, ?)
            ], $stroke_id, $point->{x}, $point->{y});
        }
        $txn->commit;
    };
    if (my $e = $@) {
        $txn->rollback;
        warn $e;
        $c->render_json({
            error => 'エラーが発生しました。'
        });
        $c->res->code(500);
        return $c->res;
    }

    my $stroke = $self->dbh->select_row(q[
        SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at`
        FROM `strokes`
        WHERE `id`= ?
    ], $stroke_id);
    $stroke->{points} = get_stroke_points($self->dbh, $stroke_id);

    return $c->render_json({
        stroke => to_stroke_json($stroke),
    });
};

1;
