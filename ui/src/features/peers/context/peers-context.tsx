import React, { useState } from 'react'
import { Peer } from '@/schema/peers.ts'
import useDialogState from '@/hooks/use-dialog-state'

type PeersDialogType =
  | 'add'
  | 'edit'
  | 'delete'
  | 'share'
  | 'qrCode'
  | 'config'
  | 'show_config'
  | 'download_config'

interface PeersContextType {
  open: PeersDialogType | null
  setOpen: (str: PeersDialogType | null) => void
  currentRow: Peer | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Peer | null>>
}

const PeersContext = React.createContext<PeersContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function PeersProvider({ children }: Props) {
  const [open, setOpen] = useDialogState<PeersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Peer | null>(null)

  return (
    <PeersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </PeersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const usePeers = () => {
  const peersContext = React.useContext(PeersContext)

  if (!peersContext) {
    throw new Error('usePeers has to be used within <PeersContext>')
  }

  return peersContext
}
