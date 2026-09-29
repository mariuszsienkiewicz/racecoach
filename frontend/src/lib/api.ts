import type { Activity, ChatMessage } from '@/types/dashboard'

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

type ActivityChatMessageDto = {
  id?: unknown
  role?: unknown
  content?: unknown
  createdAt?: unknown
}

type ActivityChatListResponse = {
  activityId?: unknown
  messages?: unknown
}

export type ActivityChatResponse = {
  reply: string
  activityId: string
  messages: ChatMessage[]
}

async function readErrorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const payload = (await response.json()) as { message?: string; detail?: string; error?: string }
    return payload.message ?? payload.detail ?? payload.error ?? fallback
  } catch {
    return fallback
  }
}

function mapChatMessage(raw: unknown): ChatMessage | null {
  if (!raw || typeof raw !== 'object') {
    return null
  }
  const dto = raw as ActivityChatMessageDto
  if (typeof dto.id !== 'string' && typeof dto.id !== 'number') {
    return null
  }
  if (dto.role !== 'athlete' && dto.role !== 'coach') {
    return null
  }
  if (typeof dto.content !== 'string' || dto.content.trim() === '') {
    return null
  }
  if (typeof dto.createdAt !== 'string' || dto.createdAt.trim() === '') {
    return null
  }

  return {
    id: String(dto.id),
    role: dto.role,
    content: dto.content,
    createdAt: dto.createdAt,
  }
}

function mapChatMessages(raw: unknown): ChatMessage[] {
  if (!Array.isArray(raw)) {
    return []
  }
  return raw.map(mapChatMessage).filter((message): message is ChatMessage => message !== null)
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

export async function fetchActivityChat(
  token: string,
  activityId: string | number,
): Promise<ChatMessage[]> {
  const response = await fetch(`${API_BASE}/api/activities/${activityId}/chat`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json',
    },
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not load the coach conversation.'))
  }

  const data = (await response.json()) as ActivityChatListResponse
  return mapChatMessages(data.messages)
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

  const data = (await response.json()) as {
    reply?: unknown
    activityId?: unknown
    messages?: unknown
  }

  const reply = typeof data.reply === 'string' ? data.reply.trim() : ''
  if (reply === '') {
    throw new Error('AI coach returned an empty reply.')
  }

  const messages = mapChatMessages(data.messages)
  if (messages.length < 2) {
    throw new Error('AI coach returned an incomplete conversation turn.')
  }

  return {
    reply,
    activityId: data.activityId != null ? String(data.activityId) : String(activityId),
    messages,
  }
}
