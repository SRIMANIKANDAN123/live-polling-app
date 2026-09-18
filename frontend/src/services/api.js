import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Normalizes the {success, data} / {success, error} envelope the backend
// always returns, so callers can just `await` and get data or a thrown
// Error with a readable message.
async function unwrap(promise) {
  try {
    const res = await promise
    return res.data.data
  } catch (err) {
    const message =
      err.response?.data?.error?.message || err.message || 'Something went wrong'
    const code = err.response?.data?.error?.code
    const wrapped = new Error(message)
    wrapped.code = code
    wrapped.status = err.response?.status
    throw wrapped
  }
}

export const authApi = {
  register: (email, password, confirmPassword) =>
    unwrap(api.post('/auth/register', { email, password, confirmPassword })),
  login: (email, password) => unwrap(api.post('/auth/login', { email, password })),
  me: () => unwrap(api.get('/auth/me')),
}

export const pollApi = {
  create: (payload) => unwrap(api.post('/polls', payload)),
  get: (id) => unwrap(api.get(`/polls/${id}`)),
  results: (id) => unwrap(api.get(`/polls/${id}/results`)),
  mine: () => unwrap(api.get('/polls/mine')),
  vote: (id, optionId) => unwrap(api.post(`/polls/${id}/vote`, { optionId })),
  close: (id) => unwrap(api.post(`/polls/${id}/close`)),
  remove: (id) => unwrap(api.delete(`/polls/${id}`)),
}

export default api
