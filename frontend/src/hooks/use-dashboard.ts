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

    let cancelled = false
    setLoading(true)
    setError(null)

    void dashboardService.getActivityCoachThread(activityId).then((thread) => {
      if (!cancelled) {
        setMessages(thread)
        setLoading(false)
      }
    })

    return () => {
      cancelled = true
    }
  }, [activityId])

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

      const athleteMessage: ChatMessage = {
        id: `local_athlete_${Date.now()}`,
        role: 'athlete',
        content: trimmed,
        createdAt: new Date().toISOString(),
      }

      setSending(true)
      setError(null)
      setMessages((prev) => [...prev, athleteMessage])

      try {
        const reply = await dashboardService.askCoach(token, activityId, trimmed)
        const coachMessage: ChatMessage = {
          id: `local_coach_${Date.now()}`,
          role: 'coach',
          content: reply,
          createdAt: new Date().toISOString(),
        }
        setMessages((prev) => [...prev, coachMessage])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Could not reach the AI coach.')
      } finally {
        setSending(false)
      }
    },
    [activityId, sending, token],
  )

  return { messages, loading, sending, error, sendMessage }
}
