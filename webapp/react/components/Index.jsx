import React from 'react';
import { Link } from 'react-router';
import fetch from 'isomorphic-fetch';

class Index extends React.Component {
  static loadProps(params, cb) {
    const apiBaseUrl = params.loadContext ? params.loadContext.apiBaseUrl : window.apiBaseUrl;
    const csrfToken = params.loadContext ? params.loadContext.csrfToken : window.csrfToken;
    fetch(`${apiBaseUrl}/api/rooms`, {
      headers: { 'x-csrf-token': csrfToken },
    })
      .then((result) => result.json())
      .then((res) => {
        cb(null, { rooms: res.rooms });
      });
  }

  render() {
    return (
      <div className="index">
        <div>
          <form method="POST" action="/rooms" className="new-room">
            <h3>新規部屋作成</h3>
            <label>
              部屋名:
              <input type="text" placeholder="例: ひたすら椅子を描く部屋" required name="name" />
            </label>
            <input type="hidden" name="token" value="" />
            <button className="create">作成</button>
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
};

export default Index;
