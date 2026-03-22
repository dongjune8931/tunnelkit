import React from 'react'
import { Tunnel } from '../store/tunnelStore'

interface Props {
  tunnel: Tunnel
  selected: boolean
  onClick: () => void
}

export function TunnelCard({ tunnel, selected, onClick }: Props) {
  return (
    <div
      onClick={onClick}
      style={{
        padding: '12px 16px',
        marginBottom: 8,
        borderRadius: 8,
        border: selected ? '2px solid #4f46e5' : '1px solid #e5e7eb',
        background: selected ? '#eef2ff' : '#fff',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
      }}
    >
      <span
        style={{
          width: 10,
          height: 10,
          borderRadius: '50%',
          background: tunnel.connected ? '#22c55e' : '#ef4444',
          display: 'inline-block',
        }}
      />
      <div>
        <div style={{ fontWeight: 600, color: '#111' }}>{tunnel.subdomain}</div>
        <div style={{ fontSize: 12, color: '#6b7280' }}>
          {tunnel.subdomain}.localhost:8080
        </div>
      </div>
    </div>
  )
}
