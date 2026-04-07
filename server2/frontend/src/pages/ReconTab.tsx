import { useState, useRef } from 'react'
import { fetchApi } from '../lib/types'

type Status = 'idle' | 'running' | 'done' | 'error'

export default function ReconTab({ domain }: { domain: string }) {
  const [target, setTarget] = useState(domain)
  const [status, setStatus] = useState<Status>('idle')
  const [errMsg, setErrMsg] = useState('')
  const [toast, setToast] = useState(false)
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  function showToast() {
    setToast(true)
    if (toastTimer.current) clearTimeout(toastTimer.current)
    toastTimer.current = setTimeout(() => setToast(false), 3000)
  }

  async function runRecon() {
    if (!target.trim()) return
    setStatus('running')
    setErrMsg('')

    try {
      const res = await fetchApi('/api/workflow', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: target.trim() }),
      })

      if (res.ok) {
        setStatus('done')
        showToast()
      } else {
        const data = await res.json().catch(() => ({}))
        setErrMsg(data.error || `server error ${res.status}`)
        setStatus('error')
      }
    } catch {
      setErrMsg('connection error')
      setStatus('error')
    }
  }

  return (
    <div style={{
      flex: 1,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'var(--bg)',
    }}>
      <div style={{
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 8,
        padding: '32px 36px',
        width: 440,
        display: 'flex',
        flexDirection: 'column',
        gap: 20,
      }}>
        {/* title */}
        <div>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', marginBottom: 4 }}>
            Automated Recon
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
            Runs the full recon pipeline. You'll be notified via Telegram when done.
          </div>
        </div>

        {/* input */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 11, color: 'var(--text-muted)' }}>target domain</label>
          <input
            type="text"
            value={target}
            onChange={e => { setTarget(e.target.value); setStatus('idle'); setErrMsg('') }}
            placeholder="example.com"
            disabled={status === 'running'}
            style={{
              background: 'var(--surface2)',
              border: '1px solid var(--border2)',
              borderRadius: 6,
              padding: '9px 12px',
              color: 'var(--text)',
              fontSize: 13,
              fontFamily: "'Fira Code', monospace",
              outline: 'none',
              opacity: status === 'running' ? 0.5 : 1,
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--accent)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border2)')}
            onKeyDown={e => e.key === 'Enter' && status !== 'running' && runRecon()}
          />
        </div>

        {/* error message */}
        {status === 'error' && (
          <div style={{
            fontSize: 12,
            color: 'var(--red)',
            background: 'var(--red-dim)',
            border: '1px solid var(--red)',
            borderRadius: 6,
            padding: '8px 12px',
          }}>
            {errMsg}
          </div>
        )}

        {/* button */}
        <button
          onClick={runRecon}
          disabled={status === 'running' || !target.trim()}
          style={{
            background: status === 'running' ? 'var(--accent-dim)' : 'var(--accent)',
            color: status === 'running' ? 'var(--text-muted)' : '#000',
            border: 'none',
            borderRadius: 6,
            padding: '10px 0',
            fontSize: 13,
            fontWeight: 600,
            cursor: status === 'running' || !target.trim() ? 'not-allowed' : 'pointer',
            opacity: !target.trim() ? 0.5 : 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
          }}
        >
          {status === 'running' ? (
            <><span className="spinner" style={{ width: 12, height: 12 }} /> running...</>
          ) : 'Run Recon'}
        </button>
      </div>
      {toast && (
        <div className="toast success">
          <span className="toast-icon success">✓</span>
          Recon started — watch Telegram for updates.
        </div>
      )}
    </div>
  )
}
