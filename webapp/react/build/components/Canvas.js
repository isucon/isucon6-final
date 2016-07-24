"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _react = require("react");

var _react2 = _interopRequireDefault(_react);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var Canvas = function (_React$Component) {
  _inherits(Canvas, _React$Component);

  function Canvas(props) {
    _classCallCheck(this, Canvas);

    var _this = _possibleConstructorReturn(this, Object.getPrototypeOf(Canvas).call(this, props));

    _this.isMouseDown = false;
    return _this;
  }

  _createClass(Canvas, [{
    key: "getCoordinates",
    value: function getCoordinates(ev) {
      var rect = this.refs.svgElement.getBoundingClientRect();
      return {
        x: ev.clientX - rect.left,
        y: ev.clientY - rect.top
      };
    }
  }, {
    key: "handleMouseDown",
    value: function handleMouseDown(ev) {
      if (!this.isMouseDown) {
        this.isMouseDown = true;
        this.props.onStrokeStart(this.getCoordinates(ev));
      }
    }
  }, {
    key: "handleMouseMove",
    value: function handleMouseMove(ev) {
      if (this.isMouseDown) {
        this.props.onStrokeMove(this.getCoordinates(ev));
      }
    }
  }, {
    key: "handleMouseUp",
    value: function handleMouseUp(ev) {
      if (this.isMouseDown) {
        this.isMouseDown = false;
        this.props.onStrokeEnd(this.getCoordinates(ev));
      }
    }
  }, {
    key: "render",
    value: function render() {
      var _this2 = this;

      // console.log(this.props.strokes);
      return _react2.default.createElement(
        "svg",
        {
          version: "1.1",
          baseProfile: "full",
          ref: "svgElement",
          width: this.props.width,
          height: this.props.height,
          style: {
            width: this.props.width,
            height: this.props.height,
            backgroundColor: 'white'
          },
          viewBox: "0 0 " + this.props.width + " " + this.props.height,
          onMouseDown: function onMouseDown(ev) {
            return _this2.handleMouseDown(ev);
          },
          onMouseUp: function onMouseUp(ev) {
            return _this2.handleMouseUp(ev);
          },
          onMouseMove: function onMouseMove(ev) {
            return _this2.handleMouseMove(ev);
          }
        },
        this.props.strokes.map(function (stroke) {
          return _react2.default.createElement("polyline", {
            key: stroke.id,
            stroke: "rgba(" + stroke.red + "," + stroke.green + "," + stroke.blue + "," + stroke.alpha + ")",
            strokeWidth: stroke.width,
            strokeLinecap: "round",
            strokeLinejoin: "round",
            fill: "none",
            points: stroke.points.map(function (point) {
              return point.x + "," + point.y;
            }).join(' ')
          });
        })
      );
    }
  }]);

  return Canvas;
}(_react2.default.Component);

Canvas.propTypes = {
  width: _react2.default.PropTypes.number,
  height: _react2.default.PropTypes.number,
  strokes: _react2.default.PropTypes.array,
  onStrokeStart: _react2.default.PropTypes.func,
  onStrokeMove: _react2.default.PropTypes.func,
  onStrokeEnd: _react2.default.PropTypes.func
};

exports.default = Canvas;