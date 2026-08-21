import React, { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import DashboardPage from '@/pages/dashboard.jsx'
import LoginPage from '@/pages/login.jsx'
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
      socket.addEventListener('error', () => logfmt('ERROR', 'websocket_error', { url: socketURL }))
    }

    connect()
    return () => {
      closed = true
      window.clearTimeout(retryTimer)
      socket?.close()
    }
  }, [])

  const healthy = useMemo(() => agents.filter((agent) => agent.status === 'healthy').length, [agents])
  const tableData = useMemo(() => agents.map((agent, index) => ({
    id: index + 1,
    header: agent.name,
    type: agent.hostname,
    status: agent.status === 'healthy' ? 'Healthy' : 'Disconnected',
    target: formatAge(agent.last_heartbeat_at, now),
    limit: String(agent.sequence),
    reviewer: agent.agent_version,
  })), [agents, now])

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
