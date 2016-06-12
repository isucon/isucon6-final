import React from 'react';
import { Link } from 'react-router';
import fetch from 'isomorphic-fetch';

class Index extends React.Component {
  static loadProps(params, cb) {
    const apiEndpoint = params.loadContext ? params.loadContext.apiEndpoint : window.apiEndpoint;
    const csrfToken = params.loadContext ? params.loadContext.csrfToken : window.csrfToken;
    fetch(`${apiEndpoint}/api/rooms`, {
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
          <p>描ける巨大匿名掲示板サイト！</p>
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
        <ul className="room-list">
          {this.props.rooms.map((room) => (
            <li className="room-info" key={room.id}>
              <Link to={`/rooms/${room.id}`}>
                <img className="thumbnail" src={`/img/${room.id}`} alt={room.name} />
                <p className="name">{room.name}</p>
                <p className="member-count">{room.watcherCount}人が参加</p>
              </Link>
            </li>
          ))}
        </ul>

      </div>
    );
  }
}

Index.propTypes = {
  rooms: React.PropTypes.array,
};

export default Index;
