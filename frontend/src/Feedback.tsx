import React, { useState, useEffect } from 'react';
import { useToast } from './ToastContext';

interface CustomerFeedback {
  id: number;
  order_id: number;
  order_number: string;
  customer_name: string;
  rating: number;
  comment: string;
  created_at: string;
}

interface FeedbackProps {
  API_BASE: string;
  token: string | null;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

const Feedback: React.FC<FeedbackProps> = ({ API_BASE, token, fetchWithAuth }) => {
  const [feedbacks, setFeedbacks] = useState<CustomerFeedback[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { error } = useToast();

  const fetchFeedbacks = async () => {
    if (!token) return;
    setIsLoading(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/feedback`);
      const data = await response.json();
      if (data.success) {
        setFeedbacks(data.feedback || []);
      } else {
        error(data.message || 'Failed to fetch feedback');
      }
    } catch (err) {
      console.error('Failed to fetch feedback:', err);
      error('Network error fetching feedback');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchFeedbacks();
  }, [token]);

  const renderStars = (rating: number) => {
    return Array.from({ length: 5 }, (_, i) => (
      <svg
        key={i}
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill={i < rating ? "#f59e0b" : "none"}
        stroke={i < rating ? "#f59e0b" : "currentColor"}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        style={{ marginRight: '2px' }}
      >
        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
      </svg>
    ));
  };

  return (
    <div className="feedback-container" style={{ padding: '0.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0 }}>Customer Sentiment</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Monitor ratings and reviews across your delivered orders</p>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <div style={{ 
            background: 'var(--bg-input)', 
            padding: '0.75rem 1rem', 
            borderRadius: '12px', 
            border: '1px solid var(--border-color)',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem'
          }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', fontWeight: 600 }}>AVERAGE RATING</span>
            <span style={{ fontWeight: 800, color: 'var(--accent-color)', fontSize: '1.1rem' }}>
              {feedbacks.length > 0 
                ? (feedbacks.reduce((acc, f) => acc + f.rating, 0) / feedbacks.length).toFixed(1) 
                : 'N/A'}
            </span>
          </div>
        </div>
      </div>

      <div style={{ background: 'var(--surface-color)', borderRadius: '16px', border: '1px solid var(--border-color)', overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--border-color)', background: 'rgba(255,255,255,0.02)' }}>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Customer</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Order</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Rating</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Message</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Date</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>Analyzing reviews...</td></tr>
            ) : feedbacks.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>No feedback collected yet.</td></tr>
            ) : (
              feedbacks.map((item, idx) => (
                <tr key={item.id} style={{ borderBottom: idx === feedbacks.length - 1 ? 'none' : '1px solid var(--border-color)' }}>
                  <td style={{ padding: '1.25rem 1.5rem' }}>
                    <div style={{ fontWeight: 600 }}>{item.customer_name}</div>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem' }}>
                    <span style={{ fontSize: '0.875rem', color: 'var(--accent-color)', fontWeight: 500 }}>{item.order_number}</span>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem' }}>
                    <div style={{ display: 'flex' }}>{renderStars(item.rating)}</div>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem' }}>
                    <div style={{ fontSize: '0.875rem', maxWidth: '300px', lineHeight: '1.4' }}>{item.comment || <em style={{opacity: 0.5}}>No comment</em>}</div>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                    {new Date(item.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Feedback;
