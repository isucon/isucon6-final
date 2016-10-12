'use strict';
import Koa from 'koa';
import Router from 'koa-router';

import convert from 'koa-convert';
import logger from 'koa-logger';
import bodyparser from 'koa-bodyparser';
import json from 'koa-json';

import mysql from 'promise-mysql';

const app = new Koa();

const getDBH = () => {
  const host = process.env.MYSQL_HOST || 'localhost';
  const port = process.env.MYSQL_PORT || '3306';
  const user = process.env.MYSQL_USER || 'root';
  const pass = process.env.MYSQL_PASS || '';
  const dbname = 'isuketch';
  return mysql.createPool({
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

const getStrokePoints = async (dbh) => {
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

const getWatcherCount = async () => {
};

const updateRoomWatcher = async () => {
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

app.on('error', (err, ctx) => {
  console.log(err)
  logger.error('server error', err, ctx);
});

const router = new Router();
router.post('/api/csrf_token', async (ctx, next) => {
  const dbh = getDBH();

  let sql = 'INSERT INTO `tokens` (`csrf_token`) VALUES (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))';
  await dbh.query(sql);
  const tokens = await dbh.query('SELECT LAST_INSERT_ID() AS lastInsertId');
  const id = tokens[0].lastInsertId;

  sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?';
  const token = await selectOne(dbh, sql, [id]);
  ctx.body = {
    token: token['csrf_token'],
  };
});

router.get('/api/rooms', async (ctx, next) => {
  const dbh = getDBH();
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
    rooms: rooms
  };
});

router.post('/api/rooms', async () => {
});

router.get('/api/rooms/:id', async () => {
});

router.post('/api/strokes/rooms/:id', async () => {
});

router.get('/api/initialize', async (ctx, next) => {
  const dbh = getDBH();
  const sqls = [
    'DELETE FROM `points` WHERE `id` > 1443000',
    'DELETE FROM `strokes` WHERE `id` > 41000',
    'DELETE FROM `rooms` WHERE `id` > 1000',
    'DELETE FROM `tokens` WHERE `id` > 0',
  ];

  for (const sql of sqls) {
    await dbh.query(sql);
  }
  ctx.body = 'ok';
});

app.use(router.routes());
app.use(router.allowedMethods());

module.exports = app;
