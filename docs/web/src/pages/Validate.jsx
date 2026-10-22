import Shot from '../components/Shot.jsx';
import Code from '../components/Code.jsx';
import { PageHead, Badge } from '../components/Ui.jsx';

export default function Validate({ rel }) {
  return (
    <>
      <PageHead
        kicker="command reference"
        title="devopsgen validate"
        lede="Statically lint DevOpsGen output — Dockerfiles, GitHub Actions workflows and Kubernetes manifests — so problems surface on your laptop instead of pipeline number forty."
      />

      <Shot
        rel={rel}
        name="validate.png"
        alt="devopsgen validate passing with the message internal validation passed and exit code 0"
        caption="Everything checks out — exit code 0."
        feature="shot-feature"
      />

      <Code label="synopsis">{`devopsgen validate [flags]          # validate the current project
devopsgen validate --output-dir ./configs`}</Code>

      <h2>What it checks</h2>
      <div className="cards">
        <div className="card">
          <h3>Docker</h3>
          <ul>
            <li><code>FROM</code> base images are pinned to digest form</li>
            <li>Required install/prune steps match the package manager</li>
            <li>Non-root user declared</li>
          </ul>
        </div>
        <div className="card">
          <h3>GitHub Actions</h3>
          <ul>
            <li>Workflow documents carry required structure and names</li>
            <li>Actions are pinned, config files complete</li>
            <li>Checkout steps reference valid refs</li>
          </ul>
        </div>
        <div className="card">
          <h3>Kubernetes</h3>
          <ul>
            <li>Resources are DNS 1123/1035 rename-clean</li>
            <li>Service ↔ Ingress wiring is consistent</li>
            <li>Probe/health requirements satisfied</li>
          </ul>
        </div>
      </div>

      <h2>Exit behaviour</h2>
      <ul>
        <li>
          <Badge tone="ok">0</Badge> — all documents passed.
        </li>
        <li>
          <Badge tone="err">4</Badge> — at least one document failed validation.
        </li>
        <li>
          <Badge tone="err">2</Badge> — malformed input (unreadable paths, invalid flags).
        </li>
      </ul>
      <p>
        Combine with <code>devopsgen init</code> in CI: generate, commit, validate, deploy — the{' '}
        <a href={`${rel}exit-codes/`}>exit codes</a> page maps every failure class.
      </p>
    </>
  );
}
