import { ROUTES, SITE_NAME } from '../routes.jsx';
import Shot from '../components/Shot.jsx';
import Code from '../components/Code.jsx';
import { PageHead } from '../components/Ui.jsx';

export default function Home({ rel }) {
  return (
    <>
      <div className="hero">
        <div>
          <p className="kicker yellow">docker · github actions · kubernetes</p>
          <h1>
            DevOps scaffolding,
            <br />
            deterministically.
          </h1>
          <p className="lede">
            DevOpsGen reads your Go or Node.js project and writes a complete, opinionated DevOps
            baseline: hardened Dockerfile, GitHub Actions pipeline and ready-to-apply Kubernetes
            manifests — with dry-run plans, automatic backups and config validation built in.
          </p>
          <div className="hero-cta">
            <a className="btn" href={`${rel}getting-started/`}>
              Get started
            </a>
            <a className="btn btn-pink" href={`${rel}wizard/`}>
              Take the wizard tour
            </a>
          </div>
        </div>
        <Shot
          rel={rel}
          name="wizard.png"
          alt="DevOpsGen interactive wizard running in a Windows terminal, asking for the project stack"
          caption="The interactive wizard — press Enter to accept smart defaults."
          feature="shot-feature"
        />
      </div>

      <h2>Why DevOpsGen</h2>
      <div className="cards">
        <div className="card">
          <h3>Deterministic</h3>
          <p>
            Same inputs, same bytes — every run. Machine-friendly for code review: a second
            generation with <code>--force</code> produces bit-identical files.
          </p>
        </div>
        <div className="card">
          <h3>Safe by default</h3>
          <p>
            Existing files are never overwritten without <code>--force</code>. Overwrites get{' '}
            <code>.bak</code> backups, failed runs roll back automatically, and{' '}
            <code>--dry-run</code> prints the full plan first.
          </p>
        </div>
        <div className="card">
          <h3>Validated output</h3>
          <p>
            Built-in validation checks Dockerfile <code>FROM</code> digests, workflow action
            pinning, Kubernetes naming and health probes before anything ships.
          </p>
        </div>
        <div className="card">
          <h3>Stack aware</h3>
          <p>
            Detects Go and Node.js projects (npm, pnpm, Express, Next.js, Vite, NestJS) from{' '}
            <code>go.mod</code> / <code>package.json</code> and tailors templates accordingly.
          </p>
        </div>
      </div>

      <h2>Everything a project needs on day one</h2>
      <p>
        One <code>init</code> command writes the whole baseline tree. Preview the real terminal
        output of every command below — all screenshots are live captures, not mockups.
      </p>
      <Code label="generated in one command">{`$ devopsgen init --stack go --ci github-actions --registry ghcr --image you/myapp --port 8080
created 5 file(s), replaced 0 file(s) in .
exit code: 0`}</Code>
      <div className="screen-strip">
        <Shot
          rel={rel}
          name="init-success.png"
          alt="Terminal showing a successful DevOpsGen init run creating five files with exit code 0"
          caption="init — first run creates the complete baseline and exits 0."
        />
        <Shot
          rel={rel}
          name="init-conflict.png"
          alt="Terminal showing DevOpsGen refusing to overwrite five files and exiting with code 3"
          caption="init on a second run refuses to overwrite files — exit code 3 without --force."
        />
        <Shot
          rel={rel}
          name="dry-run.png"
          alt="Terminal showing a DevOpsGen dry-run plan listing replace and conflict entries before exiting with code 3"
          caption="--dry-run prints the exact write plan, unresolved conflicts included."
        />
      </div>

      <h2>What gets generated</h2>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Template set</th>
              <th>Files</th>
              <th>Purpose</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>docker</code> (+ <code>.dockerignore</code>)</td>
              <td><code>Dockerfile</code></td>
              <td>
                Hardened image build with non-root user, health check hooks, install/prune
                commands derived from package manager (npm / pnpm) and lockfile detection.
              </td>
            </tr>
            <tr>
              <td><code>github-actions</code></td>
              <td><code>.github/workflows/ci.yaml</code></td>
              <td>
                Test/build across Go and Node.js versions, docker build &amp; push, coverage
                floors, manifest validation.
              </td>
            </tr>
            <tr>
              <td><code>kubernetes</code></td>
              <td><code>deployment.yaml</code>, <code>service.yaml</code>, optional <code>ingress.yaml</code></td>
              <td>
                Non-root security context, probes using your health path, RFC 1123-validated
                names, optional host-based Ingress.
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>Explore the docs</h2>
      <div className="cards core">
        {ROUTES.filter((r) => r.path !== '').map((r) => (
          <a className="card" key={r.path} style={{ display: 'block' }} href={`${rel}${r.path}/`}>
            <h3>{r.path.replace(/-/g, ' ')}</h3>
            <p>{r.description}</p>
          </a>
        ))}
      </div>

      <h2>Command cheatsheet</h2>
      <div className="screen-strip">
        <Shot
          rel={rel}
          name="help.png"
          alt="Terminal showing the full devopsgen help output listing all commands"
          caption="devopsgen --help — all commands at a glance."
        />
        <Shot
          rel={rel}
          name="list.png"
          alt="Terminal showing the devopsgen list output with available stacks, app types and registries"
          caption="devopsgen list — supported stacks, app types and registries."
        />
        <Shot
          rel={rel}
          name="version.png"
          alt="Terminal showing the devopsgen version output"
          caption="devopsgen version."
        />
      </div>
    </>
  );
}
