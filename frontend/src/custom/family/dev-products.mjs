export function familyProductEntry(url = '', method = 'GET') {
  if (method !== 'GET' && method !== 'HEAD') return null;
  const queryAt = url.indexOf('?');
  const path = queryAt < 0 ? url : url.slice(0, queryAt);
  const query = queryAt < 0 ? '' : url.slice(queryAt);
  if (!/^\/(recipes|studio)\/?$/.test(path)) return null;
  if (!path.endsWith('/')) return { redirect: true, url: path + '/' + query };
  return { redirect: false, url: path + 'index.html' + query };
}

export function familyProductDevEntries() {
  return {
    name: 'mofa-family-dev-entries',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use((request, response, next) => {
        const entry = familyProductEntry(request.url, request.method);
        if (!entry) return next();
        if (entry.redirect) {
          response.writeHead(302, { Location: entry.url });
          response.end();
          return;
        }
        request.url = entry.url;
        next();
      });
    }
  };
}
