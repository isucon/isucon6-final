import React from 'react';
import { Link } from 'react-router';
import fetchJson from '../util/fetch-json';
import NotificationSystem from 'react-notification-system';

class Index extends React.Component {
  static loadProps({ params, loadContext }, cb) {
    const apiBaseUrl = (loadContext || window).apiBaseUrl;

    fetchJson(`${apiBaseUrl}/api/rooms`)
      .then((res) => {
        cb(null, {
          rooms: res.rooms,
          csrfToken: (loadContext || window).csrfToken,
        });
      })
      .catch((err) => {
        cb(err); // TODO
      });
  }

  handleCreateNewRoom(ev) {
    ev.preventDefault();

    const room = {
      name: this.refs.newRoomName.value,
      canvas_width: 1028,
      canvas_height: 768,
    };

    if (room.name === '') {
      return; // TODO: エラーメッセージ
    }

    fetchJson('/api/rooms', {
      method: 'POST',
      body: JSON.stringify(room),
      headers: { 'x-csrf-token': this.props.csrfToken, 'content-type': 'application/json' },
    })
      .then((res) => {
        this.context.router.push({ pathname: `/rooms/${res.room.id}`, query: '', state: '' });
      })
      .catch((err) => {
        this.refs.notificationSystem.addNotification({
          title: 'エラーが発生しました',
          message: err.message,
          level: 'error',
          position: 'bc',
        });
      });
  }

  render() {
    return (
      <div className="index">
        <NotificationSystem ref="notificationSystem" />
        <div>
          <form onSubmit={(ev) => this.handleCreateNewRoom(ev)}>
            <label>
              新規部屋名:
              <input type="text" placeholder="例: ひたすら椅子を描く部屋" ref="newRoomName" />
            </label>
            <input type="hidden" name="token" value="" />
            <button type="submit">作成する</button>
          </form>
        </div>
        <div className="mdl-grid">
          {this.props.rooms.map((room) => (
            <div className="mdl-cell mdl-cell--3-col mdl-card mdl-shadow--2dp" key={room.id}>
              <div className="mdl-card__media">
                <img
                  style={{ maxWidth: '100%' }}
                  className="thumbnail"
                  src={`/img/${room.id}`}
                  alt={room.name}
                />
              </div>
              <div className="mdl-card__supporting-text">
                <h2 className="mdl-card__title-text">{room.name}</h2>
                <p>{room.watcherCount}人が参加</p>
              </div>
              <div className="mdl-card__actions mdl-card--border">
                <Link
                  to={`/rooms/${room.id}`}
                  className="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect"
                >
                  入室
                </Link>
              </div>
            </div>
          ))
          }
        </div>

      </div>
    );
  }
}

Index.propTypes = {
  rooms: React.PropTypes.array,
  csrfToken: React.PropTypes.string,
};

Index.contextTypes = {
  router: React.PropTypes.object.isRequired,
};

export default Index;
