import React from 'react';
import Canvas from './Canvas';

class Room extends React.Component {
  static loadProps({ params, loadContext }, cb) {
    const apiBaseUrl = loadContext ? loadContext.apiBaseUrl : window.apiBaseUrl;
    const csrfToken = loadContext ? loadContext.csrfToken : window.csrfToken;
    fetch(`${apiBaseUrl}/api/rooms/${params.id}`, {
      headers: { 'x-csrf-token': csrfToken },
    })
    .then((result) => result.json())
    .then((res) => {
      cb(null, { id: res.room.id, name: res.room.name, strokes: res.room.strokes });
    });
  }

  render() {
    return (
      <div className="room">
        <h2>{this.props.name}</h2>
        <Canvas strokes={this.props.strokes} followUpdates />
      </div>
    );
  }
}

Room.propTypes = {
  id: React.PropTypes.number,
  name: React.PropTypes.string,
  strokes: React.PropTypes.array,
};

export default Room;
