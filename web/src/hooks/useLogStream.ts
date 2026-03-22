import { useEffect } from 'react'
import { useTunnelStore } from '../store/tunnelStore'

export function useLogStream(subdomain: string | null) {
  const appendLogs = useTunnelStore((s) => s.appendLogs)

  useEffect(() => {
    if (!subdomain) return

    const es = new EventSource(`/api/tunnels/${subdomain}/logs/stream`)

    es.onmessage = (e) => {
      try {
        const entries = JSON.parse(e.data)
        appendLogs(entries)
      } catch {}
    }

    es.onerror = () => {
      // SSE 연결 오류 - EventSource가 자동으로 재연결 시도
    }

    return () => es.close()
  }, [subdomain, appendLogs])
}
