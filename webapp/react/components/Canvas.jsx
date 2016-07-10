import React from 'react';
import { SketchPicker } from 'react-color';
import Svg from './Svg';

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

  handleStrokeStart(point) {
    this.setState({
      tmpStroke: {
        id: Date.now(), // TODO:
        red: this.state.red,
        blue: this.state.blue,
        green: this.state.green,
        alpha: this.state.alpha,
        width: this.state.strokeWidth,
        points: [point],
      },
    });
  }

  handleStrokeMove(point) {
    this.setState({
      tmpStroke: this.addPointToStroke(this.state.tmpStroke, point),
    });
  }

  handleStrokeEnd(point) {
    const tmpStroke = this.addPointToStroke(this.state.tmpStroke, point);
    // TODO: API request
    this.setState({
      strokes: this.state.strokes.concat([tmpStroke]),
      tmpStroke: null,
    });
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
    const strokes = this.state.tmpStroke === null ?
      this.state.strokes :
      this.state.strokes.concat([this.state.tmpStroke]);

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
        <Svg
          width={this.props.width}
          height={this.props.height}
          strokes={strokes}
          onStrokeStart={(point) => this.handleStrokeStart(point)}
          onStrokeMove={(point) => this.handleStrokeMove(point)}
          onStrokeEnd={(point) => this.handleStrokeEnd(point)}
        />
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
