import React from 'react'
import { LogEntry } from '../store/tunnelStore'

interface Props {
  logs: LogEntry[]
}

function statusColor(status: number): string {
  if (status >= 500) return '#ef4444'
  if (status >= 400) return '#f59e0b'
  if (status >= 300) return '#3b82f6'
  return '#22c55e'
}

export function LogStream({ logs }: Props) {
  return (
    <div
      style={{
        fontFamily: 'monospace',
        fontSize: 13,
        background: '#0f172a',
        color: '#e2e8f0',
        borderRadius: 8,
        padding: 16,
        height: 320,
        overflowY: 'auto',
      }}
    >
      {logs.length === 0 ? (
        <div style={{ color: '#64748b' }}>요청을 기다리는 중...</div>
      ) : (
        logs.map((log) => (
          <div key={log.id} style={{ marginBottom: 4 }}>
            <span style={{ color: '#64748b' }}>
              {new Date(log.created_at).toLocaleTimeString()}
            </span>{' '}
            <span style={{ color: statusColor(log.status), fontWeight: 700 }}>
              {log.status}
            </span>{' '}
            <span style={{ color: '#7dd3fc' }}>{log.method.padEnd(6)}</span>{' '}
            <span>{log.path}</span>{' '}
            <span style={{ color: '#94a3b8' }}>{log.duration_ms}ms</span>
          </div>
        ))
      )}
    </div>
  )
}
