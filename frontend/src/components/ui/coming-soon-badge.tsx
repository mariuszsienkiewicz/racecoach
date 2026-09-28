import { cn } from '@/lib/utils'

type ComingSoonBadgeProps = {
  className?: string
}

export function ComingSoonBadge({ className }: ComingSoonBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md border border-border/80 bg-background/80 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-muted-foreground',
        className,
      )}
    >
      Coming soon
    </span>
  )
}
