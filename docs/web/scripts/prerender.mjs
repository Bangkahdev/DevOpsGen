import { mkdir, writeFile, copyFile, readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '..');

const { ROUTES, SITE_URL, renderDocument } = await import(
  pathToFileURL(path.join(root, '.ssr', 'entry-server.js')).href
);

const outDir = path.resolve(root, process.env.DG_OUT_DIR || 'dist');

for (const route of ROUTES) {
  const html = renderDocument(route);
  if (route.path === '') {
    await writeFile(path.join(outDir, 'index.html'), html, 'utf8');
  } else {
    const dir = path.join(outDir, route.path);
    await mkdir(dir, { recursive: true });
    await writeFile(path.join(dir, 'index.html'), html, 'utf8');
  }
  console.log(`prerendered: ${route.path === '' ? '/ (home)' : '/' + route.path + '/'}`);
}

// lastmod = build time (freshness signal)
const buildDate = new Date().toISOString().slice(0, 10);
const sitemapEntries = ROUTES.map((route) => {
  const url = route.path === '' ? SITE_URL : SITE_URL + route.path + '/';
  const priority = route.path === '' ? '1.0' : '0.8';
  const freq = route.path === '' ? 'weekly' : 'monthly';
  return `  <url><loc>${url}</loc><lastmod>${buildDate}</lastmod><changefreq>${freq}</changefreq><priority>${priority}</priority></url>`;
}).join('\n');

await writeFile(
  path.join(outDir, 'sitemap.xml'),
  `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${sitemapEntries}
</urlset>
`,
  'utf8',
);

await writeFile(
  path.join(outDir, 'robots.txt'),
  `User-agent: *
Allow: /

Sitemap: ${SITE_URL}sitemap.xml
`,
  'utf8',
);

// Final pages are JS-free (content fully prerendered); the compiled CSS from the
// client build is copied to a stable path so every page references it relatively.
const assetsDir = path.join(outDir, 'assets');
const assetEntries = await readdir(assetsDir);
const css = assetEntries.find((f) => f.endsWith('.css'));
if (css) {
  await copyFile(path.join(assetsDir, css), path.join(assetsDir, 'site.css'));
  console.log('site.css ready (from ' + css + ')');
} else {
  throw new Error('no css asset emitted by vite build');
}

console.log('sitemap + robots written');
