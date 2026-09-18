export default function ConnectionStatus({ status }) {
  // status: 'connecting' | 'live' | 'reconnecting'
  if (status === 'live') {
    return (
      <span className="muted" style={{ fontSize: '0.85rem', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
        <span className="live-dot" aria-hidden="true" /> Live
      </span>
    )
  }
  if (status === 'reconnecting') {
    return (
      <span style={{ fontSize: '0.85rem', color: 'var(--danger)' }} role="status">
        Reconnecting…
      </span>
    )
  }
  return (
    <span className="muted" style={{ fontSize: '0.85rem' }} role="status">
      Connecting…
    </span>
  )
}
