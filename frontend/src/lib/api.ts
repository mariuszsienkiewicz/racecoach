import type { Activity } from '@/types/dashboard'

const API_BASE = import.meta.env.VITE_API_URL ?? ''

export type AuthUser = {
  id: number
  email: string
  roles: string[]
}

type LoginResponse = {
  token: string
}

type ActivitiesResponse = {
  activities: Activity[]
}

async function readErrorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const payload = (await response.json()) as { message?: string; detail?: string; error?: string }
    return payload.message ?? payload.detail ?? payload.error ?? fallback
  } catch {
    return fallback
  }
}

export async function login(email: string, password: string): Promise<string> {
  const response = await fetch(`${API_BASE}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  if (!response.ok) {
    throw new Error('Invalid email or password.')
  }

  const data = (await response.json()) as LoginResponse
  return data.token
}

export async function fetchMe(token: string): Promise<AuthUser> {
  const response = await fetch(`${API_BASE}/api/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })

  if (!response.ok) {
    throw new Error('Your session expired. Please sign in again.')
  }

  return (await response.json()) as AuthUser
}

export async function fetchActivities(token: string): Promise<Activity[]> {
  const response = await fetch(`${API_BASE}/api/activities`, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json',
    },
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not load activities.'))
  }

  const data = (await response.json()) as ActivitiesResponse
  return Array.isArray(data.activities) ? data.activities : []
}

export async function uploadFitFile(token: string, file: File): Promise<Activity> {
  const body = new FormData()
  body.append('fitFile', file)

  const response = await fetch(`${API_BASE}/api/activities/fit`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json',
    },
    body,
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not upload the .fit file.'))
  }

  return (await response.json()) as Activity
}

export type ActivityChatResponse = {
  reply: string
  activityId: number
}

export async function postActivityChat(
  token: string,
  activityId: string | number,
  message: string,
): Promise<ActivityChatResponse> {
  const response = await fetch(`${API_BASE}/api/activities/${activityId}/chat`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify({ message }),
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not reach the AI coach.'))
  }

  const data = (await response.json()) as ActivityChatResponse
  if (typeof data.reply !== 'string' || data.reply.trim() === '') {
    throw new Error('AI coach returned an empty reply.')
  }

  return data
}
