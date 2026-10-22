import Home from './pages/Home.jsx';
import GettingStarted from './pages/GettingStarted.jsx';
import Wizard from './pages/Wizard.jsx';
import InitCmd from './pages/InitCmd.jsx';
import Detect from './pages/Detect.jsx';
import Validate from './pages/Validate.jsx';
import ExitCodes from './pages/ExitCodes.jsx';

export const SITE_URL = 'https://bangkahdev.github.io/DevOpsGen/';
export const SITE_NAME = 'DevOpsGen';

export const ROUTES = [
  {
    path: '',
    component: Home,
    title: 'DevOpsGen — DevOps scaffolding for Go & Node.js projects',
    description:
      'DevOpsGen generates production-ready Dockerfiles, GitHub Actions pipelines and Kubernetes manifests for Go and Node.js projects, with safe dry-run plans, automatic backups and config validation.',
    keywords: 'DevOpsGen, dockerfile generator, github actions generator, kubernetes manifests, go, nodejs, scaffolding, ci cd',
    ogImage: 'img/init-success.png',
  },
  {
    path: 'getting-started',
    component: GettingStarted,
    title: 'Getting Started — DevOpsGen Docs',
    description:
      'Install DevOpsGen via go install or a release binary, then generate a complete Docker, CI and Kubernetes setup for your project with a single command.',
    keywords: 'DevOpsGen install, go install, quickstart, getting started',
    ogImage: 'img/init-success.png',
  },
  {
    path: 'wizard',
    component: Wizard,
    title: 'Interactive Wizard — DevOpsGen Docs',
    description:
      'DevOpsGen launches an interactive TUI wizard that walks through stack detection, ports, container images and Kubernetes options, then writes the generated configuration.',
    keywords: 'DevOpsGen wizard, tui, bubbletea, interactive',
    ogImage: 'img/wizard.png',
  },
  {
    path: 'init',
    component: InitCmd,
    title: 'init command — DevOpsGen Docs',
    description:
      'Full reference for devopsgen init flags: stack, app type, CI provider, registry, image, port, Kubernetes ingress, health path, dry-run and force overwrite behaviour.',
    keywords: 'devopsgen init, flags, dry-run, force, image, ingress',
    ogImage: 'img/init-success.png',
  },
  {
    path: 'detect',
    component: Detect,
    title: 'detect command — DevOpsGen Docs',
    description:
      'devopsgen detect inspects the current project and reports the programming stack, the files it based its decision on, and candidate application types.',
    keywords: 'devopsgen detect, stack detection, go.mod, package.json',
    ogImage: 'img/detect.png',
  },
  {
    path: 'validate',
    component: Validate,
    title: 'validate command — DevOpsGen Docs',
    description:
      'devopsgen validate statically checks generated Dockerfiles, GitHub Actions workflows and Kubernetes manifests: FROM digests, action pinning, naming rules, probes and more.',
    keywords: 'devopsgen validate, dockerfile lint, workflow lint, kubernetes manifest lint',
    ogImage: 'img/validate.png',
  },
  {
    path: 'exit-codes',
    component: ExitCodes,
    title: 'Exit codes — DevOpsGen Docs',
    description:
      'Every devopsgen command exits with a documented status code: 0 success, 1 general error, 2 invalid configuration, 3 file conflicts, 4 validation failure, 5 ambiguous detection.',
    keywords: 'devopsgen exit codes, ci integration, pipeline status',
    ogImage: 'img/invalid-config.png',
  },
];
