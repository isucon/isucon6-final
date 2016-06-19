import express from 'express';
import path from 'path';
import React from 'react';
import { renderToString } from 'react-dom/server';
import escape from 'escape-html';
import { match, RouterContext } from 'react-router';
import routes from './routes';
import AsyncProps, { loadPropsOnServer } from 'async-props';
import fetch from 'isomorphic-fetch';
import proxy from 'http-proxy-middleware';

const apiBaseUrl = process.env.API;
if (!apiBaseUrl) {
  throw 'Please set environment variable API=http://...';
}

const app = express();

app.use(express.static(path.join(__dirname, 'public')));

app.use('/api/*', proxy({ target: apiBaseUrl, changeOrigin: true }));

app.get('*', (req, res) => {
  // https://github.com/reactjs/react-router/blob/master/docs/guides/ServerRendering.md
  match({ routes, location: req.url }, (err, redirectLocation, renderProps) => {
    if (err) {
      console.error(err)
      res.status(500).send(err.message);
    } else if (redirectLocation) {
      res.redirect(302, redirectLocation.pathname + redirectLocation.search);
    } else if (renderProps) {

      fetch(`${apiBaseUrl}/api/csrf_token`)
        .then((result) => result.json())
        .then((json) => {
          const csrfToken = json.token;
          const loadContext = { apiBaseUrl, csrfToken };

          // https://github.com/ryanflorence/async-props
          loadPropsOnServer(renderProps, loadContext, (err, asyncProps, scriptTag) => {
            if (err) {
              console.error(err)
              res.status(500).send(err.message);
            } else {
              const appHTML = renderToString(
                <AsyncProps {...renderProps} {...asyncProps} />
              );

              const html = createHtml(appHTML, scriptTag, csrfToken);

              res.status(200).send(html);
            }
          });

        })
        .catch((err) => {
          res.status(500).send(err.message);
        })
    } else {
      res.status(404).send('Not found')
    }
  });
});

const PORT = process.env.PORT || 8800;
app.listen(PORT, () => {
  console.log('Production Express server running at localhost:' + PORT);
});

function createHtml(appHtml, scriptTag, csrfToken) {
  return `<!DOCTYPE html>
<html data-csrf-token="${escape(csrfToken)}">
  <head>
    <title>SSR Sample</title>
    <link rel="stylesheet" href="/mdl/material.min.css">
    <link rel="stylesheet" href="/iconfont/material-icons.css">
    <script src="/mdl/material.min.js" async></script>
    <script src="/bundle.js" async></script>
  </head>
  <body>
    <div id="app">${appHtml}</div>
    ${scriptTag}
  </body>
</html>`;
}
