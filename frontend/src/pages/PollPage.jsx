import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { pollApi } from '../services/api'
import { connectToPoll } from '../services/ws'
import ResultsBar from '../components/ResultsBar'
import ConnectionStatus from '../components/ConnectionStatus'

const VOTED_KEY_PREFIX = 'pulse_voted_'

export default function PollPage() {
  const { id } = useParams()
  const [poll, setPoll] = useState(null)
  const [results, setResults] = useState(null) // { [optionId]: count }
  const [totalVotes, setTotalVotes] = useState(0)
  const [viewerCount, setViewerCount] = useState(null)
  const [loadError, setLoadError] = useState('')
  const [voteError, setVoteError] = useState('')
  const [selected, setSelected] = useState(null)
  const [hasVoted, setHasVoted] = useState(false)
  const [voting, setVoting] = useState(false)
  const [wsStatus, setWsStatus] = useState('connecting')
  const disconnectRef = useRef(null)

  const votedKey = `${VOTED_KEY_PREFIX}${id}`

  const applyResults = useCallback((payload) => {
    const map = {}
    for (const r of payload.results) map[r.optionId] = r.count
    setResults(map)
    setTotalVotes(payload.totalVotes)
  }, [])

  useEffect(() => {
    setHasVoted(Boolean(localStorage.getItem(votedKey)))

    pollApi
      .get(id)
      .then(setPoll)
      .catch((err) => setLoadError(err.message))

    pollApi
      .results(id)
      .then(applyResults)
      .catch(() => {})

    const disconnect = connectToPoll(id, {
      onOpen: () => setWsStatus('live'),
      onReconnecting: () => setWsStatus('reconnecting'),
      onMessage: (data) => {
        if (data.type === 'vote_update') {
          applyResults(data)
        } else if (data.type === 'viewer_count') {
          setViewerCount(data.count)
        }
      },
    })
    disconnectRef.current = disconnect
    return () => disconnect()
  }, [id, applyResults, votedKey])

  async function handleVote() {
    if (!selected) return
    setVoting(true)
    setVoteError('')
    try {
      const payload = await pollApi.vote(id, selected)
      applyResults(payload)
      localStorage.setItem(votedKey, selected)
      setHasVoted(true)
    } catch (err) {
      if (err.code === 'ALREADY_VOTED') {
        localStorage.setItem(votedKey, selected || '1')
        setHasVoted(true)
      }
      setVoteError(err.message)
    } finally {
      setVoting(false)
    }
  }

  if (loadError) {
    return (
      <div className="container" style={{ maxWidth: 480, paddingTop: 72, textAlign: 'center' }}>
        <div className="card">
          <p style={{ fontSize: '1.05rem', marginBottom: 4 }}>Poll not found</p>
          <p className="muted">{loadError}</p>
        </div>
      </div>
    )
  }

  if (!poll) {
    return (
      <div className="container" style={{ maxWidth: 480, paddingTop: 72 }}>
        <p className="muted" style={{ textAlign: 'center' }}>Loading poll…</p>
      </div>
    )
  }

  const isClosed = poll.status === 'closed'
  const showResults = hasVoted || isClosed

  return (
    <div className="container" style={{ maxWidth: 560, paddingTop: 48, paddingBottom: 72 }}>
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
          <span className={`badge ${isClosed ? 'badge-closed' : 'badge-active'}`}>
            {isClosed ? 'Closed' : 'Active'}
          </span>
          <ConnectionStatus status={wsStatus} />
        </div>

        <h2 style={{ margin: '0 0 8px' }}>{poll.question}</h2>
        {poll.description && <p className="muted" style={{ marginBottom: 24 }}>{poll.description}</p>}

        {!showResults && (
          <fieldset style={{ border: 'none', padding: 0, margin: 0 }}>
            <legend className="sr-only" style={{ position: 'absolute', width: 1, height: 1, overflow: 'hidden' }}>
              Poll options
            </legend>
            {poll.options.map((opt) => (
              <label
                key={opt.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '12px 14px',
                  border: `1px solid ${selected === opt.id ? 'var(--primary)' : 'var(--border)'}`,
                  borderRadius: 'var(--radius-sm)',
                  marginBottom: 10,
                  cursor: 'pointer',
                }}
              >
                <input
                  type="radio"
                  name="option"
                  value={opt.id}
                  checked={selected === opt.id}
                  onChange={() => setSelected(opt.id)}
                />
                {opt.text}
              </label>
            ))}

            {voteError && (
              <p className="error-text" role="alert" style={{ marginBottom: 12 }}>
                {voteError}
              </p>
            )}

            <button
              className="btn btn-primary btn-block"
              onClick={handleVote}
              disabled={!selected || voting}
            >
              {voting ? 'Submitting…' : 'Vote'}
            </button>
          </fieldset>
        )}

        {showResults && (
          <div>
            {hasVoted && !isClosed && (
              <p style={{ color: 'var(--success)', fontWeight: 600, marginBottom: 20 }}>
                ✓ Your vote has been recorded.
              </p>
            )}
            {isClosed && <p className="muted" style={{ marginBottom: 20 }}>This poll is closed.</p>}

            {poll.options.map((opt) => (
              <ResultsBar
                key={opt.id}
                label={opt.text}
                count={results?.[opt.id] ?? 0}
                total={totalVotes}
                highlight={selected === opt.id}
              />
            ))}

            <p className="muted" style={{ marginTop: 16, fontSize: '0.9rem' }}>
              Total votes: {totalVotes}
              {viewerCount != null && <> · {viewerCount} viewing</>}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
