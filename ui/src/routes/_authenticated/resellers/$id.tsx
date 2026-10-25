import { createFileRoute } from '@tanstack/react-router'
import ResellerDetail from '@/features/resellers/detail'

export const Route = createFileRoute('/_authenticated/resellers/$id')({
  component: ResellerDetailRoute,
})
function ResellerDetailRoute() {
  const { id } = Route.useParams()
  return <ResellerDetail id={Number(id)} />
}
