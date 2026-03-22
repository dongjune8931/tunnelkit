import React, { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import axios from 'axios'

interface Props {
  subdomain: string
  onClose: () => void
}

export function InviteModal({ subdomain, onClose }: Props) {
  const [inviteURL, setInviteURL] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const generate = async () => {
    setLoading(true)
    try {
      const res = await axios.post(`/api/tunnels/${subdomain}/invite`)
      setInviteURL(res.data.url)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.4)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 50,
      }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          padding: 32,
          width: 400,
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
        }}
      >
        <h2 style={{ margin: '0 0 16px', fontSize: 20, fontWeight: 700 }}>
          초대 링크 생성
        </h2>
        <p style={{ color: '#6b7280', fontSize: 14, marginBottom: 20 }}>
          이 링크를 PM이나 디자이너에게 공유하면 바로 확인할 수 있습니다.
        </p>

        {!inviteURL ? (
          <button
            onClick={generate}
            disabled={loading}
            style={{
              width: '100%',
              padding: '12px',
              background: '#4f46e5',
              color: '#fff',
              border: 'none',
              borderRadius: 8,
              fontSize: 15,
              fontWeight: 600,
              cursor: loading ? 'wait' : 'pointer',
            }}
          >
            {loading ? '생성 중...' : '초대 링크 생성'}
          </button>
        ) : (
          <div style={{ textAlign: 'center' }}>
            <div style={{ marginBottom: 16 }}>
              <QRCodeSVG value={inviteURL} size={180} />
            </div>
            <input
              readOnly
              value={inviteURL}
              onClick={(e) => (e.target as HTMLInputElement).select()}
              style={{
                width: '100%',
                padding: '10px',
                border: '1px solid #d1d5db',
                borderRadius: 8,
                fontSize: 13,
                fontFamily: 'monospace',
                boxSizing: 'border-box',
                color: '#374151',
              }}
            />
            <p style={{ fontSize: 12, color: '#9ca3af', marginTop: 8 }}>
              클릭해서 복사하세요
            </p>
          </div>
        )}

        <button
          onClick={onClose}
          style={{
            width: '100%',
            marginTop: 12,
            padding: '10px',
            background: 'transparent',
            border: '1px solid #d1d5db',
            borderRadius: 8,
            fontSize: 14,
            cursor: 'pointer',
            color: '#374151',
          }}
        >
          닫기
        </button>
      </div>
    </div>
  )
}
