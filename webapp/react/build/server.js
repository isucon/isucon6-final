'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _express = require('express');

var _express2 = _interopRequireDefault(_express);

var _path = require('path');

var _path2 = _interopRequireDefault(_path);

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _server = require('react-dom/server');

var _escapeHtml = require('escape-html');

var _escapeHtml2 = _interopRequireDefault(_escapeHtml);

var _reactRouter = require('react-router');

var _routes = require('./routes');

var _routes2 = _interopRequireDefault(_routes);

var _asyncProps = require('async-props');

var _asyncProps2 = _interopRequireDefault(_asyncProps);

var _isomorphicFetch = require('isomorphic-fetch');

var _isomorphicFetch2 = _interopRequireDefault(_isomorphicFetch);

var _httpProxyMiddleware = require('http-proxy-middleware');

var _httpProxyMiddleware2 = _interopRequireDefault(_httpProxyMiddleware);

var _Canvas = require('./components/Canvas');

var _Canvas2 = _interopRequireDefault(_Canvas);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var apiBaseUrl = process.env.API;
if (!apiBaseUrl) {
  throw 'Please set environment variable API=http://...';
}

var app = (0, _express2.default)();

app.use(_express2.default.static(_path2.default.join(__dirname, 'public')));

app.use('/api/*', (0, _httpProxyMiddleware2.default)({ target: apiBaseUrl, changeOrigin: true }));

app.get('/img/:id', function (req, res) {
  (0, _isomorphicFetch2.default)(apiBaseUrl + '/api/rooms/' + req.params.id).then(function (result) {
    return result.json();
  }).then(function (json) {
    var svg = (0, _server.renderToStaticMarkup)(_react2.default.createElement(_Canvas2.default, {
      width: 1028,
      height: 768,
      strokes: json.room.strokes
    }));
    res.type('image/svg+xml').send(
    // Waiting for React 15.3.0 https://github.com/facebook/react/pull/6471#event-722021290
    '<?xml version="1.0" standalone="no"?>' + '<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">' + svg.replace('<svg ', '<svg xmlns="http://www.w3.org/2000/svg" '));
  }).catch(function (err) {
    res.status(500).send(err.message);
  });
});

app.get('*', function (req, res) {
  // https://github.com/reactjs/react-router/blob/master/docs/guides/ServerRendering.md
  (0, _reactRouter.match)({ routes: _routes2.default, location: req.url }, function (err, redirectLocation, renderProps) {
    if (err) {
      console.error(err);
      res.status(500).send(err.message);
    } else if (redirectLocation) {
      res.redirect(302, redirectLocation.pathname + redirectLocation.search);
    } else if (renderProps) {

      (0, _isomorphicFetch2.default)(apiBaseUrl + '/api/csrf_token').then(function (result) {
        return result.json();
      }).then(function (json) {
        var csrfToken = json.token;
        var loadContext = { apiBaseUrl: apiBaseUrl, csrfToken: csrfToken };

        // https://github.com/ryanflorence/async-props
        (0, _asyncProps.loadPropsOnServer)(renderProps, loadContext, function (err, asyncProps, scriptTag) {
          if (err) {
            console.error(err);
            res.status(500).send(err.message);
          } else {
            var appHTML = (0, _server.renderToString)(_react2.default.createElement(_asyncProps2.default, _extends({}, renderProps, asyncProps)));

            var html = createHtml(appHTML, scriptTag, csrfToken);

            res.status(200).send(html);
          }
        });
      }).catch(function (err) {
        res.status(500).send(err.message);
      });
    } else {
      res.status(404).send('Not found');
    }
  });
});

var PORT = process.env.PORT || 8800;
app.listen(PORT, function () {
  console.log('Production Express server running at localhost:' + PORT);
});

function createHtml(appHtml, scriptTag, csrfToken) {
  return '<!DOCTYPE html>\n<html data-csrf-token="' + (0, _escapeHtml2.default)(csrfToken) + '">\n  <head>\n    <title>SSR Sample</title>\n    <link rel="stylesheet" href="/mdl/material.min.css">\n    <link rel="stylesheet" href="/iconfont/material-icons.css">\n    <link rel="stylesheet" href="/css/rc-color-picker.css">\n    <script src="/mdl/material.min.js" async></script>\n    <script src="/bundle.js" async></script>\n  </head>\n  <body>\n    <div id="app">' + appHtml + '</div>\n    ' + scriptTag + '\n  </body>\n</html>';
}