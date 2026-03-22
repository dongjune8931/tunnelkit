import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import { Feedback } from '../store/tunnelStore'

export function useFeedback(subdomain: string | null) {
  const qc = useQueryClient()

  const query = useQuery<Feedback[]>({
    queryKey: ['feedback', subdomain],
    queryFn: () =>
      axios.get(`/api/tunnels/${subdomain}/feedback`).then((r) => r.data),
    enabled: !!subdomain,
    refetchInterval: 5000,
  })

  const resolve = useMutation({
    mutationFn: (id: string) =>
      axios.patch(`/api/feedback/${id}/resolve`).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['feedback', subdomain] }),
  })

  return { feedbacks: query.data ?? [], resolve: resolve.mutate }
}
