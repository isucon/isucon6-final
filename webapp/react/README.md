# react-ssr-test

API -> Node (React SSR) -> User
    <- Node (proxy)     <- React DOM Update

# Server

Express web server.

```
$(npm bin)/babel-node server.js
```

# Assets

Build `browser.js` and included files into `public/bundle.js`.

```
$(npm bin)/webpack --watch --display-error-details
```

# API

Backend for both Node (Express) server and client side JS. It's proxied by the Node server to be accessed by the client.

```
php -S 127.0.0.1:9901 api.php
```

