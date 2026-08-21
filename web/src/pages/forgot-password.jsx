import { useState } from 'react'
import { Link } from 'react-router-dom'

import { useAuth } from '@/lib/auth.jsx'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function ForgotPasswordPage() {
  const { forgotPassword } = useAuth()
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    const form = new FormData(event.currentTarget)
    try {
      const data = await forgotPassword(form.get('identifier'))
      setMessage(data.message)
    } finally {
      setLoading(false)
    }
  }
  return <main className="flex min-h-svh items-center justify-center bg-black p-6 text-white"><form onSubmit={submit} className="flex w-full max-w-sm flex-col gap-4"><h1 className="text-xl font-bold">Reset your password</h1><p className="text-sm text-white/60">Enter your email or username and we will send reset instructions.</p><Input className="border-white/20 text-white" name="identifier" placeholder="admin@nodara.dev" required />{message && <p className="text-sm text-white/70">{message}</p>}<Button className="bg-white text-black hover:bg-white/90" disabled={loading}>{loading ? 'Sending…' : 'Send reset link'}</Button><Link className="text-sm text-white/70 hover:text-white" to="/login">Back to login</Link></form></main>
}
