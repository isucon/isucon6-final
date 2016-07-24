'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _reactRouter = require('react-router');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function Main(_ref) {
  var children = _ref.children;

  return _react2.default.createElement(
    'div',
    {
      className: 'main mdl-layout mdl-js-layout mdl-color--grey-100 mdl-color-text--grey-700'
    },
    _react2.default.createElement(
      'header',
      { className: 'mdl-layout__header mdl-layout__header--scroll' },
      _react2.default.createElement(
        'div',
        { className: 'mdl-layout-icon' },
        _react2.default.createElement(
          'i',
          { className: 'material-icons' },
          'border_color'
        )
      ),
      _react2.default.createElement(
        'div',
        { className: 'mdl-layout__header-row' },
        _react2.default.createElement(
          'h1',
          { className: 'mdl-layout-title' },
          _react2.default.createElement(
            _reactRouter.Link,
            { to: '/', style: { color: 'inherit', textDecoration: 'none' } },
            'ISU-Channel'
          )
        ),
        _react2.default.createElement('div', { className: 'mdl-layout-spacer' }),
        '描ける巨大匿名掲示板サイト！'
      )
    ),
    _react2.default.createElement(
      'div',
      { className: 'mdl-layout__content' },
      _react2.default.createElement(
        'div',
        { style: { width: '100%', maxWidth: '1200px', margin: '0 auto' } },
        children
      )
    ),
    _react2.default.createElement(
      'footer',
      { className: 'mdl-mini-footer' },
      _react2.default.createElement(
        'div',
        { className: 'mdl-mini-footer__left-section' },
        _react2.default.createElement(
          'div',
          { className: 'mdl-logo' },
          'by ISUCON'
        )
      )
    )
  );
}

Main.propTypes = {
  children: _react2.default.PropTypes.object
};

exports.default = Main;