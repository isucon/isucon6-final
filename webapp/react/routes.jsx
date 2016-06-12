import React from 'react';
import { Route, IndexRoute } from 'react-router';
import Index from './components/Index';
import Main from './components/Main';
import Room from './components/Room';

export default (
  <Route path="/" component={Main}>
    <IndexRoute component={Index} />
    <Route path="/rooms/:id" component={Room} />
  </Route>
);
