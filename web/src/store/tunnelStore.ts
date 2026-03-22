import { create } from 'zustand'

export interface Tunnel {
  subdomain: string
  connected: boolean
}

export interface LogEntry {
  id: string
  method: string
  path: string
  status: number
  duration_ms: number
  created_at: string
}

export interface Feedback {
  id: string
  session_id: string
  page_url: string
  element_css: string
  x_percent: number
  y_percent: number
  comment: string
  author_name: string
  resolved: boolean
  created_at: string
}

interface TunnelStore {
  tunnels: Tunnel[]
  selectedSub: string | null
  logs: LogEntry[]
  feedbacks: Feedback[]
  setTunnels: (tunnels: Tunnel[]) => void
  selectTunnel: (sub: string) => void
  appendLogs: (entries: LogEntry[]) => void
  setFeedbacks: (feedbacks: Feedback[]) => void
}

export const useTunnelStore = create<TunnelStore>((set) => ({
  tunnels: [],
  selectedSub: null,
  logs: [],
  feedbacks: [],
  setTunnels: (tunnels) => set({ tunnels }),
  selectTunnel: (sub) => set({ selectedSub: sub, logs: [], feedbacks: [] }),
  appendLogs: (entries) =>
    set((state) => ({
      logs: [...entries, ...state.logs].slice(0, 200),
    })),
  setFeedbacks: (feedbacks) => set({ feedbacks }),
}))
