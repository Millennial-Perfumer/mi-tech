import { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface User {
    id: number;
    username: string;
    role: string;
    created_at: string;
}

interface UsersProps {
    fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export function Users({ fetchWithAuth }: UsersProps) {
    const { success, error } = useToast();
    const [users, setUsers] = useState<User[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [showAddModal, setShowAddModal] = useState(false);
    
    // Form state
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [role, setRole] = useState('read');
    const [isSaving, setIsSaving] = useState(false);

    const fetchUsers = async () => {
        setIsLoading(true);
        try {
            console.log('Fetching users from:', `${API_BASE}/api/users`);
            const response = await fetchWithAuth(`${API_BASE}/api/users`);
            console.log('Response status:', response.status);
            
            const data = await response.json();
            console.log('User data received:', data);
            
            if (data.success) {
                setUsers(data.users || []);
            }
        } catch (error) {
            console.error('Error fetching users:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers();
    }, []);

    const handleAddUser = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/users`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password, role })
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Failed to create user');
            }

            success('User created successfully');
            setShowAddModal(false);
            setUsername('');
            setPassword('');
            setRole('read');
            fetchUsers();
        } catch (err: any) {
            console.error('Save error:', err);
            error(err.message || 'An error occurred while creating the user');
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <div className="tab-pane active" style={{ animation: 'fadeIn 0.4s ease-out' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '2rem' }}>
                <div style={{ display: 'flex', gap: '1rem', flex: 1, alignItems: 'center' }}>
                    <div style={{ position: 'relative', width: '300px' }}>
                        <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                        <input 
                            type="text" 
                            placeholder="Search users..." 
                            style={{ 
                                paddingLeft: '2.5rem', 
                                width: '100%', 
                                fontSize: '0.875rem',
                                backgroundColor: 'var(--bg-input)',
                                color: 'var(--text-primary)',
                                border: '1px solid var(--border-color)',
                                borderRadius: '8px'
                            }}
                        />
                    </div>
                </div>

                <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', paddingTop: '0.5rem' }}>
                    <button 
                        className="btn-primary" 
                        onClick={() => setShowAddModal(true)} 
                        style={{ 
                            padding: '0.75rem 1.5rem', 
                            fontSize: '0.875rem', 
                            fontWeight: 600,
                            background: 'var(--accent-color)', 
                            display: 'flex', 
                            alignItems: 'center', 
                            gap: '10px', 
                            borderRadius: '12px',
                            transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                            boxShadow: 'var(--shadow-glow)'
                        }}
                    >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                        Add User
                    </button>
                </div>
            </div>

            <div className="card" style={{ padding: 0, overflow: 'hidden', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
                {isLoading ? (
                    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', padding: '4rem 2rem', color: 'var(--text-secondary)' }}>
                        <svg className="spinner" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ animation: 'spin 1s linear infinite', marginRight: '0.75rem' }}><line x1="12" y1="2" x2="12" y2="6"></line><line x1="12" y1="18" x2="12" y2="22"></line><line x1="4.93" y1="4.93" x2="7.76" y2="7.76"></line><line x1="16.24" y1="16.24" x2="19.07" y2="19.07"></line><line x1="2" y1="12" x2="6" y2="12"></line><line x1="18" y1="12" x2="22" y2="12"></line><line x1="4.93" y1="19.07" x2="7.76" y2="16.24"></line><line x1="16.24" y1="7.76" x2="19.07" y2="4.93"></line></svg>
                        Loading users...
                    </div>
                ) : (
                    <div style={{ overflowX: 'auto' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                            <thead>
                                <tr style={{ backgroundColor: 'var(--bg-header)', borderBottom: '2px solid var(--border-color)' }}>
                                    <th style={{ padding: '1rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>ID</th>
                                    <th style={{ padding: '1rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Username / Email</th>
                                    <th style={{ padding: '1rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Role</th>
                                    <th style={{ padding: '1rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Created At</th>
                                </tr>
                            </thead>
                            <tbody>
                                {users.length === 0 ? (
                                    <tr>
                                        <td colSpan={4} style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                                            No users found.
                                        </td>
                                    </tr>
                                ) : (
                                    users.map(user => (
                                        <tr key={user.id} style={{ borderBottom: '1px solid var(--border-color)', transition: 'background-color 0.15s ease' }}>
                                            <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>#{user.id}</td>
                                            <td style={{ padding: '1rem', fontWeight: 500, color: 'var(--text-primary)' }}>{user.username}</td>
                                            <td style={{ padding: '1rem' }}>
                                                <span style={{
                                                    display: 'inline-flex',
                                                    alignItems: 'center',
                                                    padding: '0.25rem 0.75rem',
                                                    borderRadius: '9999px',
                                                    fontSize: '0.8rem',
                                                    fontWeight: 600,
                                                    backgroundColor: user.role === 'admin' ? 'var(--accent-subtle)' : 'var(--bg-hover)',
                                                    color: user.role === 'admin' ? 'var(--accent-color)' : 'var(--text-secondary)',
                                                }}>
                                                    {user.role === 'admin' ? 'Administrator' : 'Read-Only'}
                                                </span>
                                            </td>
                                            <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
                                                {new Date(user.created_at).toLocaleString()}
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            {/* Add User Modal */}
            {showAddModal && (
                <div className="modal-overlay" style={{ alignItems: 'flex-start', paddingTop: '4rem' }}>
                    <div className="modal-content" style={{ maxWidth: '450px', width: '100%', backgroundColor: 'var(--surface-color)', border: '1px solid var(--border-color)' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1.5rem', borderBottom: '1px solid var(--border-color)', background: 'var(--bg-header)', borderRadius: '16px 16px 0 0' }}>
                            <h2 style={{ margin: 0, fontSize: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-primary)' }}>
                                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: 'var(--accent-color)'}}><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line></svg>
                                Add New User
                            </h2>
                            <button 
                                onClick={() => setShowAddModal(false)}
                                style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', padding: '4px', display: 'flex' }}
                            >
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                            </button>
                        </div>
                        
                        <form onSubmit={handleAddUser}>
                            <div style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                                <div className="form-group" style={{ margin: 0 }}>
                                    <label style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-primary)' }}>Username / Email</label>
                                    <input 
                                        type="email" 
                                        value={username}
                                        onChange={e => setUsername(e.target.value)}
                                        required
                                        placeholder="agent@example.com"
                                        style={{ marginTop: '0.25rem' }}
                                    />
                                </div>
                                
                                <div className="form-group" style={{ margin: 0 }}>
                                    <label style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-primary)' }}>Password</label>
                                    <input 
                                        type="password" 
                                        value={password}
                                        onChange={e => setPassword(e.target.value)}
                                        required
                                        placeholder="Enter secure password"
                                        style={{ marginTop: '0.25rem' }}
                                    />
                                </div>

                                <div className="form-group" style={{ margin: 0 }}>
                                    <label style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-primary)' }}>Role Assignment</label>
                                    <select 
                                        value={role}
                                        onChange={e => setRole(e.target.value)}
                                        required
                                        style={{ marginTop: '0.25rem', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', border: '1px solid var(--border-color)' }}
                                    >
                                        <option value="read">Read-Only Agent (Default)</option>
                                        <option value="admin">Administrator (Complete Access)</option>
                                    </select>
                                    <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.5rem', lineHeight: 1.4 }}>
                                        {role === 'admin' 
                                            ? '⚠️ Administrators have complete access to add users, sync data, edit settings, and manage configurations.' 
                                            : 'ℹ️ Read-Only agents can view dashboards, reports, and data, but cannot edit or delete records.'}
                                    </p>
                                </div>
                            </div>
                            
                            <div style={{ padding: '1.25rem 1.5rem', borderTop: '1px solid var(--border-color)', background: 'var(--bg-header)', display: 'flex', justifyContent: 'flex-end', gap: '1rem', borderRadius: '0 0 16px 16px' }}>
                                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)}>
                                    Cancel
                                </button>
                                <button type="submit" className="btn-primary" disabled={isSaving} style={{ minWidth: '120px' }}>
                                    {isSaving ? 'Creating...' : 'Create User'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
