'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _reactRouter = require('react-router');

var _Index = require('./components/Index');

var _Index2 = _interopRequireDefault(_Index);

var _Main = require('./components/Main');

var _Main2 = _interopRequireDefault(_Main);

var _Room = require('./components/Room');

var _Room2 = _interopRequireDefault(_Room);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

exports.default = _react2.default.createElement(
  _reactRouter.Route,
  { path: '/', component: _Main2.default },
  _react2.default.createElement(_reactRouter.IndexRoute, { component: _Index2.default }),
  _react2.default.createElement(_reactRouter.Route, { path: '/rooms/:id', component: _Room2.default })
);