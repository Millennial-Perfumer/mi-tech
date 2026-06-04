import React, { useState, useEffect, useMemo, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import './App.css';

interface TicketsProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

interface Column {
  id: number;
  name: string;
  order: number;
}

interface Task {
  id: number;
  title: string;
  description: string;
  status: string;
  priority: string;
  column_id: number;
  order: number;
  ticket_id?: string;
  created_at: string;
  updated_at: string;
}

interface Board {
  id: number;
  name: string;
  columns: Column[];
}

export const Tickets: React.FC<TicketsProps> = ({ fetchWithAuth }) => {
  const { success, error } = useToast();
  const [board, setBoard] = useState<Board | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<string>('New Issue');
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newTicket, setNewTicket] = useState({ title: '', description: '', priority: 'medium' });
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    try {
      // 1. Find the Support Tickets board
      const boardRes = await fetchWithAuth(`${API_BASE}/api/planner/boards`);
      const boardData = await boardRes.json();
      
      if (boardData.success) {
        const supportBoard = boardData.boards.find((b: Board) => b.name === 'Support Tickets');
        if (supportBoard) {
          setBoard(supportBoard);
          // 2. Fetch tasks for this board
          const tasksRes = await fetchWithAuth(`${API_BASE}/api/planner/tasks?board_id=${supportBoard.id}`);
          const tasksData = await tasksRes.json();
          if (tasksData.success) {
            setTasks(tasksData.tasks);
          }
        }
      }
    } catch (err) {
      error('Failed to load tickets');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateTicket = async () => {
    if (!newTicket.title.trim() || !board) return;
    
    setIsSaving(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/planner/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          board_id: board.id,
          column_id: board.columns.find(c => c.name === 'New Issue')?.id || board.columns[0].id,
          title: newTicket.title,
          description: newTicket.description,
          priority: newTicket.priority
        })
      });
      
      const data = await resp.json();
      if (data.success) {
        success('Ticket raised successfully');
        setIsModalOpen(false);
        setNewTicket({ title: '', description: '', priority: 'medium' });
        loadData();
      }
    } catch (err) {
      error('Failed to create ticket');
    } finally {
      setIsSaving(false);
    }
  };

  const filteredTasks = useMemo(() => {
    return tasks.filter(t => {
      // Find the column name for this task
      const col = board?.columns.find(c => c.id === t.column_id);
      const colName = col?.name || 'Unknown';
      
      const tabMatch = colName === activeTab;
      const searchMatch = t.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
                          (t.ticket_id && t.ticket_id.toLowerCase().includes(searchQuery.toLowerCase()));
      
      return tabMatch && searchMatch;
    }).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
  }, [tasks, activeTab, board, searchQuery]);

  const tabs = board?.columns.sort((a, b) => a.order - b.order) || [];

  if (isLoading && !board) {
    return <div className="loading-shimmer" style={{height: '400px', borderRadius: '24px'}}></div>;
  }

  return (
    <div className="tickets-container page-enter">
      <div className="tickets-header-tabs">
        {tabs.map(tab => (
          <button 
            key={tab.id}
            className={`ticket-tab ${activeTab === tab.name ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.name)}
          >
            {tab.name}
            <span className="count-badge">
              {tasks.filter(t => t.column_id === tab.id).length}
            </span>
          </button>
        ))}
      </div>

      <div className="tickets-controls">
        <div className="search-box-premium">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input 
            type="text" 
            ref={searchInputRef}
            placeholder="Search tickets... (Press /)"
            aria-label="Search tickets"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ paddingRight: searchQuery ? '2.5rem' : '1rem' }}
          />
          {searchQuery && (
            <button
              onClick={() => {
                setSearchQuery('');
                searchInputRef.current?.focus();
              }}
              aria-label="Clear search"
              title="Clear search"
              className="clear-search-btn"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          )}
        </div>
        <div className="controls-actions">
          <button className="btn-secondary refresh-btn" onClick={loadData} disabled={isLoading}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className={isLoading ? 'spin' : ''}><path d="M23 4v6h-6"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            <span>Refresh</span>
          </button>
          <button className="btn-primary prestige-btn" onClick={() => setIsModalOpen(true)}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            <span>New Ticket</span>
          </button>
        </div>
      </div>

      <div className="tickets-list">
        {filteredTasks.length === 0 ? (
          <div className="empty-state-card">
            <div className="empty-icon">✓</div>
            <h3>All caught up!</h3>
            <p>No tickets found in {activeTab}.</p>
          </div>
        ) : (
          filteredTasks.map(ticket => (
            <div key={ticket.id} className="ticket-row-premium">
              <div className="ticket-id-section">
                <span className="ticket-number">{ticket.ticket_id || `#ID-${ticket.id}`}</span>
                <span className={`priority-pill priority-${ticket.priority.toLowerCase()}`}>
                  {ticket.priority}
                </span>
              </div>
              
              <div className="ticket-info">
                <h4 className="ticket-title">{ticket.title}</h4>
                <p className="ticket-desc">{ticket.description}</p>
              </div>

              <div className="ticket-meta">
                <div className="meta-item">
                   <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                   {new Date(ticket.created_at).toLocaleDateString()}
                </div>
                <div className="meta-item">
                   <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
                   WhatsApp
                </div>
              </div>

              <div className="ticket-actions">
                <button className="btn-minimal">Manage</button>
              </div>
            </div>
          ))
        )}
      </div>

      {isModalOpen && (
        <div className="modal-overlay" onClick={() => setIsModalOpen(false)}>
          <div className="modal-content glass-island-premium" onClick={e => e.stopPropagation()} style={{ maxWidth: '500px' }}>
            <div className="modal-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>Raise New Ticket</h2>
              <button className="btn-icon-minimal" aria-label="Close modal" onClick={() => setIsModalOpen(false)} style={{ border: 'none', background: 'none', cursor: 'pointer', fontSize: '1.2rem', color: 'var(--text-secondary)' }}>✕</button>
            </div>
            
            <div className="modal-body" style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', padding: '1.5rem 0' }}>
              <div className="form-group-premium">
                <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Ticket Title</label>
                <input 
                  type="text" 
                  placeholder="What is the issue?" 
                  value={newTicket.title}
                  onChange={e => setNewTicket({...newTicket, title: e.target.value})}
                  autoFocus
                  style={{ width: '100%', padding: '0.75rem', borderRadius: '10px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none' }}
                />
              </div>
              
              <div className="form-group-premium">
                <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Description (Optional)</label>
                <textarea 
                  placeholder="Provide more details about the customer request..." 
                  value={newTicket.description}
                  onChange={e => setNewTicket({...newTicket, description: e.target.value})}
                  rows={3}
                  style={{ width: '100%', padding: '0.75rem', borderRadius: '10px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none', resize: 'vertical' }}
                />
              </div>
              
              <div className="form-group-premium">
                <label style={{ display: 'block', marginBottom: '0.75rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Priority</label>
                <div className="priority-selector-premium">
                  {(['low', 'medium', 'high', 'urgent'] as const).map(p => (
                    <button 
                      key={p}
                      className={`priority-btn-expert p-${p} ${newTicket.priority === p ? 'active' : ''}`}
                      onClick={() => setNewTicket({...newTicket, priority: p})}
                    >
                      <span className="p-dot"></span>
                      {p.toUpperCase()}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            
            <div className="modal-footer" style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.25rem' }}>
              <button className="btn-secondary" onClick={() => setIsModalOpen(false)} style={{ padding: '0.6rem 1.25rem', borderRadius: '10px', fontWeight: 600 }}>Cancel</button>
              <button 
                className="btn-primary prestige-btn" 
                onClick={handleCreateTicket} 
                disabled={isSaving || !newTicket.title.trim()}
                style={{ padding: '0.6rem 1.5rem', borderRadius: '10px', fontWeight: 700 }}
              >
                {isSaving ? 'Creating...' : 'Create Ticket'}
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{`
        .tickets-container {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }
        .tickets-header-tabs {
          display: flex;
          gap: 0.5rem;
          border-bottom: 1px solid var(--border-color);
          padding-bottom: 2px;
        }
        .ticket-tab {
          background: none;
          border: none;
          padding: 0.75rem 1.5rem;
          font-weight: 600;
          font-size: 0.9rem;
          color: var(--text-secondary);
          cursor: pointer;
          position: relative;
          display: flex;
          align-items: center;
          gap: 0.75rem;
          transition: all 0.2s;
        }
        .ticket-tab.active {
          color: var(--accent-color);
        }
        .ticket-tab.active::after {
          content: '';
          position: absolute;
          bottom: -2px;
          left: 0;
          right: 0;
          height: 2px;
          background: var(--accent-color);
        }
        .count-badge {
          background: var(--bg-hover);
          padding: 2px 8px;
          borderRadius: 10px;
          fontSize: 0.75rem;
          color: var(--text-tertiary);
        }
        .ticket-tab.active .count-badge {
          background: var(--accent-subtle);
          color: var(--accent-color);
        }
        .tickets-controls {
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 1rem;
        }
        .controls-actions {
          display: flex;
          gap: 0.75rem;
          align-items: center;
        }
        .refresh-btn {
          height: 48px;
          padding: 0 1.25rem;
          display: flex;
          align-items: center;
          gap: 0.6rem;
          font-weight: 600;
        }
        .refresh-btn svg {
          transition: transform 0.3s;
        }
        .refresh-btn:hover svg {
          transform: rotate(30deg);
        }
        .spin {
          animation: spin 1s linear infinite;
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .search-box-premium {
          flex: 1;
          background: var(--surface-color);
          border: 1px solid var(--border-color);
          border-radius: 12px;
          padding: 0 1rem;
          display: flex;
          align-items: center;
          gap: 0.75rem;
          height: 48px;
          box-shadow: var(--shadow-sm);
          position: relative;
        }
        .search-box-premium input {
          background: none;
          border: none;
          width: 100%;
          color: var(--text-primary);
          font-weight: 500;
          outline: none;
        }
        .ticket-row-premium {
          background: var(--surface-color);
          border: 1px solid var(--border-color);
          border-radius: 16px;
          padding: 1.25rem;
          display: grid;
          grid-template-columns: 140px 1fr 180px 100px;
          align-items: center;
          gap: 1.5rem;
          margin-bottom: 0.75rem;
          transition: all 0.2s ease;
          box-shadow: var(--shadow-sm);
        }
        .ticket-row-premium:hover {
          border-color: var(--accent-color);
          transform: translateY(-2px);
          box-shadow: var(--shadow-md);
        }
        .ticket-id-section {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }
        .ticket-number {
          font-family: var(--font-mono);
          font-weight: 700;
          color: var(--accent-color);
          font-size: 0.95rem;
        }
        .ticket-title {
          margin: 0 0 0.25rem 0;
          font-size: 1rem;
          font-weight: 700;
          color: var(--text-primary);
        }
        .ticket-desc {
          margin: 0;
          font-size: 0.85rem;
          color: var(--text-secondary);
          display: -webkit-box;
          -webkit-line-clamp: 1;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
        .ticket-meta {
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
        }
        .meta-item {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          font-size: 0.8rem;
          color: var(--text-tertiary);
          font-weight: 500;
        }
        .btn-minimal {
          background: var(--bg-hover);
          border: none;
          padding: 0.5rem 1rem;
          border-radius: 8px;
          font-weight: 600;
          font-size: 0.85rem;
          color: var(--text-primary);
          cursor: pointer;
          transition: all 0.2s;
        }
        .btn-minimal:hover {
          background: var(--accent-color);
          color: white;
        }
        .priority-pill {
          padding: 2px 8px;
          border-radius: 6px;
          font-size: 0.7rem;
          font-weight: 800;
          text-transform: uppercase;
          width: fit-content;
        }
        .priority-urgent { background: #fee2e2; color: #ef4444; }
        .priority-high { background: #ffedd5; color: #f59e0b; }
        .priority-medium { background: #dcfce7; color: #10b981; }
        .priority-low { background: #f3f4f6; color: #6b7280; }
        
        .priority-selector-premium {
          display: grid;
          grid-template-columns: repeat(4, 1fr);
          gap: 0.5rem;
        }
        .priority-btn-expert {
          background: var(--bg-input);
          border: 1px solid var(--border-color);
          border-radius: 12px;
          padding: 0.75rem 0.5rem;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 0.5rem;
          cursor: pointer;
          transition: all 0.2s ease;
          font-size: 0.7rem;
          font-weight: 700;
          letter-spacing: 0.05em;
          color: var(--text-tertiary);
        }
        .priority-btn-expert:hover {
          background: var(--bg-hover);
          border-color: var(--text-tertiary);
        }
        .priority-btn-expert.active {
          background: var(--surface-color);
          border-color: var(--accent-color);
          color: var(--text-primary);
          box-shadow: var(--shadow-sm);
          transform: translateY(-2px);
        }
        .p-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: var(--text-tertiary);
        }
        .p-low.active .p-dot { background: #6b7280; }
        .p-medium.active .p-dot { background: #10b981; }
        .p-high.active .p-dot { background: #f59e0b; }
        .p-urgent.active .p-dot { background: #ef4444; }
        
        .active-priority {
          box-shadow: 0 0 0 2px var(--surface-color), 0 0 0 4px currentColor;
          transform: translateY(-1px);
        }
        
        .empty-state-card {
          padding: 4rem 2rem;
          text-align: center;
          background: var(--surface-color);
          border: 1px dashed var(--border-color);
          border-radius: 20px;
          color: var(--text-tertiary);
        }
        .empty-icon {
          font-size: 3rem;
          margin-bottom: 1rem;
          opacity: 0.5;
        }

        /* Modal Styles */
        .modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.65);
          backdrop-filter: blur(12px);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
          animation: fadeIn 0.3s ease;
        }
        .modal-content {
          width: 90%;
          border-radius: 24px;
          padding: 2rem;
          box-shadow: 0 20px 40px rgba(0,0,0,0.2);
          animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
        }
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes slideUp {
          from { transform: translateY(20px); opacity: 0; }
          to { transform: translateY(0); opacity: 1; }
        }
        .clear-search-btn {
          position: absolute;
          right: 12px;
          top: 50%;
          transform: translateY(-50%);
          color: var(--text-tertiary);
          padding: 4px;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 50%;
          transition: all 0.2s;
          cursor: pointer;
          border: none;
          background: transparent;
        }
        .clear-search-btn:hover {
          color: var(--text-primary);
          background: var(--bg-hover);
        }
      `}</style>
    </div>
  );
};
