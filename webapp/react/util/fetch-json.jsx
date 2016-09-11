import fetch from 'isomorphic-fetch';

export default function fetchJson(...args) {
  return fetch(...args)
    .then((response) => {
      const contentType = response.headers.get('content-type');
      if (!contentType || contentType.indexOf('application/json') === -1) {
        throw new Error(response.text());
      }
      return response.json();
    })
    .catch((err) => {
      throw new Error('Unexpected response from server');
    })
    .then((json) => {
      if (json.error) {
        throw new Error(json.error);
      }
      return json;
    });
}
