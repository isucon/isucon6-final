import React from 'react';
import { render } from 'react-dom';
import { Router, browserHistory } from 'react-router';
import routes from './routes';
import AsyncProps from 'async-props';

// for material-ui https://www.npmjs.com/package/material-ui
import injectTapEventPlugin from 'react-tap-event-plugin';
injectTapEventPlugin();

window.csrfToken = document.documentElement.dataset.csrfToken;
window.apiBaseUrl = location.origin;

const appElem = document.getElementById('app');

render((
  <Router
    history={browserHistory}
    routes={routes}
    render={(props) => <AsyncProps {...props} />}
  />
), appElem);
