import React from 'react';
import { Link } from 'react-router';
import fetchJson from '../util/fetch-json';
import { GridList, GridTile } from 'material-ui/GridList';
import IconButton from 'material-ui/IconButton';
import ModeEdit from 'material-ui/svg-icons/editor/mode-edit';
import RaisedButton from 'material-ui/RaisedButton';
import TextField from 'material-ui/TextField';
import Subheader from 'material-ui/Subheader';
import Snackbar from 'material-ui/Snackbar';

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

  constructor(props) {
    super(props);
    this.state = {
      showError: false,
      errorMessage: '',
    };
  }

  handleCreateNewRoom() {
    const room = {
      name: this.refs.newRoomName.input.value,
      canvas_width: 1028,
      canvas_height: 768,
    };

    if (!room.name) {
      this.setState({
        showError: true,
        errorMessage: '空の部屋名で作成することはできません',
      });
      return;
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
        this.setState({
          showError: true,
          errorMessage: err.message,
        });
      });
  }

  render() {
    return (
      <div className="index">
        <Snackbar
          open={this.state.showError}
          message={this.state.errorMessage}
        />

        <div>
          <TextField
            hintText="例: ひたすら椅子を描く部屋"
            floatingLabelText="新規部屋作成"
            ref="newRoomName"
            id="newRoomName"
          />
          <RaisedButton
            label="作成する"
            primary
            style={{ margin: 12 }}
            onTouchTap={(ev) => this.handleCreateNewRoom(ev)}
          />
        </div>
        <GridList
          cellHeight={222}
          cols={4}
        >
          <Subheader>新着描き込み</Subheader>
          {this.props.rooms.map((room) => (
            <GridTile
              className="room"
              id={room.id}
              key={room.id}
              title={room.name}
              subtitle={
                <span>
                  <span className="stroke_count">{room.stroke_count}</span>画
                </span>
              }
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

