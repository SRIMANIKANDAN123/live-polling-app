import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="container" style={{ textAlign: 'center', paddingTop: 96 }}>
      <h1>404</h1>
      <p className="muted" style={{ marginBottom: 24 }}>That page doesn't exist.</p>
      <Link to="/" className="btn btn-primary">Go home</Link>
    </div>
  )
}
