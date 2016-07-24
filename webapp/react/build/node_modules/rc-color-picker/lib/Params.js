'use strict';

Object.defineProperty(exports, '__esModule', {
  value: true
});

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _get = function get(_x, _x2, _x3) { var _again = true; _function: while (_again) { var object = _x, property = _x2, receiver = _x3; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x = parent; _x2 = property; _x3 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ('value' in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _colr = require('colr');

var _colr2 = _interopRequireDefault(_colr);

var colr = new _colr2['default']();
var modesMap = ['RGB', 'HSB', 'HSL'];

var Params = (function (_React$Component) {
  _inherits(Params, _React$Component);

  function Params(props) {
    var _this = this;

    _classCallCheck(this, Params);

    _get(Object.getPrototypeOf(Params.prototype), 'constructor', this).call(this, props);

    var color = colr.fromHsvObject(props.hsv);

    // 管理 input 的状态
    this.state = {
      mode: props.mode,
      color: color,
      hex: color.toHex().substr(1)
    };

    var events = ['onHexHandler', 'onAlphaHandler', 'onColorChannelChange', 'onModeChange', 'getChannelInRange', 'getColorByChannel'];

    events.forEach(function (e) {
      if (_this[e]) {
        _this[e] = _this[e].bind(_this);
      }
    });
  }

  _createClass(Params, [{
    key: 'componentWillReceiveProps',
    value: function componentWillReceiveProps(nextProps) {
      if (nextProps.hsv !== this.props.hsv) {
        var color = colr.fromHsvObject(nextProps.hsv);
        this.setState({
          hex: color.toHex().substr(1),
          color: color
        });
      }
    }
  }, {
    key: 'onHexHandler',
    value: function onHexHandler(event) {
      var value = event.target.value;
      var color = null;
      try {
        color = _colr2['default'].fromHex(value);
      } catch (e) {
        /* eslint no-empty:0 */
      }

      if (color !== null) {
        this.setState({
          color: color,
          hex: value
        });
        this.props.onChange(color.toHsvObject(), false);
      } else {
        this.setState({
          hex: value
        });
      }
    }
  }, {
    key: 'onModeChange',
    value: function onModeChange() {
      var mode = this.state.mode;
      var modeIndex = (modesMap.indexOf(mode) + 1) % modesMap.length;
      var state = this.state;

      mode = modesMap[modeIndex];
      var colorChannel = this.getColorChannel(state.color, mode);
      this.setState({
        mode: mode,
        colorChannel: colorChannel
      });
    }
  }, {
    key: 'onAlphaHandler',
    value: function onAlphaHandler(event) {
      var alpha = parseInt(event.target.value, 10);
      if (isNaN(alpha)) {
        alpha = 0;
      }
      alpha = Math.max(0, alpha);
      alpha = Math.min(alpha, 100);

      this.setState({
        alpha: alpha
      });

      this.props.onAlphaChange(alpha);
    }
  }, {
    key: 'onColorChannelChange',
    value: function onColorChannelChange(index, event) {
      var value = this.getChannelInRange(event.target.value, index);
      var colorChannel = this.getColorChannel();

      colorChannel[index] = value;

      var color = this.getColorByChannel(colorChannel);

      this.setState({
        hex: color.toHex().substr(1),
        color: color
      });
      this.props.onChange(color.toHsvObject(), false);
    }
  }, {
    key: 'getChannelInRange',
    value: function getChannelInRange(value, index) {
      var channelMap = {
        RGB: [[0, 255], [0, 255], [0, 255]],
        HSB: [[0, 360], [0, 100], [0, 100]],
        HSL: [[0, 360], [0, 100], [0, 100]]
      };
      var mode = this.state.mode;
      var range = channelMap[mode][index];
      var result = parseInt(value, 10);
      if (isNaN(result)) {
        result = 0;
      }
      result = Math.max(range[0], result);
      result = Math.min(result, range[1]);
      return result;
    }
  }, {
    key: 'getColorByChannel',
    value: function getColorByChannel(colorChannel) {
      var colorMode = this.state.mode;
      var color = undefined;
      switch (colorMode) {
        case 'RGB':
          color = colr.fromRgbArray(colorChannel);
          break;
        case 'HSB':
          color = colr.fromHsvArray(colorChannel);
          break;
        case 'HSL':
          color = colr.fromHslArray(colorChannel);
          break;
        default:
          color = colr.fromRgbArray(colorChannel);
      }
      return color;
    }
  }, {
    key: 'getPrefixCls',
    value: function getPrefixCls() {
      return this.props.rootPrefixCls + '-params';
    }
  }, {
    key: 'getColorChannel',
    value: function getColorChannel(colrInstance, mode) {
      var color = colrInstance || this.state.color;
      var colorMode = mode || this.state.mode;
      var result = undefined;
      switch (colorMode) {
        case 'RGB':
          result = color.toRgbArray();
          break;
        case 'HSB':
          result = color.toHsvArray();
          break;
        case 'HSL':
          result = color.toHslArray();
          break;
        default:
          result = color.toRgbArray();
      }
      return result;
    }
  }, {
    key: 'render',
    value: function render() {
      var prefixCls = this.getPrefixCls();
      var colorChannel = this.getColorChannel();
      return _react2['default'].createElement(
        'div',
        { className: prefixCls },
        _react2['default'].createElement(
          'div',
          { className: prefixCls + '-' + 'input' },
          _react2['default'].createElement('input', {
            className: prefixCls + '-' + 'hex',
            type: 'text',
            maxLength: '6',
            onChange: this.onHexHandler,
            value: this.state.hex.toUpperCase()
          }),
          _react2['default'].createElement('input', { type: 'number', ref: 'channel_0',
            value: colorChannel[0],
            onChange: this.onColorChannelChange.bind(null, 0) }),
          _react2['default'].createElement('input', { type: 'number', ref: 'channel_1',
            value: colorChannel[1],
            onChange: this.onColorChannelChange.bind(null, 1) }),
          _react2['default'].createElement('input', { type: 'number', ref: 'channel_2',
            value: colorChannel[2],
            onChange: this.onColorChannelChange.bind(null, 2) }),
          _react2['default'].createElement('input', { type: 'number',
            value: this.props.alpha,
            onChange: this.onAlphaHandler })
        ),
        _react2['default'].createElement(
          'div',
          { className: prefixCls + '-' + 'lable' },
          _react2['default'].createElement(
            'label',
            { className: prefixCls + '-' + 'lable-hex' },
            'Hex'
          ),
          _react2['default'].createElement(
            'label',
            { className: prefixCls + '-' + 'lable-number',
              onClick: this.onModeChange
            },
            this.state.mode[0]
          ),
          _react2['default'].createElement(
            'label',
            { className: prefixCls + '-' + 'lable-number',
              onClick: this.onModeChange
            },
            this.state.mode[1]
          ),
          _react2['default'].createElement(
            'label',
            { className: prefixCls + '-' + 'lable-number',
              onClick: this.onModeChange
            },
            this.state.mode[2]
          ),
          _react2['default'].createElement(
            'label',
            { className: prefixCls + '-' + 'lable-alpha' },
            'A'
          )
        )
      );
    }
  }]);

  return Params;
})(_react2['default'].Component);

exports['default'] = Params;

Params.propTypes = {
  onChange: _react2['default'].PropTypes.func,
  hsv: _react2['default'].PropTypes.object,
  alpha: _react2['default'].PropTypes.number,
  rootPrefixCls: _react2['default'].PropTypes.string,
  onAlphaChange: _react2['default'].PropTypes.func,
  mode: _react2['default'].PropTypes.oneOf(modesMap)
};

Params.defaultProps = {
  mode: modesMap[0]
};
module.exports = exports['default'];