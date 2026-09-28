import { mockDevices, mockPlanDays } from '@/data/mock/dashboard'
import { fetchActivities, postActivityChat, uploadFitFile as uploadFitFileApi } from '@/lib/api'
import { buildWeekSummary } from '@/lib/week-summary'
import type {
  Activity,
  ChatMessage,
  DashboardData,
  DeviceConnection,
} from '@/types/dashboard'

/**
 * Activities + coach chat hit Symfony. Plan/devices stay mocked until those APIs exist.
 */
const delay = (ms = 280) => new Promise((resolve) => setTimeout(resolve, ms))

export async function getDashboardData(token: string): Promise<DashboardData> {
  const activities = await fetchActivities(token)

  return {
    activities,
    week: buildWeekSummary(activities),
    planDays: structuredClone(mockPlanDays),
    devices: structuredClone(mockDevices),
  }
}

/** Chat is stateless on the backend for now - threads live only in the browser session. */
export async function getActivityCoachThread(_activityId: string): Promise<ChatMessage[]> {
  return []
}

export async function askCoach(
  token: string,
  activityId: string,
  content: string,
): Promise<string> {
  const { reply } = await postActivityChat(token, activityId, content)
  return reply
}

export async function uploadFitFile(token: string, file: File): Promise<Activity> {
  return uploadFitFileApi(token, file)
}

export async function connectDevice(
  provider: DeviceConnection['provider'],
): Promise<DeviceConnection> {
  await delay(500)
  return {
    id: `dev_${provider}`,
    provider,
    label: provider.charAt(0).toUpperCase() + provider.slice(1),
    connected: true,
    lastSyncAt: new Date().toISOString(),
  }
}
