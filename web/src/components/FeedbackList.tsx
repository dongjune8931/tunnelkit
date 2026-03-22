import React from 'react'
import { Feedback } from '../store/tunnelStore'

interface Props {
  feedbacks: Feedback[]
  onResolve: (id: string) => void
}

export function FeedbackList({ feedbacks, onResolve }: Props) {
  if (feedbacks.length === 0) {
    return (
      <div style={{ color: '#9ca3af', padding: '24px 0', textAlign: 'center' }}>
        아직 피드백이 없습니다
      </div>
    )
  }

  return (
    <div>
      {feedbacks.map((fb) => (
        <div
          key={fb.id}
          style={{
            border: '1px solid #e5e7eb',
            borderRadius: 8,
            padding: 14,
            marginBottom: 10,
            background: fb.resolved ? '#f9fafb' : '#fff',
            opacity: fb.resolved ? 0.6 : 1,
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
            <div>
              <span style={{ fontWeight: 600, color: '#374151' }}>
                {fb.author_name || '익명'}
              </span>
              <span style={{ fontSize: 12, color: '#9ca3af', marginLeft: 8 }}>
                {new Date(fb.created_at).toLocaleString()}
              </span>
            </div>
            {!fb.resolved && (
              <button
                onClick={() => onResolve(fb.id)}
                style={{
                  fontSize: 12,
                  padding: '4px 10px',
                  borderRadius: 6,
                  border: '1px solid #d1d5db',
                  background: '#fff',
                  cursor: 'pointer',
                  color: '#374151',
                }}
              >
                해결됨
              </button>
            )}
          </div>
          <div style={{ fontSize: 14, color: '#111827', marginBottom: 6 }}>
            {fb.comment}
          </div>
          <div style={{ fontSize: 11, color: '#9ca3af' }}>
            <span>{fb.page_url}</span>
            {fb.element_css && (
              <span style={{ marginLeft: 8, fontFamily: 'monospace' }}>
                {fb.element_css}
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}
