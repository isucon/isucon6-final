import React from 'react';
import { Link } from 'react-router';
import lightBaseTheme from 'material-ui/styles/baseThemes/lightBaseTheme';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import AppBar from 'material-ui/AppBar';

function Main({ children }) {
  return (
    <MuiThemeProvider muiTheme={getMuiTheme(lightBaseTheme, { userAgent: false })}>
      <div>
        <AppBar
          title={
            <Link to="/" style={{ color: 'inherit', textDecoration: 'none' }}>
              ISUketch 〜描ける巨大匿名掲示板サイト〜
            </Link>
          }
          showMenuIconButton={false}
        />

        <div style={{ width: '100%', maxWidth: '1200px', margin: '0 auto' }}>
          {children}
        </div>
      </div>
    </MuiThemeProvider>
  );
}

Main.propTypes = {
  children: React.PropTypes.object,
};

export default Main;
