import React from 'react';
import { Link } from 'react-router';
import fetchJson from '../util/fetch-json';
import NotificationSystem from 'react-notification-system';
import { GridList, GridTile } from 'material-ui/GridList';
import IconButton from 'material-ui/IconButton';
import ModeEdit from 'material-ui/svg-icons/editor/mode-edit';
// import StarBorder from 'material-ui/svg-icons/toggle/star-border';

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
        <GridList
          cellHeight={222}
          cols={4}
        >
          {this.props.rooms.map((room) => (
            <GridTile
              key={room.id}
              title={room.name}
              subtitle={`${room.watcherCount}人が参加`}
              actionIcon={
                  <Link to={`/rooms/${room.id}`}>
                    <IconButton>
                      <ModeEdit color="white" />
                    </IconButton>
                  </Link>
                }
            >
              <img
                style={{ maxWidth: '100%' }}
                src={`/img/${room.id}`}
                alt={room.name}
              />
            </GridTile>
          ))
          }
        </GridList>

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

