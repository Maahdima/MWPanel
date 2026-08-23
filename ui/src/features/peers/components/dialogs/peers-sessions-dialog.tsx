import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { format, parseISO } from 'date-fns'
import {
  ArrowDownIcon,
  ArrowUpIcon,
  Clock3Icon,
  HistoryIcon,
  MapPinIcon,
  RadioIcon,
  TimerIcon,
} from 'lucide-react'
import { Peer, PeerSession, PeerSessions } from '@/schema/peers.ts'
import { usePeerSessionsQuery } from '@/hooks/peers/usePeerSessionsQuery.ts'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ColoredBadge } from '@/features/shared-components/status-badge.tsx'

type SessionFilter = 'all' | 'ongoing' | 'ended'

type Props = {
  open: boolean
  onOpenChange: (state: boolean) => void
  currentRow: Peer
}

function formatDateTime(value: string) {
  try {
    return format(parseISO(value), 'MMM d, yyyy HH:mm:ss')
  } catch {
    return value
  }
}

function formatDuration(totalSeconds: number) {
  const seconds = Math.max(0, Math.floor(totalSeconds))
  const days = Math.floor(seconds / 86_400)
  const hours = Math.floor((seconds % 86_400) / 3_600)
  const minutes = Math.floor((seconds % 3_600) / 60)
  const rest = seconds % 60
  const hms = [hours, minutes, rest]
    .map((part) => String(part).padStart(2, '0'))
    .join(':')
  return days > 0 ? `${days}d ${hms}` : hms
}

function sessionSeconds(session: PeerSession, now: number) {
  const start = new Date(session.connected_at).getTime()
  const end = session.disconnected_at
    ? new Date(session.disconnected_at).getTime()
    : now
  return Math.max(0, Math.floor((end - start) / 1000))
}

export function PeersSessionsDialog({ open, onOpenChange, currentRow }: Props) {
  const [filter, setFilter] = useState<SessionFilter>('all')
  const [now, setNow] = useState(() => Date.now())

  const { data, isLoading, isFetching, isError } = usePeerSessionsQuery(
    currentRow.id,
    { enabled: open }
  )

  useEffect(() => {
    if (!open) return
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [open])

  useEffect(() => {
    if (open) return
    setFilter('all')
  }, [open])

  const sessions = data?.sessions ?? []
  const filteredSessions = useMemo(() => {
    if (filter === 'all') return sessions
    return sessions.filter((session) => session.status === filter)
  }, [filter, sessions])

  const showSkeleton = (isLoading || isFetching) && !data
  const endedCount =
    (data?.total_sessions ?? 0) - (data?.ongoing_sessions ?? 0)

  const body = (() => {
    if (showSkeleton) {
      return <SessionsSkeleton />
    }

    if (isError) {
      return (
        <div className='text-muted-foreground rounded-lg border border-dashed p-8 text-center text-sm'>
          Could not load session history. Try again in a moment.
        </div>
      )
    }

    return (
      <>
        <SessionSummary data={data} />

        <Tabs
          value={filter}
          onValueChange={(value) => setFilter(value as SessionFilter)}
        >
          <TabsList>
            <TabsTrigger value='all'>
              <span>All</span>
              <span className='text-muted-foreground'>
                {data?.total_sessions ?? 0}
              </span>
            </TabsTrigger>
            <TabsTrigger value='ongoing'>
              <span>Ongoing</span>
              <span className='text-muted-foreground'>
                {data?.ongoing_sessions ?? 0}
              </span>
            </TabsTrigger>
            <TabsTrigger value='ended'>
              <span>Ended</span>
              <span className='text-muted-foreground'>{endedCount}</span>
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {filteredSessions.length === 0 ? (
          <EmptySessions hasAny={sessions.length > 0} />
        ) : (
          <ScrollArea className='h-[min(28rem,46vh)] pr-3'>
            <ol className='relative space-y-4 pb-2'>
              {filteredSessions.map((session, index) => (
                <SessionRow
                  key={session.id}
                  session={session}
                  now={now}
                  isLast={index === filteredSessions.length - 1}
                />
              ))}
            </ol>
          </ScrollArea>
        )}
      </>
    )
  })()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='flex max-h-[85vh] flex-col gap-5 overflow-hidden sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <HistoryIcon className='size-5' />
            Session History
          </DialogTitle>
          <DialogDescription>
            Connection timeline for{' '}
            <span className='text-foreground font-medium'>
              {currentRow.name}
            </span>
            <span className='text-muted-foreground'>
              {' '}
              · {currentRow.allowed_address}
            </span>
          </DialogDescription>
        </DialogHeader>

        {body}
      </DialogContent>
    </Dialog>
  )
}

