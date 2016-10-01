import React from 'react';
import ColorPicker from 'rc-color-picker';
import Canvas from './Canvas';
import fetchJson from '../util/fetch-json';
import Slider from 'material-ui/Slider';
import Snackbar from 'material-ui/Snackbar';

class Room extends React.Component {
  static loadProps({ params, loadContext }, cb) {
    const apiBaseUrl = (loadContext || window).apiBaseUrl;

    fetchJson(`${apiBaseUrl}/api/rooms/${params.id}`)
    .then((res) => {
      cb(null, {
        id: res.room.id,
        name: res.room.name,
        strokes: res.room.strokes,
        width: res.room.canvas_width,
        height: res.room.canvas_height,
        watcherCount: res.room.watcher_count,
        csrfToken: (loadContext || window).csrfToken,
      });
    })
    .catch((err) => {
      cb(err);
    });
  }

  constructor(props) {
    super(props);
    this.state = {
      strokes: props.strokes,
      tmpStroke: null,
      strokeWidth: 20,
      red: 128,
      green: 128,
      blue: 128,
      alpha: 0.7,
      watcherCount: props.watcherCount,
      showError: false,
      errorMessage: '',
    };
  }

  componentDidMount() {
    const token = this.props.csrfToken;
    this.eventSource = new EventSource(`/api/stream/rooms/${this.props.id}?csrf_token=${token}`);
    this.eventSource.addEventListener('stroke', (ev) => {
      const strokes = this.state.strokes;
      const stroke = JSON.parse(ev.data);
      const isNew = !strokes.some((s) => s.id === stroke.id);
      if (isNew) {
        this.setState({ strokes: strokes.concat([stroke]).sort((a, b) => b.id - a.id) });
      }
    });
    this.eventSource.addEventListener('bad_request', (ev) => {
      this.setState({
        showError: true,
        errorMessage: ev.data,
      });
      this.eventSource.close();
    });
    this.eventSource.addEventListener('watcher_count', (ev) => {
      this.setState({
        watcherCount: parseInt(ev.data, 10),
      });
    });
  }

  componentWillUnmount() {
    this.eventSource.close();
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
    // TODO: return if this.state.tmpStroke already exists
    this.setState({
      tmpStroke: {
        id: 0,
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
    this.setState({
      tmpStroke: this.addPointToStroke(this.state.tmpStroke, point),
    });

    fetchJson(`/api/strokes/rooms/${this.props.id}`, {
      method: 'POST',
      body: JSON.stringify(this.state.tmpStroke),
      headers: { 'x-csrf-token': this.props.csrfToken, 'content-type': 'application/json' },
    })
      .then((res) => {
        const stroke = res.stroke;
        this.setState({
          strokes: this.state.strokes.concat([stroke]),
          tmpStroke: null,
        });
      })
      .catch((err) => {
        this.setState({
          showError: true,
          errorMessage: err.message,
        });
      });
  }

  handleChangeStrokeWidth(ev, value) {
    this.setState({
      strokeWidth: parseInt(value, 10),
    });
  }

  handleColorChange(colors) {
    if (/#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})/.test(colors.color)) {
      this.setState({
        red: parseInt(RegExp.$1, 16),
        green: parseInt(RegExp.$2, 16),
        blue: parseInt(RegExp.$3, 16),
        alpha: colors.alpha / 100,
      });
    }
  }

  makeRGBString({ red, green, blue }) {
    return `#${red.toString(16)}${green.toString(16)}${blue.toString(16)}`;
  }

  render() {
    const strokes = this.state.tmpStroke === null ?
      this.state.strokes :
      this.state.strokes.concat([this.state.tmpStroke]);

    return (
      <div className="room">
        <Snackbar
          open={this.state.showError}
          message={this.state.errorMessage}
        />

        <h2>{`${this.props.name} (${this.state.strokes.length}画)`}</h2>
        <p>{`${this.state.watcherCount}人が参加中`}</p>

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
            <Slider
              min={1}
              max={50}
              defaultValue={this.state.strokeWidth}
              step={1}
              style={{
                display: 'inline-block',
                width: 400,
                height: this.props.controlHeight,
                verticalAlign: 'sub',
              }}
              onChange={(ev, value) => this.handleChangeStrokeWidth(ev, value)}
            />
          </label>
          <span
            style={{
              display: 'inline-block',
              height: this.props.controlHeight,
              paddingLeft: 20,
              paddingRight: 20,
            }}
          >
            線の色
          </span>
          <ColorPicker
            color={this.makeRGBString(this.state)}
            alpha={this.state.alpha * 100}
            placement="topLeft"
            onChange={(ev) => this.handleColorChange(ev)}
          />
          <div style={{ border: 'solid black 1px' }}>
            <Canvas
              width={this.props.width}
              height={this.props.height}
              strokes={strokes}
              onStrokeStart={(point) => this.handleStrokeStart(point)}
              onStrokeMove={(point) => this.handleStrokeMove(point)}
              onStrokeEnd={(point) => this.handleStrokeEnd(point)}
            />
          </div>
        </div>

      </div>
    );
  }
}

Room.propTypes = {
  id: React.PropTypes.number,
  name: React.PropTypes.string,
  strokes: React.PropTypes.array,
  width: React.PropTypes.number,
  height: React.PropTypes.number,
  controlHeight: React.PropTypes.number,
  watcherCount: React.PropTypes.number,
  csrfToken: React.PropTypes.string,
};

Room.defaultProps = {
  controlHeight: 40,
};

export default Room;
