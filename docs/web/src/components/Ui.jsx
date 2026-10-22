export function Kbd({ children }) {
  return <kbd className="kbd">{children}</kbd>;
}

export function Badge({ tone = 'ok', children }) {
  return <span className={`badge badge-${tone}`}>{children}</span>;
}

export function PageHead({ kicker, title, lede }) {
  return (
    <header className="page-head">
      {kicker && <p className="kicker">{kicker}</p>}
      <h1>{title}</h1>
      {lede && <p className="lede">{lede}</p>}
    </header>
  );
}
