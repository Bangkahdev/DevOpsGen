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
            <li><code>FROM</code> instruction is present</li>
            <li>no <code>docker login</code> (credentials) inside image builds</li>
            <li><code>COPY</code>/<code>ADD</code> free of shell chaining (<code>&amp;&amp;</code>, <code>||</code>, <code>2&gt;/dev/null</code>)</li>
          </ul>
        </div>
        <div className="card">
          <h3>GitHub Actions</h3>
          <ul>
            <li>parses as a single workflow document with <code>name</code>, <code>on</code>, <code>jobs</code> and steps</li>
            <li>a <code>actions/checkout</code> step exists</li>
            <li>the image push references the commit SHA tag</li>
            <li>no unrendered template actions leak into the output</li>
          </ul>
        </div>
        <div className="card">
          <h3>Kubernetes</h3>
          <ul>
            <li>Deployment: selector ↔ template label consistency, unique container names, integer <code>containerPort</code> in 1–65535</li>
            <li>Service: selector plus <code>port</code>/<code>targetPort</code> pairs</li>
            <li>Ingress: <code>networking.k8s.io/v1</code>, valid <code>pathType</code>, wired <code>backend.service</code></li>
            <li>Service ↔ Ingress backends stay consistent</li>
          </ul>
        </div>
      </div>

      <h2>Exit behaviour</h2>
      <ul>
        <li>
          <Badge tone="ok">0</Badge> — all documents passed.
        </li>
        <li>
          <Badge tone="err">4</Badge> — a document failed lint rules, does not parse, or required artifacts are missing.
        </li>
        <li>
          <Badge tone="warn">1</Badge> — the output directory itself could not be read.
        </li>
        <li>
          <Badge tone="err">2</Badge> — malformed input (invalid flags).
        </li>
      </ul>
      <p>
        Combine with <code>devopsgen init</code> in CI: generate, commit, validate, deploy — the{' '}
        <a href={`${rel}exit-codes/`}>exit codes</a> page maps every failure class.
      </p>
      <p>
        Combine with <code>devopsgen init</code> in CI: generate, commit, validate, deploy — the{' '}
        <a href={`${rel}exit-codes/`}>exit codes</a> page maps every failure class.
      </p>
    </>
  );
}
