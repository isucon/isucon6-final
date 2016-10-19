'use strict';
import Koa from 'koa';
import Router from 'koa-router';

import convert from 'koa-convert';
import logger from 'koa-logger';
import bodyparser from 'koa-bodyparser';
import json from 'koa-json';

import mysql from 'promise-mysql';

import sse from './sse';

const app = new Koa();

const getDBH = (ctx) => {
  const host = process.env.MYSQL_HOST || 'localhost';
  const port = process.env.MYSQL_PORT || '3306';
  const user = process.env.MYSQL_USER || 'root';
  const pass = process.env.MYSQL_PASS || '';
  const dbname = 'isuketch';
  return ctx.dbh = mysql.createPool({
    host: host,
    port: port,
    user: user,
    password: pass,
    database: dbname,
    connectionLimit: 1,
    charset: 'utf8mb4',
  });
};

const selectOne = async (dbh, sql, params = []) => {
  const result = await dbh.query(sql, params);
  return result[0];
};

const selectAll = async (dbh, sql, params = []) => {
  return await dbh.query(sql, params);
};

const typeCastPointData = (data) => {
  return {
    id: data.id,
    stroke_id: data.stroke_id,
    x: data.x,
    y: data.y,
  };
};

const toRFC3339Micro = (date) => {
  return date.toISOString();
};

const typeCastStrokeData = (data) => {
  return {
    id: data.id,
    room_id: data.room_id,
    width: data.width,
    red: data.red,
    green: data.green,
    blue: data.blue,
    alpha: data.alpha,
    points: typeof(data.points) !== 'undefined' ? data.points.map(typeCastPointData) : [],
    created_at: typeof(data.created_at) !== 'undefined' ? toRFC3339Micro(data.created_at) : '',
  };
};

const typeCastRoomData = (data) => {
  return {
    id: data.id,
    name: data.name,
    canvas_width: data.canvas_width,
    canvas_height: data.canvas_height,
    created_at: typeof(data.created_at) !== 'undefined' ? toRFC3339Micro(data.created_at) : '',
    strokes: typeof(data.strokes) !== 'undefined' ? data.strokes.map(typeCastStrokeData) : [],
    stroke_count: data.stroke_count,
    watcher_count: data.watcher_count,
  };
};

class TokenException {};

const checkToken = async (dbh, csrfToken) => {
  let sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens`';
  sql    += ' WHERE `csrf_token` = ? AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY';
  const token = await selectOne(dbh, sql, [csrfToken]);
  if (typeof token === 'undefined') {
    throw new TokenException();
  }
  return token;
};

const getStrokePoints = async (dbh, strokeId) => {
  const sql = 'SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = ? ORDER BY `id` ASC';
  return await selectAll(dbh, sql, [strokeId]);
};

const getStrokes = async (dbh, roomId, greaterThanId) => {
  let sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`';
  sql +=      ' WHERE `room_id` = ? AND `id` > ? ORDER BY `id` ASC';
  return await selectAll(dbh, sql, [roomId, greaterThanId]);
};

const getRoom = async (dbh, roomId) => {
  const sql = 'SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = ?';
  return await selectOne(dbh, sql, [roomId]);
};

const getWatcherCount = async (dbh, roomId) => {
  let sql = 'SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`';
  sql +=    ' WHERE `room_id` = ? AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND';
  const result = await selectOne(dbh, sql, [roomId]);
  return result.watcher_count;
};

const updateRoomWatcher = async (dbh, roomId, tokenId) => {
  let sql = 'INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (?, ?)';
  sql +=    'ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)';
  await dbh.query(sql, [roomId, tokenId]);
};

app.use(convert(bodyparser()));
app.use(convert(json()));
app.use(convert(logger()));

// logger
app.use(async (ctx, next) => {
  const start = new Date();
  await next();
  const ms = new Date() - start;
  console.log(`[app] ${ctx.method} ${ctx.url} - ${ms}ms`);
});

app.use(async (ctx, next) => {
  await next();
  if (typeof ctx.dbh !== 'undefined') {
    await ctx.dbh.end();
    ctx.dbh = null;
  }
});

