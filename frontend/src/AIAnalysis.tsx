import React, { useState, useEffect, useRef } from 'react';
import './AIAnalysis.css';
import { useConfirm } from './ConfirmContext';

interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

interface Conversation {
  id: number;
  title: string;
  updated_at: string;
}

interface AIAnalysisProps {
  fetchWithAuth: (url: string, options?: any) => Promise<any>;
  API_BASE: string;
}

export const AIAnalysis: React.FC<AIAnalysisProps> = ({ fetchWithAuth, API_BASE }) => {
  const { confirm } = useConfirm();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConversationId, setActiveConversationId] = useState<number | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    loadConversations();
  }, []);

  useEffect(() => {
    if (activeConversationId) {
      loadMessages(activeConversationId);
    } else {
      setMessages([]);
    }
  }, [activeConversationId]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadConversations = async () => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/ai/conversations`);
      if (resp.ok) {
        const data = await resp.json();
        setConversations(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to load conversations:', err);
    }
  };

  const loadMessages = async (id: number) => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/ai/conversations?id=${id}`);
      if (resp.ok) {
        const data = await resp.json();
        if (data && data.messages) {
          setMessages(data.messages);
        }
      }
    } catch (err) {
      console.error('Failed to load messages:', err);
    }
  };

  const handleSendMessage = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage = input.trim();
    setInput('');
    // Reset textarea height
    if (textareaRef.current) textareaRef.current.style.height = 'auto';
    setIsLoading(true);

    // Add user message optimistically
    const newMessages: Message[] = [...messages, { role: 'user', content: userMessage }];
    setMessages(newMessages);

    try {
      const response = await fetchWithAuth(`${API_BASE}/api/ai/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          conversation_id: activeConversationId,
          message: userMessage
        })
      });

      if (!response.ok) throw new Error('Chat request failed');

      const reader = response.body?.getReader();
      if (!reader) throw new Error('No reader available');

      let assistantMessage = '';
      setMessages([...newMessages, { role: 'assistant', content: '' }]);

      let buffer = '';
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += new TextDecoder().decode(value, { stream: true });
        const lines = buffer.split('\n');
        
        // Keep the last partial line in the buffer
        buffer = lines.pop() || '';

        for (const line of lines) {
          const trimmedLine = line.trim();
          if (trimmedLine.startsWith('data: ')) {
            try {
              const data = JSON.parse(trimmedLine.substring(6));
              
              if (data.conversation_id && !activeConversationId) {
                setActiveConversationId(data.conversation_id);
                loadConversations();
              }

              if (data.content) {
                assistantMessage += data.content;
                setMessages(prev => {
                  const last = prev[prev.length - 1];
                  if (last && last.role === 'assistant') {
                    return [...prev.slice(0, -1), { ...last, content: assistantMessage }];
                  }
                  return prev;
                });
              }

              if (data.error) {
                setMessages(prev => [...prev, { role: 'system', content: `Error: ${data.error}` }]);
              }
            } catch (e) {
              // Not JSON or incomplete, ignore
            }
          }
        }
      }
    } catch (err: any) {
      setMessages(prev => [...prev, { role: 'system', content: `Failed to connect to AI: ${err.message}` }]);
    } finally {
      setIsLoading(false);
    }
  };

  const startNewChat = () => {
    setActiveConversationId(null);
    setMessages([]);
    setInput('');
    if (textareaRef.current) textareaRef.current.style.height = 'auto';
  };

  const deleteConversation = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    
    const confirmed = await confirm({
      title: 'Delete Analysis',
      message: 'Are you sure you want to delete this conversation? This action cannot be undone.',
      confirmLabel: 'Delete',
      variant: 'danger'
    });

    if (confirmed) {
      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/ai/conversations?id=${id}`, { method: 'DELETE' });
        if (resp.ok) {
          if (activeConversationId === id) startNewChat();
          loadConversations();
        }
      } catch (err) {
        console.error('Failed to delete conversation:', err);
      }
    }
  };

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = '24px';
      const scrollH = textareaRef.current.scrollHeight;
      if (scrollH > 24) {
        textareaRef.current.style.height = Math.min(scrollH, 150) + 'px';
      }
    }
  }, [input]);

  const quickPrompts = [
    "What's my revenue this month?",
    "Show me top 5 selling products",
    "Current inventory health report",
    "Compare today vs yesterday's sales"
  ];

  // Filter messages for display - only show user and assistant (with content)
  const displayMessages = messages.filter(msg => {
    if (msg.role === 'user' || msg.role === 'system') return true;
    if (msg.role === 'assistant' && (msg.content || isLoading)) return true;
    return false;
  });

  const renderContent = (content: string) => {
    if (!content) return null;

    // Split content into segments based on table markers
    // We look for blocks that start with | and have at least 2 lines of |
    const segments: React.ReactNode[] = [];
    const lines = content.split('\n');
    let currentBlock: string[] = [];
    let isTable = false;

    const flushBlock = (index: number) => {
      if (currentBlock.length === 0) return;
      
      const blockText = currentBlock.join('\n').trim();
      if (!blockText) return;

      if (isTable && currentBlock.length >= 3) {
        // Render Table
        const tableLines = currentBlock.filter(l => l.includes('|'));
        const headers = tableLines[0].split('|').filter(c => c.trim()).map(c => c.trim());
        const rows = tableLines.slice(2).filter(row => row.includes('|')).map(row => row.split('|').filter(c => c.trim()).map(c => c.trim()));
        
        segments.push(
          <div className="table-container" key={`table-${index}`}>
            <table>
              <thead>
                <tr>{headers.map((h, i) => <th key={i}>{formatMarkdown(h)}</th>)}</tr>
              </thead>
              <tbody>
                {rows.map((row, i) => (
                  <tr key={i}>{row.map((cell, j) => <td key={j}>{formatMarkdown(cell)}</td>)}</tr>
                ))}
              </tbody>
            </table>
          </div>
        );
      } else {
        // Render Text (lists, bold, paragraphs)
        segments.push(<div key={`text-${index}`} className="text-block">{renderText(blockText)}</div>);
      }
      currentBlock = [];
    };

    lines.forEach((line, i) => {
      const trimmed = line.trim();
      const lineIsTable = trimmed.startsWith('|') && trimmed.includes('|');

      if (lineIsTable !== isTable && currentBlock.length > 0) {
        flushBlock(i);
      }
      
      isTable = lineIsTable;
      currentBlock.push(line);
    });

    flushBlock(lines.length);
    return segments;
  };

  const renderText = (text: string) => {
    const lines = text.split('\n');
    const elements: React.ReactNode[] = [];
    let currentList: React.ReactNode[] = [];
    let listType: 'ul' | 'ol' | null = null;

    lines.forEach((line, i) => {
      const trimmed = line.trim();
      if (!trimmed) {
        if (currentList.length) {
          elements.push(listType === 'ul' ? <ul key={`list-${i}`}>{currentList}</ul> : <ol key={`list-${i}`}>{currentList}</ol>);
          currentList = [];
          listType = null;
        }
        elements.push(<br key={`br-${i}`} />);
        return;
      }

      const bulletMatch = trimmed.match(/^[\*\-•]\s+(.*)/);
      const numberMatch = trimmed.match(/^\d+\.\s+(.*)/);

      if (bulletMatch) {
        if (listType === 'ol') {
          elements.push(<ol key={`list-${i}`}>{currentList}</ol>);
          currentList = [];
        }
        listType = 'ul';
        currentList.push(<li key={`li-${i}`}>{formatMarkdown(bulletMatch[1])}</li>);
      } else if (numberMatch) {
        if (listType === 'ul') {
          elements.push(<ul key={`list-${i}`}>{currentList}</ul>);
          currentList = [];
        }
        listType = 'ol';
        currentList.push(<li key={`li-${i}`}>{formatMarkdown(numberMatch[1])}</li>);
      } else {
        if (currentList.length) {
          elements.push(listType === 'ul' ? <ul key={`list-${i}`}>{currentList}</ul> : <ol key={`list-${i}`}>{currentList}</ol>);
          currentList = [];
          listType = null;
        }
        elements.push(<p key={`p-${i}`}>{formatMarkdown(trimmed)}</p>);
      }
    });

    if (currentList.length) {
      elements.push(listType === 'ul' ? <ul key="last-list">{currentList}</ul> : <ol key="last-list">{currentList}</ol>);
    }

    return elements;
  };

  const formatMarkdown = (text: string) => {
    if (!text) return "";
    // Match **bold** or *italic*
    const parts = text.split(/(\*\*.*?\*\*|\*.*?\*)/g);
    return parts.map((part, i) => {
      if (part.startsWith('**') && part.endsWith('**')) {
        return <strong key={i}>{part.slice(2, -2)}</strong>;
      }
      if (part.startsWith('*') && part.endsWith('*')) {
        return <em key={i}>{part.slice(1, -1)}</em>;
      }
      return part;
    });
  };

  return (
    <div className={`ai_analysis_container ${isSidebarOpen ? 'sidebar-open' : ''}`}>
      {/* Sidebar */}
      <aside className="ai-history-sidebar">
        <div className="sidebar-header">
          <button className="new-chat-btn" onClick={startNewChat}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            New Analysis
          </button>
          <button className="toggle-sidebar-btn" onClick={() => setIsSidebarOpen(false)} title="Collapse Sidebar">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
          </button>
        </div>
        
        <div className="history-list">
          {conversations.map(conv => (
            <div 
              key={conv.id} 
              className={`history-item ${activeConversationId === conv.id ? 'active' : ''}`}
              onClick={() => setActiveConversationId(conv.id)}
            >
              <div className="history-item-left">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="chat-icon">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
                <span className="chat-title">{conv.title}</span>
              </div>
              <button className="delete-conv-btn" onClick={(e) => deleteConversation(conv.id, e)} title="Delete Analysis">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="3 6 5 6 21 6"></polyline>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                </svg>
              </button>
            </div>
          ))}
        </div>
      </aside>

      {/* Main Chat */}
      <main className="ai-chat-main">
        {!isSidebarOpen && (
          <button className="open-sidebar-btn" onClick={() => setIsSidebarOpen(true)}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '6px'}}>
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            History
          </button>
        )}

        <div className="chat-messages">
          {displayMessages.length === 0 ? (
            <div className="empty-state">
              <div className="ai-brain-icon">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 2v4"></path>
                  <path d="M12 18v4"></path>
                  <path d="M4.93 4.93l2.83 2.83"></path>
                  <path d="M16.24 16.24l2.83 2.83"></path>
                  <path d="M2 12h4"></path>
                  <path d="M18 12h4"></path>
                  <path d="M4.93 19.07l2.83-2.83"></path>
                  <path d="M16.24 7.76l2.83-2.83"></path>
                </svg>
              </div>
              <h2>Business Intelligence</h2>
              <p>Analyze your sales, revenue, and inventory with AI.</p>
              <div className="quick-prompts">
                {quickPrompts.map(prompt => (
                  <button key={prompt} onClick={() => { setInput(prompt); }}>
                    {prompt}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            displayMessages.map((msg, idx) => (
              <div key={idx} className={`message-wrapper ${msg.role}`}>
                <div className="message-avatar">
                  {msg.role === 'user' ? (
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                      <circle cx="12" cy="7" r="4"></circle>
                    </svg>
                  ) : msg.role === 'assistant' ? (
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="3" y="11" width="18" height="10" rx="2"></rect>
                      <circle cx="12" cy="5" r="2"></circle>
                      <path d="M12 7v4"></path>
                      <line x1="8" y1="16" x2="8" y2="16"></line>
                      <line x1="16" y1="16" x2="16" y2="16"></line>
                    </svg>
                  ) : (
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <circle cx="12" cy="12" r="3"></circle>
                      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1-2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
                    </svg>
                  )}
                </div>
                <div className="message-bubble">
                  <div className="message-content">
                    {msg.content ? renderContent(msg.content) : (
                      <div className="typing-indicator">
                        <span></span><span></span><span></span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="ai-chat-input-area">
          <div className="ai-input-pill">
            <button className="input-utility-btn" title="Add context">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
            </button>

            <textarea 
              ref={textareaRef}
              className="ai-textarea"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask anything..."
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSendMessage();
                }
              }}
            />

            <div className="input-actions-right">
              <button 
                className={`ai-send-btn ${input.trim() ? 'active' : ''}`}
                onClick={handleSendMessage}
                disabled={!input.trim() || isLoading}
                aria-label="Send message"
              >
                {isLoading ? (
                  <div className="typing-indicator small"><span></span><span></span><span></span></div>
                ) : (
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="22" y1="2" x2="11" y2="13"></line>
                    <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                  </svg>
                )}
              </button>
            </div>
          </div>
          <p className="ai-disclaimer">AI can make mistakes. Check important info.</p>
        </div>
      </main>
    </div>
  );
};
