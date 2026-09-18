import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { pollApi } from '../services/api'

const CREATE_BG =
  'https://media.istockphoto.com/id/2163061817/vector/voting-election-ballot-box-vote-security-voting-rights-concept.jpg?s=612x612&w=0&k=20&c=X62l3ThDz3ExSDFTmLfh-CfwB9npMbI3tqv0MhrDH68='

export default function CreatePoll() {
  const navigate = useNavigate()
  const [question, setQuestion] = useState('')
  const [description, setDescription] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [expiresAt, setExpiresAt] = useState('')
  const [anonymousVoting, setAnonymousVoting] = useState(true)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [created, setCreated] = useState(null)
  const [copied, setCopied] = useState(false)

  function updateOption(i, value) {
    setOptions((prev) => prev.map((o, idx) => (idx === i ? value : o)))
  }

  function addOption() {
    if (options.length >= 10) return
    setOptions((prev) => [...prev, ''])
  }

  function removeOption(i) {
    if (options.length <= 2) return
    setOptions((prev) => prev.filter((_, idx) => idx !== i))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    const cleanOptions = options.map((o) => o.trim()).filter(Boolean)
    if (!question.trim()) {
      setError('A question is required')
      return
    }
    if (cleanOptions.length < 2) {
      setError('Add at least 2 options')
      return
    }

    setSubmitting(true)
    try {
      const poll = await pollApi.create({
        question: question.trim(),
        description: description.trim(),
        options: cleanOptions,
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : null,
        anonymousVoting,
      })
      setCreated(poll)
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  const pageBgStyle = {
    backgroundImage: `linear-gradient(rgba(20, 15, 50, 0.72), rgba(20, 15, 50, 0.72)), url(${CREATE_BG})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
    minHeight: '100%',
  }

  if (created) {
    const shareUrl = `${window.location.origin}/poll/${created.publicId}`
    return (
      <div style={pageBgStyle}>
        <div className="container" style={{ maxWidth: 520, paddingTop: 72, paddingBottom: 72 }}>
          <div className="card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '2rem', marginBottom: 8 }}>🎉</div>
            <h2 style={{ margin: '0 0 8px' }}>Poll created successfully!</h2>
            <p className="muted" style={{ marginBottom: 24 }}>Share your poll:</p>

            <div
              className="field"
              style={{ flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 24 }}
            >
              <input className="input" readOnly value={shareUrl} onFocus={(e) => e.target.select()} />
              <button
                className="btn btn-secondary"
                onClick={() => {
                  navigator.clipboard.writeText(shareUrl)
                  setCopied(true)
                  setTimeout(() => setCopied(false), 1500)
                }}
              >
                {copied ? 'Copied!' : 'Copy link'}
              </button>
            </div>

            <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
              <button className="btn btn-primary" onClick={() => navigate(`/poll/${created.publicId}`)}>
                Open poll
              </button>
              <button className="btn btn-secondary" onClick={() => navigate('/dashboard')}>
                Back to dashboard
              </button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div style={pageBgStyle}>
      <div className="container" style={{ maxWidth: 560, paddingTop: 48, paddingBottom: 72 }}>
        <h1 style={{ color: '#ffffff' }}>Create a poll</h1>
        <form onSubmit={handleSubmit} noValidate className="card">
          <div className="field">
            <label htmlFor="question">Question</label>
            <input
              id="question"
              className="input"
              required
              maxLength={300}
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="Which programming language do you prefer?"
            />
          </div>

          <div className="field">
            <label htmlFor="description">Description (optional)</label>
            <textarea
              id="description"
              className="input"
              rows={2}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Vote for your favorite."
            />
          </div>

          <div className="field">
            <label>Options</label>
            {options.map((opt, i) => (
              <div key={i} style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
                <input
                  className="input"
                  required
                  maxLength={120}
                  value={opt}
                  onChange={(e) => updateOption(i, e.target.value)}
                  placeholder={`Option ${i + 1}`}
                />
                {options.length > 2 && (
                  <button
                    type="button"
                    className="btn btn-secondary"
                    aria-label={`Remove option ${i + 1}`}
                    onClick={() => removeOption(i)}
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
            {options.length < 10 && (
              <button type="button" className="btn btn-secondary" onClick={addOption} style={{ marginTop: 4 }}>
                + Add option
              </button>
            )}
          </div>

          <div className="field">
            <label htmlFor="expiresAt">Closing time (optional)</label>
            <input
              id="expiresAt"
              className="input"
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
            />
          </div>

          <div className="field" style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
            <input
              id="anonymousVoting"
              type="checkbox"
              checked={anonymousVoting}
              onChange={(e) => setAnonymousVoting(e.target.checked)}
            />
            <label htmlFor="anonymousVoting" style={{ margin: 0 }}>
              Allow anonymous voting (no login required to vote)
            </label>
          </div>

          {error && <p className="error-text" role="alert" style={{ marginBottom: 16 }}>{error}</p>}

          <button className="btn btn-primary btn-block" type="submit" disabled={submitting}>
            {submitting ? 'Creating…' : 'Create poll'}
          </button>
        </form>
      </div>
    </div>
  )
}