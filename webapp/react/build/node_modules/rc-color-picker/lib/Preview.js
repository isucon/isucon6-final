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

var Preview = (function (_React$Component) {
  _inherits(Preview, _React$Component);

  function Preview() {
    _classCallCheck(this, Preview);

    _get(Object.getPrototypeOf(Preview.prototype), 'constructor', this).apply(this, arguments);
  }

  _createClass(Preview, [{
    key: 'onChange',
    value: function onChange(e) {
      var value = e.target.value;
      var color = colr.fromHex(value);
      this.props.onChange(color.toHsvObject());
      e.stopPropagation();
    }
  }, {
    key: 'getPrefixCls',
    value: function getPrefixCls() {
      return this.props.rootPrefixCls + '-preview';
    }
  }, {
    key: 'render',
    value: function render() {
      var prefixCls = this.getPrefixCls();
      var hex = colr.fromHsvObject(this.props.hsv).toHex();
      return _react2['default'].createElement(
        'div',
        { className: prefixCls },
        _react2['default'].createElement('span', { style: { backgroundColor: hex, opacity: this.props.alpha / 100 } }),
        _react2['default'].createElement('input', {
          type: 'color',
          value: hex,
          onChange: this.onChange.bind(this),
          onClick: this.props.onInputClick
        })
      );
    }
  }]);

  return Preview;
})(_react2['default'].Component);

exports['default'] = Preview;

Preview.propTypes = {
  rootPrefixCls: _react2['default'].PropTypes.string,
  hsv: _react2['default'].PropTypes.object,
  alpha: _react2['default'].PropTypes.number,
  onChange: _react2['default'].PropTypes.func,
  onInputClick: _react2['default'].PropTypes.func
};
module.exports = exports['default'];