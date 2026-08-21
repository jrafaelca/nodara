import { LoginForm } from '@/components/login-form'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/lib/auth.jsx'

export default function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (event) => {
    event.preventDefault()
    setError('')
    setLoading(true)
    const form = new FormData(event.currentTarget)
    try {
      const user = await login(form.get('identifier'), form.get('password'))
      navigate(user.password_change_required ? '/change-password' : '/')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="flex min-h-svh flex-col items-center justify-center bg-black p-6 text-white md:p-10">
      <div className="w-full max-w-sm"><LoginForm onSubmit={handleSubmit} error={error} loading={loading} /></div>
    </main>
  )
}