app.on('error', (err, ctx) => {
  console.log(err)
  logger.error('server error', err, ctx);
});

const router = new Router();
router.post('/api/csrf_token', async (ctx, next) => {
  const dbh = getDBH(ctx);

  let sql = 'INSERT INTO `tokens` (`csrf_token`) VALUES (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))';
  const result = await dbh.query(sql);
  const id = result.insertId;

  sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?';
  const token = await selectOne(dbh, sql, [id]);
  ctx.body = {
    token: token['csrf_token'],
  };
});

router.get('/api/rooms', async (ctx, next) => {
  const dbh = getDBH(ctx);
  const sql = 'SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes` GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100';
  const results = await selectAll(dbh, sql);
  const rooms = [];
  for (const result of results) {
    const room = await getRoom(dbh, result['room_id']);
    const strokes = await getStrokes(dbh, room['id'], 0);
    room['stroke_count'] = strokes.length;
    rooms.push(room);
  }
  ctx.body = {
    rooms: rooms.map(typeCastRoomData)
  };
});

router.post('/api/rooms', async (ctx, next) => {
  const dbh = getDBH(ctx);

  let token = null;
  try {
    token = await checkToken(dbh, ctx.headers['x-csrf-token']);
  } catch (e) {
    if (e instanceof TokenException) {
      console.error(e);
      ctx.status = 400;
      ctx.body = {
        error: 'トークンエラー。ページを再読み込みしてください。'
      };
      return;
    } else {
      throw e;
    }
  }

  if (!ctx.request.body.name || !ctx.request.body.canvas_width || !ctx.request.body.canvas_height) {
    ctx.status = 400;
    ctx.body = {
      error: 'リクエストが正しくありません。'
    };
    return;
  }

  let roomId;
  await dbh.query('BEGIN');
  try {
    let sql = 'INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)';
    sql +=    ' VALUES (?, ?, ?)';
    const result = await dbh.query(sql, [ctx.request.body.name, ctx.request.body.canvas_width, ctx.request.body.canvas_height]);
    roomId = result.insertId;

    sql = 'INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (?, ?)';
    await dbh.query(sql, [roomId, token.id]);
    await dbh.query('COMMIT');
  } catch (e) {
    await dbh.query('ROLLBACK');
    console.error(e);
    ctx.status = 500;
    ctx.body = {
      error: 'エラーが発生しました。'
    };
    return;
  }

  const room = await getRoom(dbh, roomId);
  ctx.body = {
    room: typeCastRoomData(room),
  };
});

router.get('/api/rooms/:id', async (ctx, next) => {
  const dbh = getDBH(ctx);

  const room = await getRoom(dbh, ctx.params.id);
  if (typeof room === 'undefined') {
    ctx.status = 404;
    ctx.body = {
      error: 'この部屋は存在しません。',
    };
    return;
  }

  const strokes = await getStrokes(dbh, room.id, 0);
  let i = 0;
  for ( const stroke of strokes ) {
    strokes[i].points = await getStrokePoints(dbh, stroke.id);
    i++;
  }

  room.strokes = strokes;
  room.watcher_count = await getWatcherCount(dbh, room.id);

  ctx.body = {
    room: typeCastRoomData(room),
  };
});

