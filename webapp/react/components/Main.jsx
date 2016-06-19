import React from 'react';
import { Link } from 'react-router';

function Main({ children }) {
  return (
    <div
      className="main mdl-layout mdl-js-layout mdl-color--grey-100 mdl-color-text--grey-700"
    >
      <header className="mdl-layout__header mdl-layout__header--scroll">
        <div className="mdl-layout-icon">
          <i className="material-icons">border_color</i>
        </div>
        <div className="mdl-layout__header-row">
          <h1 className="mdl-layout-title">
            <Link to="/" style={{ color: 'inherit', textDecoration: 'none' }}>
              ISU-Channel
            </Link>
          </h1>
          <div className="mdl-layout-spacer"></div>
          描ける巨大匿名掲示板サイト！
        </div>
      </header>

      <div className="mdl-layout__content">
        <div style={{ width: '100%', maxWidth: '1200px', margin: '0 auto' }}>
          {children}
        </div>
      </div>

      <footer className="mdl-mini-footer">
        <div className="mdl-mini-footer__left-section">
          <div className="mdl-logo">by ISUCON</div>
        </div>
      </footer>
    </div>
  );
}

Main.propTypes = {
  children: React.PropTypes.object,
};

export default Main;
