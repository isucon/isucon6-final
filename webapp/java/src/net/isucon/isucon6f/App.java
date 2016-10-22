package net.isucon.isucon6f;
import static spark.Spark.*;

import java.io.OutputStream;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.sql.Connection ;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.logging.Logger;

import spark.utils.StringUtils;

import com.google.gson.Gson;

public class App {
    private static class DBConfig {
        static final String HOST = "localhost";
        static final int PORT = 3306;
        static final String USER = "root";
        static final String PASS = "";
        static final String DBNAME = "isuketch";
        public static String getDSN() {
            return String.format("jdbc:mysql://%s:%d/%s?user=%s&password=%s", HOST, PORT, DBNAME, USER, PASS);
        }
    }

    private static final Logger logger = Logger.getLogger("App");

    public static void main(String[] args) throws ClassNotFoundException, SQLException {
        Class.forName("com.mysql.jdbc.Driver");
        try (Connection conn = DriverManager.getConnection(DBConfig.getDSN())) {
            Gson gson = new Gson();

            // $app->post('/api/csrf_token', ...
            post("/api/csrf_token", (request, response) -> {
                Statement s = conn.createStatement();

                int id = s.executeUpdate("INSERT INTO `tokens` (`csrf_token`) VALUES (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))", Statement.RETURN_GENERATED_KEYS);
                PreparedStatement ps = conn.prepareStatement("SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?");
                ps.setInt(1, id);
                ResultSet rs = ps.executeQuery();
                String token = rs.getString("csrf_token");
                rs.close();
                ps.close();

                Map<String, String> map = new HashMap<String, String>();
                map.put("token", token);
                return map;
            }, gson::toJson);

            // $app->get('/api/rooms', ...
            get("/api/rooms", (request, response) -> {
                String sql = "SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes` GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100";
                PreparedStatement ps = conn.prepareStatement(sql);
                ResultSet rs = ps.executeQuery();

                List<Room> rooms = new ArrayList<>();
                while (rs.next()) {
                    Room room = getRoom(conn, rs.getInt("room_id"));
                    room.setStrokeCount(getStrokes(conn, rs.getInt("id"), 0).length);
                    rooms.add(room);
                }

                Map<String, Room[]> map = new HashMap<String, Room[]>();
                map.put("rooms", rooms.toArray(new Room[0]));
                return map;
            }, gson::toJson);

            // $app->post('/api/rooms'
            post("/api/rooms", (request, response) -> {
                Token token;
                Map<String, String> map = new HashMap<String, String>();
                try {
                    token = checkToken(conn, request.headers("x-csrf-token"));
                } catch (TokenException $e) {
                    response.status(400);
                    map.put("error", "トークンエラー。ページを再読み込みしてください。");
                    return map;
                }
                if (StringUtils.isEmpty(request.queryParams("name"))
                        || StringUtils.isEmpty(request.queryParams("canvas_width"))
                        || StringUtils.isEmpty(request.queryParams("canvas_height"))
                   ) {
                    response.status(400);
                    map.put("error", "リクエストが正しくありません。");
                    return map;
                   }
                conn.setAutoCommit(false);
                int room_id;
                try {
                    PreparedStatement ps = conn.prepareStatement("INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`) VALUES (?, ?, ?)", Statement.RETURN_GENERATED_KEYS);
                    ps.setString(1, request.queryParams("name"));
                    ps.setString(2, request.queryParams("canvas_width"));
                    ps.setString(3, request.queryParams("canvas_height"));
                    room_id = ps.executeUpdate();

                    PreparedStatement ps2 = conn.prepareStatement("INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (?, ?)");
                    ps2.setInt(1, room_id);
                    ps2.setInt(2, token.id);
                    ps2.executeUpdate();

                    conn.commit();
                } catch (SQLException e) {
                    logger.warning(e.toString());
                    response.status(500);
                    conn.rollback();
                    map.put("error", "エラーが発生しました。");
                    return map;
                } finally {
                    conn.setAutoCommit(true);
                }
                Map<String, Room> map2 = new HashMap<String, Room>();
                map2.put("room", getRoom(conn, room_id));
                return map2;
            }, gson::toJson);

            get("/api/stream/rooms/:id", (request, response) -> {
                response.raw().setContentType("text/event-stream");

                Token token;
                try {
                    token = checkToken(conn, request.params("csrf_token"));
                } catch (TokenException e) {
                    try (OutputStream os = response.raw().getOutputStream()) {
                        response.status(400);
                        os.write("event:bad_request\ndata:トークンエラー。ページを再読み込みしてください。\n\n".getBytes());
                        os.flush();
                    }
                    return "";
                }
                Room room = getRoom(conn, request.params(":id"));
                if (room == null) {
                    try (OutputStream os = response.raw().getOutputStream()) {
                        os.write("event:bad_request\ndata:この部屋は存在しません\n\n".getBytes());
                        os.flush();
                        return "";
                    }
                }

                updateRoomWatcher(conn, room.id, token.id);
                int watcher_count = getWatcherCount(conn, room.id);
                try (OutputStream os = response.raw().getOutputStream()) {
                    os.write(("retry:500\n\nevent:watcher_count\ndata:" + watcher_count + "\n\n").getBytes());
                    os.flush();
                }

                int last_stroke_id = 0;
                if (!StringUtils.isEmpty(request.headers("Last-Event-ID"))) {
                    last_stroke_id = (int) Integer.valueOf(request.headers("Last-Event-ID"));
                }

                int loop = 6;
                int new_watcher_count = 0;
                while (loop > 0) {
                    loop--;
                    Thread.sleep(500); // 500ms

                    StrokeData[] strokes = getStrokes(conn, room.id, last_stroke_id);

                    for (StrokeData stroke : strokes) {
                        stroke.setPoints(getStrokePoints(conn, stroke.id));
                        try (OutputStream os = response.raw().getOutputStream()) {
                            os.write((
                                        "id:" + stroke.id + "\n\nevent:stroke\ndata:" + gson.toJson(stroke) + "\n\n"
                                     ).getBytes());
                            os.flush();
                        }
                        last_stroke_id = stroke.id;        	        	
                    }

                    updateRoomWatcher(conn, room.id, token.id);
                    new_watcher_count = getWatcherCount(conn, room.id);
                    if (new_watcher_count != watcher_count) {
                        watcher_count = new_watcher_count;
                        try (OutputStream os = response.raw().getOutputStream()) {
                            os.write(("event:watcher_count\ndata:" + watcher_count + "\n\n").getBytes());
                            os.flush();
                        }
                    }
                }

                return "";
            });

            post("/api/stream/rooms/:id", (request, response) -> {
                Map<String, String> map = new HashMap<String, String>();
                Token token;
                try {
                    token = checkToken(conn, request.params("x-csrf-token"));
                } catch (TokenException e) {
                    response.status(400);
                    map.put("error", "トークンエラー。ページを再読み込みしてください。");
                    return map;
                }

                Room room = getRoom(conn, request.params(":id"));
                if (room == null) {
                    response.status(400);
                    map.put("error", "この部屋は存在しません。");
                    return map;
                }

                if (StringUtils.isEmpty(request.params("width")) || StringUtils.isEmpty(request.params("points"))) {
                    response.status(400);
                    map.put("error", "リクエストが正しくありません。");
                    return map;
                }

                int stroke_count = getStrokes(conn, room.id, 0).length;
                if (stroke_count == 0) {
                    String sql = "SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = ? AND `token_id` = ?";
                    try (PreparedStatement ps = conn.prepareStatement(sql)) {
                        ps.setInt(1, room.id);
                        ps.setInt(2, token.id);
                        try (ResultSet rs = ps.executeQuery()) {
                            int cnt = rs.getInt("cnt");
                            if (cnt == 0) {
                                response.status(400);
                                map.put("error", "他人の作成した部屋に1画目を描くことはできません");
                            }
                        }
                    }
                }

                // TODO:
                /*
                   conn.setAutoCommit(false);
                   try {
                   String sql1 = "INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`) VALUES(?, ?, ?, ?, ?, ?)";
                   try (PreparedStatement ps = conn.prepareStatement(sql1, Statement.RETURN_GENERATED_KEYS)) {
                   ps.setInt(1, room.id);
                   ps.setString(2, request.params("width"));
                   ps.setString(3, request.params("red"));
                   ps.setString(4, request.params("green"));
                   ps.setString(5, request.params("blue"));
                   ps.setString(6, request.params("alpha"));
                   int id = ps.executeUpdate();

                   }
                   } catch (SQLException e) {

                   }
                   String sql2 = "INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (?, ?, ?)";
                   try (PreparedStatement ps = conn.prepareStatement(sql2)) {
                   request.params("points");
                   ps.setString(stroke.id);
                   ps.setInt(point.x);
                   ps.setInt(point.y);
                   ps.executeUpdate();
                   }
                   conn.commit();
                   } catch (Exception e) {
                   conn.rollback();
                   logger.warning(e.toString());
                   response.status(500);
                   map.put("error", "エラーが発生しました。");
                   return map;
                   } finally  {
                   conn.setAutoCommit(true);
                   }
                /*
                foreach ($postedStroke['points'] as $point) {
                execute($dbh, $sql, [
                ':stroke_id' => $stroke_id,
                ':x' => $point['x'],
                ':y' => $point['y']
                ]);
                }

                $dbh->commit();
                } catch (Exception $e) {
                $dbh->rollback();
                $this->logger->error($e->getMessage());
                return $response->withStatus(500)->withJson(['error' => 'エラーが発生しました。']);
                }

                $sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`';
                $sql .= ' WHERE `id` = :stroke_id';
                $stroke = selectOne($dbh, $sql, [':stroke_id' => $stroke_id]);

                $stroke['points'] = getStrokePoints($dbh, $stroke_id);

                return $response->withJson(['stroke' => typeCastStrokeData($stroke)]);
                */                

                return "";
    });

}
}

private static Point[] getStrokePoints(Connection conn, int room_id) throws SQLException {
    String sql = "SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = ? ORDER BY `id` ASC";
    List<Point> points = new ArrayList<Point>();
    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        ps.setInt(1, room_id);
        try (ResultSet rs = ps.executeQuery()) {
            while (rs.next()) {
                points.add(new Point(rs.getInt("id"), rs.getInt("stroke_id"), rs.getInt("x"), rs.getInt("y")));
            }
        }
    }
    return null;
}

