import React from 'react';

class Room extends React.Component {
  static loadProps({ params, loadContext }, cb) {
    const apiBaseUrl = loadContext ? loadContext.apiBaseUrl : window.apiBaseUrl;
    const csrfToken = loadContext ? loadContext.csrfToken : window.csrfToken;
    fetch(`${apiBaseUrl}/api/rooms/${params.id}`, {
      headers: { 'x-csrf-token': csrfToken },
    })
    .then((result) => result.json())
    .then((res) => {
      cb(null, { id: res.room.id, name: res.room.name });
    });
  }

  render() {
    return (
      <div className="room">
        <p>{this.props.name}</p>
        <div className="canvas-column">
          <div id="canvas"></div>
        </div>
      </div>
    );
  }
}

Room.propTypes = {
  id: React.PropTypes.number,
  name: React.PropTypes.string,
};

export default Room;
