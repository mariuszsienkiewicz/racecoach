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

export async function reprocessActivity(
  token: string,
  activityId: string | number,
  mode: 'full' | 'from_structure' = 'from_structure',
): Promise<Activity> {
  const response = await fetch(`${API_BASE}/api/activities/${activityId}/reprocess`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify({ mode }),
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not refresh analysis.'))
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

type PostActivityChatOptions = {
  onToken?: (text: string) => void
  signal?: AbortSignal
}

/**
 * POST /api/activities/{id}/chat - SSE stream (token / done / error).
 * Uses fetch + manual SSE parse (EventSource cannot send POST + JWT).
 */
export async function postActivityChat(
  token: string,
  activityId: string | number,
  message: string,
  options: PostActivityChatOptions = {},
): Promise<ActivityChatResponse> {
  const response = await fetch(`${API_BASE}/api/activities/${activityId}/chat`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
    },
    body: JSON.stringify({ message }),
    signal: options.signal,
  })

  if (response.status === 401) {
    throw new Error('Your session expired. Please sign in again.')
  }

  // Validation / ownership errors still arrive as JSON before the stream starts.
  const contentType = response.headers.get('content-type') ?? ''
  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Could not reach the AI coach.'))
  }
  if (!contentType.includes('text/event-stream') || !response.body) {
    throw new Error('AI coach returned an unexpected response.')
  }

  return consumeActivityChatSse(response.body, String(activityId), options.onToken)
}

async function consumeActivityChatSse(
  body: ReadableStream<Uint8Array>,
  fallbackActivityId: string,
  onToken?: (text: string) => void,
): Promise<ActivityChatResponse> {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let eventName: string | null = null
  let dataLines: string[] = []
  let result: ActivityChatResponse | null = null

  const handleEvent = (event: string, data: string) => {
    let payload: unknown
    try {
      payload = JSON.parse(data) as unknown
    } catch {
      throw new Error('AI coach returned invalid stream data.')
    }
    if (!payload || typeof payload !== 'object') {
      throw new Error('AI coach returned invalid stream data.')
    }
    const record = payload as Record<string, unknown>

    if (event === 'token') {
      const text = record.text
      if (typeof text === 'string' && text !== '') {
        onToken?.(text)
      }
      return
    }

    if (event === 'error') {
      const message =
        typeof record.message === 'string' && record.message.trim() !== ''
          ? record.message.trim()
          : 'Could not reach the AI coach.'
      throw new Error(message)
    }

    if (event === 'done') {
      const reply = typeof record.reply === 'string' ? record.reply.trim() : ''
      if (reply === '') {
        throw new Error('AI coach returned an empty reply.')
      }
      const messages = mapChatMessages(record.messages)
      if (messages.length < 2) {
        throw new Error('AI coach returned an incomplete conversation turn.')
      }
      result = {
        reply,
        activityId:
          record.activityId != null ? String(record.activityId) : fallbackActivityId,
        messages,
      }
    }
  }

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })

    while (true) {
      const newline = buffer.indexOf('\n')
      if (newline === -1) {
        break
      }

      let line = buffer.slice(0, newline)
      buffer = buffer.slice(newline + 1)
      if (line.endsWith('\r')) {
        line = line.slice(0, -1)
      }

      if (line === '') {
        if (eventName === null && dataLines.length === 0) {
          continue
        }
        const event = eventName ?? 'message'
        const data = dataLines.join('\n')
        eventName = null
        dataLines = []
        handleEvent(event, data)
        continue
      }

      if (line.startsWith(':')) {
        continue
      }
      if (line.startsWith('event:')) {
        eventName = line.slice('event:'.length).trim()
        continue
      }
      if (line.startsWith('data:')) {
        dataLines.push(line.slice('data:'.length).replace(/^\s/, ''))
      }
    }
  }

  if (eventName !== null || dataLines.length > 0) {
    handleEvent(eventName ?? 'message', dataLines.join('\n'))
  }

  if (result === null) {
    throw new Error('AI coach stream ended without a reply.')
  }

  return result
}
