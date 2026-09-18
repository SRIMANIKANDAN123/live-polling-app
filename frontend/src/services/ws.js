const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'

/**
 * Opens a WebSocket connection to a poll's live-updates room, with
 * automatic reconnect (exponential-ish backoff) if the connection drops.
 * Returns a cleanup function to close it for good.
 *
 * handlers: { onMessage(data), onOpen(), onClose(), onReconnecting() }
 */
export function connectToPoll(pollId, handlers = {}) {
  let socket = null
  let closedByCaller = false
  let attempt = 0
  let reconnectTimer = null

  function connect() {
    socket = new WebSocket(`${WS_URL}/ws/polls/${pollId}`)

    socket.onopen = () => {
      attempt = 0
      handlers.onOpen?.()
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handlers.onMessage?.(data)
      } catch {
        // ignore malformed frames
      }
    }

    socket.onclose = () => {
      handlers.onClose?.()
      if (!closedByCaller) {
        attempt += 1
        const delay = Math.min(1000 * attempt, 5000)
        handlers.onReconnecting?.()
        reconnectTimer = setTimeout(connect, delay)
      }
    }

    socket.onerror = () => {
      socket.close()
    }
  }

  connect()

  return function disconnect() {
    closedByCaller = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    socket?.close()
  }
}
