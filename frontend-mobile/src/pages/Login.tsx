import React, { useState } from 'react';
import { useAuthStore } from '../store/useAuthStore';
import { API_BASE } from '../api';
import { LogIn, ShieldCheck, KeyRound, ArrowRight } from 'lucide-react';

export const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [showOTP, setShowOTP] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const setToken = useAuthStore((state) => state.setToken);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch(`${API_BASE}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      
      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || 'Login failed');
      }

      const data = await response.json();

      if (data.requires_2fa) {
        setShowOTP(true);
      } else if (data.token) {
        setToken(data.token);
      } else {
        setError('Invalid server response');
      }
    } catch (err: any) {
      setError(err.message || 'Network error. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch(`${API_BASE}/api/auth/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, otp }),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || 'Verification failed');
      }

      const data = await response.json();
      if (data.token) {
        setToken(data.token);
      } else {
        setError('Invalid OTP or session expired');
      }
    } catch (err: any) {
      setError(err.message || 'Network error. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="mobile-shell" style={{ justifyContent: 'center', padding: '0 1.5rem', minHeight: '100vh' }}>
      <div className="glass-card" style={{ width: '100%', padding: '3rem 1.75rem 2.5rem' }}>
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <div style={{ 
            width: '72px', 
            height: '72px', 
            background: 'var(--accent-subtle)', 
            borderRadius: '22px', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            margin: '0 auto 1.75rem',
            color: 'var(--accent-color)',
            boxShadow: 'var(--glow-accent)',
            border: '1px solid rgba(16, 185, 129, 0.2)'
          }}>
            {showOTP ? <KeyRound size={36} strokeWidth={1.5} /> : <ShieldCheck size={36} strokeWidth={1.5} />}
          </div>
          <h1 style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>
            {showOTP ? 'Security' : 'Mi Tech'}
          </h1>
          <p style={{ fontSize: '0.9rem', opacity: 0.8 }}>
            {showOTP ? 'Enter the verification code' : 'Sign in to your mobile console'}
          </p>
        </div>

        {!showOTP ? (
          <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div className="form-group">
              <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(255,255,255,0.02)',
                  border: '1px solid var(--glass-border)',
                  color: '#fff',
                  padding: '1.1rem',
                  borderRadius: '16px',
                  fontSize: '1rem',
                  outline: 'none',
                }}
                required
              />
            </div>
            <div className="form-group">
              <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(255,255,255,0.02)',
                  border: '1px solid var(--glass-border)',
                  color: '#fff',
                  padding: '1.1rem',
                  borderRadius: '16px',
                  fontSize: '1rem',
                  outline: 'none',
                }}
                required
              />
            </div>

            {error && <p style={{ color: '#ef4444', fontSize: '0.8rem', textAlign: 'center', marginTop: '0.5rem', fontWeight: 600 }}>{error}</p>}

            <button
              type="submit"
              disabled={isLoading}
              style={{
                background: 'linear-gradient(135deg, var(--accent-color), #059669)',
                color: '#fff',
                border: 'none',
                padding: '1.1rem',
                borderRadius: '16px',
                fontWeight: 700,
                fontSize: '1.05rem',
                marginTop: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.75rem',
                boxShadow: '0 8px 24px rgba(16, 185, 129, 0.3)',
              }}
            >
              {isLoading ? 'Verifying...' : (
                <>
                  <LogIn size={20} />
                  Access Console
                </>
              )}
            </button>
          </form>
        ) : (
          <form onSubmit={handleVerifyOTP} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div className="form-group">
              <input
                type="text"
                placeholder="Verification Code"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(255,255,255,0.02)',
                  border: '1px solid var(--glass-border)',
                  color: '#fff',
                  padding: '1.1rem',
                  borderRadius: '16px',
                  fontSize: '1.2rem',
                  textAlign: 'center',
                  letterSpacing: '0.5em',
                  outline: 'none',
                }}
                maxLength={6}
                autoFocus
                required
              />
            </div>

            {error && <p style={{ color: '#ef4444', fontSize: '0.8rem', textAlign: 'center', marginTop: '0.5rem', fontWeight: 600 }}>{error}</p>}

            <button
              type="submit"
              disabled={isLoading}
              style={{
                background: 'linear-gradient(135deg, var(--secondary-color), #0891b2)',
                color: '#fff',
                border: 'none',
                padding: '1.1rem',
                borderRadius: '16px',
                fontWeight: 700,
                fontSize: '1.05rem',
                marginTop: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.75rem',
                boxShadow: '0 8px 24px rgba(6, 182, 212, 0.3)',
              }}
            >
              {isLoading ? 'Verifying Code...' : (
                <>
                  <ArrowRight size={20} />
                  Verify & Log In
                </>
              )}
            </button>
            <button 
              type="button" 
              onClick={() => setShowOTP(false)} 
              style={{ background: 'transparent', border: 'none', color: 'var(--text-tertiary)', fontSize: '0.85rem' }}
            >
              Back to Login
            </button>
          </form>
        )}
      </div>
    </div>
  );
};
