import { useState } from 'react'
import type { FormEvent } from 'react'
import { ArrowLeft, ArrowRight, Eye, EyeOff, ShieldCheck } from 'lucide-react'
import fullLogo from '../assets/mi-black-full.png'
import { API_BASE } from '../lib/api'

type AuthScreenProps = {
  onLogin: (token: string) => void
}

function getErrorMessage(response: Response, fallback: string) {
  return response.text().then((body) => {
    if (!body) return fallback
    try {
      const parsed = JSON.parse(body) as { message?: string; error?: string }
      return parsed.message || parsed.error || fallback
    } catch {
      return body
    }
  })
}

export function AuthScreen({ onLogin }: AuthScreenProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [otp, setOtp] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [requiresOtp, setRequiresOtp] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const handleLogin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      const response = await fetch(`${API_BASE}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      if (!response.ok) {
        throw new Error(await getErrorMessage(response, 'Unable to sign in'))
      }

      const data = (await response.json()) as { token?: string; requires_2fa?: boolean }
      if (data.requires_2fa) {
        setRequiresOtp(true)
      } else if (data.token) {
        onLogin(data.token)
      } else {
        throw new Error('The sign-in response did not include a session token')
      }
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to sign in')
    } finally {
      setIsLoading(false)
    }
  }

  const handleOtp = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      const response = await fetch(`${API_BASE}/api/auth/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, otp }),
      })

      if (!response.ok) {
        throw new Error(await getErrorMessage(response, 'That verification code is invalid or expired'))
      }

      const data = (await response.json()) as { token?: string }
      if (!data.token) throw new Error('The verification response did not include a session token')
      onLogin(data.token)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to verify the code')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="auth-shell">
      <section className="auth-card" aria-labelledby="auth-heading">
        <img className="auth-logo" src={fullLogo} alt="Millennial Perfumer" />
        <p className="eyebrow">MP Workspace</p>
        <h1 id="auth-heading">{requiresOtp ? 'Verify your identity' : 'Welcome back'}</h1>
        <p className="auth-intro">
          {requiresOtp
            ? 'Enter the six-digit code sent to your registered WhatsApp number.'
            : 'Sign in to manage orders, GST, inventory, and customer operations.'}
        </p>

        {requiresOtp ? (
          <form className="auth-form" onSubmit={handleOtp}>
            <label className="form-field" htmlFor="auth-otp">
              <span>Verification code</span>
              <input
                id="auth-otp"
                value={otp}
                onChange={(event) => setOtp(event.target.value)}
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="000000"
                maxLength={6}
                required
              />
            </label>
            {error && <p className="auth-error" role="alert">{error}</p>}
            <button className="auth-submit" type="submit" disabled={isLoading}>
              <ShieldCheck size={17} aria-hidden="true" />
              {isLoading ? 'Verifying…' : 'Verify and continue'}
              <ArrowRight size={17} aria-hidden="true" />
            </button>
            <button className="auth-back" type="button" onClick={() => setRequiresOtp(false)}>
              <ArrowLeft size={15} aria-hidden="true" />
              Use a different account
            </button>
          </form>
        ) : (
          <form className="auth-form" onSubmit={handleLogin}>
            <label className="form-field" htmlFor="auth-username">
              <span>Username</span>
              <input
                id="auth-username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                autoComplete="username"
                placeholder="admin"
                required
              />
            </label>
            <label className="form-field" htmlFor="auth-password">
              <span>Password</span>
              <span className="password-field">
                <input
                  id="auth-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  placeholder="Enter your password"
                  required
                />
                <button
                  className="password-toggle"
                  type="button"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  onClick={() => setShowPassword((visible) => !visible)}
                >
                  {showPassword ? <EyeOff size={17} aria-hidden="true" /> : <Eye size={17} aria-hidden="true" />}
                </button>
              </span>
            </label>
            {error && <p className="auth-error" role="alert">{error}</p>}
            <button className="auth-submit" type="submit" disabled={isLoading}>
              {isLoading ? 'Signing in…' : 'Sign in'}
              <ArrowRight size={17} aria-hidden="true" />
            </button>
          </form>
        )}
      </section>
    </main>
  )
}
