export default function ResultsBar({ label, count, total, highlight }) {
  const pct = total > 0 ? Math.round((count / total) * 1000) / 10 : 0

  return (
    <div style={{ marginBottom: 14 }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginBottom: 6,
          fontSize: '0.92rem',
        }}
      >
        <span style={{ fontWeight: highlight ? 700 : 500 }}>{label}</span>
        <span className="muted">
          {count} vote{count === 1 ? '' : 's'} · {pct}%
        </span>
      </div>
      <div
        style={{
          height: 10,
          borderRadius: 999,
          background: 'var(--surface-hover)',
          overflow: 'hidden',
          border: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            height: '100%',
            width: `${pct}%`,
            background: highlight ? 'var(--primary)' : 'var(--text-muted)',
            borderRadius: 999,
            transition: 'width 0.5s cubic-bezier(0.22, 1, 0.36, 1)',
          }}
        />
      </div>
    </div>
  )
}
