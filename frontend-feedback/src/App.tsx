import { useState, useEffect } from 'react'
import logo from './assets/logo.png'

function App() {
  const [rating, setRating] = useState(0)
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [isValidating, setIsValidating] = useState(true)
  const [isValidated, setIsValidated] = useState(false)
  const [error, setError] = useState('')
  const [hoverRating, setHoverRating] = useState(0)
  
  const [orderId, setOrderId] = useState<string | null>(null)
  const [phone, setPhone] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const o = params.get('o') || params.get('order_id')
    const p = params.get('p') || params.get('phone')
    setOrderId(o)
    setPhone(p)

    if (o && p) {
      validateRequest(o, p)
    } else {
      setIsValidating(false)
      setIsValidated(false)
    }
  }, [])

  const validateRequest = async (o: string, p: string) => {
    try {
      const apiUrl = import.meta.env.VITE_API_URL || 'https://mi-tech-api.millennialperfumer.in';
      const response = await fetch(`${apiUrl}/api/feedback/validate?o=${encodeURIComponent(o)}&p=${encodeURIComponent(p)}`)
      const data = await response.json()

      if (response.ok) {
        if (data.already_submitted) {
          setSubmitted(true)
        }
        setIsValidated(true)
      } else {
        setIsValidated(false)
      }
    } catch (err) {
      setError('Connection error. Please try again.')
    } finally {
      setIsValidating(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (rating === 0) {
      setError('Please select a rating')
      return
    }

    setLoading(true)
    setError('')

    try {
      const apiUrl = import.meta.env.VITE_API_URL || 'https://mi-tech.millennialperfumer.in';
      const response = await fetch(`${apiUrl}/api/feedback/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          order_id: orderId ? parseInt(orderId) : 0,
          rating: rating,
          message: message,
          phone: phone,
        }),
      })

      if (response.ok) {
        setSubmitted(true)
      } else {
        setError('Submission failed. Please try again.')
      }
    } catch (err) {
      setError('Technical error. Check your connection.')
    } finally {
      setLoading(false)
    }
  }

  // Unauthorized / Invalid Link State
  if (!isValidating && !isValidated) {
    return (
      <div className="min-h-screen w-full flex items-center justify-center p-6 bg-[#0a0a0c] font-montserrat">
        <div className="max-w-md w-full bg-white/5 backdrop-blur-3xl rounded-[2.5rem] p-10 shadow-2xl border border-rose-500/20 text-center animate-in fade-in zoom-in duration-500">
          <div className="mb-8 inline-flex items-center justify-center w-24 h-24 bg-rose-500/10 rounded-full border border-rose-500/20">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-10 w-10 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white mb-4 tracking-tight uppercase">Access Denied</h2>
          <p className="text-zinc-400 text-base font-light leading-relaxed">
            This feedback link is invalid or has expired. Please check your latest message from Millennial Perfumer.
          </p>
        </div>
      </div>
    )
  }

  // Loading / Validating State
  if (isValidating) {
    return (
      <div className="min-h-screen w-full flex items-center justify-center bg-[#0a0a0c]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-[#10b981]/20 border-t-[#10b981] rounded-full animate-spin"></div>
          <p className="text-zinc-500 text-xs uppercase tracking-[0.2em] animate-pulse">Authenticating...</p>
        </div>
      </div>
    )
  }

  if (submitted) {
    return (
      <div className="min-h-screen w-full flex items-center justify-center p-6 bg-[#0a0a0c] font-montserrat text-center">
        <div className="max-w-md w-full bg-white/[0.03] backdrop-blur-3xl rounded-[2.5rem] p-10 shadow-2xl border border-[#10b981]/20">
          <div className="mb-8 inline-flex items-center justify-center w-24 h-24 bg-[#10b981]/10 rounded-full border border-[#10b981]/20">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-10 w-10 text-[#10b981]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-3xl font-bold text-white mb-4 tracking-tight uppercase">EXPERIENCE SAVED</h2>
          <p className="text-zinc-200 text-lg font-light leading-relaxed mb-8">
            Your words truly motivate us. Thank you for helping us refine our craft.
          </p>
          <button 
            onClick={() => setSubmitted(false)}
            className="text-[#10b981] text-xs uppercase tracking-[0.2em] font-semibold hover:opacity-80 transition-all underline underline-offset-8"
          >
            Redo your experience
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen w-full flex flex-col items-center justify-center p-6 bg-[#0a0a0c] font-montserrat selection:bg-white/20">
      {/* Dynamic Background Accents */}
      <div className="fixed top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] pointer-events-none"></div>
      <div className="fixed bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-indigo-600/10 rounded-full blur-[120px] pointer-events-none"></div>

      {/* Brand Header */}
      <div className="mb-6 text-center animate-in fade-in slide-in-from-top-4 duration-700">
        <img src={logo} alt="Millennial Perfumer" className="h-20 md:h-24 w-auto opacity-100 drop-shadow-[0_0_30px_rgba(16,185,129,0.15)]" />
      </div>

      {/* Main Experience Card */}
      <div className="max-w-md w-full bg-white/[0.03] backdrop-blur-[40px] rounded-[3rem] p-8 md:p-10 shadow-[0_32px_64px_-16px_rgba(0,0,0,0.6)] border border-[#10b981]/20 relative overflow-hidden transition-all duration-500 hover:border-[#10b981]/40">
        <div className="relative z-10 text-center">
          <h2 className="text-2xl md:text-3xl font-bold text-white mb-4 leading-tight tracking-[0.05em] uppercase text-center">
            Your words truly motivate us.
          </h2>
          
          <p className="text-zinc-200 mb-10 text-base font-light leading-relaxed text-center">
            If you loved your latest fragrance, we'd appreciate a rating and a few words about your journey with it.
          </p>

          <form onSubmit={handleSubmit} className="space-y-8">
            {/* Elegant Star Rating */}
            <div className="flex justify-between items-center px-1">
              {[1, 2, 3, 4, 5].map((s) => (
                <button
                  key={s}
                  type="button"
                  onMouseEnter={() => setHoverRating(s)}
                  onMouseLeave={() => setHoverRating(0)}
                  onClick={() => setRating(s)}
                  className="relative group transition-all duration-300 transform"
                >
                  <span className={`text-5xl transition-all duration-300 ${
                    (hoverRating || rating) >= s 
                      ? 'text-[#10b981] drop-shadow-[0_0_15px_rgba(16,185,129,0.4)]' 
                      : 'text-zinc-800'
                  }`}>
                    ★
                  </span>
                  {rating === s && (
                    <span className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-1 h-1 bg-[#10b981] rounded-full"></span>
                  )}
                </button>
              ))}
            </div>

            {/* Premium Message Input */}
            <div className="space-y-2 text-left">
              <label className="text-[10px] uppercase tracking-[0.2em] text-zinc-500 font-bold ml-1">Comments</label>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="Our creations are for those who seek the extraordinary. Tell us your experience..."
                className="w-full bg-white/[0.04] border border-white/5 rounded-2xl p-5 min-h-[140px] text-white placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-[#10b981]/30 transition-all duration-300 shadow-inner resize-none font-light leading-relaxed"
              />
            </div>

            {error && (
              <p className="text-rose-500 text-xs text-center font-medium animate-pulse">{error}</p>
            )}

            <button
              type="submit"
              disabled={loading}
              className={`group relative w-full py-5 rounded-2xl bg-white text-black font-semibold shadow-2xl hover:bg-zinc-100 active:scale-[0.98] transition-all duration-300 overflow-hidden ${
                loading ? 'opacity-50 cursor-not-allowed' : ''
              }`}
            >
              <span className="relative z-10 flex items-center justify-center">
                {loading ? (
                  <svg className="animate-spin h-5 w-5 text-black" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : 'Submit Experience'}
              </span>
            </button>
          </form>
        </div>
      </div>
      
      {/* Minimal Footer */}
      <footer className="mt-12 text-zinc-600 text-[10px] uppercase tracking-[0.3em] flex items-center gap-4 opacity-50">
        <span className="w-8 h-[1px] bg-zinc-800"></span>
        Handcrafted with Passion
        <span className="w-8 h-[1px] bg-zinc-800"></span>
      </footer>
    </div>
  )
}

export default App
