import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export default function OnlineUsersSkeleton() {
  return (
    <Card className='col-span-1 lg:col-span-2'>
      <CardContent className='px-6'>
        <h2 className='mb-4 border-white/10 pb-2 text-lg font-semibold'>
          Recently Online Users
        </h2>
        <table className='w-full text-sm'>
          <thead>
            <tr className='text-center text-gray-400'>
              <th className='justify-center py-2'>Name</th>
              <th className='justify-center py-2'>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {[...Array(5)].map((_, idx) => (
              <tr
                key={idx}
                className='justify-center border-t border-white/5 transition'
              >
                <td className='flex items-center justify-center gap-3 py-4'>
                  <Skeleton className='h-8 w-8 rounded-full' />
                  <Skeleton className='h-4 w-24' />
                </td>
                <td className='py-2'>
                  <Skeleton className='mx-auto h-4 w-16' />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </CardContent>
    </Card>
  )
}
