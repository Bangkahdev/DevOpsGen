import { defineConfig } from "cf/config";

export default defineConfig({
  worker: {
    name: "devopsgen-docs",
    compatibilityDate: "2026-09-21",
    observability: {
      enabled: true,
    },
    workersDev: false,
    domains: ["devopsgen.bangkah.tech"],
    assets: {
      notFoundHandling: "single-page-application",
    },
  },
});
