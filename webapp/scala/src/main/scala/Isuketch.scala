import java.sql.Timestamp

import akka.NotUsed
import akka.actor.{Props, Actor, ActorSystem}
import akka.http.scaladsl.Http
import akka.http.scaladsl.marshallers.sprayjson.SprayJsonSupport
import akka.http.scaladsl.model.StatusCodes
import akka.http.scaladsl.server.{Directives, StandardRoute}
import akka.stream.ActorMaterializer
import akka.stream.actor.ActorPublisher
import akka.stream.scaladsl.Source
import com.typesafe.config.ConfigFactory
import de.heikoseeberger.akkasse.EventStreamMarshalling._
import de.heikoseeberger.akkasse.ServerSentEvent
import scalikejdbc.config.DBs
import scalikejdbc.interpolation.Implicits._
import scalikejdbc.{DB, WrappedResultSet}
import spray.json._

import scala.concurrent.duration._
import scala.util.Try

trait IsuketchJsonSupport extends SprayJsonSupport with DefaultJsonProtocol {

  implicit object TimestampFormat extends JsonFormat[Timestamp] {

    def write(ts: Timestamp) = JsString(ts.toInstant.toString)

    def read(json: JsValue) = json match {
      case _ => throw new DeserializationException("Parsing not supported!!") // このアプリケーションでは不要
    }
  }

  implicit val printer = CompactPrinter

  implicit val pointFormat = jsonFormat4(Point.apply)
  implicit val strokeFormat = jsonFormat9(Stroke.apply)
  implicit val postedStrokePointFormat = jsonFormat2(PostedStrokePoint.apply)
  implicit val postedStrokeFormat = jsonFormat6(PostedStroke.apply)
  implicit val roomFormat = jsonFormat8(Room.apply)
  implicit val postedRoomFormat = jsonFormat3(PostedRoom.apply)
  implicit val apiResponseFormat = jsonFormat5(ApiResponse.apply)
}

object Main extends Directives with IsuketchJsonSupport {

  implicit val system = ActorSystem()
  implicit val materializer = ActorMaterializer()
  implicit val executionContext = system.dispatcher

  def mapToken = (rs: WrappedResultSet) => Token(rs.long(1), rs.string(2), rs.timestamp(3))

  def mapPoint = (rs: WrappedResultSet) => Point(rs.long(1), rs.long(2), rs.double(3), rs.double(4))

  def mapStroke = (rs: WrappedResultSet) => Stroke(rs.long(1), rs.long(2), rs.int(3), rs.int(4), rs.int(5), rs.int(6), rs.double(7), rs.timestamp(8))

  def mapRoom = (rs: WrappedResultSet) => Room(rs.long(1), rs.string(2), rs.int(3), rs.int(4), rs.timestamp(5))

  // Utilities

  def checkToken(optToken: Option[String]): Option[Token] = {
    optToken.flatMap {
      tok =>
        DB.readOnly { implicit s =>
          sql"""
              SELECT `id`, `csrf_token`, `created_at` FROM `tokens`
              WHERE `csrf_token` = ${tok} AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY
          """.map(mapToken).single.apply
        }
    }
  }

  def getStrokePoints(strokeId: Long): Seq[Point] = {
    DB.readOnly { implicit s =>
      sql"""
            SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = ${strokeId} ORDER BY `id` ASC
        """.map(mapPoint).list.apply
    }
  }

  def getStrokes(roomId: Long, greaterThanId: Long): Seq[Stroke] = {
    DB.readOnly { implicit s =>
      sql"""
            SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`
            WHERE `room_id` = ${roomId} AND `id` > ${greaterThanId} ORDER BY `id` ASC
        """.map(mapStroke).list.apply
    }
  }

  def getRoom(roomId: Long): Option[Room] = {
    DB.readOnly { implicit s =>
      sql"""
            SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = ${roomId}
        """.map(mapRoom).single.apply
    }
  }

  def getWatcherCount(roomId: Long): Int = {
    DB.readOnly { implicit s =>
      sql"""
            SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`
            WHERE `room_id` = ${roomId} AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND
        """.map(_.int(1)).single.apply.get
    }
  }

