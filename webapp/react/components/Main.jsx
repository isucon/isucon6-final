import React from 'react';
import { Link } from 'react-router';

function Main({ children }) {
  return (
    <div className="main">
      <header>
        <h1><Link to="/">ISU-Channel</Link></h1>
      </header>
      {children}
    </div>
  );
}

Main.propTypes = {
  children: React.PropTypes.object,
};

export default Main;
