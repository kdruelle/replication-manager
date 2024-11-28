const toBase64 = (str) => btoa(unescape(encodeURIComponent(str)));

const authConfig = {
  1: { // Local calls
    resolveUrl: (apiUrl) => `/api/${apiUrl}`,
    getToken: () => localStorage.getItem('user_token')
  },
  2: { // Peer calls
    resolveUrl: (apiUrl, baseUrl) => `/peer/${baseUrl}/api/${apiUrl}`,
    getToken: (baseUrl) => {
      return localStorage.getItem(`user_token_${baseUrl}`);
    }
  },
  3: { // Mattermost API
    resolveUrl: (apiUrl) => `https://meet.signal18.io/api/v4/${apiUrl}`,
    getToken: () => localStorage.getItem('meet_token')
  }
};

const getContentType = (type) => type === 'json' ? { 'Content-Type': 'application/json; charset="utf-8"' } : {};

const buildHeaders = (authValue, contentType, baseUrl) => {
  const encodedBaseUrl = toBase64(baseUrl);
  const { getToken } = authConfig[authValue] || {};
  const token = getToken ? getToken(encodedBaseUrl) : null;

  return {
    ...getContentType(contentType),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    Accept: '*/*',
  };
};

const resolveUrl = (apiUrl, authValue, baseUrl) => {
  const encodedBaseUrl = toBase64(baseUrl);
  const { resolveUrl } = authConfig[authValue] || {};
  return resolveUrl ? resolveUrl(apiUrl, encodedBaseUrl) : apiUrl;
};

const handleResponse = async (response) => {
  const contentType = response.headers.get('Content-Type')
  let data = null
  if (contentType && contentType.includes('application/json')) {
    data = await response.json()
  } else if (contentType && contentType.includes('text/plain')) {
    data = await response.text()
    if (data.startsWith('[') || data.startsWith('{')) {
      try {
        data = JSON.parse(data);
      } catch (e) {
        throw new Error("Failed to parse JSON: " + e.message);
      }
    }
  }

  return { data, status: response.status };
};

const performRequest = async (method, apiUrl, params, authValue, baseUrl = '') => {
  const url = resolveUrl(apiUrl, authValue, baseUrl);
  const headers = buildHeaders(authValue, 'json', baseUrl);

  const options = {
    method,
    headers,
    ...(params ? { body: JSON.stringify(params) } : {})
  };

  if (apiUrl == 'login' || apiUrl == 'login-git') {
    delete options.headers.Authorization
  }

  try {
    const response = await fetch(url, options);
    if (response.status === 401) {
      localStorage.removeItem('user_token')
      localStorage.removeItem('username')
      window.location.reload()
    } else {
      return handleResponse(response);
    }
  } catch (error) {
    console.error(`${method} Request Error:`, error);
    throw error;
  }
};

const requestWrapper = (authValue, baseUrl = '') => ({
  get: (apiUrl, params) => performRequest('GET', apiUrl, params, authValue, baseUrl),
  post: (apiUrl, params) => performRequest('POST', apiUrl, params, authValue, baseUrl),
  getAll: (urls, params) => {
    const requests = urls.map((url) => {
      const resolvedUrl = resolveUrl(url, authValue, baseUrl);
      const headers = buildHeaders(authValue, 'json', baseUrl);

      const options = {
        method: 'GET',
        headers,
        ...(params ? { body: JSON.stringify(params) } : {})
      };

      return fetch(resolvedUrl, options);
    });

    return Promise.allSettled(requests).then((results) =>
      results.map((result, idx) =>
        result.status === 'fulfilled'
          ? { url: urls[idx], data: result.value }
          : { url: urls[idx], error: result.reason }
      )
    );
  }
});

export const localApi = requestWrapper(1); // Wrapper for local API calls
export const peerApi = (baseUrl) => requestWrapper(2, baseUrl); // Wrapper for peer API calls
export const meetApi = requestWrapper(3); // Wrapper for Mattermost (meetAPI) calls

export const getApi = (baseURL = '') => {
  return baseURL ? peerApi(baseURL) : localApi;
};