private static int getWatcherCount(Connection conn, int room_id) throws SQLException {
    String sql = "SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`"
        + " WHERE `room_id` = ? AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND";

    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        ps.setInt(1, room_id);
        try (ResultSet rs = ps.executeQuery()) {
            return rs.getInt("watcher_count");
        }
    }
}

private static void updateRoomWatcher(Connection conn, int room_id, int token_id) throws SQLException {
    String sql = "INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (?, ?)"
        + " ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)";
    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        ps.setInt(1, room_id);
        ps.setInt(2, token_id);
        try (ResultSet rs = ps.executeQuery()) {}
    }
    return;
}

private static Token checkToken(Connection conn, String csrf_token) throws TokenException {
    String sql = "SELECT `id`, `csrf_token`, `created_at` FROM `tokens`"
        +" WHERE `csrf_token` = ? AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY";
    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        ps.setString(1, csrf_token);
        try (ResultSet rs = ps.executeQuery()) {
            if (!rs.isBeforeFirst() ) {    
                throw new TokenException();
            }
            return new Token(rs.getInt("id"), rs.getString("csrf_token"), rs.getString("created_at"));
        }
    } catch (SQLException e) {
        logger.warning(e.toString());
        throw new TokenException();
    }
}

private static StrokeData[] getStrokes(Connection conn, int room_id, int greater_than_id) throws SQLException {
    String sql = "SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`"
        +" WHERE `room_id` = ? AND `id` > ? ORDER BY `id` ASC";
    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        ps.setInt(1, room_id);
        ps.setInt(2, greater_than_id);
        try (ResultSet rs = ps.executeQuery()) {
            return null;
        }
    }
}

private static Room getRoom(Connection conn, int room_id) {
    return getRoom(conn, String.valueOf(room_id));
}

private static Room getRoom(Connection conn, String room_id) {
    String sql = "SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = :room_id";
    try (PreparedStatement ps = conn.prepareStatement(sql)) {
        try (ResultSet rs = ps.executeQuery()) {
            return new Room(rs.getInt("id"), rs.getString("name"), rs.getInt("canvas_width"), rs.getInt("canvas_height"), rs.getDate("created_at").toInstant());	
        }
    } catch (SQLException e) {
        logger.warning(e.toString());
    }
    return null;
}
}
