import { useEffect, useMemo, useState } from 'react'
import { logfmt } from './log.js'

const socketURL = import.meta.env.VITE_CORE_WS_URL || 'ws://localhost:8080/ws'

function formatAge(date, now) {
  if (!date) return 'nunca'
  const seconds = Math.max(0, Math.floor((now - new Date(date).getTime()) / 1000))
  return `${seconds}s`
}

function App() {
  const [agents, setAgents] = useState([])
  const [connected, setConnected] = useState(false)
  const [lastEvent, setLastEvent] = useState(null)
  const [now, setNow] = useState(Date.now())

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [])

  useEffect(() => {
    let socket
    let retryTimer
    let closed = false

    const connect = () => {
      socket = new WebSocket(socketURL)
      socket.addEventListener('open', () => {
        setConnected(true)
        logfmt('INFO', 'websocket_connected', { url: socketURL })
      })
      socket.addEventListener('message', (message) => {
        const event = JSON.parse(message.data)
        setLastEvent(event)
        if (event.type === 'agent.snapshot') {
          setAgents(event.agents || [])
        } else if (event.agent) {
          setAgents((current) => {
            const index = current.findIndex((agent) => agent.id === event.agent.id)
            if (index < 0) return [...current, event.agent]
            const next = [...current]
            next[index] = event.agent
            return next
          })
        }
      })
      socket.addEventListener('close', () => {
        setConnected(false)
        logfmt('WARN', 'websocket_disconnected', { url: socketURL })
        if (!closed) retryTimer = window.setTimeout(connect, 2000)
      })
      socket.addEventListener('error', () => {
        logfmt('ERROR', 'websocket_error', { url: socketURL })
      })
    }

    connect()
    return () => {
      closed = true
      window.clearTimeout(retryTimer)
      socket?.close()
    }
  }, [])

  const healthy = useMemo(() => agents.filter((agent) => agent.status === 'healthy').length, [agents])

  return (
    <main className="shell">
      <header className="header">
        <div>
          <p className="eyebrow">NODARA CONSOLE</p>
          <h1>Salud de agentes</h1>
          <p className="subtitle">Heartbeat en tiempo real sobre gRPC, PostgreSQL y WebSocket.</p>
        </div>
        <div className={`connection ${connected ? 'online' : 'offline'}`}>
          <span className="dot" /> {connected ? 'Consola conectada' : 'Conectando...'}
        </div>
      </header>

      <section className="summary">
        <div><span>Total</span><strong>{agents.length}</strong></div>
        <div><span>Healthy</span><strong className="healthy-text">{healthy}</strong></div>
        <div><span>Último evento</span><strong>{lastEvent?.type || '—'}</strong></div>
      </section>

      <section className="panel">
        <div className="panel-heading">
          <h2>Agentes</h2>
          <span>actualización automática</span>
        </div>
        {agents.length === 0 ? (
          <div className="empty">Esperando el primer heartbeat...</div>
        ) : (
          <div className="table-wrap">
            <table>
              <thead><tr><th>Agente</th><th>Hostname</th><th>Estado</th><th>Último heartbeat</th><th>Secuencia</th><th>Versión</th></tr></thead>
              <tbody>
                {agents.map((agent) => (
                  <tr key={agent.id}>
                    <td><strong>{agent.name}</strong><small>{agent.id}</small></td>
                    <td>{agent.hostname}</td>
                    <td><span className={`status ${agent.status}`}>{agent.status}</span></td>
                    <td>{formatAge(agent.last_heartbeat_at, now)} atrás</td>
                    <td>{agent.sequence}</td>
                    <td>{agent.agent_version}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  )
}

export default App
