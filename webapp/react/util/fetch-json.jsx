import fetch from 'isomorphic-fetch';

export default function fetchJson(...args) {
  return fetch(...args)
    .then((response) => {
      const contentType = response.headers.get('content-type');
      if (!contentType || contentType.indexOf('application/json') === -1) {
        return response.text().catch(() => {
          // console.error(text);
          throw new Error('サーバーエラー');
        });
      }
      return response.json();
    })
    .then((json) => {
      if (json.error) {
        throw new Error(json.error);
      }
      return json;
    });
}
