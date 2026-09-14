import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { SocialQueueDashboard } from './SocialQueueDashboard';
import { SocialQueueUploaderModal } from './SocialQueueUploaderModal';

interface SocialQueuePageProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  appConfigs?: Record<string, string>;
  onNavigate?: (tab: string) => void;
}

export const SocialQueuePage: React.FC<SocialQueuePageProps> = ({ fetchWithAuth, appConfigs, onNavigate }) => {
  const [queuePosts, setQueuePosts] = useState<any[]>([]);
  const [isQueueModalOpen, setIsQueueModalOpen] = useState(false);

  const fetchQueuePosts = async () => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/marketing/smm/queue`);
      const data = await res.json();
      if (data.success && data.posts) {
        setQueuePosts(data.posts);
      }
    } catch (err) {
      console.error('Failed to fetch queue posts:', err);
    }
  };

  useEffect(() => {
    fetchQueuePosts();
  }, []);

  const gdriveFolderUrl = appConfigs?.gdrive_automation_folder_url || 'https://drive.google.com/drive/folders/1djXkok8cuP3efyurTd2nOwoKRo-HpEC3';

  return (
    <div className="social-queue-page tab-content-fade" style={{ animation: 'fadeIn 0.4s ease-out' }}>
      <SocialQueueDashboard
        queuePosts={queuePosts}
        gdriveFolderUrl={gdriveFolderUrl}
        onOpenUploader={() => setIsQueueModalOpen(true)}
        onRefresh={fetchQueuePosts}
        onNavigate={onNavigate}
      />

      <SocialQueueUploaderModal
        isOpen={isQueueModalOpen}
        gdriveFolderUrl={gdriveFolderUrl}
        onClose={() => setIsQueueModalOpen(false)}
        onSuccess={fetchQueuePosts}
        fetchWithAuth={fetchWithAuth}
      />
    </div>
  );
};
