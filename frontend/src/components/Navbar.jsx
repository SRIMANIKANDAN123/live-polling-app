import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useTheme } from '../context/ThemeContext'

export default function Navbar() {
  const { user, logout } = useAuth()
  const { theme, toggle } = useTheme()
  const navigate = useNavigate()
  const [dark, setDark] = useState(false)

  return (
    <header
      style={{
        borderBottom: '1px solid var(--border)',
        background: 'var(--surface)',
      }}
    >
      <div
        className="container navbar-row"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: 12,
          minHeight: 64,
          padding: '12px 24px',
        }}
      >
        <Link
          to="/"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            fontWeight: 800,
            fontSize: '1.15rem',
          }}
        >
          <img
            src="https://img.icons8.com/?size=100&id=Hj0KDL60qVg2&format=png&color=000000"
            alt="voting"
            style={{ width: 28, height: 28 }}
          />
          <span className="brand-gradient">SnapVote</span>
        </Link>

        <nav style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <button
            className="btn btn-secondary"
            aria-label="Toggle dark mode"
            style={{ padding: '8px 12px' }}
            onClick={() => {
              setDark(!dark)
              toggle()
            }}
          >
            {dark ? 'Switch to Light' : 'Switch to Dark'}
            {theme === 'light' ? (
              <img
                src="https://img.icons8.com/?size=100&id=tlXZO3hU3TSG&format=png&color=000000"
                alt="Moon icon"
                style={{ width: '20px', marginLeft: '8px' }}
              />
            ) : (
              <img
                src="https://img.icons8.com/?size=100&id=WWRW41cpXB4o&format=png&color=000000"
                alt="Sun icon"
                style={{ width: '20px', marginLeft: '8px' }}
              />
            )}
          </button>

          {user ? (
            <>
              <Link to="/dashboard" className="btn btn-secondary">
                Dashboard
              </Link>
              <button
                className="btn btn-secondary"
                onClick={() => {
                  logout()
                  navigate('/')
                }}
              >
                Log out
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-secondary">
                Log in
              </Link>
              <Link to="/register" className="btn btn-primary">
                Sign up
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}