# -*- coding: utf-8 -*-
import json
import os
from datetime import datetime, timezone
import decimal
import time

import MySQLdb.cursors

from flask import Flask, jsonify, request, Response
from flask.json import JSONEncoder


def get_db():
    host = os.environ.get('MYSQL_HOST', 'localhost')
    port = int(os.environ.get('MYSQL_PORT', 3306))
    user = os.environ.get('MYSQL_USER', 'root')
    passwd = os.environ.get('MYSQL_PASS', '')
    dbname = 'isuketch'
    charset = 'utf8mb4'
    cursorclass = MySQLdb.cursors.DictCursor
    autocommit = True
    return MySQLdb.connect(host=host, port=port, user=user, passwd=passwd, db=dbname, cursorclass=cursorclass, charset=charset, autocommit=autocommit)


def execute(db, sql, params={}):
    cursor = db.cursor()
    cursor.execute(sql, params)
    return cursor.lastrowid


def selectOne(db, sql, params={}):
    cursor = db.cursor()
    cursor.execute(sql, params)
    return cursor.fetchone()


def selectAll(db, sql, params={}):
    cursor = db.cursor()
    cursor.execute(sql, params)
    return cursor.fetchall()


def printAndFlush(content):
    return content

def toRFC3339Micro(date):
    # RFC3339では+00:00のときはZにするという仕様だが、pythonは準拠していないため
    return date.replace(tzinfo=timezone.utc).isoformat().replace('+00:00', 'Z')


class CustomJSONEncoder(JSONEncoder):
    def default(self, obj):
        try:
            if isinstance(obj, datetime):
                return toRFC3339Micro(obj)
            if isinstance(obj, decimal.Decimal):
                return float(obj)

            iterable = iter(obj)
        except TypeError:
            pass
        else:
            return list(iterable)
        return JSONEncoder.default(self, obj)


class TokenException(Exception):
    pass


def checkToken(db, csrf_token):
    sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens`'
    sql += ' WHERE `csrf_token` = %(csrf_token)s AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY'
    token = selectOne(db, sql, {'csrf_token': csrf_token})
    if not token:
        raise TokenException()
    return token


def getStrokePoints(db, stroke_id):
    sql = 'SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = %(stroke_id)s ORDER BY `id` ASC'
    return selectAll(db, sql, {'stroke_id': stroke_id})


def getStrokes(db, room_id, greater_than_id):
    sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`'
    sql += ' WHERE `room_id` = %(room_id)s AND `id` > %(greater_than_id)s ORDER BY `id` ASC'
    return selectAll(db, sql, {'room_id': room_id, 'greater_than_id': greater_than_id})


def getRoom(db, room_id):
    sql = 'SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = %(room_id)s'
    return selectOne(db, sql, {'room_id': room_id})


def getWatcherCount(db, room_id):
    sql = 'SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`'
    sql += ' WHERE `room_id` = %(room_id)s AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND'
    result = selectOne(db, sql, {'room_id': room_id})
    return result['watcher_count']


def updateRoomWatcher(db, room_id, token_id):
    sql = 'INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (%(room_id)s, %(token_id)s)'
    sql += ' ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)'
    return execute(db, sql, {'room_id': room_id, 'token_id': token_id})


app = Flask(__name__)
app.json_encoder = CustomJSONEncoder
app.config['JSONIFY_PRETTYPRINT_REGULAR'] = False

# Routes


@app.route('/api/csrf_token', methods=['POST'])
def csrf_token():
    db = get_db()

    sql = 'INSERT INTO `tokens` (`csrf_token`) VALUES'
    sql += ' (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))'

    id = execute(db, sql)

    sql = 'SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = %(id)s'
    token = selectOne(db, sql, {'id': id})

    return jsonify({'token': token['csrf_token']})


@app.route('/api/rooms', methods=['GET'])
def get_rooms():

    db = get_db()

    sql = 'SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`'
    sql += ' GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100'
    results = selectAll(db, sql)

    rooms = []
    for result in results:
        room = getRoom(db, result['room_id'])
        strokes = getStrokes(db, room['id'], 0)
        room['stroke_count'] = len(strokes)
        rooms.append(room)

    return jsonify({'rooms': rooms})


@app.route('/api/rooms', methods=['POST'])
def post_rooms():
    db = get_db()
    try:
        token = checkToken(db, request.headers.get('x-csrf-token'))
    except TokenException:
        res = jsonify({'error': 'トークンエラー。ページを再読み込みしてください。'})
        res.status_code = 400
        return res

    posted_room = request.form

    if 'name' not in posted_room or 'canvas_width' not in posted_room or 'canvas_height' not in posted_room:
        res = jsonify({'error': 'リクエストが正しくありません。'})
        res.status_code = 400
        return res

    cursor = db.cursor()
    try:
        sql = 'INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)'
        sql += ' VALUES (%(name)s, %(canvas_width)s, %(canvas_height)s)'
        cursor.execute(sql, {
            'name': posted_room.get('name'),
            'canvas_width': posted_room.get('canvas_width'),
            'canvas_height': posted_room.get('canvas_height'),
        })
        room_id = cursor.lastrowid

        sql = 'INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (%(room_id)s, %(token_id)s)'
        cursor.execute(sql, {
            'room_id': room_id,
            'token_id': token['id'],
        })
    except Exception as e:
        db.rollback()
        app.logger.error(e)
        res = jsonify({'error': 'エラーが発生しました。'})
        res.status_code = 500
        return res

    room = getRoom(db, room_id)
    return jsonify({'room': room})