  def updateRoomWatcher(roomId: Long, tokenId: Long): Unit = {
    DB.autoCommit { implicit s =>
      sql"""
            INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (${roomId}, ${tokenId})
            ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)
        """.update.apply
    }
  }

  // API endpoints

  def apiPostCsrfToken(): StandardRoute = {
    val id = DB.autoCommit { implicit s =>
      sql"""
            INSERT INTO `tokens` (`csrf_token`) VALUES (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))
        """.updateAndReturnGeneratedKey(1).apply
    }
    DB.readOnly { implicit s =>
      sql"""
            SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ${id}
        """.map(mapToken).single.apply
    }.fold(throw new RuntimeException) {
      token =>
        complete(ApiResponse(token = Some(token.csrf_token)).toJson)
    }
  }

  def apiGetRooms(): StandardRoute = {
    val roomIds: Seq[Long] = DB.readOnly { implicit s =>
      sql"""
            SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`
            GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100
        """.map(_.long(1)).list.apply
    }

    val rooms = roomIds.map { roomId =>
      getRoom(roomId).fold(throw new RuntimeException) {
        room => room.copy(stroke_count = getStrokes(roomId, 0).length)
      }
    }

    complete(ApiResponse(rooms = Some(rooms)).toJson)
  }

  def apiPostRooms(postedRoom: PostedRoom, optToken: Option[String]): StandardRoute = {
    if (postedRoom.name == "" || postedRoom.canvas_width == 0 || postedRoom.canvas_height == 0) {
      complete(StatusCodes.BadRequest, ApiResponse(error = Some("リクエストが正しくありません。")).toJson)
    } else {
      checkToken(optToken) match {
        case None =>
          complete(StatusCodes.BadRequest, ApiResponse(error = Some("トークンエラー。ページを再読み込みしてください。")).toJson)

        case Some(token) =>
          val roomId = DB.localTx { implicit s =>
            val id =
              sql"""
                  INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)
                  VALUES (${postedRoom.name}, ${postedRoom.canvas_width}, ${postedRoom.canvas_height})
              """.updateAndReturnGeneratedKey(1).apply

            sql"""
                INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (${id}, ${token.id})
            """.update.apply

            id
          }

          getRoom(roomId).fold(throw new RuntimeException) { room =>
            complete(ApiResponse(room = Some(room)).toJson)
          }
      }
    }
  }


  def apiGetRoomsId(id: Int): StandardRoute = {
    getRoom(id) match {
      case None =>
        complete(StatusCodes.NotFound, ApiResponse(error = Some("この部屋は存在しません")).toJson)

      case Some(room) =>
        val strokes = getStrokes(id, 0).map {
          stroke =>
            stroke.copy(points = getStrokePoints(stroke.id))
        }

        complete(ApiResponse(room = Some(room.copy(strokes = strokes))).toJson)
    }
  }


  def apiStreamRoom(id: Int, optLastEventId: Option[String], optToken: Option[String]): StandardRoute = {
    checkToken(optToken) match {
      case None =>
        complete{
          Source(List(ServerSentEvent(eventType = "bad_request", data = "トークンエラー。ページを再読み込みしてください。")))
        }

      case Some(token) =>
        getRoom(id) match {
          case None =>
            complete{
              Source(List(ServerSentEvent(eventType = "bad_request", data = "この部屋は存在しません")))
            }

          case Some(room) =>

            updateRoomWatcher(room.id, token.id)
            var watcherCount = getWatcherCount(room.id)

            var firstTick = true
            var lastId: Long = optLastEventId.flatMap{ n => Try{n.toLong}.toOption }.getOrElse(0L)

            complete {

              Source
                .tick(initialDelay = 0.seconds, interval = 500.milliseconds, tick = NotUsed)
                .take(7)
                .map { _ =>

                  if (firstTick) {
                    // 初回
                    firstTick = false
                    List(new ServerSentEvent(eventType = Some("watcher_count"), data = watcherCount.toString, retry = Some(500)))

                  } else {
                    // 2回目以降
                    var ls = getStrokes(room.id, lastId).map {
                      stroke =>
                        lastId = stroke.id
                        ServerSentEvent(eventType = "stroke", data = stroke.toJson, id = lastId.toString)
                    }.toList

                    updateRoomWatcher(room.id, token.id)
                    val newWatcherCount = getWatcherCount(room.id)

                    if (watcherCount != newWatcherCount) {
                      watcherCount = newWatcherCount
                      ls ::= ServerSentEvent(eventType = "watcher_count", data = watcherCount.toString)
                    }

                    ls
                  }
                }
                .mapConcat(identity)
            }

        }
    }
  }

