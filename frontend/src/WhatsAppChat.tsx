import { useState, useEffect, useRef } from 'react';
import { API_BASE } from './api';
import './WhatsAppChat.css';

interface Conversation {
  id: number;
  phone_number: string;
  contact_name: string;
  last_message: string;
  last_message_at: string;
  mode: 'auto' | 'human';
}

interface ChatMessage {
  id: number;
  message_id: string;
  text: string;
  type: string;
  direction: 'incoming' | 'outgoing';
  sender_role: string;
  status: string;
  sent_at: string;
}

interface WhatsAppChatProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export function WhatsAppChat({ fetchWithAuth }: WhatsAppChatProps) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputText, setInputText] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesAreaRef = useRef<HTMLDivElement>(null);
  const pollingRef = useRef<number | null>(null);

  const scrollToBottom = (behavior: ScrollBehavior = 'smooth') => {
    if (messagesAreaRef.current) {
      messagesAreaRef.current.scrollTo({
        top: messagesAreaRef.current.scrollHeight,
        behavior
      });
    }
  };

  useEffect(() => {
    fetchConversations();
    
    // Start global polling for conversations
    const interval = setInterval(fetchConversations, 5000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (selectedConversation) {
      setMessages([]);
      setOffset(0);
      setHasMore(true);
      fetchMessages(selectedConversation.id, 30, 0, false);
      
      // Clear old polling and start new one for specific chat
      if (pollingRef.current) clearInterval(pollingRef.current);
      pollingRef.current = window.setInterval(() => {
        // Only poll for latest messages if at the bottom or near it
        const area = messagesAreaRef.current;
        if (area) {
          const isAtBottom = area.scrollHeight - area.scrollTop <= area.clientHeight + 100;
          if (isAtBottom) {
             fetchMessages(selectedConversation.id, 20, 0, false, true);
          }
        }
      }, 3000);
    }
    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current);
    };
  }, [selectedConversation?.id]);

  useEffect(() => {
    // Only scroll to bottom on initial load or new messages (not on lazy load)
    if (!isLoadingMore) {
      scrollToBottom();
    }
  }, [messages, isLoadingMore]);

  const fetchConversations = async () => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/conversations`);
      const data = await resp.json();
      if (Array.isArray(data)) {
        setConversations(data);
      }
    } catch (err) {
      console.error('Failed to fetch conversations:', err);
    } finally {
      setIsInitialLoading(false);
    }
  };

  const fetchMessages = async (convId: number, limit = 30, currentOffset = 0, append = false, silent = false) => {
    if (isLoadingMore && append) return;
    if (append) setIsLoadingMore(true);

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/chat?conversation_id=${convId}&limit=${limit}&offset=${currentOffset}`);
      const data = await resp.json();
      
      if (Array.isArray(data)) {
        if (append) {
          // Prepend older messages
          if (data.length > 0) {
            const oldScrollHeight = messagesAreaRef.current?.scrollHeight || 0;
            
            setMessages((prev: ChatMessage[]) => [...data, ...prev]);
            setOffset((prev: number) => prev + data.length);
            
            // Adjust scroll after render
            setTimeout(() => {
              if (messagesAreaRef.current) {
                const newScrollHeight = messagesAreaRef.current.scrollHeight;
                messagesAreaRef.current.scrollTop = newScrollHeight - oldScrollHeight;
              }
            }, 0);
          }
          if (data.length < limit) setHasMore(false);
        } else {
          // Initial or polling update
          setMessages((prev: ChatMessage[]) => {
            // Simple check to see if we have new messages at the end
            if (prev.length > 0 && data.length > 0) {
              const lastPrevId = prev[prev.length - 1].id;
              const lastDataId = data[data.length - 1].id;
              if (lastPrevId === lastDataId) return prev; // No new messages
            }
            // If polling, we might need to merge or just replace if small
            if (!silent) setOffset(data.length);
            return data;
          });
          if (!silent) setHasMore(data.length === limit);
        }
      }
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      if (append) setIsLoadingMore(false);
    }
  };

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    if (target.scrollTop === 0 && hasMore && !isLoadingMore && selectedConversation) {
      fetchMessages(selectedConversation.id, 20, offset, true);
    }
  };

  const activeConversation = conversations.find(c => c.id === selectedConversation?.id) || selectedConversation;

  const sendMessage = async () => {
    if (!activeConversation || !inputText.trim() || isSending) return;

    setIsSending(true);
    const textToSend = inputText.trim();
    setInputText('');

    // Optimistic update
    const tempMsg: ChatMessage = {
      id: Date.now(),
      message_id: 'temp-' + Date.now(),
      text: textToSend,
      type: 'text',
      direction: 'outgoing',
      sender_role: 'human',
      status: 'sending',
      sent_at: new Date().toISOString()
    };
    setMessages((prev: ChatMessage[]) => [...prev, tempMsg]);

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/send-message`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phone_number: activeConversation.phone_number,
          text: textToSend
        })
      });

      if (resp.ok) {
        fetchMessages(activeConversation.id);
        fetchConversations();
      } else {
        throw new Error('Failed to send');
      }
    } catch (err) {
      console.error('Send error:', err);
      // Remove optimistic message or mark as failed
      setMessages((prev: ChatMessage[]) => prev.map((m: ChatMessage) => m.id === tempMsg.id ? { ...m, status: 'failed' } : m));
    } finally {
      setIsSending(false);
    }
  };

  const toggleMode = async (mode: 'auto' | 'human') => {
    if (!activeConversation) return;

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/conversations/mode`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: activeConversation.id,
          mode: mode
        })
      });

      if (resp.ok) {
        if (selectedConversation) {
          setSelectedConversation({ ...selectedConversation, mode });
        }
        fetchConversations();
      }
    } catch (err) {
      console.error('Failed to toggle mode:', err);
    }
  };

  const filteredConversations = conversations.filter(c => 
    c.phone_number.includes(searchQuery) || 
    (c.contact_name && c.contact_name.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const formatTime = (dateStr: string) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getInitials = (name: string, phone: string) => {
    if (name) return name.substring(0, 1).toUpperCase();
    return phone ? phone.substring(phone.length - 2) : '?';
  };

  if (isInitialLoading) {
    return <div className="chat-empty-state">Loading chats...</div>;
  }

  return (
    <div className="whatsapp-chat-container">
      {/* Decorative background blobs for glassmorphism */}
      <div className="chat-glass-blob blob-1"></div>
      <div className="chat-glass-blob blob-2"></div>
      <div className="chat-glass-blob blob-3"></div>

      {/* Sidebar: Conversations List */}
      <aside className="chat-sidebar">
        <div className="chat-sidebar-header">
          <div className="chat-search-wrapper">
            <input 
              type="text" 
              className="chat-search-input" 
              placeholder="Search chats..." 
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            <svg className="chat-search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
          </div>
        </div>
        
        <div className="conversation-list">
          {filteredConversations.length === 0 ? (
            <div style={{ padding: '2rem', textAlign: 'center', opacity: 0.5 }}>No conversations found</div>
          ) : (
            filteredConversations.map(conv => (
              <div 
                key={conv.id} 
                className={`conversation-item ${activeConversation?.id === conv.id ? 'active' : ''}`}
                onClick={() => setSelectedConversation(conv)}
              >
                <div className="conv-avatar">
                  {getInitials(conv.contact_name, conv.phone_number)}
                </div>
                <div className="conv-info">
                  <div className="conv-header">
                    <span className="conv-name">{conv.contact_name || conv.phone_number}</span>
                    <span className="conv-time">{formatTime(conv.last_message_at)}</span>
                  </div>
                  <div className="conv-last-msg">{conv.last_message}</div>
                  <div style={{ marginTop: '6px' }}>
                    <span className={`conv-mode-badge mode-${conv.mode}`}>
                      {conv.mode.toUpperCase()}
                    </span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </aside>

      {/* Main Panel: Chat Thread */}
      <main className="chat-panel">
        {activeConversation ? (
          <>
            <header className="chat-panel-header">
              <div className="chat-header-info">
                <div className="conv-avatar" style={{ width: '36px', height: '36px', fontSize: '0.9rem' }}>
                   {getInitials(activeConversation.contact_name, activeConversation.phone_number)}
                </div>
                <div>
                  <div className="chat-header-name">{activeConversation.contact_name || activeConversation.phone_number}</div>
                  <div style={{ fontSize: '0.75rem', opacity: 0.6 }}>{activeConversation.phone_number}</div>
                </div>
              </div>
              
              <div className="mode-toggle-group">
                <button 
                  className={`mode-btn ${activeConversation.mode === 'auto' ? 'active' : ''}`}
                  onClick={() => toggleMode('auto')}
                >
                  Auto
                </button>
                <button 
                  className={`mode-btn ${activeConversation.mode === 'human' ? 'active' : ''}`}
                  onClick={() => toggleMode('human')}
                >
                  Human
                </button>
              </div>
            </header>

            <div 
              className="messages-area" 
              ref={messagesAreaRef}
              onScroll={handleScroll}
            >
              {isLoadingMore && (
                <div style={{ textAlign: 'center', padding: '1rem', fontSize: '0.8rem', opacity: 0.5 }}>
                  Loading older messages...
                </div>
              )}
              {messages.map(msg => (
                <div key={msg.id} className={`message-bubble message-${msg.direction}`}>
                  <div className="message-text">{msg.text}</div>
                  <div className="message-info">
                    <span className="message-time">{formatTime(msg.sent_at)}</span>
                    {msg.direction === 'outgoing' && (
                      <span className="message-status">
                        {msg.status === 'read' ? '✓✓' : msg.status === 'delivered' ? '✓✓' : '✓'}
                      </span>
                    )}
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>

            <div className="chat-input-area">
              <input 
                type="text" 
                className="chat-msg-input" 
                placeholder="Type a message..." 
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
              />
              <button 
                className="chat-send-btn" 
                onClick={sendMessage}
                disabled={!inputText.trim() || isSending}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
              </button>
            </div>
          </>
        ) : (
          <div className="chat-empty-state">
             <svg className="chat-empty-icon" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
             <h3>Select a conversation</h3>
             <p>Choose a contact from the sidebar to start messaging</p>
          </div>
        )}
      </main>
    </div>
  );
}
