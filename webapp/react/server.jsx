import express from 'express';
import https from 'https';
import fs from 'fs';
import path from 'path';
import React from 'react';
import { renderToString, renderToStaticMarkup } from 'react-dom/server';
import escape from 'escape-html';
import { match, RouterContext } from 'react-router';
import routes from './routes';
import AsyncProps, { loadPropsOnServer } from 'async-props';
import fetchJson from './util/fetch-json';
import proxy from 'http-proxy-middleware';
import Canvas from './components/Canvas';

// for material-ui https://www.npmjs.com/package/material-ui
import injectTapEventPlugin from 'react-tap-event-plugin';
injectTapEventPlugin();

const apiBaseUrl = process.env.API;
if (!apiBaseUrl) {
  throw 'Please set environment variable API=http://...';
}
if (!process.env.SSL_KEY) {
  throw 'Please set environment variable SSL_KEY=/path/to/server.key';
}
if (!process.env.SSL_CERT) {
  throw 'Please set environment variable SSL_CERT=/path/to/server.crt';
}

const options = {
  key: fs.readFileSync(process.env.SSL_KEY),
  cert: fs.readFileSync(process.env.SSL_CERT),
};

const app = express();

app.use(express.static('public'));

app.use('/api/*', proxy({ target: apiBaseUrl, changeOrigin: true }));

app.get('/img/:id', (req, res) => {
  fetchJson(`${apiBaseUrl}/api/rooms/${req.params.id}`)
    .then((json) => {
      const svg = renderToStaticMarkup(
        <Canvas
          width={json.room.canvas_width}
          height={json.room.canvas_height}
          strokes={json.room.strokes}
        />
      );
      res.type('image/svg+xml').send(
        '<?xml version="1.0" standalone="no"?>' +
        '<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">' +
        svg
      );
    })
    .catch((err) => {
      console.log(`error: ${err.message}`);
      return res.status(500);
    });
});

app.get('*', (req, res) => {
  // https://github.com/reactjs/react-router/blob/master/docs/guides/ServerRendering.md
  match({ routes, location: req.url }, (err, redirectLocation, renderProps) => {
    if (err) {
      console.error(err);
      return res.status(500).send('Internal Server Error');
    } else if (redirectLocation) {
      return res.redirect(302, redirectLocation.pathname + redirectLocation.search);
    } else if (!renderProps) {
      return res.status(404).send('Not found');
    }

    fetch(`${apiBaseUrl}/api/csrf_token`, {
      method: 'POST',
    })
      .then((result) => result.json())
      .then((json) => {
        const csrfToken = json.token;
        const loadContext = { apiBaseUrl, csrfToken };

        // https://github.com/ryanflorence/async-props
        loadPropsOnServer(renderProps, loadContext, (err, asyncProps, scriptTag) => {
          if (err) {
            console.error(err);
            return res.status(500).send('Internal Server Error');
          } else {
            const appHTML = renderToString(
              <AsyncProps {...renderProps} {...asyncProps} />
            );

            const html = createHtml(appHTML, scriptTag, csrfToken);

            return res.status(200).send(html);
          }
        });

      })
      .catch((err) => {
        console.error(err);
        return res.status(500).send('Internal Server Error');
      });
  });
});

const PORT = process.env.PORT || 443;
https.createServer(options, app).listen(PORT);

function createHtml(appHtml, scriptTag, csrfToken) {
  return `<!DOCTYPE html>
<html data-csrf-token="${escape(csrfToken)}">
  <head>
    <title>ISUketch</title>
    <meta name="viewport" content="width=device-width,initial-scale=1">
    <link rel="stylesheet" href="/css/rc-color-picker.css">
    <link rel="stylesheet" href="/css/sanitize.css">
    <script src="/bundle.js" async></script>
  </head>
  <body>
    <div id="app">${appHtml}</div>
    ${scriptTag}
  </body>
</html>`;
}
