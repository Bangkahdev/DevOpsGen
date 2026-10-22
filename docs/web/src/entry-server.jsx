import { renderToStaticMarkup } from 'react-dom/server';
import React from 'react';
import { ROUTES, SITE_URL, SITE_NAME } from './routes.jsx';
import Layout from './components/Layout.jsx';

export { ROUTES, SITE_URL, SITE_NAME };

const escapeHtml = (s) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export function relPrefix(path) {
  return path === '' ? '' : '../';
}

export function pageUrl(path) {
  return `${SITE_URL}${path === '' ? '' : path + '/'}`;
}

export function renderDocument(route) {
  const rel = relPrefix(route.path);
  const url = pageUrl(route.path);
  const image = `${SITE_URL}${route.ogImage}`;
  const body = renderToStaticMarkup(
    <Layout rel={rel} activePath={route.path}>
      <route.component rel={rel} />
    </Layout>,
  );
  const r = route.path === '' ? [{ '@type': 'ListItem', position: 1, name: 'Home', item: url }] : [
    { '@type': 'ListItem', position: 1, name: 'Home', item: SITE_URL },
    {
      '@type': 'ListItem',
      position: 2,
      name: route.path.replace(/-/g, ' '),
      item: url,
    },
  ];
  const jsonLd = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'SoftwareApplication',
        name: 'DevOpsGen',
        applicationCategory: 'DeveloperApplication',
        operatingSystem: 'Linux, Windows, macOS',
        description: route.description,
        keywords: route.keywords,
        url: SITE_URL,
        image: [`${SITE_URL}img/og.png`, image],
        logo: { '@type': 'ImageObject', url: `${SITE_URL}logo.svg` },
        author: { '@type': 'Person', name: 'Bangkah', url: 'https://github.com/Bangkahdev' },
        downloadUrl: 'https://github.com/Bangkahdev/DevOpsGen/releases',
        license: 'https://github.com/Bangkahdev/DevOpsGen#readme',
        offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
      },
      { '@type': 'BreadcrumbList', itemListElement: r },
    ],
  };

  const ogSizeMeta = route.ogImage === 'img/og.png'
    ? `  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
`
    : '';

  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>${escapeHtml(route.title)}</title>
  <meta name="description" content="${escapeHtml(route.description)}" />
  <meta name="keywords" content="${escapeHtml(route.keywords)}" />
  <meta name="robots" content="index, follow" />
  <meta name="author" content="Bangkah" />
  <meta name="theme-color" content="#f3ede3" />
  <link rel="icon" type="image/svg+xml" href="${rel}favicon.svg" />
  <link rel="apple-touch-icon" href="${rel}apple-touch-icon.png" />
  <link rel="canonical" href="${url}" />
  <meta property="og:type" content="website" />
  <meta property="og:site_name" content="${escapeHtml(SITE_NAME)}" />
  <meta property="og:locale" content="en_US" />
  <meta property="og:url" content="${url}" />
  <meta property="og:title" content="${escapeHtml(route.title)}" />
  <meta property="og:description" content="${escapeHtml(route.description)}" />
  <meta property="og:image" content="${image}" />
  <meta property="og:image:alt" content="${escapeHtml(SITE_NAME)} — DevOps scaffolding for Go and Node.js projects" />
${ogSizeMeta}  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="${escapeHtml(route.title)}" />
  <meta name="twitter:description" content="${escapeHtml(route.description)}" />
  <meta name="twitter:image" content="${image}" />
  <meta name="twitter:image:alt" content="${escapeHtml(SITE_NAME)} — DevOps scaffolding for Go and Node.js projects" />
  <link rel="stylesheet" crossorigin href="${rel}assets/site.css" />
  <script type="application/ld+json">${JSON.stringify(jsonLd)}</script>
</head>
<body>
${body}
</body>
</html>`;
}
