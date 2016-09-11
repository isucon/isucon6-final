import fetch from 'isomorphic-fetch';

export default function fetchJson(...args) {
  return fetch(...args)
    .then((response) => {
      const contentType = response.headers.get('content-type');
      if (response.status !== 200
        || !contentType
        || contentType.indexOf('application/json') === -1
      ) {
        throw new Error(`Bad response ${response.status}: ${response.text()}`);
      }
      return response.json();
    });
}
