import Shot from '../components/Shot.jsx';
import Code from '../components/Code.jsx';
import { PageHead } from '../components/Ui.jsx';

export default function ExitCodes({ rel }) {
  return (
    <>
      <PageHead
        kicker="ci integration"
        title="Exit codes"
        lede="Every DevOpsGen command exits with a stable, documented status code so shells, scripts and CI jobs can branch deterministically — no output parsing required."
      />

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Code</th>
              <th>Classification</th>
              <th>Meaning</th>
              <th>Typical example</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="num">0</td>
              <td><span className="badge badge-ok">success</span></td>
              <td>Everything requested completed; dry-run plans that resolved cleanly also end here.</td>
              <td><code>init</code> created 5 files; <code>--dry-run --force</code> plan is green.</td>
            </tr>
            <tr>
              <td className="num">1</td>
              <td><span className="badge badge-err">general error</span></td>
              <td>
                Unexpected failure — I/O errors, template issues, or a run that failed mid-write
                and rolled back.
              </td>
              <td>Disk full while writing <code>deployment.yaml</code>.</td>
            </tr>
            <tr>
              <td className="num">2</td>
              <td><span className="badge badge-err">invalid input</span></td>
              <td>Configuration failed validation before anything was planned or written.</td>
              <td><code>--image BANGKAH/my-app</code>; port <code>-1</code>; Ingress without host.</td>
            </tr>
            <tr>
              <td className="num">3</td>
              <td><span className="badge badge-warn">conflict</span></td>
              <td>
                Target files already exist and would be overwritten; nothing was written.
                Retried with <code>--force</code>.
              </td>
              <td>Second <code>init</code> without <code>--force</code>.</td>
            </tr>
            <tr>
              <td className="num">4</td>
              <td><span className="badge badge-err">validation failed</span></td>
              <td>
                Rejected output: a template rendering failure, required artifacts missing, or
                generated documents failing the static lint rules.
              </td>
              <td>A workflow that never publishes the commit SHA tag.</td>
            </tr>
            <tr>
              <td className="num">5</td>
              <td><span className="badge badge-warn">detection</span></td>
              <td>Stack detection was ambiguous or found nothing at all.</td>
              <td>Empty directory without <code>go.mod</code> / <code>package.json</code>.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>In practice</h2>
      <div className="screen-strip">
        <Shot
          rel={rel}
          name="invalid-config.png"
          alt="devopsgen exiting with code 2 after failing image validation"
          caption="Exit 2 — config rejected before planning."
        />
        <Shot
          rel={rel}
          name="init-conflict.png"
          alt="devopsgen exiting with code 3 on existing files"
          caption="Exit 3 — conflict, nothing written."
        />
      </div>

      <Code label="shell script branching">{`devopsgen init --stack go --registry ghcr --image you/myapp --port 8080
case $? in
  0) echo "ready" ;;
  3) devopsgen init --force --stack go --registry ghcr --image you/myapp --port 8080 ;;
  2) echo "fix your flags" >&2 ;;
  *) echo "see devopsgen validate output" >&2 ;;
esac`}</Code>

      <div className="callout callout-info">
        Codes are part of DevOpsGen's compatibility promise: bumping a major version may add new
        codes, but existing ones keep their meaning.
      </div>
    </>
  );
}
