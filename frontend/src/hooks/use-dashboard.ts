import { useCallback, useEffect, useRef, useState } from 'react'

import { useAuth } from '@/context/auth-context'
import { buildWeekSummary, hasPendingActivities } from '@/lib/week-summary'
import * as dashboardService from '@/services/dashboard-service'
import type {
  Activity,
  ChatMessage,
  DashboardData,
  DeviceConnection,
} from '@/types/dashboard'

const PENDING_POLL_MS = 3000
const MAX_CHAT_MESSAGE_LENGTH = 2000

function mapCoachSendError(err: unknown): string {
  const raw = err instanceof Error ? err.message : 'Could not reach the AI coach.'
  if (/session expired/i.test(raw)) {
    return raw
  }
  if (/message is too long/i.test(raw)) {
    return 'Your message is too long. Please shorten it and try again.'
  }
  if (/AI coach unavailable/i.test(raw) || /Invalid chat response/i.test(raw)) {
    return raw
  }
  if (/history/i.test(raw)) {
    return 'Could not reach the AI coach. Please try again.'
  }
  return raw
}

type DashboardState = {
  data: DashboardData | null
  loading: boolean
  refreshing: boolean
  error: string | null
  refresh: () => Promise<void>
  addActivity: (activity: Activity) => void
  upsertDevice: (device: DeviceConnection) => void
  uploadFit: (file: File) => Promise<Activity>
  connectProvider: (provider: DeviceConnection['provider']) => Promise<void>
}

export function useDashboard(): DashboardState {
  const { token } = useAuth()
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const requestId = useRef(0)

  const refresh = useCallback(
    async (opts?: { silent?: boolean }) => {
      if (!token) {
        setData(null)
        setLoading(false)
        setError('Your session expired. Please sign in again.')
        return
      }

      const id = ++requestId.current
      const silent = opts?.silent === true

      if (silent) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      setError(null)

      try {
        const next = await dashboardService.getDashboardData(token)
        if (id !== requestId.current) {
          return
        }
        setData(next)
      } catch (err) {
        if (id !== requestId.current) {
          return
        }
        setError(err instanceof Error ? err.message : 'Could not load dashboard data.')
      } finally {
        if (id === requestId.current) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    },
    [token],
  )

  useEffect(() => {
    void refresh()
  }, [refresh])

  useEffect(() => {
    if (!token || !data || !hasPendingActivities(data.activities)) {
      return
    }

    const timer = window.setInterval(() => {
      void refresh({ silent: true })
    }, PENDING_POLL_MS)

    return () => window.clearInterval(timer)
  }, [token, data, refresh])

  const addActivity = useCallback((activity: Activity) => {
    setData((current) => {
      if (!current) {
        return current
      }
      const activities = [activity, ...current.activities.filter((item) => item.id !== activity.id)]
      return {
        ...current,
        activities,
        week: buildWeekSummary(activities),
      }
    })
  }, [])

  const upsertDevice = useCallback((device: DeviceConnection) => {
    setData((current) => {
      if (!current) {
        return current
      }
      const exists = current.devices.some((item) => item.provider === device.provider)
      return {
        ...current,
        devices: exists
          ? current.devices.map((item) => (item.provider === device.provider ? device : item))
          : [...current.devices, device],
      }
    })
  }, [])

  const uploadFit = useCallback(
    async (file: File) => {
      if (!token) {
        throw new Error('Your session expired. Please sign in again.')
      }
      const activity = await dashboardService.uploadFitFile(token, file)
      addActivity(activity)
      return activity
    },
    [addActivity, token],
  )

  const connectProvider = useCallback(
    async (provider: DeviceConnection['provider']) => {
      const device = await dashboardService.connectDevice(provider)
      upsertDevice(device)
    },
    [upsertDevice],
  )

  return {
    data,
    loading,
    refreshing,
    error,
    refresh: () => refresh(),
    addActivity,
    upsertDevice,
    uploadFit,
    connectProvider,
  }
}

export function useActivityCoach(activityId: string | null) {
  const { token } = useAuth()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!activityId) {
      setMessages([])
      setError(null)
      setLoading(false)
      return
    }

    if (!token) {
      setMessages([])
      setError('Your session expired. Please sign in again.')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    void dashboardService
      .getActivityCoachThread(token, activityId)
      .then((thread) => {
        if (!cancelled) {
          setMessages(thread)
          setLoading(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setMessages([])
          setError(err instanceof Error ? err.message : 'Could not load the coach conversation.')
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [activityId, token])

  const sendMessage = useCallback(
    async (content: string) => {
      const trimmed = content.trim()
      if (!activityId || !trimmed || sending) {
        return
      }
      if (!token) {
        setError('Your session expired. Please sign in again.')
        return
      }
      if ([...trimmed].length > MAX_CHAT_MESSAGE_LENGTH) {
        setError('Your message is too long. Please shorten it and try again.')
        return
      }

      const now = Date.now()
      const optimisticAthleteId = `local_athlete_${now}`
      const streamingCoachId = `local_coach_${now}`
      const athleteMessage: ChatMessage = {
        id: optimisticAthleteId,
        role: 'athlete',
        content: trimmed,
        createdAt: new Date().toISOString(),
      }

      setSending(true)
      setError(null)
      setMessages((prev) => [...prev, athleteMessage])

      try {
        const turn = await dashboardService.askCoach(token, activityId, trimmed, {
          onToken: (text) => {
            setMessages((prev) => {
              const existing = prev.find((message) => message.id === streamingCoachId)
              if (existing) {
                return prev.map((message) =>
                  message.id === streamingCoachId
                    ? { ...message, content: message.content + text }
                    : message,
                )
              }
              return [
                ...prev,
                {
                  id: streamingCoachId,
                  role: 'coach' as const,
                  content: text,
                  createdAt: new Date().toISOString(),
                },
              ]
            })
          },
        })
        setMessages((prev) => {
          const withoutLocal = prev.filter(
            (message) =>
              message.id !== optimisticAthleteId && message.id !== streamingCoachId,
          )
          return [...withoutLocal, ...turn]
        })
      } catch (err) {
        // Athlete turn is already persisted server-side; drop the streaming coach bubble.
        setMessages((prev) => prev.filter((message) => message.id !== streamingCoachId))
        setError(mapCoachSendError(err))
      } finally {
        setSending(false)
      }
    },
    [activityId, sending, token],
  )

  return { messages, loading, sending, error, sendMessage }
}
