import { useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { syncSelectedPeers } from '@/api/peers.ts'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery.ts'
import { useSyncPeersQuery } from '@/hooks/peers/useSyncPeersQuery.ts'
import { useServersListQuery } from '@/hooks/servers/useServersListQuery'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog.tsx'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table.tsx'
import { ServerSelect } from '@/components/server-select'
import { usePeers } from '@/features/peers/context/peers-context.tsx'

export function PeersSyncDialog() {
  const { open, setOpen } = usePeers()
  const queryClient = useQueryClient()
  const enabled = open === 'sync'
  const { data: interfaces = [] } = useInterfacesListQuery()
  const { data: servers = [] } = useServersListQuery()
  const [selectedServer, setSelectedServer] = useState<number>()
  const serverID =
    selectedServer ?? servers.find((server) => server.is_active)?.id
  const [selectedInterface, setSelectedInterface] = useState('all')
  const { data, isLoading, error } = useSyncPeersQuery(
    enabled,
    selectedInterface === 'all' ? undefined : selectedInterface,
    serverID
  )
  const [selected, setSelected] = useState<Record<string, boolean>>({})

  useEffect(() => {
    if (!enabled) return
    setSelected({})
  }, [enabled, selectedInterface, serverID])

  const peerIds = useMemo(
    () =>
      Object.entries(selected)
        .filter(([, checked]) => checked)
        .map(([id]) => id),
    [selected]
  )

  const handleSubmit = async () => {
    try {
      await syncSelectedPeers({ peer_ids: peerIds, server_id: serverID })
      await queryClient.invalidateQueries({ queryKey: ['peers_list'] })
      toast.success('Selected peers synced successfully', { duration: 5000 })
      setOpen(null)
    } catch (error) {
      toast.error('Failed to sync selected peers')
      throw error
    }
  }

  return (
    <Dialog open={enabled} onOpenChange={(value) => !value && setOpen(null)}>
      <DialogContent className='sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>Sync Peers</DialogTitle>
          <DialogDescription>
            Filter by interface, then select only the peers you want to sync.
          </DialogDescription>
        </DialogHeader>

        <DialogBody className='space-y-3'>
          <ServerSelect
            value={serverID}
            onValueChange={(value) => {
              setSelectedServer(value)
              setSelectedInterface('all')
            }}
          />
          {error && (
            <p role='alert' className='text-destructive text-sm'>
              Unable to load peers from this server.
            </p>
          )}
          <div className='w-full max-w-xs'>
            <Select
              value={selectedInterface}
              onValueChange={setSelectedInterface}
            >
              <SelectTrigger className='w-full'>
                <SelectValue placeholder='Filter by interface' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='all'>All interfaces</SelectItem>
                {interfaces
                  .filter((iface) => iface.server_id === serverID)
                  .map((iface) => (
                    <SelectItem key={iface.id} value={iface.name}>
                      {iface.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <div className='overflow-hidden rounded-xl border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='w-12' />
                  <TableHead>Name</TableHead>
                  <TableHead>Interface</TableHead>
                  <TableHead>Allowed Address</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell colSpan={4}>Loading peers...</TableCell>
                  </TableRow>
                ) : data?.length ? (
                  data.map((peer) => (
                    <TableRow key={peer.id}>
                      <TableCell>
                        <Checkbox
                          checked={selected[peer.id] ?? false}
                          onCheckedChange={(checked) =>
                            setSelected((prev) => ({
                              ...prev,
                              [peer.id]: Boolean(checked),
                            }))
                          }
                        />
                      </TableCell>
                      <TableCell>{peer.name}</TableCell>
                      <TableCell>{peer.interface}</TableCell>
                      <TableCell>{peer.allowed_address}</TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={4}>
                      No peers found for this interface.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </DialogBody>

        <DialogFooter>
          <Button variant='outline' onClick={() => setOpen(null)}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!serverID || !peerIds.length}
          >
            Sync Selected
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
