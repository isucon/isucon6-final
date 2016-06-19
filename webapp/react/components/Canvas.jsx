import React from 'react';

class Canvas extends React.Component {
  // static loadProps(params, cb) {
    // const apiBaseUrl = params.loadContext ? params.loadContext.apiBaseUrl : window.apiBaseUrl;
    // const csrfToken = params.loadContext ? params.loadContext.csrfToken : window.csrfToken;
    // fetch(`${apiBaseUrl}/api/rooms`, {
    //   headers: { 'x-csrf-token': csrfToken },
    // })
    //   .then((result) => result.json())
    //   .then((res) => {
    //     cb(null, { rooms: res.rooms });
    //   });
  // }

  constructor(props) {
    super(props);
    // console.log(props.followUpdates);
    this.state = { strokes: props.strokes, tmpStroke: null };
  }

  getCoordinates(ev) {
    const rect = this.refs.svgElement.getBoundingClientRect();
    return {
      x: ev.clientX - rect.left,
      y: ev.clientY - rect.top,
    };
  }

  addPointToStroke(orig, point) {
    return {
      id: orig.id,
      red: orig.red,
      blue: orig.blue,
      green: orig.green,
      alpha: orig.alpha,
      width: orig.width,
      points: orig.points.concat([point]),
    };
  }

  handleMouseDown(ev) {
    this.mouseDown = true;
    this.setState({
      tmpStroke: this.addPointToStroke({
        id: Date.now(), // TODO:
        red: 128,
        blue: 128,
        green: 128,
        alpha: 0.5,
        width: 5,
        points: [],
      }, this.getCoordinates(ev)),
    });
  }

  handleMouseUp(ev) {
    if (this.mouseDown && this.state.tmpStroke) {
      const tmpStroke = this.addPointToStroke(this.state.tmpStroke, this.getCoordinates(ev));
      this.setState({
        strokes: this.state.strokes.concat(tmpStroke),
        tmpStroke: null,
      });
    }
    this.mouseDown = false;
  }

  handleMouseMove(ev) {
    if (this.mouseDown && this.state.tmpStroke) {
      this.setState({
        tmpStroke: this.addPointToStroke(this.state.tmpStroke, this.getCoordinates(ev)),
      });
    }
  }

  render() {
    return (
      <div className="canvas">
        <svg
          ref="svgElement"
          width={this.props.width}
          height={this.props.height}
          style={{
            width: `${this.props.width}px`,
            height: `${this.props.height}px`,
            border: 'solid black 1px',
          }}
          viewBox={`0 0 ${this.props.width} ${this.props.height}`}
          onMouseDown={(ev) => this.handleMouseDown(ev)}
          onMouseUp={(ev) => this.handleMouseUp(ev)}
          onMouseMove={(ev) => this.handleMouseMove(ev)}
        >
          {this.state.strokes
            .concat([this.state.tmpStroke].filter((s) => s !== null))
            .map((stroke) => (
              <polyline
                key={stroke.id}
                stroke={`rgba(${stroke.red},${stroke.green},${stroke.blue},${stroke.alpha})`}
                strokeWidth={stroke.width}
                strokeLinecap="round"
                fill="none"
                points={stroke.points.map((point) => `${point.x},${point.y}`).join(' ')}
              />
          ))}
        </svg>
      </div>
    );
  }
}

Canvas.propTypes = {
  width: React.PropTypes.number,
  height: React.PropTypes.number,
  strokes: React.PropTypes.array,
  followUpdates: React.PropTypes.bool,
};

Canvas.defaultProps = {
  width: 1028,
  height: 768,
};

export default Canvas;
