'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _reactRouter = require('react-router');

var _isomorphicFetch = require('isomorphic-fetch');

var _isomorphicFetch2 = _interopRequireDefault(_isomorphicFetch);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var Index = function (_React$Component) {
  _inherits(Index, _React$Component);

  function Index() {
    _classCallCheck(this, Index);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(Index).apply(this, arguments));
  }

  _createClass(Index, [{
    key: 'render',
    value: function render() {
      return _react2.default.createElement(
        'div',
        { className: 'index' },
        _react2.default.createElement(
          'div',
          null,
          _react2.default.createElement(
            'form',
            { method: 'POST', action: '/rooms', className: 'new-room' },
            _react2.default.createElement(
              'h3',
              null,
              '新規部屋作成'
            ),
            _react2.default.createElement(
              'label',
              null,
              '部屋名:',
              _react2.default.createElement('input', { type: 'text', placeholder: '例: ひたすら椅子を描く部屋', required: true, name: 'name' })
            ),
            _react2.default.createElement('input', { type: 'hidden', name: 'token', value: '' }),
            _react2.default.createElement(
              'button',
              { className: 'create' },
              '作成'
            )
          )
        ),
        _react2.default.createElement(
          'div',
          { className: 'mdl-grid' },
          this.props.rooms.map(function (room) {
            return _react2.default.createElement(
              'div',
              { className: 'mdl-cell mdl-cell--3-col mdl-card mdl-shadow--2dp', key: room.id },
              _react2.default.createElement(
                'div',
                { className: 'mdl-card__media' },
                _react2.default.createElement('img', {
                  style: { maxWidth: '100%' },
                  className: 'thumbnail',
                  src: '/img/' + room.id,
                  alt: room.name
                })
              ),
              _react2.default.createElement(
                'div',
                { className: 'mdl-card__supporting-text' },
                _react2.default.createElement(
                  'h2',
                  { className: 'mdl-card__title-text' },
                  room.name
                ),
                _react2.default.createElement(
                  'p',
                  null,
                  room.watcherCount,
                  '人が参加'
                )
              ),
              _react2.default.createElement(
                'div',
                { className: 'mdl-card__actions mdl-card--border' },
                _react2.default.createElement(
                  _reactRouter.Link,
                  {
                    to: '/rooms/' + room.id,
                    className: 'mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect'
                  },
                  '入室'
                )
              )
            );
          })
        )
      );
    }
  }], [{
    key: 'loadProps',
    value: function loadProps(params, cb) {
      var apiBaseUrl = params.loadContext ? params.loadContext.apiBaseUrl : window.apiBaseUrl;
      var csrfToken = params.loadContext ? params.loadContext.csrfToken : window.csrfToken;
      (0, _isomorphicFetch2.default)(apiBaseUrl + '/api/rooms', {
        headers: { 'x-csrf-token': csrfToken }
      }).then(function (result) {
        return result.json();
      }).then(function (res) {
        cb(null, { rooms: res.rooms });
      });
    }
  }]);

  return Index;
}(_react2.default.Component);

Index.propTypes = {
  rooms: _react2.default.PropTypes.array
};

exports.default = Index;