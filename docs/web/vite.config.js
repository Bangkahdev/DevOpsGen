import { cloudflare } from '@cloudflare/vite-plugin';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { spawnSync } from 'node:child_process';
import path from 'node:path';

const prerenderer = () => ({
  name: 'devopsgen-prerenderer',
  apply: 'build',
  closeBundle() {
    if (process.env.DG_CHILD === '1') return;
    const targets = [path.resolve(process.cwd(), 'dist')];
    if (process.env.CLOUDFLARE_VITE_FORCE_BUILD_OUTPUT) {
      targets.push(
        path.resolve(process.cwd(), '.cloudflare/output/v0/workers/default/assets'),
      );
    }
    for (const outDir of targets) {
      const env = { ...process.env, DG_CHILD: '1', DG_OUT_DIR: outDir };
      spawnSync(
        process.execPath,
        ['node_modules/vite/bin/vite.js', 'build', '--ssr', 'src/entry-server.jsx', '--outDir', '.ssr'],
        { env, stdio: 'inherit' },
      );
      spawnSync(process.execPath, ['scripts/prerender.mjs'], { env, stdio: 'inherit' });
    }
  },
});

export default defineConfig({
  base: './',
  plugins:
    process.env.DG_CHILD === '1'
      ? [react(), prerenderer()]
      : [cloudflare(), react(), prerenderer()],
  build: {
    outDir: 'dist',
  },
});
