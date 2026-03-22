import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import { Tunnel, useTunnelStore } from '../store/tunnelStore'
import { TunnelCard } from '../components/TunnelCard'
import { LogStream } from '../components/LogStream'
import { FeedbackList } from '../components/FeedbackList'
import { InviteModal } from '../components/InviteModal'
import { useLogStream } from '../hooks/useLogStream'
import { useFeedback } from '../hooks/useFeedback'

export function DashboardPage() {
  const { tunnels, selectedSub, logs, setTunnels, selectTunnel } = useTunnelStore()
  const [showInvite, setShowInvite] = useState(false)
  const [tab, setTab] = useState<'logs' | 'feedback'>('logs')

  // 터널 목록 폴링
  useQuery<Tunnel[]>({
    queryKey: ['tunnels'],
    queryFn: () => axios.get('/api/tunnels').then((r) => r.data),
    refetchInterval: 3000,
    select: (data) => {
      setTunnels(data)
      return data
    },
  })

  useLogStream(selectedSub)
  const { feedbacks, resolve } = useFeedback(selectedSub)

  return (
    <div style={{ display: 'flex', height: '100vh', fontFamily: 'system-ui, sans-serif' }}>
      {/* 사이드바 */}
      <aside
        style={{
          width: 260,
          borderRight: '1px solid #e5e7eb',
          padding: '20px 16px',
          background: '#f9fafb',
          overflowY: 'auto',
        }}
      >
        <h1 style={{ fontSize: 18, fontWeight: 800, color: '#4f46e5', marginBottom: 20 }}>
          previewd
        </h1>
        <div style={{ fontSize: 12, color: '#9ca3af', marginBottom: 10, fontWeight: 600 }}>
          활성 터널
        </div>
        {tunnels.length === 0 ? (
          <div style={{ color: '#d1d5db', fontSize: 13 }}>
            CLI를 실행하면 터널이 나타납니다
          </div>
        ) : (
          tunnels.map((t) => (
            <TunnelCard
              key={t.subdomain}
              tunnel={t}
              selected={t.subdomain === selectedSub}
              onClick={() => selectTunnel(t.subdomain)}
            />
          ))
        )}
      </aside>

      {/* 메인 영역 */}
      <main style={{ flex: 1, padding: 28, overflowY: 'auto' }}>
        {!selectedSub ? (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              color: '#9ca3af',
              fontSize: 16,
            }}
          >
            좌측에서 터널을 선택하세요
          </div>
        ) : (
          <>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 24,
              }}
            >
              <div>
                <h2 style={{ fontSize: 22, fontWeight: 700, margin: 0 }}>
                  {selectedSub}
                </h2>
                <a
                  href={`http://${selectedSub}.localhost:8080`}
                  target="_blank"
                  rel="noreferrer"
                  style={{ fontSize: 13, color: '#4f46e5' }}
                >
                  {selectedSub}.localhost:8080 →
                </a>
              </div>
              <button
                onClick={() => setShowInvite(true)}
                style={{
                  padding: '10px 18px',
                  background: '#4f46e5',
                  color: '#fff',
                  border: 'none',
                  borderRadius: 8,
                  fontSize: 14,
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                초대 링크 생성
              </button>
            </div>

            {/* 탭 */}
            <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
              {(['logs', 'feedback'] as const).map((t) => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  style={{
                    padding: '8px 16px',
                    borderRadius: 6,
                    border: 'none',
                    background: tab === t ? '#4f46e5' : '#e5e7eb',
                    color: tab === t ? '#fff' : '#374151',
                    fontWeight: 600,
                    cursor: 'pointer',
                    fontSize: 14,
                  }}
                >
                  {t === 'logs' ? '요청 로그' : `피드백 (${feedbacks.length})`}
                </button>
              ))}
            </div>

            {tab === 'logs' ? (
              <LogStream logs={logs} />
            ) : (
              <FeedbackList feedbacks={feedbacks} onResolve={resolve} />
            )}
          </>
        )}
      </main>

      {showInvite && selectedSub && (
        <InviteModal subdomain={selectedSub} onClose={() => setShowInvite(false)} />
      )}
    </div>
  )
}