function SessionSummary({ data }: { data?: PeerSessions }) {
  return (
    <div className='grid gap-3 sm:grid-cols-3'>
      <SummaryCard
        icon={<RadioIcon className='size-4' />}
        label='Sessions'
        value={String(data?.total_sessions ?? 0)}
        hint={
          data?.ongoing_sessions
            ? `${data.ongoing_sessions} ongoing`
            : 'No active session'
        }
      />
      <SummaryCard
        icon={<TimerIcon className='size-4' />}
        label='Time online'
        value={data?.total_duration_label ?? '00:00:00'}
        hint='All recorded sessions'
      />
      <SummaryCard
        icon={<ArrowDownIcon className='size-4' />}
        label='Session traffic'
        value={data?.total_usage ?? '0 B'}
        hint={`${data?.total_download_usage ?? '0 B'} down · ${data?.total_upload_usage ?? '0 B'} up`}
      />
    </div>
  )
}

function SummaryCard({
  icon,
  label,
  value,
  hint,
}: {
  icon: ReactNode
  label: string
  value: string
  hint: string
}) {
  return (
    <div className='bg-muted/40 rounded-xl border p-4'>
      <div className='text-muted-foreground flex items-center gap-2 text-xs font-medium tracking-wide uppercase'>
        {icon}
        {label}
      </div>
      <div className='mt-2 text-lg font-semibold tracking-tight'>{value}</div>
      <div className='text-muted-foreground mt-1 text-xs'>{hint}</div>
    </div>
  )
}

function SessionRow({
  session,
  now,
  isLast,
}: {
  session: PeerSession
  now: number
  isLast: boolean
}) {
  const ongoing = session.status === 'ongoing'
  const duration = formatDuration(sessionSeconds(session, now))

  return (
    <li className='relative flex gap-4'>
      <div className='flex w-4 flex-col items-center'>
        <span className='relative mt-1.5 flex size-3 items-center justify-center'>
          <span
            className={
              ongoing
                ? 'size-3 rounded-full bg-green-500 shadow-[0_0_0_4px] shadow-green-500/20'
                : 'bg-muted-foreground/40 size-3 rounded-full'
            }
          />
          {ongoing && (
            <span className='absolute size-3 animate-ping rounded-full bg-green-500/70' />
          )}
        </span>
        {!isLast && <span className='bg-border mt-1 w-px flex-1' />}
      </div>

      <div className='bg-card mb-1 flex-1 rounded-xl border p-4'>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div className='space-y-1'>
            <div className='flex flex-wrap items-center gap-2'>
              {ongoing ? (
                <ColoredBadge color='green' text='ongoing' />
              ) : (
                <ColoredBadge color='gray' text='ended' />
              )}
              <span className='text-sm font-medium'>{duration}</span>
            </div>
            <div className='text-muted-foreground flex flex-wrap items-center gap-1 text-sm'>
              <Clock3Icon className='size-3.5' />
              <span>{formatDateTime(session.connected_at)}</span>
              <span>→</span>
              <span>
                {session.disconnected_at
                  ? formatDateTime(session.disconnected_at)
                  : 'now'}
              </span>
            </div>
          </div>
          <Badge variant='outline' className='font-normal'>
            {session.total_usage}
          </Badge>
        </div>

        <div className='mt-3 flex flex-wrap gap-x-4 gap-y-2 text-sm'>
          <span className='text-muted-foreground inline-flex items-center gap-1.5'>
            <MapPinIcon className='size-3.5' />
            {session.endpoint ?? 'Endpoint unknown'}
          </span>
          <span className='inline-flex items-center gap-1.5'>
            <ArrowDownIcon className='text-muted-foreground size-3.5' />
            {session.download_usage}
          </span>
          <span className='inline-flex items-center gap-1.5'>
            <ArrowUpIcon className='text-muted-foreground size-3.5' />
            {session.upload_usage}
          </span>
        </div>
      </div>
    </li>
  )
}

function EmptySessions({ hasAny }: { hasAny: boolean }) {
  return (
    <div className='flex flex-col items-center justify-center rounded-xl border border-dashed px-6 py-12 text-center'>
      <div className='bg-muted mb-3 rounded-full p-3'>
        <HistoryIcon className='text-muted-foreground size-5' />
      </div>
      <p className='text-sm font-medium'>
        {hasAny ? 'No sessions in this filter' : 'No sessions recorded yet'}
      </p>
      <p className='text-muted-foreground mt-1 max-w-sm text-sm'>
        {hasAny
          ? 'Try another tab to see the rest of this peer’s history.'
          : 'History starts automatically the next time this peer connects.'}
      </p>
    </div>
  )
}

function SessionsSkeleton() {
  return (
    <div className='space-y-4'>
      <div className='grid gap-3 sm:grid-cols-3'>
        <Skeleton className='h-24 rounded-xl' />
        <Skeleton className='h-24 rounded-xl' />
        <Skeleton className='h-24 rounded-xl' />
      </div>
      <Skeleton className='h-9 w-64' />
      <div className='space-y-3'>
        <Skeleton className='h-28 rounded-xl' />
        <Skeleton className='h-28 rounded-xl' />
        <Skeleton className='h-28 rounded-xl' />
      </div>
    </div>
  )
}
