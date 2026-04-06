import React, { useState, useEffect, useMemo } from 'react';
import { 
  DndContext, 
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  closestCorners,
  useDroppable,
} from '@dnd-kit/core';
import type { DragStartEvent, DragOverEvent, DragEndEvent } from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  arrayMove,
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, 
  PieChart, Pie, Cell, Legend, AreaChart, Area 
} from 'recharts';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import './App.css';

interface PlannerProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

interface Board {
  id: number;
  name: string;
  description: string;
  columns: Column[];
}

interface Column {
  id: number;
  name: string;
  order: number;
}

interface Subtask {
  id: string;
  title: string;
  completed: boolean;
}

interface TaskMetadata {
  subtasks?: Subtask[];
}

interface Task {
  id: number;
  title: string;
  description: string;
  status: string;
  priority: string;
  column_id: number;
  order: number;
  sprint_id?: number;
  metadata?: TaskMetadata | any;
  created_at: string;
}

interface Sprint {
  id: number;
  name: string;
  goal: string;
  start_date: string;
  end_date: string;
  status: string;
}

interface Analytics {
  sprint_velocity: number;
  task_lead_time_days: number;
}

// Draggable Task Card Component
const SortableTaskCard = ({ task, onClick }: { task: Task, onClick?: (task: Task) => void }) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging
  } = useSortable({ id: task.id });

  const style = {
    transform: CSS.Translate.toString(transform),
    transition,
    opacity: isDragging ? 0.3 : 1,
    zIndex: isDragging ? 1000 : 1,
  };

  const subtasksCount = task.metadata?.subtasks?.length || 0;
  const completedCount = task.metadata?.subtasks?.filter((s: Subtask) => s.completed).length || 0;

  return (
    <div 
      ref={setNodeRef} 
      style={style} 
      {...attributes} 
      {...listeners}
      className={`task-card ${isDragging ? 'grabbing' : ''}`}
      onClick={(e) => {
          e.stopPropagation();
          onClick?.(task);
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
        <span className={`priority-tag priority-${task.priority.toLowerCase()}`}>
          {task.priority}
        </span>
        {subtasksCount > 0 && (
           <span className="subtask-mini-pill">
              {completedCount}/{subtasksCount}
           </span>
        )}
      </div>
      <h4 style={{ margin: 0, fontSize: '0.9rem', color: 'var(--text-primary)', fontWeight: 600 }}>{task.title}</h4>
      {task.description && (
         <p style={{ margin: '0.5rem 0 0', fontSize: '0.75rem', color: 'var(--text-secondary)', overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', lineHeight: '1.4' }}>
          {task.description}
        </p>
      )}
    </div>
  );
};

// Droppable Column Component
const DroppableColumn = ({ col, tasks, openEditModal }: { col: Column, tasks: Task[], openEditModal: (t: Task) => void }) => {
  const { setNodeRef } = useDroppable({ 
    id: `col-${col.id}`,
    data: { type: 'column', columnId: col.id }
  });

  return (
    <div ref={setNodeRef} className="kanban-col-wrapper">
      <div className="kanban-col-header">
        <div className="col-title-group">
          <span className="col-dot"></span>
          <span className="col-name">{col.name}</span>
          <span className="col-count">{tasks.length}</span>
        </div>
      </div>

      <SortableContext 
        id={`col-${col.id}`}
        items={tasks.map(t => t.id)}
        strategy={verticalListSortingStrategy}
      >
        <div className="task-drop-zone">
          {tasks.sort((a,b) => a.order - b.order).map(task => (
            <SortableTaskCard key={task.id} task={task} onClick={openEditModal} />
          ))}
          {tasks.length === 0 && (
            <div className="empty-col-state">Drop tasks here</div>
          )}
        </div>
      </SortableContext>
    </div>
  );
};

export const Planner: React.FC<PlannerProps> = ({ fetchWithAuth }) => {
  const { success, error } = useToast();
  const [activeBoard, setActiveBoard] = useState<Board | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [sprints, setSprints] = useState<Sprint[]>([]);
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [activeId, setActiveId] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [view, setView] = useState<'kanban' | 'planning' | 'analytics'>('kanban');
  
  // Modals
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [isSprintModalOpen, setIsSprintModalOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editingTask, setEditingTask] = useState<Partial<Task>>({ 
    title: '', 
    description: '', 
    priority: 'low',
    column_id: undefined,
    metadata: { subtasks: [] }
  });
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('');
  const [selectedSprintId, setSelectedSprintId] = useState<number | null>(null);
  const [editingSprint, setEditingSprint] = useState({
    name: '',
    goal: '',
    start_date: new Date().toISOString().split('T')[0],
    end_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  });
  const [isEditingSprint, setIsEditingSprint] = useState(false);
  const [editingSprintId, setEditingSprintId] = useState<number | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  useEffect(() => {
    loadAllData();
  }, []);

  const loadAllData = async () => {
    setIsLoading(true);
    await Promise.all([
      fetchBoards(),
      fetchSprints(),
      fetchAnalytics()
    ]);
    setIsLoading(false);
  };

  const fetchBoards = async () => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/boards`);
      const data = await res.json();
      if (data.success && data.boards.length > 0) {
        setActiveBoard(data.boards[0]);
        await fetchTasks(data.boards[0].id);
      }
    } catch (err) {
      error('Failed to load boards');
    }
  };

  const fetchTasks = async (boardId: number) => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/tasks?board_id=${boardId}`);
      const data = await res.json();
      if (data.success) {
        setTasks(data.tasks);
      }
    } catch (err) {
      error('Failed to load tasks');
    }
  };

  const fetchSprints = async () => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/sprints`);
      const data = await res.json();
      if (data.success) setSprints(data.sprints);
    } catch (err) {}
  };

  const fetchAnalytics = async () => {
    try {
      // If we are in a board but no sprint is active, we just get aggregate
      const url = `${API_BASE}/api/planner/analytics`;
      const res = await fetchWithAuth(url);
      const data = await res.json();
      if (data.success) {
        setAnalytics(data.analytics);
      }
    } catch (err) {
      console.error('Analytics fetch failed:', err);
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as number);
  };

  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over) return;

    const activeId = active.id as number;
    const overId = over.id as number;

    const activeTask = tasks.find(t => t.id === activeId);
    if (!activeTask) return;

    // Handle moving between columns
    const overStr = over.id.toString();
    if (overStr.startsWith('col-')) {
      const newColId = parseInt(overStr.replace('col-', ''));
      if (activeTask.column_id !== newColId) {
        setTasks(prev => prev.map(t => t.id === activeId ? { ...t, column_id: newColId } : t));
      }
      return;
    }

    const overTask = tasks.find(t => t.id === overId);
    if (overTask && activeTask.column_id !== overTask.column_id) {
       setTasks(prev => prev.map(t => t.id === activeId ? { ...t, column_id: overTask.column_id } : t));
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);
    
    if (!over) {
      fetchTasks(activeBoard!.id);
      return;
    }

    const activeId = active.id as number;
    const overId = over.id;
    
    const activeTask = tasks.find(t => t.id === activeId);
    if (!activeTask) return;

    let newColumnId = activeTask.column_id;
    let newOrder = 1;

    const overIdStr = overId.toString();
    
    if (overIdStr.startsWith('col-')) {
      newColumnId = parseInt(overIdStr.replace('col-', ''));
      // Find the max order in this column and add 1
      const colTasks = tasks.filter(t => t.column_id === newColumnId);
      newOrder = colTasks.length > 0 ? Math.max(...colTasks.map(t => t.order || 0)) + 1 : 1;
    } else {
      const overTask = tasks.find(t => t.id === (overId as number));
      if (overTask) {
        newColumnId = overTask.column_id;
        
        // Calculate new order based on position relative to overTask
        const colTasks = tasks.filter(t => t.column_id === newColumnId).sort((a, b) => (a.order || 0) - (b.order || 0));
        const oldIdx = colTasks.findIndex(t => t.id === activeId);
        const newIdx = colTasks.findIndex(t => t.id === (overId as number));
        
        // If we are moving within the same column, we can use arrayMove logic
        if (oldIdx !== -1) {
          const reordered = arrayMove(colTasks, oldIdx, newIdx);
          newOrder = newIdx + 1;
          // Update local state for immediate feedback
          setTasks(prev => {
            const otherTasks = prev.filter(t => t.column_id !== newColumnId);
            const updatedColTasks = reordered.map((t, i) => ({ ...t, order: i + 1 }));
            return [...otherTasks, ...updatedColTasks];
          });
        } else {
          // Moving from different column
          newOrder = newIdx + 1;
          // Update local state
          setTasks(prev => {
            const taskToMove = { ...activeTask, column_id: newColumnId, order: newOrder };
            const otherTasks = prev.filter(t => t.id !== activeId);
            // Insert at newIdx
            const targetColTasks = otherTasks.filter(t => t.column_id === newColumnId).sort((a, b) => (a.order || 0) - (b.order || 0));
            targetColTasks.splice(newIdx, 0, taskToMove);
            const finalizedTargetTasks = targetColTasks.map((t, i) => ({ ...t, order: i + 1 }));
            const rest = otherTasks.filter(t => t.column_id !== newColumnId);
            return [...rest, ...finalizedTargetTasks];
          });
        }
      }
    }

    console.log(`[MoveTask] Task ${activeId} -> Col ${newColumnId}, Order ${newOrder}`);

    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/tasks/move`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          task_id: activeId,
          to_column_id: newColumnId,
          new_order: newOrder
        })
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.message || 'Server error');
      
      success('Task moved successfully');
      // fetchTasks(activeBoard!.id); // Avoid flicker if local state is good
    } catch (err: any) {
      error(`Conflict: ${err.message || 'Failed to move'}`);
      fetchTasks(activeBoard!.id);
    }
  };

  const handleSaveSprint = async () => {
    if (!editingSprint.name) {
      error('Sprint name is required');
      return;
    }

    try {
      const method = isEditingSprint ? 'PUT' : 'POST';
      const url = isEditingSprint 
        ? `${API_BASE}/api/planner/sprints?id=${editingSprintId}` 
        : `${API_BASE}/api/planner/sprints`;

      const res = await fetchWithAuth(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...editingSprint,
          start_date: new Date(editingSprint.start_date).toISOString(),
          end_date: new Date(editingSprint.end_date).toISOString()
        })
      });
      const data = await res.json();
      if (data.success) {
        success(isEditingSprint ? 'Sprint updated' : 'Sprint created');
        setIsSprintModalOpen(false);
        setIsEditingSprint(false);
        setEditingSprintId(null);
        setEditingSprint({
          name: '',
          goal: '',
          start_date: new Date().toISOString().split('T')[0],
          end_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
        });
        fetchSprints();
      }
    } catch (err) {
      error('Failed to save sprint');
    }
  };

  const handleDeleteSprint = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this sprint? Tasks will move to backlog.')) return;
    
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/sprints?id=${id}`, {
        method: 'DELETE'
      });
      const data = await res.json();
      if (data.success) {
        success('Sprint deleted');
        setIsSprintModalOpen(false);
        setIsEditingSprint(false);
        setEditingSprintId(null);
        if (selectedSprintId === id) setSelectedSprintId(null);
        fetchSprints();
        fetchTasks(activeBoard!.id);
      }
    } catch (err) {
      error('Failed to delete sprint');
    }
  };

  const handleSaveTask = async () => {
    if (!editingTask.title) {
      error('Title is required');
      return;
    }

    try {
      const method = isEditing ? 'PUT' : 'POST';
      const url = isEditing ? `${API_BASE}/api/planner/tasks?id=${editingTask.id}` : `${API_BASE}/api/planner/tasks`;
      
      const payload = isEditing ? {
          title: editingTask.title,
          description: editingTask.description,
          priority: editingTask.priority,
          metadata: editingTask.metadata,
          sprint_id: editingTask.sprint_id || null
      } : {
          board_id: activeBoard!.id,
          column_id: editingTask.column_id || activeBoard!.columns[0].id,
          title: editingTask.title,
          description: editingTask.description,
          priority: editingTask.priority,
          metadata: editingTask.metadata,
          sprint_id: editingTask.sprint_id || null
      };

      const res = await fetchWithAuth(url, {
        method,
        headers: { 
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (data.success) {
        success(isEditing ? 'Task updated' : 'Task created');
        setIsTaskModalOpen(false);
        fetchTasks(activeBoard!.id);
      }
    } catch (err) {
      error(`Failed to ${isEditing ? 'update' : 'create'} task`);
    }
  };

  const handleDeleteTask = async () => {
    if (!editingTask.id) return;
    if (!window.confirm('Are you sure you want to delete this task?')) return;

    try {
      const res = await fetchWithAuth(`${API_BASE}/api/planner/tasks?id=${editingTask.id}`, {
        method: 'DELETE'
      });
      const data = await res.json();
      if (data.success) {
        success('Task deleted');
        setIsTaskModalOpen(false);
        fetchTasks(activeBoard!.id);
      }
    } catch (err) {
      error('Failed to delete task');
    }
  };

  const openEditModal = (task: Task) => {
    setEditingTask({
        ...task,
        metadata: task.metadata || { subtasks: [] }
    });
    setIsEditing(true);
    setIsTaskModalOpen(true);
  };

  const openCreateModal = () => {
    setEditingTask({ 
        title: '', 
        description: '', 
        priority: 'medium', 
        column_id: activeBoard?.columns[0].id,
        metadata: { subtasks: [] }
    });
    setIsEditing(false);
    setIsTaskModalOpen(true);
  };

  const addSubtask = () => {
    if (!newSubtaskTitle) return;
    const newSub: Subtask = {
        id: Math.random().toString(36).substr(2, 9),
        title: newSubtaskTitle,
        completed: false
    };
    setEditingTask({
        ...editingTask,
        metadata: {
            ...editingTask.metadata,
            subtasks: [...(editingTask.metadata?.subtasks || []), newSub]
        }
    });
    setNewSubtaskTitle('');
  };

  const toggleSubtask = (subId: string) => {
    const updatedSubtasks = (editingTask.metadata?.subtasks || []).map((s: Subtask) => 
        s.id === subId ? { ...s, completed: !s.completed } : s
    );
    setEditingTask({
        ...editingTask,
        metadata: { ...editingTask.metadata, subtasks: updatedSubtasks }
    });
  };

  const deleteSubtask = (subId: string) => {
     const updatedSubtasks = (editingTask.metadata?.subtasks || []).filter((s: Subtask) => s.id !== subId);
     setEditingTask({
        ...editingTask,
        metadata: { ...editingTask.metadata, subtasks: updatedSubtasks }
    });
  };

  // Pre-aggregated data for charts
  const statusLabels = useMemo(() => {
    if (!activeBoard) return [];
    return activeBoard.columns.map(col => ({
      name: col.name,
      value: tasks.filter(t => t.column_id === col.id).length
    }));
  }, [tasks, activeBoard]);

  const priorityData = useMemo(() => {
    const counts: any = { low: 0, medium: 0, high: 0, urgent: 0 };
    tasks.forEach(t => {
      const p = t.priority.toLowerCase();
      if (counts[p] !== undefined) counts[p]++;
    });
    return Object.keys(counts).map(k => ({ name: k.toUpperCase(), value: counts[k] }));
  }, [tasks]);

  const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

  if (isLoading && !activeBoard) return <div className="loading-shimmer" style={{height: '500px', borderRadius: '24px'}}></div>;

  return (
    <div className="planner-container">
      <div className="planner-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <div className="view-switcher-pill">
            {(['kanban', 'planning', 'analytics'] as const).map(v => (
              <button
                key={v}
                className={`view-pill-btn ${view === v ? 'active' : ''}`}
                onClick={() => setView(v)}
              >
                {v.charAt(0).toUpperCase() + v.slice(1)}
              </button>
            ))}
          </div>
        </div>
        
        <div style={{ display: 'flex', gap: '0.75rem' }}>
          <button className="btn-secondary" style={{ padding: '0.6rem 1rem' }} onClick={loadAllData}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M23 4v6h-6"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          </button>
          <button 
            className="btn-primary prestige-btn" 
            onClick={openCreateModal}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            New Task
          </button>
        </div>
      </div>

      <div className="planner-viewport">
        {view === 'kanban' && activeBoard && (
          <DndContext 
            sensors={sensors}
            collisionDetection={closestCorners}
            onDragStart={handleDragStart}
            onDragOver={handleDragOver}
            onDragEnd={handleDragEnd}
          >
            <div className="kanban-scroller">
              {activeBoard.columns.sort((a, b) => a.order - b.order).map(col => (
                <DroppableColumn 
                  key={col.id} 
                  col={col} 
                  tasks={tasks.filter(t => t.column_id === col.id).sort((a,b) => (a.order || 0) - (b.order || 0))} 
                  openEditModal={openEditModal} 
                />
              ))}
            </div>
            
            <DragOverlay dropAnimation={null}>
              {activeId ? (
                <div className="task-card dragging-overlay">
                  <h4 style={{ margin: 0, fontSize: '0.9rem', color: 'var(--text-primary)', fontWeight: 600 }}>
                    {tasks.find(t => t.id === activeId)?.title}
                  </h4>
                </div>
              ) : null}
            </DragOverlay>
          </DndContext>
        )}

        {view === 'planning' && (
          <div className="planning-grid">
            <div className="planning-sidebar">
              <div className="section-subtitle">
                <span>Active Sprints</span>
                <span className="col-count">{sprints.length}</span>
              </div>
              <div className="sprint-list">
                <div 
                  className={`sprint-item ${selectedSprintId === null ? 'active' : ''}`}
                   onClick={(e) => {
                     e.stopPropagation();
                     setSelectedSprintId(null);
                   }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <span className="sprint-name">Product Backlog</span>
                    <span className="status-badge planned">Grooming</span>
                  </div>
                  <div className="sprint-date">Unassigned tasks</div>
                </div>

                {sprints.map(s => (
                  <div 
                    key={s.id} 
                    className={`sprint-item ${selectedSprintId === s.id ? 'active' : ''}`}
                    onClick={(e) => {
                      e.stopPropagation();
                      setSelectedSprintId(s.id);
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                      <div style={{ flex: 1 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <span className="sprint-name">{s.name}</span>
                          <button 
                            className="icon-btn-subtle" 
                            onClick={(e) => {
                              e.stopPropagation();
                              setIsEditingSprint(true);
                              setEditingSprintId(s.id);
                              setEditingSprint({
                                name: s.name,
                                goal: s.goal || '',
                                start_date: s.start_date.split('T')[0],
                                end_date: s.end_date.split('T')[0]
                              });
                              setIsSprintModalOpen(true);
                            }}
                            title="Edit Sprint"
                          >
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                            </svg>
                          </button>
                        </div>
                        <span className="sprint-date">{new Date(s.start_date).toLocaleDateString(undefined, { day: 'numeric', month: 'short' })} - {new Date(s.end_date).toLocaleDateString(undefined, { day: 'numeric', month: 'short' })}</span>
                      </div>
                      <span className={`status-badge ${s.status}`}>{s.status}</span>
                    </div>
                  </div>
                ))}
                {sprints.length === 0 && (
                  <div className="empty-state" style={{ padding: '2rem 1rem', textAlign: 'center' }}>
                    <p style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)', marginBottom: '1rem' }}>No active sprints</p>
                  </div>
                )}
                <button 
                  className="btn-secondary prestige-btn" 
                  style={{ width: '100%', marginTop: '0.5rem', background: 'var(--bg-input)', border: '1px dashed var(--border-color)', color: 'var(--text-secondary)' }}
                  onClick={() => {
                    setIsEditingSprint(false);
                    setEditingSprint({ name: '', goal: '', start_date: '', end_date: '' });
                    setIsSprintModalOpen(true);
                  }}
                >
                  + Create Sprint
                </button>
              </div>
            </div>
            <div className="planning-main">
               <div className="section-subtitle">
                  <span>{selectedSprintId ? sprints.find(s => s.id === selectedSprintId)?.name : 'Backlog Grooming'}</span>
                  <span className="col-count">
                    {tasks.filter(t => selectedSprintId ? t.sprint_id === selectedSprintId : !t.sprint_id).length}
                  </span>
               </div>
               <div className="backlog-container">
                  {tasks.filter(t => selectedSprintId ? t.sprint_id === selectedSprintId : !t.sprint_id).sort((a,b) => (a.order || 0) - (b.order || 0)).map(t => (
                    <div key={t.id} className="backlog-item" onClick={() => openEditModal(t)}>
                       <div className={`priority-mini ${t.priority.toLowerCase()}`}></div>
                       <div className="backlog-title">{t.title}</div>
                       <div className="backlog-meta">
                          <span>{t.priority.toUpperCase()}</span>
                          <span>•</span>
                          <span>{new Date(t.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}</span>
                       </div>
                    </div>
                  ))}
                  {tasks.filter(t => selectedSprintId ? t.sprint_id === selectedSprintId : !t.sprint_id).length === 0 && (
                    <div style={{ padding: '4rem', textAlign: 'center' }}>
                       <p style={{ color: 'var(--text-tertiary)', fontSize: '0.9rem', fontWeight: 600 }}>
                         {selectedSprintId ? 'No tasks in this sprint.' : 'Backlog is crystal clear.'}
                       </p>
                       <p style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                         {selectedSprintId ? 'Add tasks to this sprint to see them here.' : 'All tasks are assigned to sprints.'}
                       </p>
                    </div>
                  )}
               </div>
            </div>
          </div>
        )}

        {view === 'analytics' && (
          <div className="analytics-dashboard">
            <div className="analytics-row">
               <div className="analytics-card glass-island-premium">
                  <h4>Task Distribution</h4>
                  <div style={{ height: '250px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={statusLabels}
                          innerRadius={60}
                          outerRadius={80}
                          paddingAngle={5}
                          dataKey="value"
                        >
                          {statusLabels.map((_, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </div>
               </div>

               <div className="analytics-card glass-island-premium">
                  <h4>Priority Heatmap</h4>
                  <div style={{ height: '250px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={priorityData}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
                        <XAxis dataKey="name" fontSize={10} axisLine={false} tickLine={false} />
                        <YAxis fontSize={10} axisLine={false} tickLine={false} />
                        <Tooltip 
                          contentStyle={{ background: 'var(--surface-color)', border: '1px solid var(--border-color)', borderRadius: '12px' }}
                        />
                        <Bar dataKey="value" fill="var(--accent-color)" radius={[4, 4, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
               </div>
            </div>

            <div className="analytics-row" style={{ marginTop: '1.5rem' }}>
               <div className="analytics-card glass-island-premium" style={{ flex: 1 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                    <h4>Execution Velocity</h4>
                    <div className="velocity-badge">
                      <span className="v-label">Avg Lead Time:</span>
                      <span className="v-value">{analytics?.task_lead_time_days?.toFixed(1) || '0.0'} Days</span>
                    </div>
                  </div>
                  <div style={{ height: '300px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={[
                        { name: 'Mon', v: 4 }, { name: 'Tue', v: 7 }, { name: 'Wed', v: 5 }, 
                        { name: 'Thu', v: 8 }, { name: 'Fri', v: 12 }, { name: 'Sat', v: 6 }, { name: 'Sun', v: 9 }
                      ]}>
                        <defs>
                          <linearGradient id="colorVel" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="var(--accent-color)" stopOpacity={0.3}/>
                            <stop offset="95%" stopColor="var(--accent-color)" stopOpacity={0}/>
                          </linearGradient>
                        </defs>
                        <XAxis dataKey="name" fontSize={10} axisLine={false} tickLine={false} />
                        <Tooltip />
                        <Area type="monotone" dataKey="v" stroke="var(--accent-color)" fillOpacity={1} fill="url(#colorVel)" strokeWidth={3} />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
               </div>
            </div>
          </div>
        )}
      </div>

      {isTaskModalOpen && (
        <div className="modal-overlay" onClick={() => setIsTaskModalOpen(false)}>
          <div className="premium-modal wide-modal-lux" onClick={e => e.stopPropagation()}>
            <div className="modal-header-glass">
                <div className="modal-header-flex">
                   <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
                      <div className={`priority-indicator-lux ${(editingTask.priority || 'medium').toLowerCase()}`}></div>
                      <h2 style={{ margin: 0, fontWeight: 900, fontSize: '1.1rem', letterSpacing: '-0.02em', textTransform: 'uppercase' }}>
                        {isEditing ? 'Task Architecture' : 'Draft Objective'}
                      </h2>
                   </div>
                   {isEditing && (
                      <button className="delete-btn-lux" onClick={handleDeleteTask}>
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                      </button>
                   )}
                </div>
            </div>
            
            <div className="modal-layout-lux">
                <div className="modal-main-lux">
                    <div className="input-group-lux">
                        <label className="input-label-premium">System Title</label>
                        <input 
                          type="text" 
                          className="premium-input-lux headline-input-lux w-full" 
                          placeholder="What is the mission?"
                          value={editingTask.title}
                          onChange={e => setEditingTask({...editingTask, title: e.target.value})}
                          autoFocus
                        />
                    </div>

                    <div className="input-group-lux">
                        <label className="input-label-premium">Operational Context</label>
                        <textarea 
                          className="premium-input-lux w-full" 
                          rows={5} 
                          placeholder="Define the requirements and constraints..."
                          value={editingTask.description}
                          onChange={e => setEditingTask({...editingTask, description: e.target.value})}
                        />
                    </div>
                    
                    <div className="subtasks-section-lux">
                        <div className="subtask-header-lux">
                            <label className="input-label-premium" style={{ margin: 0 }}>Sub-Components</label>
                            <div className="progress-pill-lux">
                                {(editingTask.metadata?.subtasks || []).filter((s: Subtask) => s.completed).length}/{(editingTask.metadata?.subtasks || []).length} Complete
                            </div>
                        </div>
                        <div className="subtask-list-lux">
                            {(editingTask.metadata?.subtasks || []).map((sub: Subtask) => (
                                <div key={sub.id} className={`subtask-card-lux ${sub.completed ? 'completed' : ''}`}>
                                    <div className="subtask-check-wrapper" onClick={() => toggleSubtask(sub.id)}>
                                        <div className={`custom-checkbox-lux ${sub.completed ? 'checked' : ''}`}>
                                            {sub.completed && <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4"><polyline points="20 6 9 17 4 12"/></svg>}
                                        </div>
                                    </div>
                                    <span className="subtask-label-lux">{sub.title}</span>
                                    <button className="subtask-remove-lux" onClick={() => deleteSubtask(sub.id)}>
                                       <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                                    </button>
                                </div>
                            ))}
                        </div>
                        <div className="add-subtask-container-lux">
                            <input 
                                type="text"
                                className="premium-input-lux mini-input-lux"
                                placeholder="+ Integrate new component"
                                value={newSubtaskTitle}
                                onChange={e => setNewSubtaskTitle(e.target.value)}
                                onKeyPress={e => e.key === 'Enter' && addSubtask()}
                            />
                            <button className="add-action-btn-lux" onClick={addSubtask}>Deploy</button>
                        </div>
                    </div>
                </div>

                <div className="modal-sidebar-lux">
                    <div className="sidebar-segment-lux">
                        <label className="input-label-premium">Priority Level</label>
                        <select 
                            className="premium-input-lux w-full"
                            value={editingTask.priority}
                            onChange={e => setEditingTask({...editingTask, priority: e.target.value})}
                        >
                            <option value="low">Low Priority</option>
                            <option value="medium">Medium Priority</option>
                            <option value="high">High Priority</option>
                            <option value="urgent">Urgent Overclock</option>
                        </select>
                    </div>

                    {!isEditing && (
                        <div className="sidebar-segment-lux">
                            <label className="input-label-premium">Initial Status</label>
                            <select 
                                className="premium-input-lux w-full"
                                value={editingTask.column_id}
                                onChange={e => setEditingTask({...editingTask, column_id: parseInt(e.target.value)})}
                            >
                                {activeBoard?.columns.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                            </select>
                        </div>
                    )}

                    {activeBoard?.columns && isEditing && (
                         <div className="sidebar-segment-lux">
                            <label className="input-label-premium">Phase Indicator</label>
                            <div className="status-indicator-pill-lux">
                                <span className={`status-dot-lux ${activeBoard.columns.find(c => c.id === editingTask.column_id)?.name.toLowerCase().replace(' ', '-')}`}></span>
                                {activeBoard.columns.find(c => c.id === editingTask.column_id)?.name || 'STAGING'}
                            </div>
                         </div>
                    )}

                    <div className="sidebar-segment-lux">
                        <label className="input-label-premium">Deployment Sprint</label>
                        <select 
                            className="premium-input-lux w-full"
                            value={editingTask.sprint_id || ''}
                            onChange={e => setEditingTask({...editingTask, sprint_id: e.target.value ? parseInt(e.target.value) : undefined})}
                        >
                            <option value="">Global Backlog</option>
                            {sprints.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                        </select>
                    </div>

                    <div className="sidebar-footer-lux">
                       <p className="timestamp-lux">Created: {editingTask.created_at ? new Date(editingTask.created_at).toLocaleDateString() : 'N/A'}</p>
                    </div>
                </div>
            </div>

            <div className="modal-actions-footer-lux">
                <button className="btn-secondary-lux" onClick={() => setIsTaskModalOpen(false)}>Dismiss</button>
                <button className="btn-primary-lux prestige-btn" onClick={handleSaveTask}>
                    {isEditing ? 'Commit Changes' : 'Initialize Task'}
                </button>
            </div>
          </div>
        </div>
      )}

      {isSprintModalOpen && (
        <div className="modal-overlay" onClick={() => setIsSprintModalOpen(false)}>
          <div className="premium-modal wide-modal-lux" onClick={e => e.stopPropagation()}>
              <div className="modal-header-glass">
                <div className="modal-header-flex">
                   <div style={{ display: 'flex', gap: '1rem', width: '100%' }}>
                  {isEditingSprint && (
                    <button 
                      className="delete-btn-lux"
                      style={{ padding: '0 1.5rem', height: '48px', display: 'flex', alignItems: 'center', gap: '8px' }}
                      onClick={() => editingSprintId && handleDeleteSprint(editingSprintId)}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                        <path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                      </svg>
                      Delete
                    </button>
                  )}
                      <div className="header-icon-pill-mini">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
                      </div>
                      <h2 style={{ margin: 0, fontWeight: 900, fontSize: '1rem', letterSpacing: '0.05em', textTransform: 'uppercase' }}>{isEditingSprint ? 'Update Sprint' : 'Initialize Sprint'}</h2>
                   </div>
                   <button className="close-btn-lux" onClick={() => setIsSprintModalOpen(false)}>×</button>
                </div>
              </div>

             <div className="modal-body-premium">
                <div style={{ marginBottom: '1.5rem' }}>
                    <label className="input-label-premium">Sprint Name</label>
                    <input 
                      type="text" 
                      className="premium-input-lux w-full" 
                      placeholder="e.g. Q2 Launch"
                      value={editingSprint.name}
                      onChange={e => setEditingSprint({...editingSprint, name: e.target.value})}
                      autoFocus
                    />
                </div>
                <div style={{ marginBottom: '1.5rem' }}>
                    <label className="input-label-premium">Goal / Objective</label>
                    <textarea 
                      className="premium-input-lux w-full" 
                      rows={3}
                      placeholder="Define the core mission of this sprint..."
                      value={editingSprint.goal}
                      onChange={e => setEditingSprint({...editingSprint, goal: e.target.value})}
                    />
                </div>
                <div className="date-range-grid">
                    <div className="date-input-wrapper">
                        <label className="input-label-premium">Start Date</label>
                        <div className="custom-date-container">
                          <input 
                            type="date" 
                            className="premium-date-input" 
                            value={editingSprint.start_date}
                            onChange={e => setEditingSprint({...editingSprint, start_date: e.target.value})}
                          />
                          <div className="date-overlay-icon">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
                          </div>
                        </div>
                    </div>
                    <div className="date-input-wrapper">
                        <label className="input-label-premium">End Date</label>
                        <div className="custom-date-container">
                          <input 
                            type="date" 
                            className="premium-date-input" 
                            value={editingSprint.end_date}
                            onChange={e => setEditingSprint({...editingSprint, end_date: e.target.value})}
                          />
                          <div className="date-overlay-icon">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
                          </div>
                        </div>
                    </div>
                </div>
             </div>

             <div className="modal-actions-footer-lux">
                <button className="btn-secondary-lux" onClick={() => setIsSprintModalOpen(false)}>Dismiss</button>
                <button className="btn-primary-lux prestige-btn" onClick={handleSaveSprint}>
                  <span>{isEditingSprint ? 'Update Changes' : 'Initialize Sprint'}</span>
                  {!isEditingSprint && <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>}
                </button>
             </div>
          </div>
        </div>
      )}

      <style dangerouslySetInnerHTML={{ __html: `
        .planner-container {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
          height: calc(100vh - 180px);
          overflow: hidden;
          padding: 0 0.5rem;
        }
        .planner-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0.5rem 0;
        }
        .view-switcher-pill {
          display: flex;
          background: var(--bg-input);
          padding: 4px;
          border-radius: 14px;
          gap: 4px;
          box-shadow: inset 0 2px 4px rgba(0,0,0,0.05);
        }
        .view-pill-btn {
          padding: 0.5rem 1.5rem;
          border-radius: 10px;
          border: none;
          background: transparent;
          color: var(--text-tertiary);
          font-weight: 700;
          font-size: 0.8rem;
          cursor: pointer;
          transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        }
        .view-pill-btn.active {
          background: var(--surface-color);
          color: var(--accent-color);
          box-shadow: 0 4px 12px rgba(0,0,0,0.08);
        }
        .planner-viewport {
          flex: 1;
          min-height: 0;
          overflow: hidden;
        }
        .kanban-scroller {
          display: flex;
          gap: 1.5rem;
          height: 100%;
          overflow-x: auto;
          padding-bottom: 1rem;
          scrollbar-width: thin;
        }
        .kanban-col-wrapper {
          flex: 0 0 320px;
          background: var(--bg-input);
          border-radius: 20px;
          display: flex;
          flex-direction: column;
          border: 1px solid var(--border-color);
          transition: transform 0.2s;
        }
        .kanban-col-header {
          padding: 1.25rem;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }
        .col-title-group {
          display: flex;
          align-items: center;
          gap: 0.75rem;
        }
        .col-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: var(--accent-color);
        }
        .col-name {
          font-weight: 800;
          font-size: 0.8rem;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          color: var(--text-primary);
        }
        .col-count {
          font-size: 0.7rem;
          background: var(--surface-color);
          padding: 2px 8px;
          border-radius: 8px;
          color: var(--text-tertiary);
          font-weight: 700;
        }
        .task-drop-zone {
          flex: 1;
          padding: 0 1rem 1.5rem;
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
          overflow-y: auto;
        }
        .task-card {
          background: var(--surface-color);
          padding: 1.25rem;
          border-radius: 16px;
          border: 1px solid var(--border-color);
          box-shadow: 0 4px 6px rgba(0,0,0,0.02);
          cursor: grab;
          transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
          user-select: none;
        }
        .task-card:hover {
          border-color: var(--accent-color);
          transform: translateY(-2px);
          box-shadow: 0 8px 16px rgba(0,0,0,0.06);
        }
        .task-card:active {
          cursor: grabbing;
        }
        .dragging-overlay {
          width: 280px;
          border: 2px solid var(--accent-color);
          box-shadow: 0 20px 40px rgba(0,0,0,0.15);
          opacity: 0.9;
          transform: scale(1.05);
        }
        .priority-tag {
          font-size: 0.65rem;
          text-transform: uppercase;
          font-weight: 800;
          letter-spacing: 0.05em;
          padding: 2px 8px;
          border-radius: 4px;
          background: var(--bg-hover);
        }
        .priority-urgent { color: #ef4444; border: 1px solid #fee2e2; }
        .priority-high { color: #f59e0b; border: 1px solid #fef3c7; }
        .priority-medium { color: var(--accent-color); border: 1px solid var(--nav-active-bg); }
        .priority-low { color: var(--text-tertiary); }
        
        .subtask-mini-pill {
            font-size: 0.65rem;
            color: var(--text-tertiary);
            font-weight: 700;
            background: var(--bg-input);
            padding: 2px 6px;
            border-radius: 6px;
        }

        .empty-col-state {
          border: 2px dashed var(--border-color);
          border-radius: 12px;
          padding: 2rem;
          text-align: center;
          color: var(--text-tertiary);
          font-size: 0.8rem;
          font-weight: 600;
        }

        /* Modal Layout */
        .wide-modal {
            max-width: 800px !important;
            width: 90% !important;
        }
        .modal-layout {
            display: grid;
            grid-template-columns: 1fr 240px;
            gap: 2.5rem;
        }
        .headline-input {
            font-size: 1.25rem;
            font-weight: 700;
            padding: 1rem;
            border-radius: 16px;
            background: var(--bg-input);
        }
        .modal-actions-footer {
            display: flex;
            justify-content: flex-end;
            gap: 1rem;
            margin-top: 2.5rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--border-color);
        }
        .status-indicator-pill {
            background: var(--nav-active-bg);
            color: var(--accent-color);
            padding: 8px 12px;
            border-radius: 10px;
            font-size: 0.8rem;
            font-weight: 800;
            text-transform: uppercase;
        }

        /* Subtasks Styling */
        .subtasks-section {
            background: var(--bg-input);
            padding: 1.5rem;
            border-radius: 20px;
            border: 1px solid var(--border-color);
        }
        .subtasks-count {
            font-size: 0.7rem;
            color: var(--text-tertiary);
            font-weight: 800;
        }
        .subtask-list {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
            margin: 1rem 0;
        }
        .subtask-item {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            background: var(--surface-color);
            padding: 0.75rem 1rem;
            border-radius: 12px;
            border: 1px solid var(--border-color);
            transition: all 0.2s;
        }
        .subtask-item.completed {
            opacity: 0.6;
        }
        .subtask-item.completed .subtask-title {
            text-decoration: line-through;
            color: var(--text-tertiary);
        }
        .subtask-checkbox {
            width: 18px;
            height: 18px;
            accent-color: var(--accent-color);
            cursor: pointer;
        }
        .subtask-title {
            flex: 1;
            font-size: 0.85rem;
            font-weight: 600;
            color: var(--text-primary);
        }
        .delete-sub-btn {
            background: transparent;
            border: none;
            color: var(--text-tertiary);
            font-size: 1.25rem;
            cursor: pointer;
            padding: 0 4px;
        }
        .delete-sub-btn:hover { color: #ef4444; }

        .add-subtask-wrapper {
            display: flex;
            gap: 0.5rem;
            margin-top: 1rem;
        }
        .premium-input-mini {
            flex: 1;
            background: var(--surface-color);
            border: 1px solid var(--border-color);
            padding: 0.6rem 1rem;
            border-radius: 10px;
            font-size: 0.85rem;
        }
        .btn-add-sub {
            background: var(--surface-color);
            border: 1px solid var(--accent-color);
            color: var(--accent-color);
            padding: 0 1rem;
            border-radius: 10px;
            font-size: 0.75rem;
            font-weight: 800;
            cursor: pointer;
        }

        /* Planning View (Backlog Item) */
        .backlog-item {
            cursor: pointer;
            transition: transform 0.2s;
        }
        .backlog-item:hover {
            transform: translateX(4px);
            border-color: var(--accent-color);
        }

        /* Planning View & General */
        /* Planning View & General */
        .planning-grid {
          display: grid;
          grid-template-columns: 350px 1fr;
          gap: 2rem;
          height: 100%;
          padding-top: 0.5rem;
        }
        .planning-sidebar {
          background: rgba(248, 250, 252, 0.02);
          border-radius: 24px;
          padding: 2rem;
          border: 1px solid var(--border-color);
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }
        .section-subtitle {
          font-size: 0.75rem;
          font-weight: 800;
          text-transform: uppercase;
          letter-spacing: 0.15em;
          color: var(--text-tertiary);
          margin-bottom: 1rem;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }
        .sprint-list {
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }
        .sprint-item {
          padding: 1.25rem;
          border-radius: 18px;
          background: var(--surface-color);
          border: 1px solid var(--border-color);
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
          cursor: pointer;
          position: relative;
          overflow: hidden;
        }
        .sprint-item::before {
          content: '';
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          width: 4px;
          background: transparent;
          transition: background 0.3s;
        }
        .sprint-item.active {
          border-color: var(--accent-color);
          box-shadow: 0 10px 25px -5px rgba(99, 102, 241, 0.15);
        }
        .sprint-item.active::before {
          background: var(--accent-color);
        }
        .sprint-item:hover {
          transform: translateY(-2px);
          border-color: var(--accent-color);
        }
        .sprint-name { 
          font-weight: 700; 
          font-size: 0.95rem;
          color: var(--text-primary); 
          display: block;
          margin-bottom: 0.25rem;
        }
        .status-badge { 
          font-size: 0.6rem; 
          padding: 2px 8px; 
          border-radius: 20px; 
          text-transform: uppercase; 
          font-weight: 800; 
          letter-spacing: 0.05em;
        }
        .status-badge.active { 
          background: rgba(16, 185, 129, 0.1); 
          color: #10b981; 
          border: 1px solid rgba(16, 185, 129, 0.2);
        }
        .status-badge.planned { 
          background: var(--bg-input); 
          color: var(--text-tertiary); 
          border: 1px solid var(--border-color);
        }
        .sprint-date { 
          font-size: 0.75rem; 
          color: var(--text-tertiary); 
          font-weight: 500;
        }

        .planning-main {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }
        .backlog-container {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
          background: rgba(248, 250, 252, 0.01);
          padding: 0.5rem;
          border-radius: 24px;
        }
        .backlog-item {
          display: flex;
          align-items: center;
          gap: 1.25rem;
          padding: 1rem 1.5rem;
          background: var(--surface-color);
          border-radius: 16px;
          border: 1px solid var(--border-color);
          transition: all 0.2s ease;
          cursor: pointer;
        }
        .backlog-item:hover {
          border-color: var(--accent-color);
          transform: translateX(6px);
          box-shadow: var(--shadow-sm);
        }
        .priority-mini {
          width: 4px;
          height: 24px;
          border-radius: 2px;
          flex-shrink: 0;
        }
        .priority-mini.urgent { background: #ef4444; }
        .priority-mini.high { background: #f59e0b; }
        .priority-mini.medium { background: var(--accent-color); }
        .priority-mini.low { background: var(--text-tertiary); }

        .backlog-title {
          font-weight: 600;
          font-size: 0.9rem;
          color: var(--text-primary);
          flex: 1;
        }
        .backlog-meta {
          display: flex;
          align-items: center;
          gap: 1rem;
          font-size: 0.75rem;
          color: var(--text-tertiary);
        }

        .btn-danger-outline {
            background: transparent;
            border: 1px solid rgba(239, 68, 68, 0.3);
            color: #ef4444;
            padding: 6px 12px;
            border-radius: 10px;
            font-size: 0.75rem;
            font-weight: 700;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        .btn-danger-outline:hover {
            background: #fef2f2;
            border-color: #ef4444;
        }

        /* Analytics View */
        .analytics-dashboard {
          padding: 1rem;
          height: 100%;
          overflow-y: auto;
        }
        .analytics-row {
          display: flex;
          gap: 1.5rem;
          flex-wrap: wrap;
        }
        .analytics-card {
          flex: 1;
          min-width: 300px;
          padding: 1.5rem;
          background: var(--surface-color);
          border-radius: 24px;
          border: 1px solid var(--border-color);
        }
        .analytics-card h4 {
          margin-bottom: 1.5rem;
          font-weight: 800;
          text-transform: uppercase;
          font-size: 0.8rem;
          letter-spacing: 0.1em;
          color: var(--text-secondary);
        }
        .velocity-badge {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          background: var(--bg-input);
          padding: 6px 14px;
          border-radius: 20px;
        }
        .v-label { font-size: 0.7rem; color: var(--text-tertiary); font-weight: 700; }
        .v-value { font-size: 0.8rem; color: var(--accent-color); font-weight: 800; }

        .prestige-btn {
          box-shadow: 0 10px 20px -5px rgba(99, 102, 241, 0.3);
          transition: all 0.3s;
        }
        .prestige-btn:hover {
          transform: translateY(-2px) scale(1.02);
          box-shadow: 0 15px 30px -5px rgba(99, 102, 241, 0.4);
        }

        /* Lux Modals Update */
        .modal-header-glass {
          padding: 2rem 2.5rem;
          background: linear-gradient(to right, rgba(99, 102, 241, 0.08), transparent);
          display: flex;
          align-items: center;
          gap: 1.25rem;
          border-bottom: 1px solid var(--border-color);
        }
        .header-icon-pill {
          width: 48px;
          height: 48px;
          background: var(--nav-active-bg);
          color: var(--accent-color);
          border-radius: 14px;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        /* Task Details Modal Premium Styles */
         .wide-modal-lux {
           width: 900px;
           max-width: 95vw;
           background: var(--surface-color);
           border-radius: 32px;
           border: 1px solid var(--border-color);
           box-shadow: var(--shadow-xl), 0 0 0 1px rgba(255,255,255,0.05);
           overflow: hidden;
           display: flex;
           flex-direction: column;
           max-height: 90vh;
         }
         .modal-header-flex {
           display: flex;
           justify-content: space-between;
           align-items: center;
           width: 100%;
           padding: 1.5rem 2rem;
         }
         .modal-layout-lux {
           display: grid;
           grid-template-columns: 1.8fr 1fr;
           flex: 1;
           overflow: hidden;
         }
         .modal-main-lux {
           padding: 2rem;
           overflow-y: auto;
           scrollbar-width: none;
         }
         .modal-sidebar-lux {
           padding: 2rem;
           border-left: 1px solid var(--border-color);
           background: rgba(0,0,0,0.02);
           overflow-y: auto;
         }
         .input-group-lux { margin-bottom: 2rem; }
         .headline-input-lux {
           font-size: 1.5rem;
           font-weight: 800;
           letter-spacing: -0.01em;
           border-color: transparent !important;
           padding-left: 0 !important;
         }
         .headline-input-lux:focus { background: transparent !important; }
         
         .priority-indicator-lux {
           width: 12px;
           height: 12px;
           border-radius: 4px;
         }
         .priority-indicator-lux.low { background: #10b981; }
         .priority-indicator-lux.medium { background: #6366f1; }
         .priority-indicator-lux.high { background: #f59e0b; }
         .priority-indicator-lux.urgent { background: #ef4444; box-shadow: 0 0 12px rgba(239, 68, 68, 0.4); }

         .delete-btn-lux {
           background: rgba(239, 68, 68, 0.1);
           color: #ef4444;
           border: 1px solid rgba(239, 68, 68, 0.2);
           padding: 0.75rem;
           border-radius: 14px;
           cursor: pointer;
           transition: all 0.2s;
         }
         .delete-btn-lux:hover { background: #ef4444; color: white; }

         .subtasks-section-lux {
           margin-top: 2rem;
           background: var(--bg-input);
           border-radius: 24px;
           padding: 1.5rem;
           border: 1px solid var(--border-color);
         }
         .subtask-header-lux {
           display: flex;
           justify-content: space-between;
           align-items: center;
           margin-bottom: 1.25rem;
         }
         .progress-pill-lux {
           font-size: 0.7rem;
           font-weight: 900;
           color: var(--accent-color);
           background: rgba(99, 102, 241, 0.1);
           padding: 0.4rem 0.8rem;
           border-radius: 10px;
           text-transform: uppercase;
         }
         .subtask-list-lux {
           display: flex;
           flex-direction: column;
           gap: 0.75rem;
           margin-bottom: 1.5rem;
         }
         .subtask-card-lux {
           background: var(--surface-color);
           padding: 1rem 1.25rem;
           border-radius: 16px;
           display: flex;
           align-items: center;
           gap: 1rem;
           border: 1px solid var(--border-color);
           transition: all 0.2s;
         }
         .subtask-card-lux:hover { transform: translateX(4px); border-color: var(--accent-color); }
         .subtask-card-lux.completed { opacity: 0.6; background: rgba(0,0,0,0.02); }
         .subtask-card-lux.completed .subtask-label-lux { text-decoration: line-through; }
         .subtask-check-wrapper { cursor: pointer; }

          .icon-btn-subtle {
            background: transparent;
            border: none;
            color: var(--text-tertiary);
            padding: 4px;
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            justify-content: center;
            opacity: 0.6;
          }
          .icon-btn-subtle:hover {
            background: var(--nav-active-bg);
            color: var(--accent-color);
            opacity: 1;
            transform: scale(1.1);
          }
         .custom-checkbox-lux {
           width: 22px;
           height: 22px;
           border-radius: 7px;
           border: 2px solid var(--border-color);
           display: flex;
           align-items: center;
           justify-content: center;
           transition: all 0.2s;
         }
         .custom-checkbox-lux.checked {
           background: var(--accent-color);
           border-color: var(--accent-color);
           color: white;
         }
         .subtask-label-lux {
           flex: 1;
           font-size: 0.85rem;
           font-weight: 600;
           color: var(--text-primary);
         }
         .subtask-remove-lux {
           background: transparent;
           border: none;
           color: var(--text-tertiary);
           cursor: pointer;
           padding: 4px;
           border-radius: 6px;
         }
         .subtask-remove-lux:hover { color: #ef4444; background: rgba(239, 68, 68, 0.05); }

         .add-subtask-container-lux {
           display: flex;
           gap: 0.75rem;
         }
         .mini-input-lux { font-size: 0.85rem !important; padding: 0.8rem 1.25rem !important; }
         .add-action-btn-lux {
           background: var(--text-primary);
           color: var(--surface-color);
           border: none;
           padding: 0 1.25rem;
           border-radius: 14px;
           font-weight: 800;
           font-size: 0.75rem;
           cursor: pointer;
         }

         .sidebar-segment-lux { margin-bottom: 2.25rem; }
         .status-dot-lux {
           width: 8px;
           height: 8px;
           border-radius: 50%;
           background: #6366f1;
         }
         .status-dot-lux.todo { background: #94a3b8; }
         .status-dot-lux.in-progress { background: #6366f1; }
         .status-dot-lux.done { background: #10b981; }
         
         .sidebar-footer-lux {
           margin-top: auto;
           padding-top: 2rem;
         }
         .timestamp-lux {
           font-size: 0.65rem;
           color: var(--text-tertiary);
           font-weight: 700;
           text-transform: uppercase;
           letter-spacing: 0.05em;
         }
        .modal-body-premium {
          padding: 2.5rem;
        }
        .input-label-premium {
          display: block;
          font-size: 0.75rem;
          font-weight: 800;
          text-transform: uppercase;
          letter-spacing: 0.1em;
          color: var(--text-tertiary);
          margin-bottom: 0.75rem;
        }
        .premium-input-lux {
          background: var(--bg-input);
          border: 1px solid var(--border-color);
          padding: 1rem 1.25rem;
          border-radius: 16px;
          color: var(--text-primary);
          font-family: inherit;
          font-size: 0.95rem;
          transition: all 0.3s;
          box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
        }
        .premium-input-lux:focus {
          outline: none;
          border-color: var(--accent-color);
          background: var(--surface-color);
          box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.05);
        }
        .date-range-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 1.5rem;
        }
        .custom-date-container {
          position: relative;
          display: flex;
          align-items: center;
        }
        .premium-date-input {
          width: 100%;
          background: var(--bg-input);
          border: 1px solid var(--border-color);
          padding: 1rem 1.25rem;
          padding-right: 3rem;
          border-radius: 16px;
          color: var(--text-primary);
          font-family: inherit;
          cursor: pointer;
          transition: all 0.3s;
          -webkit-appearance: none;
        }
        .premium-date-input:hover { border-color: var(--accent-color); }
        .date-overlay-icon {
          position: absolute;
          right: 1.25rem;
          color: var(--text-tertiary);
          pointer-events: none;
        }
         .modal-header-glass {
           padding: 0;
           background: rgba(var(--accent-color-rgb), 0.04);
           border-bottom: 1px solid var(--border-color);
         }
        .modal-actions-footer-lux {
          padding: 1.5rem 2.5rem 2.5rem;
          display: flex;
          gap: 1.25rem;
          border-top: 1px solid var(--border-color);
        }
        .btn-primary-lux {
          background: var(--accent-color);
          color: white;
          border: none;
          padding: 1.1rem 2rem;
          border-radius: 16px;
          font-weight: 800;
          font-size: 0.95rem;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.75rem;
          flex: 1.5;
        }
        .btn-secondary-lux {
          background: var(--bg-input);
          color: var(--text-secondary);
          border: 1px solid var(--border-color);
          padding: 1.1rem 1.5rem;
          border-radius: 16px;
          font-weight: 700;
          cursor: pointer;
          flex: 1;
          transition: all 0.2s;
        }
        .btn-secondary-lux:hover {
          background: var(--surface-color);
          color: var(--text-primary);
        }

        /* Task Details Pill Updates */
        .status-indicator-pill-lux {
          display: inline-flex;
          align-items: center;
          gap: 0.6rem;
          padding: 0.6rem 1.25rem;
          border-radius: 14px;
          font-size: 0.75rem;
          font-weight: 800;
          background: var(--bg-input);
          border: 1px solid var(--border-color);
          color: var(--text-primary);
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }
        .status-pill-blue { background: rgba(99, 102, 241, 0.1); color: #6366f1; border: 1px solid rgba(99, 102, 241, 0.2); }
        .status-pill-green { background: rgba(16, 185, 129, 0.1); color: #10b981; border: 1px solid rgba(16, 185, 129, 0.2); }
        .status-pill-orange { background: rgba(245, 158, 11, 0.1); color: #f59e0b; border: 1px solid rgba(245, 158, 11, 0.2); }
        .status-pill-slate { background: var(--bg-input); color: var(--text-tertiary); border: 1px solid var(--border-color); }
      `}} />
    </div>
  );
};
