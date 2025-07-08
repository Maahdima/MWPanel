import React, { useState } from 'react'
import { Interface } from '@/schema/interfaces.ts'
import useDialogState from '@/hooks/use-dialog-state'

type InterfacesDialogType = 'add' | 'edit' | 'delete'

interface InterfacesContextType {
  open: InterfacesDialogType | null
  setOpen: (str: InterfacesDialogType | null) => void
  currentRow: Interface | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Interface | null>>
}

const InterfacesContext = React.createContext<InterfacesContextType | null>(
  null
)

interface Props {
  children: React.ReactNode
}

export default function InterfacesProvider({ children }: Props) {
  const [open, setOpen] = useDialogState<InterfacesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Interface | null>(null)

  return (
    <InterfacesContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </InterfacesContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useInterfaces = () => {
  const interfacesContext = React.useContext(InterfacesContext)

  if (!interfacesContext) {
    throw new Error('useInterfaces has to be used within <InterfacesContext>')
  }

  return interfacesContext
}
