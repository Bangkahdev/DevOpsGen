export default function Code({ children, label }) {
  return (
    <div className="code-block">
      {label && <span className="code-label">{label}</span>}
      <pre>
        <code>{children}</code>
      </pre>
    </div>
  );
}
