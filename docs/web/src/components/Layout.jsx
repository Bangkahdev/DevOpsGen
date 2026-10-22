import { ROUTES, SITE_URL } from '../routes.jsx';

const GROUPS = [
  { label: 'Overview', items: ['getting-started', 'wizard'] },
  { label: 'Command reference', items: ['init', 'detect', 'validate'] },
  { label: 'Operations', items: ['exit-codes'] },
];

export default function Layout({ rel, activePath, children }) {
  const homeActive = activePath === '';
  return (
    <>
      <a className="skip" href="#main">skip to content</a>
      <header className="topbar">
        <a className="brand" href={rel || './'}>
          <img className="brand-mark" src={`${rel}logo.svg`} width="36" height="36" alt="" />
          <span className="brand-name">DevOpsGen</span>
          <span className="brand-sub">DOCS</span>
        </a>
        <a className="gh-pill" href="https://github.com/Bangkahdev/DevOpsGen" rel="noopener noreferrer" target="_blank">
          GitHub
        </a>
      </header>

      <div className="shell">
        <aside className="sidebar" aria-label="Documentation navigation">
          <a className={`side-link side-home${homeActive ? ' is-active' : ''}`} href={rel || './'}>
            Home
          </a>
          {GROUPS.map((g) => (
            <div className="side-group" key={g.label}>
              <p className="group-label">{g.label}</p>
              <ul>
                {g.items.map((p) => {
                  const r = ROUTES.find((x) => x.path === p);
                  return (
                    <li key={p}>
                      <a
                        className={`side-link${activePath === p ? ' is-active' : ''}`}
                        href={`${rel}${p}/`}
                        aria-current={activePath === p ? 'page' : undefined}
                      >
                        {r ? r.path.replace(/-/g, ' ') : p}
                      </a>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </aside>

        <main id="main">{children}</main>
      </div>

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