  def apiPostStrokeRoomsId(id: Int, postedStroke: PostedStroke, optToken: Option[String]): StandardRoute = {
    checkToken(optToken) match {
      case None =>
        complete(StatusCodes.BadRequest, ApiResponse(error = Some("トークンエラー。ページを再読み込みしてください。")).toJson)

      case Some(token) =>
        getRoom(id) match {
          case None =>
            complete(StatusCodes.NotFound, ApiResponse(error = Some("この部屋は存在しません")).toJson)

          case Some(room) =>

            if (postedStroke.width == 0 || postedStroke.points.isEmpty) {
              complete(StatusCodes.BadRequest, ApiResponse(error = Some("リクエストが正しくありません。")).toJson)
            } else {
              if (
                getStrokes(room.id, 0).isEmpty &&
                  0 == DB.readOnly { implicit s =>
                    sql"""
                      SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = ${room.id} AND `token_id` = ${token.id}
                  """.map(_.long(1)).single.apply.get
                  }
              ) {

                complete(StatusCodes.BadRequest, ApiResponse(error = Some("他人の作成した部屋に1画目を描くことはできません")).toJson)

              } else {
                val strokeId = DB.localTx { implicit s =>
                  val strokeId = sql"""
                   INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)
                   VALUES(${id}, ${postedStroke.width}, ${postedStroke.red}, ${postedStroke.green}, ${postedStroke.blue}, ${postedStroke.alpha})
                  """.updateAndReturnGeneratedKey(1).apply

                  postedStroke.points.map { point =>
                    sql"""
                        INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (${strokeId}, ${point.x}, ${point.y})
                    """.update.apply
                  }

                  strokeId
                }

                DB.readOnly { implicit s =>
                  sql"""
                      SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`
                      WHERE `id` = ${strokeId}
                  """.map(mapStroke).single.apply
                }.fold(throw new RuntimeException) { stroke =>
                  complete(ApiResponse(stroke = Some(stroke.copy(points = getStrokePoints(stroke.id)))).toJson)
                }
              }
            }
        }
    }
  }

  val routes =
    (post & path("api" / "csrf_token")) {
      apiPostCsrfToken()
    } ~
    (get & path("api" / "rooms")) {
      apiGetRooms()
    } ~
    (post & path ("api" / "rooms")) {
      entity(as[PostedRoom]) { postedRoom =>
        optionalHeaderValueByName("x-csrf-token") { optToken =>
          apiPostRooms(postedRoom, optToken)
        }
      }
    } ~
    (get & path ("api" / "rooms" / IntNumber)) { id =>
      apiGetRoomsId(id)
    } ~
    (get & path ("api" / "stream" / "rooms" / IntNumber)) { id =>
      optionalHeaderValueByName("Last-Event-ID") { optLastEventId =>
        parameters('csrf_token.?) { optToken =>
          apiStreamRoom(id, optLastEventId, optToken)
        }
      }
    } ~
    (post & path ("api" / "strokes" / "rooms" / IntNumber)) { id =>
      entity(as[PostedStroke]) { postedStroke =>
        optionalHeaderValueByName("x-csrf-token") { optToken =>
          apiPostStrokeRoomsId(id, postedStroke, optToken)
        }
      }
    }

  def main(args: Array[String]) {

    val config = ConfigFactory.load()
    //      override val logger = Logging(system, getClass)

    DBs.setupAll

    Http().bindAndHandle(routes, config.getString("http.interface"), config.getInt("http.port"))
  }
}
