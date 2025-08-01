import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

export default function OnlineUsersSkeleton() {
  return (
    <Card className='col-span-1 flex h-full flex-col lg:col-span-2'>
      <CardHeader>
        <h2 className='mb-4 border-white/10 pb-2 text-lg font-semibold'>
          Recently Online Users
        </h2>
      </CardHeader>
      <CardContent className='p-0'>
        <div className='-mt-4'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className='pl-6'>User</TableHead>
                <TableHead className='pr-6 text-right'>Last Seen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {[...Array(5)].map((_, idx) => (
                <TableRow key={idx}>
                  <TableCell className='py-4 pl-6'>
                    <div className='flex items-center gap-3'>
                      <Skeleton className='h-9 w-9 rounded-full' />
                      <Skeleton className='h-4 w-32 rounded-md' />
                    </div>
                  </TableCell>
                  <TableCell className='pr-6 text-right'>
                    <Skeleton className='h-4 w-24 rounded-md' />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
