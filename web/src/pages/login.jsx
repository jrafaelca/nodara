import { LoginForm } from '@/components/login-form'

export default function LoginPage() {
  return (
    <main className="flex min-h-svh flex-col items-center justify-center bg-black p-6 text-white md:p-10">
      <div className="w-full max-w-sm"><LoginForm /></div>
    </main>
  )
}