@app.route('/api/rooms/<id>')
def rooms(id):
    db = get_db()
    room = getRoom(db, id)

    if room is None:
        res = jsonify({'error': 'この部屋は存在しません。'})
        res.status__code = 500
        return res

    strokes = getStrokes(db, room['id'], 0)

    for i, stroke in enumerate(strokes):
        strokes[i]['points'] = getStrokePoints(db, stroke['id'])

    room['strokes'] = strokes
    room['watcher_count'] = getWatcherCount(db, room['id'])

    return jsonify({'room': room})


@app.route('/api/stream/rooms/<id>')
def stream_rooms(id):
    db = get_db()

    try:
        token = checkToken(db, request.args.get('csrf_token'))
    except TokenException:
        return printAndFlush(
            'event:bad_request\n' +
            'data:トークンエラー。ページを再読み込みしてください。\n\n'
        ), 200, {'Content-Type': 'text/event-stream'}

    room = getRoom(db, id)

    if room is None:
        return printAndFlush(
            'event:bad_request\n' +
            'data:この部屋は存在しません\n\n'
        ), 200, {'Content-Type': 'text/event-stream'}

    last_stroke_id = 0
    if 'Last-Event-ID' in request.headers:
        last_stroke_id = request.headers.get('Last-Event-ID')

    def gen(db, room, token, last_stroke_id):

        updateRoomWatcher(db, room['id'], token['id'])
        watcher_count = getWatcherCount(db, room['id'])

        yield printAndFlush(
            'retry:500\n\n' +
            'event:watcher_count\n' +
            'data:%d\n\n' % (watcher_count)
        )

        for loop in range(6):
            time.sleep(0.5)  # 500ms

            strokes = getStrokes(db, room['id'], last_stroke_id)
            # app.logger.info(strokes)

            for stroke in strokes:
                stroke['points'] = getStrokePoints(db, stroke['id'])
                yield printAndFlush(
                    'id:' + str(stroke['id']) + '\n\n' +
                    'event:stroke\n' +
                    'data:' + json.dumps(stroke, cls=CustomJSONEncoder) + '\n\n'
                )
                last_stroke_id = stroke['id']

            updateRoomWatcher(db, room['id'], token['id'])
            new_watcher_count = getWatcherCount(db, room['id'])
            if new_watcher_count != watcher_count:
                yield printAndFlush(
                    'event:watcher_count\n' +
                    'data:%d\n\n' % (watcher_count)
                )
                watcher_count = new_watcher_count

    return Response(gen(db, room, token, last_stroke_id), mimetype='text/event-stream')


@app.route('/api/strokes/rooms/<id>', methods=['POST'])
def post_strokes_rooms(id):
    db = get_db()

    try:
        token = checkToken(db, request.headers.get('x-csrf-token'))
    except TokenException:
        res = jsonify({'error': 'トークンエラー。ページを再読み込みしてください。'})
        res.status_code = 400
        return res

    room = getRoom(db, id)

    if room is None:
        res = jsonify({'error': 'この部屋は存在しません。'})
        res.status_code = 404
        return res

    postedStroke = request.get_json()
    if 'width' not in postedStroke or 'points' not in postedStroke:
        res = jsonify({'error': 'リクエストが正しくありません。'})
        res.status_code = 400
        return res

    stroke_count = len(getStrokes(db, room['id'], 0))
    if stroke_count > 1000:
        res = jsonify({'error': '1000画を超えました。これ以上描くことはできません。'})
        res.status_code = 400
        return res
    if stroke_count == 0:
        sql = 'SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = %(room_id)s AND `token_id` = %(token_id)s'
        result = selectOne(db, sql, {'room_id': room['id'], 'token_id': token['id']})
        if result['cnt'] == 0:
            res = jsonify({'error': '他人の作成した部屋に1画目を描くことはできません'})
            res.status_code = 400
            return res

    cursor = db.cursor()
    try:
        sql = 'INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)'
        sql += ' VALUES(%(room_id)s, %(width)s, %(red)s, %(green)s, %(blue)s, %(alpha)s)'
        cursor.execute(sql, {
            'room_id': room['id'],
            'width': postedStroke.get('width'),
            'red': postedStroke.get('red'),
            'green': postedStroke.get('green'),
            'blue': postedStroke.get('blue'),
            'alpha': postedStroke.get('alpha'),
        })
        stroke_id = cursor.lastrowid

        sql = 'INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (%(stroke_id)s, %(x)s, %(y)s)'
        for point in postedStroke.get('points'):
            cursor.execute(sql, {
                'stroke_id': stroke_id,
                'x': point['x'],
                'y': point['y']
            })
    except Exception as e:
        db.rollback()
        app.logger.error(e)
        res = jsonify({'error': 'エラーが発生しました。'})
        res.status_code = 500
        return res

    sql = 'SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`'
    sql += ' WHERE `id` = %(stroke_id)s'
    stroke = selectOne(db, sql, {'stroke_id': stroke_id})

    stroke['points'] = getStrokePoints(db, stroke_id)

    return jsonify({'stroke': stroke})


@app.route('/api/initialize')
def initialize():
    db = get_db()

    sqls = [
        'DELETE FROM `points` WHERE `id` > 1443000',
        'DELETE FROM `strokes` WHERE `id` > 41000',
        'DELETE FROM `rooms` WHERE `id` > 1000',
        'DELETE FROM `tokens` WHERE `id` > 0',
    ]

    for sql in sqls:
        execute(db, sql)

    return 'ok'


if __name__ == '__main__':
    debug = os.environ.get('ISUCON_ENV') != 'production'
    app.run(host='', port=80, debug=debug)
