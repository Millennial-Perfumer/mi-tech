import { API_BASE } from './api';
import React, { useState } from 'react';
import fullLogoDark from './assets/full_logo_dark_theme.png';

interface LoginProps {
  onLogin: (token: string) => void;
}

export const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await fetch(`${API_BASE}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const msg = await response.text();
        throw new Error(msg || 'Invalid credentials');
      }

      const data = await response.json();
      onLogin(data.token);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      background: 'radial-gradient(circle at top right, #1e293b, #0f172a, #020617)',
      padding: '2rem',
      position: 'relative',
      overflow: 'hidden'
    }}>
      {/* Decorative Blur Blobs */}
      <div style={{ position: 'absolute', top: '-10%', right: '-10%', width: '40%', height: '40%', background: 'rgba(14, 165, 233, 0.05)', filter: 'blur(120px)', borderRadius: '50%', zIndex: 1 }}></div>
      <div style={{ position: 'absolute', bottom: '-10%', left: '-10%', width: '30%', height: '30%', background: 'rgba(56, 189, 248, 0.03)', filter: 'blur(100px)', borderRadius: '50%', zIndex: 1 }}></div>

      <div className="card" style={{
        maxWidth: '420px',
        width: '100%',
        padding: '3rem 2.5rem',
        background: 'rgba(255, 255, 255, 0.015)',
        backdropFilter: 'blur(24px)',
        WebkitBackdropFilter: 'blur(24px)',
        border: '1px solid rgba(255, 255, 255, 0.08)',
        borderRadius: '28px',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
        textAlign: 'center',
        zIndex: 2,
        animation: 'fadeIn 0.6s ease-out'
      }}>
        <img src={fullLogoDark} alt="mi-tech" style={{ width: '200px', marginBottom: '2.5rem', filter: 'drop-shadow(0 0 8px rgba(255,255,255,0.1))' }} />
        
        <h2 style={{ marginBottom: '0.5rem', fontSize: '1.75rem', fontWeight: 700, letterSpacing: '-0.02em', color: '#fff' }}>Welcome Back</h2>
        <p style={{ color: 'rgba(255, 255, 255, 0.5)', marginBottom: '2.5rem', fontSize: '0.95rem' }}>
          Please enter your credentials to continue.
        </p>

        <form onSubmit={handleSubmit} style={{ textAlign: 'left' }}>
          <div className="form-group" style={{ marginBottom: '1.5rem' }}>
            <label>Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              required
            />
          </div>
          <div className="form-group" style={{ marginBottom: '2rem' }}>
            <label>Password</label>
            <div style={{ position: 'relative' }}>
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                style={{ paddingRight: '3rem' }}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                style={{
                  position: 'absolute',
                  right: '0.75rem',
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-secondary)',
                  padding: '4px',
                  display: 'flex',
                  alignItems: 'center',
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer'
                }}
              >
                {showPassword ? (
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>
                ) : (
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
                )}
              </button>
            </div>
          </div>

          {error && (
            <div style={{
              backgroundColor: '#fee2e2',
              color: '#991b1b',
              padding: '0.75rem',
              borderRadius: '8px',
              fontSize: '0.875rem',
              marginBottom: '1.5rem',
              textAlign: 'center',
              border: '1px solid #fecaca'
            }}>
              {error}
            </div>
          )}

          <button
            type="submit"
            className="btn-primary"
            style={{
              width: '100%',
              padding: '1rem',
              fontSize: '1rem',
              fontWeight: 600,
              borderRadius: '14px',
              marginTop: '1rem',
              background: 'linear-gradient(135deg, #0ea5e9, #0284c7)',
              border: 'none',
              boxShadow: '0 8px 20px -4px rgba(14, 165, 233, 0.4)',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
            disabled={loading}
          >
            {loading ? 'Authenticating...' : 'Sign In'}
          </button>
        </form>
      </div>
      
      <p style={{ marginTop: '2.5rem', color: 'rgba(255, 255, 255, 0.3)', fontSize: '0.85rem', zIndex: 2 }}>
          &copy; {new Date().getFullYear()} mi-tech. All rights reserved.
      </p>
    </div>
  );
};
