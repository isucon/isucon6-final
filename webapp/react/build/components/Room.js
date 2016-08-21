'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _rcColorPicker = require('rc-color-picker');

var _rcColorPicker2 = _interopRequireDefault(_rcColorPicker);

var _Canvas = require('./Canvas');

var _Canvas2 = _interopRequireDefault(_Canvas);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var Room = function (_React$Component) {
  _inherits(Room, _React$Component);

  _createClass(Room, null, [{
    key: 'loadProps',
    value: function loadProps(_ref, cb) {
      var params = _ref.params;
      var loadContext = _ref.loadContext;

      var apiBaseUrl = loadContext ? loadContext.apiBaseUrl : window.apiBaseUrl;
      var csrfToken = loadContext ? loadContext.csrfToken : window.csrfToken;
      fetch(apiBaseUrl + '/api/rooms/' + params.id, {
        headers: { 'x-csrf-token': csrfToken }
      }).then(function (result) {
        return result.json();
      }).then(function (res) {
        cb(null, { id: res.room.id, name: res.room.name, strokes: res.room.strokes });
      });
    }
  }]);

  function Room(props) {
    _classCallCheck(this, Room);

    var _this = _possibleConstructorReturn(this, Object.getPrototypeOf(Room).call(this, props));

    _this.state = {
      strokes: props.strokes,
      tmpStroke: null,
      strokeWidth: 20,
      red: 128,
      green: 128,
      blue: 128,
      alpha: 0.7
    };
    return _this;
  }

  _createClass(Room, [{
    key: 'addPointToStroke',
    value: function addPointToStroke(orig, point) {
      return {
        id: orig.id,
        red: orig.red,
        blue: orig.blue,
        green: orig.green,
        alpha: orig.alpha,
        width: orig.width,
        points: orig.points.concat([point])
      };
    }
  }, {
    key: 'handleStrokeStart',
    value: function handleStrokeStart(point) {
      // TODO: return if this.state.tmpStroke already exists
      this.setState({
        tmpStroke: {
          id: 'tmp',
          red: this.state.red,
          blue: this.state.blue,
          green: this.state.green,
          alpha: this.state.alpha,
          width: this.state.strokeWidth,
          points: [point]
        }
      });
    }
  }, {
    key: 'handleStrokeMove',
    value: function handleStrokeMove(point) {
      this.setState({
        tmpStroke: this.addPointToStroke(this.state.tmpStroke, point)
      });
    }
  }, {
    key: 'handleStrokeEnd',
    value: function handleStrokeEnd(point) {
      var _this2 = this;

      this.setState({
        tmpStroke: this.addPointToStroke(this.state.tmpStroke, point)
      });

      var apiBaseUrl = window.apiBaseUrl;
      var csrfToken = window.csrfToken;

      fetch(apiBaseUrl + '/api/strokes/rooms/' + this.props.id, {
        method: 'POST',
        body: JSON.stringify(this.state.tmpStroke),
        headers: { 'x-csrf-token': csrfToken, 'content-type': 'application/json' }
      }).then(function (result) {
        if (result.status === 200) {
          return result.json();
        }
        throw result.json() || 'status ' + result.status;
      }).then(function (res) {
        var stroke = res.stroke;
        // TODO: check response
        _this2.setState({
          strokes: _this2.state.strokes.concat([stroke]),
          tmpStroke: null
        });
      }).catch(function (error) {
        // TODO: Flash
        console.log(error.message || 'Unknown error');
      });
    }
  }, {
    key: 'handleChangeStrokeWidth',
    value: function handleChangeStrokeWidth(ev) {
      this.setState({
        strokeWidth: parseInt(ev.target.value, 10)
      });
    }
  }, {
    key: 'handleColorChange',
    value: function handleColorChange(colors) {
      if (/#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})/.test(colors.color)) {
        this.setState({
          red: parseInt(RegExp.$1, 16),
          green: parseInt(RegExp.$2, 16),
          blue: parseInt(RegExp.$3, 16),
          alpha: colors.alpha / 100
        });
      }
    }
  }, {
    key: 'makeRGBString',
    value: function makeRGBString(_ref2) {
      var red = _ref2.red;
      var green = _ref2.green;
      var blue = _ref2.blue;

      return '#' + red.toString(16) + green.toString(16) + blue.toString(16);
    }
  }, {
    key: 'render',
    value: function render() {
      var _this3 = this;

      var strokes = this.state.tmpStroke === null ? this.state.strokes : this.state.strokes.concat([this.state.tmpStroke]);

      return _react2.default.createElement(
        'div',
        { className: 'room' },
        _react2.default.createElement(
          'h2',
          null,
          this.props.name
        ),
        _react2.default.createElement(
          'div',
          { className: 'canvas', style: { width: this.props.width + 2, margin: '0 auto' } },
          _react2.default.createElement(
            'label',
            null,
            _react2.default.createElement(
              'span',
              {
                style: {
                  display: 'inline-block',
                  width: 100,
                  height: this.props.controlHeight
                }
              },
              '線の太さ (',
              this.state.strokeWidth,
              ')'
            ),
            _react2.default.createElement('input', {
              type: 'range',
              min: '1',
              max: '50',
              value: this.state.strokeWidth,
              style: {
                width: 400,
                height: this.props.controlHeight,
                verticalAlign: 'middle'
              },
              onChange: function onChange(ev) {
                return _this3.handleChangeStrokeWidth(ev);
              }
            })
          ),
          _react2.default.createElement(
            'span',
            {
              style: {
                display: 'inline-block',
                height: this.props.controlHeight,
                paddingLeft: 20,
                paddingRight: 20
              }
            },
            '線の色'
          ),
          _react2.default.createElement(_rcColorPicker2.default, {
            color: this.makeRGBString(this.state),
            alpha: this.state.alpha * 100,
            placement: 'topLeft',
            onChange: function onChange(ev) {
              return _this3.handleColorChange(ev);
            }
          }),
          _react2.default.createElement(
            'div',
            { style: { border: 'solid black 1px' } },
            _react2.default.createElement(_Canvas2.default, {
              width: this.props.width,
              height: this.props.height,
              strokes: strokes,
              onStrokeStart: function onStrokeStart(point) {
                return _this3.handleStrokeStart(point);
              },
              onStrokeMove: function onStrokeMove(point) {
                return _this3.handleStrokeMove(point);
              },
              onStrokeEnd: function onStrokeEnd(point) {
                return _this3.handleStrokeEnd(point);
              }
            })
          )
        )
      );
    }
  }]);

  return Room;
}(_react2.default.Component);

Room.propTypes = {
  id: _react2.default.PropTypes.number,
  name: _react2.default.PropTypes.string,
  strokes: _react2.default.PropTypes.array,
  width: _react2.default.PropTypes.number,
  height: _react2.default.PropTypes.number,
  controlHeight: _react2.default.PropTypes.number
};

Room.defaultProps = {
  width: 1028,
  height: 768,
  controlHeight: 40
};

exports.default = Room;