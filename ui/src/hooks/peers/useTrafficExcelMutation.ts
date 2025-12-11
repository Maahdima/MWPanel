import { useMutation } from '@tanstack/react-query'
import { exportTrafficExcel } from '@/api/peers.ts'

export const useTrafficExcelMutation = () => {
  return useMutation({
    mutationFn: exportTrafficExcel,
  })
}
