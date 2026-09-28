import { useRef, useState, type ChangeEvent, type DragEvent } from 'react'
import { Link2, Upload } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { ComingSoonBadge } from '@/components/ui/coming-soon-badge'
import type { DeviceConnection } from '@/types/dashboard'
import { cn } from '@/lib/utils'

type ImportPanelProps = {
  devices: DeviceConnection[]
  onUpload: (file: File) => Promise<void>
  /** Reserved for Coros/Garmin OAuth - UI is ready, wiring comes later. */
  onConnect?: (provider: DeviceConnection['provider']) => Promise<void>
}

export function ImportPanel({ devices, onUpload }: ImportPanelProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  async function handleFile(file: File | undefined) {
    if (!file) {
      return
    }
    if (!file.name.toLowerCase().endsWith('.fit')) {
      setMessage('Please choose a .fit file.')
      return
    }

    setBusy(true)
    setMessage(null)
    try {
      await onUpload(file)
      setMessage(`Uploaded “${file.name}”, queued for analysis.`)
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Upload failed.')
    } finally {
      setBusy(false)
    }
  }

  function onDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault()
    setDragging(false)
    void handleFile(event.dataTransfer.files?.[0])
  }

  function onBrowse(event: ChangeEvent<HTMLInputElement>) {
    void handleFile(event.target.files?.[0])
    event.target.value = ''
  }

  return (
    <section
      id="import"
      className="scroll-mt-8 space-y-6 rounded-3xl border border-border/70 bg-background/75 p-6 backdrop-blur sm:p-8"
    >
      <div>
        <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Import</p>
        <h2 className="mt-2 font-display text-2xl font-semibold tracking-tight">
          Bring workouts in
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
          Upload a `.fit` file now. Provider sync (Coros, Garmin, …) keeps the same desk layout -
          OAuth hooks up later without redesigning this panel.
        </p>
      </div>

      <div
        onDragOver={(event) => {
          event.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
        className={cn(
          'rounded-2xl border border-dashed px-5 py-8 text-center transition-colors',
          dragging ? 'border-primary bg-primary/5' : 'border-border/80 bg-[#f7f4ec]/40',
        )}
      >
        <Upload className="mx-auto h-6 w-6 text-primary" />
        <p className="mt-3 font-medium">Drop a .fit file here</p>
        <p className="mt-1 text-sm text-muted-foreground">or browse from your computer</p>
        <Button
          type="button"
          className="mt-4"
          disabled={busy}
          onClick={() => inputRef.current?.click()}
        >
          {busy ? 'Uploading…' : 'Choose .fit file'}
        </Button>
        <input
          ref={inputRef}
          type="file"
          accept=".fit,application/octet-stream"
          className="hidden"
          onChange={onBrowse}
        />
      </div>

      {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}

      <div className="space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
            Device sync
          </p>
          <ComingSoonBadge />
        </div>
        {devices.map((device) => (
          <div
            key={device.id}
            className="flex flex-col gap-3 rounded-2xl border border-border/70 px-4 py-4 opacity-80 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p className="font-medium">{device.label}</p>
              <p className="text-sm text-muted-foreground">
                {device.connected
                  ? `Connected · last sync ${
                      device.lastSyncAt
                        ? new Intl.DateTimeFormat('en', {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          }).format(new Date(device.lastSyncAt))
                        : 'just now'
                    }`
                  : 'Not connected'}
              </p>
            </div>
            <Button type="button" variant="outline" disabled>
              <Link2 className="h-4 w-4" />
              {device.connected ? 'Connected' : `Connect ${device.label}`}
            </Button>
          </div>
        ))}
      </div>
    </section>
  )
}
