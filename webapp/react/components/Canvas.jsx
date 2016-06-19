import React from 'react';
import { SketchPicker } from 'react-color';

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
    this.state = {
      strokes: props.strokes,
      tmpStroke: null,
      strokeWidth: 10,
      red: 128,
      green: 128,
      blue: 128,
      alpha: 0.5,
    };
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
      red: this.state.red,
      blue: this.state.blue,
      green: this.state.green,
      alpha: this.state.alpha,
      width: this.state.strokeWidth,
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
        width: this.state.strokeWidth,
        points: [],
      }, this.getCoordinates(ev)),
    });
  }

  handleMouseUp(ev) {
    if (this.mouseDown && this.state.tmpStroke) {
      this.setState({
        strokes: this.state.strokes.concat(
          this.addPointToStroke(this.state.tmpStroke, this.getCoordinates(ev))
        ),
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

  handleChangeStrokeWidth(ev) {
    this.setState({
      strokeWidth: parseInt(ev.target.value, 10),
    });
  }

  handleClickColorPickerToggle(ev) {
    if (ev.target === ev.currentTarget) {
      if (this.refs.colorPicker.style.display === 'none') {
        this.refs.colorPicker.style.display = 'block';
      } else {
        this.refs.colorPicker.style.display = 'none';
      }
    }
  }

  handleColorChange(ev) {
    this.setState({
      red: ev.rgb.r,
      green: ev.rgb.g,
      blue: ev.rgb.b,
      alpha: ev.rgb.a,
    });
  }

  makeRGBAString({ red, green, blue, alpha }) {
    return `rgba(${red},${green},${blue},${alpha})`;
  }

  render() {
    return (
      <div className="canvas" style={{ width: this.props.width + 2, margin: '0 auto' }}>
        <label>
          <span
            style={{
              display: 'inline-block',
              width: 100,
              height: this.props.controlHeight,
            }}
          >
            線の太さ ({this.state.strokeWidth})
          </span>
          <input
            type="range"
            min="1"
            max="50"
            value={this.state.strokeWidth}
            style={{
              width: 400,
              height: this.props.controlHeight,
              verticalAlign: 'middle',
            }}
            onChange={(ev) => this.handleChangeStrokeWidth(ev)}
          />
        </label>
        <span
          style={{
            display: 'inline-block',
            width: 220,
            height: this.props.controlHeight,
            paddingLeft: 30,
          }}
        >
          線の色 ({this.makeRGBAString(this.state)})
        </span>
        <div
          style={{
            display: 'inline-block',
            width: 60,
            height: this.props.controlHeight,
            position: 'relative',
            backgroundColor: this.makeRGBAString(this.state),
            verticalAlign: 'middle',
          }}
          onClick={(ev) => this.handleClickColorPickerToggle(ev)}
        >
          <div
            ref="colorPicker"
            style={{ display: 'none', position: 'absolute', top: this.props.controlHeight }}
          >
            <SketchPicker
              color={this.makeRGBAString(this.state)}
              onChange={(ev) => this.handleColorChange(ev)}
            />
          </div>
        </div>
        <svg
          ref="svgElement"
          width={this.props.width}
          height={this.props.height}
          style={{
            width: this.props.width,
            height: this.props.height,
            border: 'solid black 1px',
          }}
          viewBox={`0 0 ${this.props.width} ${this.props.height}`}
          onMouseDown={(ev) => this.handleMouseDown(ev)}
          onMouseUp={(ev) => this.handleMouseUp(ev)}
          onMouseMove={(ev) => this.handleMouseMove(ev)}
        >
          {this.state.strokes
            .concat(this.state.tmpStroke ? [this.state.tmpStroke] : [])
            .map((stroke) => (
              <polyline
                key={stroke.id}
                stroke={this.makeRGBAString(stroke)}
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
  controlHeight: React.PropTypes.number,
  strokes: React.PropTypes.array,
  followUpdates: React.PropTypes.bool,
};

Canvas.defaultProps = {
  width: 1028,
  height: 768,
  controlHeight: 20,
};

export default Canvas;
