import { Link } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import AuthLayout from '../auth-layout'
import { OtpForm } from './components/otp-form'

export default function Otp() {
  return (
    <AuthLayout>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-base tracking-tight'>
            Two-factor Authentication
          </CardTitle>
          <CardDescription>
            Enter the 6-digit code from your authenticator app, or use a
            recovery code.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <OtpForm />
        </CardContent>
        <CardFooter className='justify-center'>
          <p className='text-muted-foreground text-center text-sm'>
            <Link
              to='/sign-in'
              className='hover:text-primary underline underline-offset-4'
            >
              Back to sign in
            </Link>
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
