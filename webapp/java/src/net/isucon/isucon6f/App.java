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
                    room.setStrokeCount(getStrokes(conn, rs.getInt("id"), 0));
                    rooms.add(room);
                }

                Map<String, Room[]> map = new HashMap<String, Room[]>();
                map.put("rooms", rooms.toArray(new Room[0]));
                return map;
            }, gson::toJson);

            // $app->post('/api/rooms'
            post("/api/rooms", (request, response) -> {
                // TODO: Implements
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
            
            get("/api/stream/rooms/[{id}]", (request, response) -> {
            	response.raw().setContentType("text/event-stream");
            	
            	Token token;
            	try {
            		token = checkToken(conn, request.params("csrf_token"));
            	} catch (TokenException e) {
            		try (OutputStream os = response.raw().getOutputStream()) {
            			os.write("event:bad_request\ndata:トークンエラー。ページを再読み込みしてください。\n\n".getBytes());
            			os.flush();
            		}
            		return "";
            	}
            	
            	response.raw();
            	return "";
            });
        }
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

    private static int getStrokes(Connection conn, int int1, int i) {
        // TODO Auto-generated method stub
        return 0;
    }

    private static Room getRoom(Connection conn, int room_id) {
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
