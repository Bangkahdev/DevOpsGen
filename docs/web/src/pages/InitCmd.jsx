import Shot from '../components/Shot.jsx';
import Code from '../components/Code.jsx';
import { PageHead, Badge } from '../components/Ui.jsx';

export default function InitCmd({ rel }) {
  return (
    <>
      <PageHead
        kicker="command reference"
        title="devopsgen init"
        lede="Initialize DevOps configuration. In an interactive terminal with no complete configuration the TUI wizard is launched; when all required parameters are supplied via flags the process runs headless and never waits for stdin."
      />

      <Code label="synopsis">{`devopsgen init [flags]`}</Code>

      <h2>Flags</h2>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Flag</th>
              <th>Default</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="flag-name">--stack</td>
              <td>auto-detected</td>
              <td>Project stack: <code>go</code> or <code>nodejs</code>.</td>
            </tr>
            <tr>
              <td className="flag-name">--app-type</td>
              <td>auto-detected</td>
              <td>Application type (see <code>devopsgen list</code>).</td>
            </tr>
            <tr>
              <td className="flag-name">--ci</td>
              <td><code>github-actions</code></td>
              <td>CI provider.</td>
            </tr>
            <tr>
              <td className="flag-name">--registry</td>
              <td><code>dockerhub</code></td>
              <td>Container registry: <code>dockerhub</code> or <code>ghcr</code>.</td>
            </tr>
            <tr>
              <td className="flag-name">--image</td>
              <td>—</td>
              <td>
                Image repository as <code>namespace/repository</code> — no host and no tag.
              </td>
            </tr>
            <tr>
              <td className="flag-name">--tag</td>
              <td><code>latest</code></td>
              <td>Additional image tag alias.</td>
            </tr>
            <tr>
              <td className="flag-name">--port</td>
              <td>—</td>
              <td>Application port, 1–65535.</td>
            </tr>
            <tr>
              <td className="flag-name">--k8s</td>
              <td><code>true</code></td>
              <td>Generate Kubernetes manifests.</td>
            </tr>
            <tr>
              <td className="flag-name">--ingress</td>
              <td>off</td>
              <td>Enable Kubernetes Ingress (requires <code>--ingress-host</code>).</td>
            </tr>
            <tr>
              <td className="flag-name">--ingress-host</td>
              <td>—</td>
              <td>Ingress hostname, validated as RFC 1123.</td>
            </tr>
            <tr>
              <td className="flag-name">--health-path</td>
              <td>—</td>
              <td>HTTP health probe path wired into manifests.</td>
            </tr>
            <tr>
              <td className="flag-name">--build-cmd</td>
              <td>—</td>
              <td>Build command.</td>
            </tr>
            <tr>
              <td className="flag-name">--build-path</td>
              <td><code>.</code></td>
              <td>Go package path to build inside the container.</td>
            </tr>
            <tr>
              <td className="flag-name">--start-cmd</td>
              <td>—</td>
              <td>Startup command.</td>
            </tr>
            <tr>
              <td className="flag-name">--test-cmd</td>
              <td>—</td>
              <td>Test command.</td>
            </tr>
            <tr>
              <td className="flag-name">--output-dir</td>
              <td><code>.</code></td>
              <td>Output directory.</td>
            </tr>
            <tr>
              <td className="flag-name">--dry-run</td>
              <td>off</td>
              <td>Simulate without writing anything.</td>
            </tr>
            <tr>
              <td className="flag-name">--force</td>
              <td>off</td>
              <td>Overwrite existing regular files.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>Success &amp; conflicts</h2>
      <p>
        A clean run writes every template and exits with{' '}
        <Badge tone="ok">0</Badge>. A second run over an existing project stops before touching
        anything and exits with <Badge tone="warn">3</Badge> until you add <code>--force</code>:
      </p>
      <div className="screen-strip">
        <Shot
          rel={rel}
          name="init-success.png"
          alt="devopsgen init success output with exit code 0"
          caption="First run: five files created, exit code 0."
        />
        <Shot
          rel={rel}
          name="init-conflict.png"
          alt="devopsgen init conflict error refusing to overwrite files with exit code 3"
          caption="Re-run without --force: exit code 3, zero files touched."
        />
      </div>

      <h2>Preview plans with --dry-run</h2>
      <p>
        <code>--dry-run</code> prints the byte-plan: which files would be created, which would be
        replaced, and which are in conflict. With a conflict present it exits{' '}
        <Badge tone="warn">3</Badge>; combined with <code>--force</code> a fully green plan exits{' '}
        <Badge tone="ok">0</Badge> without writing anything:
      </p>
      <div className="screen-strip">
        <Shot
          rel={rel}
          name="dry-run.png"
          alt="devopsgen dry-run plan output showing replace and conflict entries with exit code 3"
          caption="--dry-run over an existing project: conflict entries, exit code 3."
        />
        <Shot
          rel={rel}
          name="dry-run-force.png"
          alt="devopsgen dry-run force plan output where every file would be replaced, exit code 0"
          caption="--dry-run --force: every file replaced in plan form, exit code 0."
        />
      </div>

      <h2>Config validation</h2>
      <p>
        Invalid configuration fails fast with <Badge tone="err">2</Badge> before any file is
        planned — e.g. an image that must be <code>namespace/repository</code> without host or
        tag, a port outside 1–65535, or an Ingress without a host:
      </p>
      <Shot
        rel={rel}
        name="invalid-config.png"
        alt="devopsgen init failing config validation on an invalid image with exit code 2"
        caption="Config errors exit with code 2 and never write files."
        feature="shot-feature"
      />
      <div className="callout callout-ok">
        Overwrites performed with <code>--force</code> keep the previous content as{' '}
        <code>&lt;file&gt;.devopsgen.bak</code> backups, and any failure mid-run triggers a
        best-effort rollback to the pre-run state.
      </div>
    </>
  );
}
