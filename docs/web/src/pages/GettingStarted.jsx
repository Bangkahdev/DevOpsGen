import Shot from '../components/Shot.jsx';
import Code from '../components/Code.jsx';
import { PageHead } from '../components/Ui.jsx';

export default function GettingStarted({ rel }) {
  return (
    <>
      <PageHead
        kicker="getting started"
        title="Zero to CI in one command"
        lede="Install DevOpsGen, point it at any Go or Node.js project, and get a full Docker + GitHub Actions + Kubernetes baseline in seconds."
      />

      <h2>Install</h2>
      <Code label="go install (go 1.22+)">{`go install devopsgen.dev/devopsgen@latest`}</Code>
      <p>
        Or grab a prebuilt binary from the project{' '}
        <a href="https://github.com/Bangkahdev/DevOpsGen/releases" target="_blank" rel="noopener noreferrer">
          GitHub releases
        </a>
        , or build from source:
      </p>
      <Code label="from source">{`git clone https://github.com/Bangkahdev/DevOpsGen.git
cd DevOpsGen
make build`}</Code>

      <h2>Two ways to run it</h2>
      <ul className="step-list">
        <li>
          <strong>Interactive:</strong> run <code>devopsgen init</code> in a terminal — the wizard
          detects your stack and interviews you step by step.
        </li>
        <li>
          <strong>Headless:</strong> pass every required parameter as flags — the process never
          waits for stdin, which makes it script- and CI-friendly.
        </li>
      </ul>

      <h2>Headless quickstart</h2>
      <Code label="inside your project">{`devopsgen init \
  --stack go \
  --ci github-actions \
  --registry ghcr \
  --image you/myapp \
  --port 8080`}</Code>
      <Shot
        rel={rel}
        name="init-success.png"
        alt="Terminal capture of devopsgen init completing successfully and reporting five files created"
        caption="A successful first run — five files created, exit code 0."
        feature="shot-feature"
      />
      <p>Inside a Go project this writes:</p>
      <Code label="what you get">{`.github/workflows/ci.yml   # tests (when configured), docker build & push
kubernetes/deployment.yaml  # image, port, resources, optional probes
kubernetes/service.yaml     # cluster service for the app
kubernetes/ingress.yaml     # only with --ingress
Dockerfile                  # hardened multi-stage build
.dockerignore               # slim build context, secrets excluded`}</Code>
      <div className="callout">
        Second runs are safe: DevOpsGen refuses to overwrite without <code>--force</code> and
        exits with code 3 — see <a href={`${rel}init/`}>the init reference</a>.
      </div>

      <h2>Next steps</h2>
      <ul>
        <li>
          Prefer point-and-press? Tour the <a href={`${rel}wizard/`}>interactive wizard</a>.
        </li>
        <li>
          Every <code>init</code> flag explained: <a href={`${rel}init/`}>init command</a>.
        </li>
        <li>
          Building your own CI around it? Read the <a href={`${rel}exit-codes/`}>exit codes</a>.
        </li>
      </ul>
    </>
  );
}
