import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { pollApi } from '../services/api'

export default function Dashboard() {
  const [polls, setPolls] = useState(null)
  const [error, setError] = useState('')
  const [copiedId, setCopiedId] = useState(null)

  useEffect(() => {
    pollApi
      .mine()
      .then(setPolls)
      .catch((err) => setError(err.message))
  }, [])

  async function handleClose(id) {
    try {
      await pollApi.close(id)
      setPolls((prev) => prev.map((p) => (p.publicId === id ? { ...p, status: 'closed' } : p)))
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleDelete(id) {
    if (!confirm('Delete this poll permanently? This cannot be undone.')) return
    try {
      await pollApi.remove(id)
      setPolls((prev) => prev.filter((p) => p.publicId !== id))
    } catch (err) {
      setError(err.message)
    }
  }

  function copyLink(id) {
    const url = `${window.location.origin}/poll/${id}`
    navigator.clipboard.writeText(url)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 1500)
  }

  const active = polls?.filter((p) => p.status === 'active') ?? []
  const closed = polls?.filter((p) => p.status === 'closed') ?? []

  return (
    <div style={{ position: 'relative', minHeight: '100%', overflow: 'hidden' }}>
      <video
        autoPlay
        muted
        loop
        playsInline
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          objectFit: 'cover',
          zIndex: -2,
        }}
      >
        <source src="/videos/bg.mp4" type="video/mp4" />
      </video>
      <div
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          background: 'rgba(15, 15, 25, 0.6)',
          zIndex: -1,
        }}
      />

      <div className="container" style={{ paddingTop: 48, paddingBottom: 72, position: 'relative', zIndex: 1 }}>
        <div className="page-header">
          <h1 style={{ color: '#ffffff' }}>My polls</h1>
          <Link to="/dashboard/create" className="btn btn-primary">
            + Create poll
          </Link>
        </div>

        {error && <p className="error-text">{error}</p>}

        {polls === null && !error && <p style={{ color: 'rgba(255,255,255,0.85)' }}>Loading your polls…</p>}

        {polls && polls.length === 0 && (
          <div className="card" style={{ textAlign: 'center', padding: 48 }}>
            <p style={{ fontSize: '1.05rem', marginBottom: 8 }}>No polls yet</p>
            <p className="muted" style={{ marginBottom: 24 }}>
              Create your first poll and share the link with your audience.
            </p>
            <Link to="/dashboard/create" className="btn btn-primary">
              Create a poll
            </Link>
          </div>
        )}

        {active.length > 0 && (
          <>
            <h3 style={{ fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.04em', color: 'rgba(255,255,255,0.75)' }}>
              Active
            </h3>
            <div style={{ display: 'grid', gap: 16, marginBottom: 32 }}>
              {active.map((p) => (
                <PollRow
                  key={p.publicId}
                  poll={p}
                  copied={copiedId === p.publicId}
                  onCopy={() => copyLink(p.publicId)}
                  onClose={() => handleClose(p.publicId)}
                  onDelete={() => handleDelete(p.publicId)}
                />
              ))}
            </div>
          </>
        )}

        {closed.length > 0 && (
          <>
            <h3 style={{ fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.04em', color: 'rgba(255,255,255,0.75)' }}>
              Closed
            </h3>
            <div style={{ display: 'grid', gap: 16 }}>
              {closed.map((p) => (
                <PollRow
                  key={p.publicId}
                  poll={p}
                  copied={copiedId === p.publicId}
                  onCopy={() => copyLink(p.publicId)}
                  onDelete={() => handleDelete(p.publicId)}
                />
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

function PollRow({ poll, copied, onCopy, onClose, onDelete }) {
  return (
    <div className="card" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
      <div style={{ minWidth: 0 }}>
        <span className={`badge ${poll.status === 'active' ? 'badge-active' : 'badge-closed'}`}>
          {poll.status}
        </span>
        <h3 style={{ margin: '10px 0 4px', overflow: 'hidden', textOverflow: 'ellipsis' }}>{poll.question}</h3>
        <p className="muted" style={{ margin: 0, fontSize: '0.85rem' }}>
          Created {new Date(poll.createdAt).toLocaleDateString()}
        </p>
      </div>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <Link to={`/poll/${poll.publicId}`} className="btn btn-secondary">
          View
        </Link>
        <button className="btn btn-secondary" onClick={onCopy}>
          {copied ? 'Copied!' : 'Copy link'}
        </button>
        {poll.status === 'active' && onClose && (
          <button className="btn btn-secondary" onClick={onClose}>
            Close
          </button>
        )}
        <button className="btn btn-danger" onClick={onDelete}>
          Delete
        </button>
      </div>
    </div>
  )
}