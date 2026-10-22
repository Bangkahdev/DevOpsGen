import { ROUTES, SITE_URL } from '../routes.jsx';

export default function Layout({ rel, children }) {
  return (
    <>
      <a className="skip" href="#main">skip to content</a>
      <header className="topbar">
        <a className="brand" href={rel || './'}>
          <span className="brand-mark">DG</span>
          <span>DevOpsGen<span className="brand-sub">DOCS</span></span>
        </a>
        <nav aria-label="Main navigation">
          <ul>
            {ROUTES.filter((r) => r.path !== '').map((r) => (
              <li key={r.path}>
                <a
                  className={rel === '' ? 'nav-link' : 'nav-link nav-sub'}
                  href={`${rel}${r.path}/`}
                >
                  {r.path.replace(/-/g, ' ')}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <a className="gh-pill" href="https://github.com/Bangkahdev/DevOpsGen" rel="noopener noreferrer" target="_blank">
          GitHub
        </a>
      </header>
      <main id="main">{children}</main>
      <footer className="footer">
        <p>
          DevOpsGen — deterministic DevOps scaffolding for Go &amp; Node.js.{' '}
          <a href={`${rel}getting-started/`}>Get started</a> or browse the{' '}
          <a href={`${rel}exit-codes/`}>exit codes</a>.
        </p>
        <p className="footer-links">
          <a href={SITE_URL}>canonical site</a> · <a href="https://github.com/Bangkahdev/DevOpsGen">source code</a>
        </p>
      </footer>
    </>
  );
}
