import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from './api'

type Msg = { id: number; role: string; content: string; created_at: string }

export function AiAssistantPage() {
  const [messages, setMessages] = useState<Msg[]>([])
  const [input, setInput] = useState('')
  const [pending, setPending] = useState(false)
  const endRef = useRef<HTMLDivElement>(null)

  async function loadHistory() {
    const r = await api<{ messages: Msg[] }>(
      '/api/hr/chat/history?page=1&page_size=200'
    )
    setMessages(r.messages || [])
  }

  useEffect(() => {
    api<{ messages: Msg[] }>('/api/hr/chat/history?page=1&page_size=200')
      .then((r) => setMessages(r.messages || []))
      .catch(() => {})
  }, [])

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function send(e: FormEvent) {
    e.preventDefault()
    if (!input.trim() || pending) return
    setPending(true)
    try {
      await api<{ answer: string }>('/api/hr/chat', {
        method: 'POST',
        body: JSON.stringify({ content: input }),
      })
      setInput('')
      await loadHistory()
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="page-shell-inner">
      <div className="content-card">
        <h1 className="page-title">AI 助手</h1>
        <p className="muted page-lead">
          可询问投递统计、岗位热度等问题；历史记录会显示在下方。
        </p>
        <div className="ai-panel">
          <div className="ai-msgs" role="log" aria-live="polite">
            {messages.map((m) => (
              <div key={m.id} className={`ai-bubble ai-bubble--${m.role}`}>
                {m.content}
              </div>
            ))}
            <div ref={endRef} />
          </div>
          <form onSubmit={send} className="ai-input-row">
            <input
              value={input}
              placeholder="例如：全平台投递总数？岗位热度？"
              onChange={(e) => setInput(e.target.value)}
              aria-label="输入问题"
            />
            <button type="submit" disabled={pending}>
              发送
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