router.get('/api/stream/rooms/:id', async (ctx, next) => {
  ctx.type = 'text/event-stream';
  ctx.req.setTimeout(Number.MAX_VALUE);
  ctx.body = new sse();

  const dbh = await getDBH(ctx);
  let token;
  try {
    token = await checkToken(dbh, ctx.query.csrf_token);
  } catch (e) {
    if (e instanceof TokenException) {
      ctx.body.write(
        "event:bad_request\n" +
        "data:トークンエラー。ページを再読みこみしてください。\n\n");
      ctx.body.end();
      return;
    } else {
      throw e;
    }
  }

  const room = await getRoom(dbh, ctx.params.id);
  if ( typeof room === 'undefined' ) {
    ctx.body.write(
      "event:bad_request\n" +
      'data:この部屋は存在しません\n\n');
    ctx.body.end();
    return;
  }

  await updateRoomWatcher(dbh, room.id, token.id);
  let watcherCount = await getWatcherCount(dbh, room.id);

  ctx.body.write(
    "retry:500\n\n" +
    "event:watcher_count\n" +
    `data:${watcherCount}\n\n`
  );

  let lastStrokeId = 0;
  if (ctx.headers['last-event-id']) {
    lastStrokeId = parseInt(ctx.headers['last-event-id']);
  }

  await new Promise((resolve, reject) => {
    let loop = 6;
    const interval = async () => {
      try {
        loop--;
        const strokes = await getStrokes(dbh, room.id, lastStrokeId);
        for (const stroke of strokes) {
          stroke.points = await getStrokePoints(dbh, stroke.id);
          ctx.body.write(
            `id:${stroke.id}\n\n` +
            "event:stroke\n" +
            `data:${JSON.stringify(typeCastStrokeData(stroke))}\n\n`
          );
          lastStrokeId = stroke.id;
        }

        await updateRoomWatcher(dbh, room.id, token.id);
        const newWatcherCount = await getWatcherCount(dbh, room.id);
        if (newWatcherCount !== watcherCount) {
          watcherCount = newWatcherCount;
          ctx.body.write(
            "event:watcher_count\n" +
            `data:${watcherCount}\n\n`
          );
        }

        if ( loop === 0 ) {
          resolve();
        } else {
          intervalId = setTimeout(interval, 500);
        }
      } catch(e) {
        console.error(e);
        reject(e);
      }
    };
    let intervalId = setTimeout(interval, 500);
  });
  ctx.body.end();
});

router.post('/api/strokes/rooms/:id', async (ctx, next) => {
  const dbh = await getDBH(ctx);

  let token;
  try {
    token = await checkToken(dbh, ctx.headers['x-csrf-token']);
  } catch (e) {
    if (e instanceof TokenException) {
      ctx.status = 400;
      ctx.body = {
        error: 'トークンエラー。ページを再読み込みしてください。',
      };
    } else {
      throw e;
    }
  }

  const room = await getRoom(dbh, ctx.params.id);
  if (typeof room === 'undefined') {
    ctx.status = 400;
    ctx.body = {
      error: 'この部屋は存在しません。'
    };
    return;
  }

  if (!ctx.request.body.width || !ctx.request.body.points) {
    ctx.status = 400;
    ctx.body = {
      error: 'リクエストが正しくありません。'
    };
    return;
  }

  const strokes = await getStrokes(dbh, room.id, 0);
  const strokeCount = strokes.length;
  // TODO:
  if (strokeCount === 0) {
    const sql = 'SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = ? AND `token_id` = ?';
    const result = await selectOne(dbh, sql, [room.id, token.id]);
    if (result.cnt === 0) {
      ctx.status = 400;
      ctx.body = {
        error: '他人の作成した部屋に1画目を描くことはできません'
      };
      return;
    }
  }

  await dbh.query('BEGIN');
  let strokeId;
  try {
    let sql = 'INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)';
    sql +=    'VALUES(?, ?, ?, ?, ?, ?)';
    const result = await dbh.query(sql, [
      room.id,
      ctx.request.body.width,
      ctx.request.body.red,
      ctx.request.body.green,
      ctx.request.body.blue,
      ctx.request.body.alpha
    ]);
    strokeId = result.insertId;

    sql = 'INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (?, ?, ?)';
    for (let point of ctx.request.body.points) {
      await dbh.query(sql, [strokeId, point.x, point.y]);
    }
    await dbh.query('COMMIT');
  } catch (e) {
    await dbh.query('ROLLBACK');
    console.error(e);
    ctx.status = 500;
    ctx.body = {
      error: 'エラーが発生しました。'
    };
  }

  let sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`';
  sql +=    ' WHERE `id` = ?';
  const stroke = await selectOne(dbh, sql, [strokeId]);
  stroke.points = await getStrokePoints(dbh, strokeId);
  ctx.body = {
    stroke: typeCastStrokeData(stroke)
  };

});

app.use(router.routes());
app.use(router.allowedMethods());

module.exports = app;
