import java.sql.Timestamp

case class ApiResponse(
  error: Option[String] = None,
  token: Option[String] = None,
  rooms: Option[Seq[Room]] = None,
  room: Option[Room] = None,
  stroke: Option[Stroke] = None
)

final case class Token(
  id: Long,
  csrf_token: String,
  created_at: Timestamp
)

final case class Point(
  id: Long,
  stroke_id: Long,
  x: Double,
  y: Double
)

final case class PostedStrokePoint(
  x: Double,
  y: Double
)

final case class Stroke(
  id: Long,
  room_id: Long,
  width: Int,
  red: Int,
  green: Int,
  blue: Int,
  alpha: Double,
  created_at: Timestamp,
  points: Seq[Point] = Seq()
)

final case class PostedStroke(
  width: Int,
  red: Int,
  green: Int,
  blue: Int,
  alpha: Double,
  points: Seq[PostedStrokePoint] = Seq()
)

final case class Room(
  id: Long,
  name: String,
  canvas_width: Int,
  canvas_height: Int,
  created_at: Timestamp,
  strokes: Seq[Stroke] = Seq(),
  stroke_count: Int = 0,
  watcher_count: Int = 0
)

final case class PostedRoom(
  name: String,
  canvas_width: Int,
  canvas_height: Int
)
