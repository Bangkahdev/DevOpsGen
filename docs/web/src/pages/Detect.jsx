import Shot from '../components/Shot.jsx';
import { PageHead } from '../components/Ui.jsx';

export default function Detect({ rel }) {
  return (
    <>
      <PageHead
        kicker="command reference"
        title="devopsgen detect"
        lede="Inspect the current project and report what DevOpsGen would scaffold for it: the programming stack, the exact files that shaped the decision, and the candidate application types."
      />

      <Shot
        rel={rel}
        name="detect.png"
        alt="devopsgen detect output in a terminal reporting stack go with indicators go.mod"
        caption="devopsgen detect inside a Go module — chain of evidence straight from the filesystem."
        feature="shot-feature"
      />

      <h2>Reading the output</h2>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Line</th>
              <th>Meaning</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>stack: go</code></td>
              <td>The detected programming stack (<code>go</code> or <code>nodejs</code>).</td>
            </tr>
            <tr>
              <td><code>indicators: go.mod, ...</code></td>
              <td>The files that produced evidence — your audit trail for why the stack matched.</td>
            </tr>
            <tr>
              <td><code>app-type candidates: binary</code></td>
              <td>Application types plausible for this project; the wizard default comes from here too.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>How it decides</h2>
      <ul>
        <li>
          Go — <code>go.mod</code> and its module directive looking like{' '}
          <code>module example.com/app</code>.
        </li>
        <li>
          Node.js — <code>package.json</code>, plus its dependencies, and lockfile flavours
          (npm / pnpm) that later steer Docker image commands.
        </li>
        <li>
          Ambiguity or zero indicators produce a distinct exit code (see{' '}
          <a href={`${rel}exit-codes/`}>exit codes</a>), so scripts can branch safely.
        </li>
      </ul>

      <div className="callout">
        Detection semantics: DevOpsGen detects in the current working directory, not in{' '}
        <code>--output-dir</code>. Run it from the project you want scanned.
      </div>
    </>
  );
}
