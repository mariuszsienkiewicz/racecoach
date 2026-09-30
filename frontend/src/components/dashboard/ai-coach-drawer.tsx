import { useEffect, useRef, useState, type FormEvent } from 'react'
import { MessageSquare, SendHorizontal, X } from 'lucide-react'

import { useActivityCoach } from '@/hooks/use-dashboard'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Bubble, BubbleContent } from '@/components/ui/bubble'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Marker, MarkerContent, MarkerIcon } from '@/components/ui/marker'
import {
  Message,
  MessageAvatar,
  MessageContent,
  MessageHeader,
} from '@/components/ui/message'
import { Spinner } from '@/components/ui/spinner'
import type { Activity, ChatMessage } from '@/types/dashboard'
import { cn } from '@/lib/utils'

type AiCoachDrawerProps = {
  activity: Activity | null
  onClose: () => void
}

function CoachMessage({ message }: { message: ChatMessage }) {
  const isAthlete = message.role === 'athlete'

  return (
    <Message
      align={isAthlete ? 'end' : 'start'}
      className="rc-fade-up-soft"
    >
      <MessageAvatar>
        <Avatar size="sm" className={cn(isAthlete ? 'bg-foreground/10' : 'bg-primary/15')}>
          <AvatarFallback
            className={cn(
              'font-display text-[11px] font-semibold',
              isAthlete ? 'text-foreground' : 'bg-primary text-primary-foreground',
            )}
          >
            {isAthlete ? 'You' : 'RC'}
          </AvatarFallback>
        </Avatar>
      </MessageAvatar>
      <MessageContent>
        <MessageHeader>{isAthlete ? 'You' : 'RaceCoach'}</MessageHeader>
        <Bubble variant={isAthlete ? 'outline' : 'default'} align={isAthlete ? 'end' : 'start'}>
          <BubbleContent className="whitespace-pre-wrap">{message.content}</BubbleContent>
        </Bubble>
      </MessageContent>
    </Message>
  )
}

export function AiCoachDrawer({ activity, onClose }: AiCoachDrawerProps) {
  const { messages, loading, sending, error, sendMessage } = useActivityCoach(activity?.id ?? null)
  const [draft, setDraft] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const streamingCoach = messages.find(
    (message) => message.role === 'coach' && message.id.startsWith('local_coach_'),
  )
  const waitingForFirstToken = sending && !streamingCoach

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [messages, sending, streamingCoach?.content, activity?.id])

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    const content = draft
    if (!content.trim() || sending) {
      return
    }
    setDraft('')
    await sendMessage(content)
  }

  return (
    <div
      className={cn(
        'fixed inset-0 z-40 transition',
        activity ? 'pointer-events-auto' : 'pointer-events-none',
      )}
      aria-hidden={!activity}
    >
      <button
        type="button"
        className={cn(
          'absolute inset-0 bg-[#1a2a20]/40 backdrop-blur-[2px] transition-opacity',
          activity ? 'opacity-100' : 'opacity-0',
        )}
        onClick={onClose}
        aria-label="Close AI coach"
      />

      <aside
        className={cn(
          'absolute inset-y-0 right-0 flex w-full max-w-xl flex-col border-l border-border/70 bg-[#f8f5ee] shadow-2xl transition-transform duration-300 sm:max-w-2xl',
          activity ? 'translate-x-0' : 'translate-x-full',
        )}
      >
        <div className="relative overflow-hidden border-b border-border/70 px-5 py-4 sm:px-6">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,oklch(0.42_0.12_155_/_0.12),transparent_55%)]"
          />
          <div className="relative flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                <MessageSquare className="size-3.5" />
                AI coach
              </p>
              <h2 className="mt-1 truncate font-display text-xl font-semibold tracking-tight sm:text-2xl">
                {activity?.title ?? 'Activity'}
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Ask about pacing, effort, or what to do next for this session.
              </p>
            </div>
            <Button type="button" variant="ghost" size="sm" onClick={onClose} aria-label="Close">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <div className="flex-1 space-y-4 overflow-y-auto px-5 py-5 sm:px-6">
          {loading ? (
            <Marker role="status">
              <MarkerIcon>
                <Spinner />
              </MarkerIcon>
              <MarkerContent>Loading conversation…</MarkerContent>
            </Marker>
          ) : null}

          {!loading && messages.length === 0 && !sending ? (
            <div className="rounded-2xl border border-dashed border-border/80 bg-background/50 px-4 py-8 text-center">
              <p className="font-display text-base font-semibold tracking-tight">Start the debrief</p>
              <p className="mt-2 text-sm text-muted-foreground">
                Your message appears instantly. RaceCoach answers from this workout&apos;s analysis.
              </p>
            </div>
          ) : null}

          {messages.map((message) => (
            <CoachMessage key={message.id} message={message} />
          ))}

          {waitingForFirstToken ? (
            <Message align="start" className="rc-fade-up-soft">
              <MessageAvatar>
                <Avatar size="sm" className="bg-primary/15">
                  <AvatarFallback className="bg-primary font-display text-[11px] font-semibold text-primary-foreground">
                    RC
                  </AvatarFallback>
                </Avatar>
              </MessageAvatar>
              <MessageContent>
                <MessageHeader>RaceCoach</MessageHeader>
                <Marker role="status" className="rounded-2xl bg-background/70 px-3 py-2.5 shadow-sm">
                  <MarkerIcon>
                    <Spinner className="text-primary" />
                  </MarkerIcon>
                  <MarkerContent className="rc-coach-shimmer font-medium text-foreground/80">
                    Reading your session…
                  </MarkerContent>
                </Marker>
              </MessageContent>
            </Message>
          ) : null}

          <div ref={bottomRef} />
        </div>

        <form onSubmit={onSubmit} className="border-t border-border/70 bg-[#f3efe6]/80 p-4 backdrop-blur sm:p-5">
          <div className="flex gap-2">
            <Input
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder="Ask the coach about this activity…"
              disabled={!activity || sending}
              className="h-11 bg-background/90"
            />
            <Button
              type="submit"
              disabled={!activity || sending || !draft.trim()}
              className="h-11 shrink-0 px-4"
              aria-label="Send message"
            >
              {sending ? <Spinner className="size-4" /> : <SendHorizontal className="size-4" />}
            </Button>
          </div>
          {error ? (
            <p className="mt-2 text-xs text-destructive" role="alert">
              {error}
            </p>
          ) : (
            <p className="mt-2 text-xs text-muted-foreground">
              Grounded in this session&apos;s features and summary. Saved for this activity.
            </p>
          )}
        </form>
      </aside>
    </div>
  )
}
