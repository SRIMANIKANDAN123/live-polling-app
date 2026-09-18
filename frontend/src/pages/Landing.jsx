import { Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import ResultsBar from '../components/ResultsBar'

const HERO_IMAGE =
  'https://media.istockphoto.com/id/2169856307/photo/a-person-casting-a-ballot-into-a-gray-box-against-a-vibrant-blue-background-with-a-sunburst.jpg?b=1&s=612x612&w=0&k=20&c=484527OkgXNevsTjThNmdw6urHKDlcvjwX8vSpzr0wI='

export default function Landing() {
  const { user } = useAuth()

  return (
    <div
      className="hero-section"
      style={{
        backgroundImage: `linear-gradient(rgba(10, 20, 30, 0.55), rgba(10, 20, 30, 0.7)), url(${HERO_IMAGE})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
        paddingTop: 96,
        paddingBottom: 96,
      }}
    >
      <div className="container">
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1.1fr 0.9fr',
            gap: 48,
            alignItems: 'center',
          }}
          className="hero-grid"
        >
          <div>
            <span
              className="badge"
              style={{
                marginBottom: 20,
                background: 'rgba(255, 255, 255, 0.16)',
                color: '#ffffff',
                border: '1px solid rgba(255, 255, 255, 0.3)',
              }}
            >
              Live · Instant · No refresh
            </span>
            <h1 style={{ fontSize: '2.6rem', lineHeight: 1.15, margin: '16px 0', color: '#ffffff' }}>
              Ask a question.
              <br />
              Watch the answers arrive.
            </h1>
            <p style={{ fontSize: '1.05rem', maxWidth: 460, marginBottom: 32, color: 'rgba(255, 255, 255, 0.82)' }}>
              Pulse is a live polling tool. Create a poll, share one link, and
              every viewer sees results update in real time — no page
              refresh, ever.
            </p>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
              <Link to={user ? '/dashboard/create' : '/register'} className="btn btn-primary">
                Create a poll
              </Link>
              <Link
                to="/login"
                className="btn btn-secondary"
                style={{ background: 'rgba(255,255,255,0.1)', color: '#ffffff', borderColor: 'rgba(255,255,255,0.4)' }}
              >
                I have a poll link
              </Link>
            </div>

            <div style={{ display: 'flex', gap: 32, marginTop: 48, flexWrap: 'wrap' }}>
              <Feature title="Realtime" text="Redis pub/sub pushes every vote over WebSocket." />
              <Feature title="Fair" text="One vote per person, enforced server-side." />
              <Feature title="Simple" text="Create a poll in under a minute." />
            </div>
          </div>

          <div className="card">
            <p className="muted" style={{ marginBottom: 4, fontSize: '0.85rem' }}>
              Example poll
            </p>
            <h3 style={{ margin: '0 0 20px' }}>Which language should we use for the new service?</h3>
            <ResultsBar label="Go" count={42} total={101} highlight />
            <ResultsBar label="TypeScript" count={31} total={101} />
            <ResultsBar label="Rust" count={28} total={101} />
            <p className="muted" style={{ fontSize: '0.85rem', marginTop: 8 }}>
              Total votes: 101 · <span style={{ color: 'var(--success)' }}>● Live</span>
            </p>
          </div>
        </div>
      </div>

      <style>{`
        @media (max-width: 800px) {
          .hero-grid { grid-template-columns: 1fr !important; }
        }
        @media (max-width: 640px) {
          .hero-section { padding-top: 56px !important; padding-bottom: 56px !important; }
        }
      `}</style>
    </div>
  )
}

function Feature({ title, text }) {
  return (
    <div>
      <p style={{ fontWeight: 700, marginBottom: 4, color: '#ffffff' }}>{title}</p>
      <p style={{ fontSize: '0.9rem', maxWidth: 160, color: 'rgba(255, 255, 255, 0.75)' }}>
        {text}
      </p>
    </div>
  )
}