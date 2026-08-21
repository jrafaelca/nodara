import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useAuth } from '@/lib/auth.jsx'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function ChangePasswordPage() {
  const { changePassword } = useAuth()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    setLoading(true)
    setError('')
    try {
      await changePassword(form.get('password'), form.get('password_confirmation'))
      navigate('/')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setLoading(false)
    }
  }

  return <main className="flex min-h-svh items-center justify-center bg-black p-6 text-white"><form onSubmit={submit} className="flex w-full max-w-sm flex-col gap-4"><h1 className="text-xl font-bold">Set a new password</h1><p className="text-sm text-white/60">Your temporary password must be replaced before using Nodara.</p><Input className="border-white/20 text-white" name="password" type="password" autoComplete="new-password" placeholder="New password" required /><Input className="border-white/20 text-white" name="password_confirmation" type="password" autoComplete="new-password" placeholder="Confirm password" required />{error && <p className="text-sm text-red-300">{error}</p>}<Button className="bg-white text-black hover:bg-white/90" disabled={loading}>{loading ? 'Saving…' : 'Save password'}</Button></form></main>
}
