import React, { useState, useEffect, useMemo } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';
import './JudgeMeReviews.css';

interface DraftReview {
  id: string;
  product_id: string;
  product_title: string;
  reviewer_name: string;
  gender: string;
  email: string;
  rating: number;
  title: string;
  body: string;
  shop_domain: string;
}

interface PublishedReview {
  id: number;
  review_id: string;
  product_id: string;
  product_title: string;
  reviewer_name: string;
  gender: string;
  email: string;
  rating: number;
  title: string;
  body: string;
  shop_domain: string;
  status: string;
  status_code: number;
  published_at: string;
}

interface Product {
  id: number;
  title: string;
  mi_sku: string;
}

interface SubmissionResult {
  index: number;
  product_id: string;
  product_title: string;
  reviewer_name: string;
  email: string;
  status: string;
  status_code: number;
  response_body: string;
}

const MAX_SELECTED_PRODUCTS = 10;
const PUBLISHED_REVIEWS_PAGE_SIZE = 10;

const getInitials = (name: string) => {
  if (!name) return '??';
  const parts = name.split(' ').filter(p => p.length > 0);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase();
};

export const JudgeMeReviews: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm } = useConfirm();

  // Active Sub-Tab state: 'drafts' | 'published'
  const [activeSubTab, setActiveSubTab] = useState<'drafts' | 'published'>('drafts');

  // Configuration state
  const [shopDomain, setShopDomain] = useState('4296fb-8e.myshopify.com');
  const [email, setEmail] = useState('hari.crze.101@gmail.com');
  const [aliasEmail, setAliasEmail] = useState(false);
  const [reviewsPerProduct, setReviewsPerProduct] = useState(1);
  const [selectedProductIDs, setSelectedProductIDs] = useState<string[]>([]);
  const [selectionWarning, setSelectionWarning] = useState<string | null>(null);
  const [rangeFrom, setRangeFrom] = useState('');
  const [rangeTo, setRangeTo] = useState('');

  const handleApplySkuRange = (fromStr: string, toStr: string) => {
    if (!fromStr && !toStr) return;

    const parseNum = (s: string) => {
      const match = s.match(/\d+/);
      return match ? parseInt(match[0], 10) : null;
    };

    const startNum = parseNum(fromStr);
    const endNum = parseNum(toStr);

    if (startNum === null && endNum === null) return;

    const minVal = startNum !== null && endNum !== null ? Math.min(startNum, endNum) : (startNum ?? endNum!);
    const maxVal = startNum !== null && endNum !== null ? Math.max(startNum, endNum) : (endNum ?? startNum!);

    const matchedIDs = products.filter((p, index) => {
      const skuNum = parseNum(p.mi_sku || '');
      const idxNum = index + 1;
      const matchSku = skuNum !== null && skuNum >= minVal && skuNum <= maxVal;
      const matchIdx = idxNum >= minVal && idxNum <= maxVal;
      return matchSku || matchIdx;
    }).map(p => p.id.toString());

    const cappedIDs = matchedIDs.slice(0, MAX_SELECTED_PRODUCTS);
    setSelectedProductIDs(cappedIDs);
    setSelectionWarning(
      matchedIDs.length > MAX_SELECTED_PRODUCTS
        ? `Only the first ${MAX_SELECTED_PRODUCTS} products in this range were selected.`
        : null
    );
  };

  // UI Modal & Drawer states
  const [isGenerateModalOpen, setIsGenerateModalOpen] = useState(false);
  const [selectedPublishedReview, setSelectedPublishedReview] = useState<PublishedReview | null>(null);

  // Data state
  const [products, setProducts] = useState<Product[]>([]);
  const [reviews, setReviews] = useState<DraftReview[]>([]);
  const [selectedReviewIDs, setSelectedReviewIDs] = useState<string[]>([]);
  const [publishedReviews, setPublishedReviews] = useState<PublishedReview[]>([]);
  const [publishedTotal, setPublishedTotal] = useState<number>(0);
  const [publishedPage, setPublishedPage] = useState<number>(1);

  // Execution & Loading states
  const [isGenerating, setIsGenerating] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoadingPublished, setIsLoadingPublished] = useState(false);
  const [dryRun, setDryRun] = useState(false);
  const [delayMs, setDelayMs] = useState(1200);
  const [ratingFilter, setRatingFilter] = useState<number | 0>(0);
  const [submissionProgress, setSubmissionProgress] = useState<{ current: number; total: number; statusText: string } | null>(null);
  const [resultsReport, setResultsReport] = useState<SubmissionResult[] | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    return fetch(url, { ...options, headers });
  };

  // Fetch catalog products
  useEffect(() => {
    const loadProducts = async () => {
      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/inventory`);
        if (resp.ok) {
          const data = await resp.json();
          setProducts(data);
        }
      } catch (err) {
        console.error('Failed to load products', err);
      }
    };
    loadProducts();
  }, []);

  // Fetch published reviews history from DB
  const loadPublishedReviews = async (page = publishedPage) => {
    setIsLoadingPublished(true);
    try {
      const queryParams = new URLSearchParams({
        page: page.toString(),
        limit: PUBLISHED_REVIEWS_PAGE_SIZE.toString(),
        search: searchQuery
      });
      const resp = await fetchWithAuth(`${API_BASE}/api/marketing/judgeme/published?${queryParams.toString()}`);
      if (resp.ok) {
        const data = await resp.json();
        setPublishedReviews(data.reviews || []);
        setPublishedTotal(data.total || 0);
      }
    } catch (err) {
      console.error('Failed to load published reviews', err);
    } finally {
      setIsLoadingPublished(false);
    }
  };

  useEffect(() => {
    if (activeSubTab === 'published') {
      loadPublishedReviews();
    }
  }, [activeSubTab, publishedPage, searchQuery]);

  // Max 10 calculation logic
  const calculatedTotalBatchSize = useMemo(() => {
    const targetProductCount = selectedProductIDs.length === 0 ? MAX_SELECTED_PRODUCTS : selectedProductIDs.length;
    return Math.min(MAX_SELECTED_PRODUCTS, targetProductCount * reviewsPerProduct);
  }, [selectedProductIDs, reviewsPerProduct]);

  const toggleSelectProductInModal = (idStr: string) => {
    setSelectedProductIDs(prev => {
      if (prev.includes(idStr)) {
        setSelectionWarning(null);
        return prev.filter(i => i !== idStr);
      }
      if (prev.length >= MAX_SELECTED_PRODUCTS) {
        setSelectionWarning(`You can select a maximum of ${MAX_SELECTED_PRODUCTS} products per batch.`);
        return prev;
      }
      setSelectionWarning(null);
      return [...prev, idStr];
    });
  };

  const handleGenerateReviews = async () => {
    setIsGenerating(true);
    setResultsReport(null);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/marketing/judgeme/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          shop_domain: shopDomain,
          email: email,
          alias_email: aliasEmail,
          reviews_per_product: reviewsPerProduct,
          product_ids: selectedProductIDs.slice(0, MAX_SELECTED_PRODUCTS)
        })
      });

      if (resp.ok) {
        const data: DraftReview[] = await resp.json();
        setReviews(data);
        setIsGenerateModalOpen(false);
        setActiveSubTab('drafts');
        toastSuccess(`Generated ${data.length} review drafts (Max 10 batch limit)!`);
      } else {
        toastError('Failed to generate review drafts');
      }
    } catch (err) {
      toastError('Error generating reviews');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleUpdateReview = (id: string, field: keyof DraftReview, value: any) => {
    setReviews(prev => prev.map(r => r.id === id ? { ...r, [field]: value } : r));
  };

  const handleDeleteReview = (id: string) => {
    setReviews(prev => prev.filter(r => r.id !== id));
    setSelectedReviewIDs(prev => prev.filter(i => i !== id));
  };

  const handleBulkDelete = () => {
    setReviews(prev => prev.filter(r => !selectedReviewIDs.includes(r.id)));
    setSelectedReviewIDs([]);
    toastSuccess('Selected reviews deleted');
  };

  const handleAddManualReview = () => {
    if (reviews.length >= 10) {
      toastError('Maximum 10 draft reviews allowed in approval queue at a time');
      return;
    }
    const firstProduct = products.length > 0 ? products[0] : null;
    const shopifyMapping = (firstProduct as any)?.mappings?.find((m: any) => m.platform === 'shopify');
    const defaultProductID = shopifyMapping?.external_variant_id || '';

    const newDraft: DraftReview = {
      id: `custom_${Date.now()}`,
      product_id: defaultProductID,
      product_title: firstProduct ? firstProduct.title : 'Aeros',
      reviewer_name: 'Hemanth Kumar',
      gender: 'male',
      email: email,
      rating: 5,
      title: 'Amazing drydown & projection!',
      body: 'Genuinely impressed with the scent quality. Lasts easily 7+ hours on clothes.',
      shop_domain: shopDomain
    };
    setReviews(prev => [newDraft, ...prev]);
  };

  const handleSubmitToJudgeMe = async () => {
    const targetReviews = selectedReviewIDs.length > 0
      ? reviews.filter(r => selectedReviewIDs.includes(r.id))
      : reviews;

    if (targetReviews.length === 0) {
      toastError('No reviews selected to submit');
      return;
    }

    const confirmed = await confirm({
      title: dryRun ? 'Test Submission (Dry Run)' : 'Publish Reviews to Judge.me',
      message: `Are you sure you want to ${dryRun ? 'test submit' : 'live publish'} ${targetReviews.length} reviews to Judge.me?`,
      confirmLabel: dryRun ? 'Run Test' : 'Publish Live'
    });

    if (!confirmed) return;

    setIsSubmitting(true);
    setSubmissionProgress({ current: 0, total: targetReviews.length, statusText: 'Connecting to Judge.me API...' });

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/marketing/judgeme/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          reviews: targetReviews,
          delay_ms: delayMs,
          dry_run: dryRun
        })
      });

      if (resp.ok) {
        const report = await resp.json();
        setResultsReport(report.results || []);

        // Remove submitted items from drafts queue
        const submittedIDs = targetReviews.map(r => r.id);
        setReviews(prev => prev.filter(r => !submittedIDs.includes(r.id)));
        setSelectedReviewIDs([]);

        toastSuccess(`Processed ${report.successful}/${report.total_processed} reviews! Saved to database.`);
        setPublishedPage(1);
        loadPublishedReviews(1);
      } else {
        toastError('Failed to submit reviews to Judge.me');
      }
    } catch (err) {
      toastError('Submission error occurred');
    } finally {
      setIsSubmitting(false);
      setSubmissionProgress(null);
    }
  };

  const filteredDraftReviews = useMemo(() => {
    return reviews.filter(r => {
      if (ratingFilter > 0 && r.rating !== ratingFilter) return false;
      if (searchQuery) {
        const q = searchQuery.toLowerCase();
        return (
          r.reviewer_name.toLowerCase().includes(q) ||
          r.product_title.toLowerCase().includes(q) ||
          r.title.toLowerCase().includes(q) ||
          r.body.toLowerCase().includes(q)
        );
      }
      return true;
    });
  }, [reviews, searchQuery, ratingFilter]);

  const uniqueProductCount = useMemo(() => {
    return new Set(reviews.map(r => r.product_id)).size;
  }, [reviews]);

  const toggleSelectAllDrafts = () => {
    if (selectedReviewIDs.length === filteredDraftReviews.length) {
      setSelectedReviewIDs([]);
    } else {
      setSelectedReviewIDs(filteredDraftReviews.map(r => r.id));
    }
  };

  const toggleSelectReview = (id: string) => {
    setSelectedReviewIDs(prev =>
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    );
  };

  return (
    <div className="tab-content staggered-fade-in judgeme-page" style={{ paddingBottom: '5rem', maxWidth: '1440px', margin: '0 auto' }}>

      <header className="page-header judgeme-page__page-header">
        <div>
          <h1 className="page-title">Judge.me Reviews</h1>
          <p className="page-subtitle">Generate, review, and publish customer feedback for your store.</p>
        </div>
      </header>

      <section className="glass-island-premium judgeme-page__header" aria-label="Judge.me review workspace">
        <div className="judgeme-page__header-info">
          <div className="judgeme-page__header-icon" aria-hidden="true">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M7 18.5 3.5 21l1-4.2A8.5 8.5 0 1 1 7 18.5Z" />
              <path d="M8 11h.01M12 11h.01M16 11h.01" />
            </svg>
          </div>
          <div>
            <h2 className="judgeme-page__title">Judge.me Reviews</h2>
            <div className="judgeme-page__status-line">
              <span>Review workspace</span>
              <span className="judgeme-page__status-separator" aria-hidden="true" />
              <span className="judgeme-page__status-active">Batch limit: 10</span>
            </div>
          </div>
        </div>
        <div className="judgeme-page__header-actions">
          <button
            type="button"
            className="btn-secondary"
            onClick={handleAddManualReview}
          >
            + Single Custom Review
          </button>

          <button
            type="button"
            className="btn-primary"
            onClick={() => setIsGenerateModalOpen(true)}
          >
            Generate Reviews
          </button>
        </div>
      </section>

      {/* Sub-Tab Navigation (Clean Underline Bar) */}
      <div className="judgeme-tabs-nav">
        <button
          className={`judgeme-tab-btn ${activeSubTab === 'drafts' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('drafts')}
        >
          <span>Draft Queue</span>
          <span className="judgeme-tab-badge">{reviews.length}</span>
        </button>
        <button
          className={`judgeme-tab-btn ${activeSubTab === 'published' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('published')}
        >
          <span>Published History</span>
          <span className="judgeme-tab-badge">{publishedTotal}</span>
        </button>
      </div>

      {/* KPI Stats Summary Cards */}
      <div className="judgeme-page__metrics" style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
        gap: '1.25rem',
        marginBottom: '1.75rem'
      }}>
        <div className="metric-card judgeme-page__metric" style={{ padding: '1.25rem 1.5rem', margin: 0 }}>
          <div style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Draft Queue
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '0.25rem' }}>
            {reviews.length} / 10
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            {uniqueProductCount} products covered
          </div>
        </div>

        <div className="metric-card judgeme-page__metric" style={{ padding: '1.25rem 1.5rem', margin: 0 }}>
          <div style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Total Published
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: '#10b981', marginTop: '0.25rem' }}>
            {publishedTotal}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Stored in PostgreSQL DB
          </div>
        </div>

        <div className="metric-card judgeme-page__metric" style={{ padding: '1.25rem 1.5rem', margin: 0 }}>
          <div style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Active Store
          </div>
          <div style={{ fontSize: '0.95rem', fontWeight: 800, color: '#047857', marginTop: '0.5rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {shopDomain}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Judge.me API Verified
          </div>
        </div>

        <div className="metric-card judgeme-page__metric" style={{ padding: '1.25rem 1.5rem', margin: 0 }}>
          <div style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Products Catalog
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '0.25rem' }}>
            {products.length}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Available for review target
          </div>
        </div>
      </div>

      {/* Submission Progress Bar */}
      {submissionProgress && (
        <div style={{
          background: 'var(--bg-card)',
          border: '1px solid rgba(5, 150, 105, 0.42)',
          borderRadius: '14px',
          padding: '1.25rem 1.5rem',
          marginBottom: '1.75rem',
          boxShadow: '0 8px 24px rgba(5, 150, 105, 0.12)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem', fontWeight: 800, marginBottom: '0.5rem' }}>
            <span style={{ color: '#047857', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.25"><path d="M12 20V10"/><path d="m18 14-6-6-6 6"/><path d="M5 4h14"/></svg>
              {submissionProgress.statusText}
            </span>
            <span style={{ color: 'var(--text-secondary)' }}>
              {submissionProgress.current} / {submissionProgress.total} Processed
            </span>
          </div>
          <div style={{ height: '8px', background: 'var(--bg-input)', borderRadius: '4px', overflow: 'hidden' }}>
            <div style={{
              width: `${(submissionProgress.current / submissionProgress.total) * 100}%`,
              height: '100%',
              background: 'linear-gradient(90deg, #047857, #10b981)',
              transition: 'width 0.3s ease'
            }} />
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* SUB-TAB 1: DRAFT QUEUE & APPROVAL */}
      {/* ========================================================================= */}
      {activeSubTab === 'drafts' && (
        <>
          <section className="judgeme-page__toolbar" aria-label="Draft review controls">
            <div className="judgeme-page__filters">
              <div className="judgeme-page__search">
                <label className="judgeme-page__sr-only" htmlFor="review-search">Search draft reviews</label>
                <input
                  id="review-search"
                  type="text"
                  placeholder="Search reviewer, product, text..."
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                />
                <svg className="judgeme-page__search-icon" aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.25"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              </div>

              <label className="judgeme-page__select-field">
                <span className="judgeme-page__sr-only">Filter by rating</span>
              <select
                aria-label="Filter by rating"
                value={ratingFilter}
                onChange={e => setRatingFilter(parseInt(e.target.value))}
              >
                <option value={0}>All Ratings</option>
                <option value={5}>5 Stars ★★★★★</option>
                <option value={4}>4 Stars ★★★★</option>
              </select>
              </label>
            </div>

            <div className="judgeme-page__actions">
              <label className={`judgeme-page__dry-run ${dryRun ? 'is-active' : ''}`}>
                <input
                  type="checkbox"
                  checked={dryRun}
                  onChange={e => setDryRun(e.target.checked)}
                />
                <span className="judgeme-page__switch" aria-hidden="true" />
                <span>Dry run</span>
              </label>

              <label className="judgeme-page__select-field judgeme-page__delay-field">
                <span className="judgeme-page__sr-only">Submission delay</span>
              <select
                aria-label="Submission delay"
                value={delayMs}
                onChange={e => setDelayMs(parseInt(e.target.value))}
              >
                <option value={500}>500ms Delay</option>
                <option value={1200}>1200ms Delay</option>
                <option value={2000}>2000ms Delay</option>
              </select>
              </label>

              {selectedReviewIDs.length > 0 && (
                <button
                  type="button"
                  onClick={handleBulkDelete}
                  className="judgeme-page__delete-btn"
                >
                  Delete ({selectedReviewIDs.length})
                </button>
              )}

              <button
                type="button"
                className="judgeme-page__submit-btn"
                onClick={handleSubmitToJudgeMe}
                disabled={isSubmitting || reviews.length === 0}
              >
                {isSubmitting ? 'Publishing…' : selectedReviewIDs.length > 0 ? `Publish selected (${selectedReviewIDs.length})` : 'Publish drafts'}
              </button>
            </div>
          </section>

          {/* Draft Queue Table */}
          {reviews.length === 0 ? (
            <div className="judgeme-page__empty" style={{
              background: 'var(--bg-card)',
              border: '1px dashed var(--border-color)',
              textAlign: 'center',
              padding: '5rem 2rem',
              borderRadius: '20px'
            }}>
              <div style={{
                width: '56px',
                height: '56px',
                borderRadius: '16px',
                background: 'rgba(5, 150, 105, 0.08)',
                color: '#047857',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '1.5rem',
                margin: '0 auto 1.25rem auto'
              }}>
                <svg aria-hidden="true" width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                  <path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/>
                  <path d="M14.5 4.5 19.5 9.5"/>
                  <path d="m13 16 6.7-6.7a1.4 1.4 0 0 0-2-2L11 14l-1 3Z"/>
                </svg>
              </div>
              <h3 style={{ margin: 0, fontWeight: 800, fontSize: '1.15rem', color: 'var(--text-primary)' }}>
                No Draft Reviews in Approval Queue
              </h3>
              <p style={{ color: 'var(--text-tertiary)', maxWidth: '440px', margin: '0.5rem auto 1.5rem auto', fontSize: '0.875rem', lineHeight: 1.5 }}>
                Click <strong>"Generate Reviews"</strong> to select products and generate a max batch of 10 reviews for inspection.
              </p>
              <button
                type="button"
                className="btn-primary"
                onClick={() => setIsGenerateModalOpen(true)}
                style={{
                  height: '42px',
                  padding: '0 1.5rem',
                  borderRadius: '10px',
                  fontWeight: 800,
                  background: 'linear-gradient(135deg, #059669, #047857)'
                }}
              >
                Generate Reviews
              </button>
            </div>
          ) : (
            <div className="table-container glass-card-premium judgeme-page__table" style={{ borderRadius: '16px', overflow: 'hidden' }}>
              <table className="premium-table">
                <colgroup>
                  <col className="judgeme-page__select-column" />
                  <col className="judgeme-page__product-column" />
                  <col className="judgeme-page__reviewer-column" />
                  <col className="judgeme-page__rating-column" />
                  <col className="judgeme-page__title-column" />
                  <col className="judgeme-page__body-column" />
                  <col className="judgeme-page__email-column" />
                  <col className="judgeme-page__delete-column" />
                </colgroup>
                <thead>
                  <tr>
                    <th>
                      <input
                        type="checkbox"
                        checked={selectedReviewIDs.length > 0 && selectedReviewIDs.length === filteredDraftReviews.length}
                        onChange={toggleSelectAllDrafts}
                        style={{ width: '16px', height: '16px', accentColor: '#059669', borderRadius: '4px' }}
                      />
                    </th>
                    <th>Target Product</th>
                    <th>Reviewer Name</th>
                    <th>Rating</th>
                    <th>Review Title</th>
                    <th>Body Text</th>
                    <th>Email</th>
                    <th aria-label="Actions"></th>
                  </tr>
                </thead>
                <tbody>
                  {filteredDraftReviews.map(r => {
                    const isSelected = selectedReviewIDs.includes(r.id);
                    return (
                      <tr key={r.id} className="hover-row" style={{ background: isSelected ? 'rgba(5, 150, 105, 0.05)' : undefined }}>
                        <td>
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => toggleSelectReview(r.id)}
                            style={{ width: '16px', height: '16px', accentColor: '#059669', borderRadius: '4px' }}
                          />
                        </td>

                        {/* Target Product */}
                        <td>
                          <div style={{ fontWeight: 800, fontSize: '0.85rem', color: 'var(--text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '150px' }}>
                            {r.product_title}
                          </div>
                          <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                            ID: {r.product_id}
                          </div>
                        </td>

                        {/* Reviewer Name */}
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <div style={{
                              width: '30px',
                              height: '30px',
                              borderRadius: '8px',
                              background: r.gender === 'female' ? 'rgba(236, 72, 153, 0.12)' : 'rgba(59, 130, 246, 0.12)',
                              color: r.gender === 'female' ? '#ec4899' : '#3b82f6',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 800,
                              fontSize: '0.72rem',
                              flexShrink: 0
                            }}>
                              {getInitials(r.reviewer_name)}
                            </div>
                            <input
                              type="text"
                              value={r.reviewer_name}
                              onChange={e => handleUpdateReview(r.id, 'reviewer_name', e.target.value)}
                              style={{ width: '100%', height: '34px', fontSize: '0.82rem', fontWeight: 700, borderRadius: '6px', padding: '0 0.55rem', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none' }}
                            />
                          </div>
                        </td>

                        {/* Rating */}
                        <td>
                          <select
                            value={r.rating}
                            onChange={e => handleUpdateReview(r.id, 'rating', parseInt(e.target.value))}
                            style={{ width: '100%', height: '34px', fontSize: '0.82rem', fontWeight: 800, color: '#f59e0b', borderRadius: '6px', padding: '0 0.4rem', border: '1px solid var(--border-color)', background: 'var(--bg-input)', outline: 'none' }}
                          >
                            <option value={5}>5 ★★★★★</option>
                            <option value={4}>4 ★★★★</option>
                            <option value={3}>3 ★★★</option>
                          </select>
                        </td>

                        {/* Title */}
                        <td>
                          <input
                            type="text"
                            value={r.title}
                            onChange={e => handleUpdateReview(r.id, 'title', e.target.value)}
                            style={{ width: '100%', height: '34px', fontSize: '0.82rem', fontWeight: 700, borderRadius: '6px', padding: '0 0.55rem', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none' }}
                          />
                        </td>

                        {/* Body */}
                        <td>
                          <textarea
                            value={r.body}
                            rows={2}
                            onChange={e => handleUpdateReview(r.id, 'body', e.target.value)}
                            style={{ width: '100%', fontSize: '0.8rem', padding: '0.4rem 0.55rem', resize: 'vertical', fontFamily: 'inherit', borderRadius: '6px', lineHeight: 1.35, border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none' }}
                          />
                        </td>

                        {/* Email */}
                        <td>
                          <input
                            type="text"
                            value={r.email}
                            onChange={e => handleUpdateReview(r.id, 'email', e.target.value)}
                            style={{ width: '100%', height: '34px', fontSize: '0.78rem', borderRadius: '6px', padding: '0 0.55rem', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', outline: 'none' }}
                          />
                        </td>

                        {/* Delete */}
                        <td className="judgeme-page__delete-cell">
                          <button
                            onClick={() => handleDeleteReview(r.id)}
                            title="Remove Review"
                            style={{
                              background: 'transparent',
                              color: 'var(--text-tertiary)',
                              border: 'none',
                              borderRadius: '6px',
                              width: '28px',
                              height: '28px',
                              cursor: 'pointer',
                              display: 'inline-flex',
                              alignItems: 'center',
                              justifyContent: 'center'
                            }}
                            onMouseEnter={e => e.currentTarget.style.color = 'var(--status-danger)'}
                            onMouseLeave={e => e.currentTarget.style.color = 'var(--text-tertiary)'}
                          >
                            ✕
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {/* ========================================================================= */}
      {/* SUB-TAB 2: PUBLISHED REVIEWS HISTORY */}
      {/* ========================================================================= */}
      {activeSubTab === 'published' && (
        <>
          {/* History Search Bar */}
          <div className="judgeme-page__toolbar" style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-color)',
            borderRadius: '16px',
            padding: '1rem 1.25rem',
            marginBottom: '1.5rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '1rem'
          }}>
            <div style={{ position: 'relative', width: '320px' }}>
              <input
                type="text"
                placeholder="Filter published reviews history..."
                className="search-input"
                value={searchQuery}
                onChange={e => {
                  setSearchQuery(e.target.value);
                  setPublishedPage(1);
                }}
                style={{ width: '100%', height: '38px', fontSize: '0.85rem', paddingLeft: '2.2rem', borderRadius: '8px' }}
              />
              <svg style={{ position: 'absolute', left: '10px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            </div>

            <div style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
              Total Records Logged: <strong style={{ color: 'var(--text-primary)' }}>{publishedTotal}</strong>
            </div>
          </div>

          {/* Published Table */}
          {isLoadingPublished ? (
            <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--text-tertiary)' }}>
              Loading published history...
            </div>
          ) : publishedReviews.length === 0 ? (
            <div className="judgeme-page__empty" style={{
              background: 'var(--bg-card)',
              border: '1px dashed var(--border-color)',
              textAlign: 'center',
              padding: '5rem 2rem',
              borderRadius: '20px'
            }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.75rem' }}>📜</div>
              <h3 style={{ margin: 0, fontWeight: 800, fontSize: '1.15rem', color: 'var(--text-primary)' }}>
                No Published Review Logs Found
              </h3>
              <p style={{ color: 'var(--text-tertiary)', maxWidth: '440px', margin: '0.5rem auto 0 auto', fontSize: '0.875rem' }}>
                All reviews submitted to Judge.me will automatically be logged and tracked permanently here in PostgreSQL DB.
              </p>
            </div>
          ) : (
            <div className="table-container glass-card-premium judgeme-page__table" style={{ borderRadius: '16px', overflow: 'hidden' }}>
              <table className="premium-table">
                <thead>
                  <tr>
                    <th style={{ paddingLeft: '1.5rem', width: '180px' }}>Product</th>
                    <th style={{ width: '180px' }}>Reviewer Name</th>
                    <th style={{ width: '100px' }}>Rating</th>
                    <th>Review Title</th>
                    <th style={{ width: '200px' }}>Email</th>
                    <th style={{ width: '150px' }}>Published Date</th>
                    <th style={{ paddingRight: '1.5rem', textAlign: 'right', width: '90px' }}>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {publishedReviews.map(r => (
                    <tr
                      key={r.id}
                      className="hover-row"
                      onClick={() => setSelectedPublishedReview(r)}
                      style={{ cursor: 'pointer' }}
                    >
                      <td style={{ paddingLeft: '1.5rem' }}>
                        <div style={{ fontWeight: 800, fontSize: '0.85rem', color: 'var(--text-primary)' }}>
                          {r.product_title}
                        </div>
                        <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                          ID: {r.product_id}
                        </div>
                      </td>
                      <td>
                        <div style={{ fontWeight: 700, fontSize: '0.85rem' }}>{r.reviewer_name}</div>
                      </td>
                      <td>
                        <span className="star-rating-badge">{r.rating} ★</span>
                      </td>
                      <td>
                        <div style={{ fontWeight: 700, fontSize: '0.85rem', color: 'var(--text-primary)' }}>{r.title}</div>
                      </td>
                      <td style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{r.email}</td>
                      <td style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>
                        {new Date(r.published_at).toLocaleDateString('en-IN', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
                      </td>
                      <td style={{ paddingRight: '1.5rem', textAlign: 'right' }}>
                        <span className="badge-pill" style={{
                          background: r.status === 'SUCCESS' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(245, 158, 11, 0.15)',
                          color: r.status === 'SUCCESS' ? 'var(--status-active)' : 'var(--status-warning)',
                          fontWeight: 800,
                          fontSize: '0.75rem',
                          padding: '0.2rem 0.5rem'
                        }}>
                          {r.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {/* Pagination Controls */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem 1.5rem', borderTop: '1px solid var(--border-color)' }}>
                <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                  Loaded {publishedReviews.length} review{publishedReviews.length === 1 ? '' : 's'} · Page <strong>{publishedPage}</strong> of {Math.ceil(publishedTotal / PUBLISHED_REVIEWS_PAGE_SIZE) || 1} ({publishedTotal} total)
                </span>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setPublishedPage(prev => Math.max(1, prev - 1))}
                    disabled={publishedPage === 1}
                    style={{ height: '34px', padding: '0 1rem', fontSize: '0.8rem', borderRadius: '8px' }}
                  >
                    Previous
                  </button>
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setPublishedPage(prev => prev + 1)}
                    disabled={publishedPage * PUBLISHED_REVIEWS_PAGE_SIZE >= publishedTotal}
                    style={{ height: '34px', padding: '0 1rem', fontSize: '0.8rem', borderRadius: '8px' }}
                  >
                    Next
                  </button>
                </div>
              </div>
            </div>
          )}
        </>
      )}

      {/* PRO GENERATION CONFIGURATION MODAL (MAX 10 TOTAL BATCH) */}
      {isGenerateModalOpen && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.5)' }} onClick={() => setIsGenerateModalOpen(false)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{
            maxWidth: '740px',
            width: '92%',
            maxHeight: '88vh',
            overflowY: 'auto',
            borderRadius: '24px',
            padding: '2.25rem'
          }}>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.75rem' }}>
              <div>
                <h3 style={{ margin: 0, fontSize: '1.35rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.01em' }}>
                  Generate AI Review Drafts (Max 10 Batch)
                </h3>
                <p style={{ margin: '6px 0 0 0', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                  Pick products and review count. Maximum 10 reviews generated per run.
                </p>
              </div>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => setIsGenerateModalOpen(false)}
                style={{ width: '36px', height: '36px', padding: 0, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '1rem' }}
              >
                ✕
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>

              {/* Target Products Selector Grid */}
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    Target Products ({selectedProductIDs.length === 0 ? `All ${MAX_SELECTED_PRODUCTS} Catalog Items` : `${selectedProductIDs.length} Selected`})
                  </label>
                  {selectedProductIDs.length > 0 && (
                    <button
                      type="button"
                      onClick={() => {
                        setSelectedProductIDs([]);
                        setSelectionWarning(null);
                      }}
                      style={{ background: 'none', border: 'none', color: '#059669', fontSize: '0.8rem', fontWeight: 700, cursor: 'pointer' }}
                    >
                      Clear Selection
                    </button>
                  )}
                </div>
                {selectionWarning && (
                  <div className="judgeme-page__selection-warning" role="alert">
                    <svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.25">
                      <path d="M12 9v4" />
                      <path d="M12 17h.01" />
                      <path d="M10.3 3.7 2.9 16.4A2 2 0 0 0 4.6 19h14.8a2 2 0 0 0 1.7-2.6L13.7 3.7a2 2 0 0 0-3.4 0Z" />
                    </svg>
                    <span>{selectionWarning}</span>
                  </div>
                )}

                {/* SKU / Index Range Selection Inputs (Sleek Toolbar) */}
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: '0.85rem',
                  marginBottom: '1rem',
                  background: 'rgba(5, 150, 105, 0.04)',
                  padding: '0.85rem 1.15rem',
                  borderRadius: '14px',
                  border: '1px solid rgba(5, 150, 105, 0.18)',
                  flexWrap: 'wrap'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flex: 1, minWidth: '280px' }}>
                    <span style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-primary)', whiteSpace: 'nowrap' }}>
                      Range Selection:
                    </span>
                    <input
                      type="text"
                      placeholder="From (e.g. mi-001)"
                      value={rangeFrom}
                      onChange={e => setRangeFrom(e.target.value)}
                      onKeyDown={e => e.key === 'Enter' && handleApplySkuRange(rangeFrom, rangeTo)}
                      style={{
                        padding: '0.45rem 0.75rem',
                        fontSize: '0.85rem',
                        borderRadius: '8px',
                        border: '1px solid var(--border-color)',
                        background: 'var(--bg-card)',
                        color: 'var(--text-primary)',
                        width: '135px',
                        outline: 'none'
                      }}
                    />
                    <span style={{ fontSize: '0.85rem', color: 'var(--text-tertiary)', fontWeight: 600 }}>to</span>
                    <input
                      type="text"
                      placeholder="To (e.g. mi-010)"
                      value={rangeTo}
                      onChange={e => setRangeTo(e.target.value)}
                      onKeyDown={e => e.key === 'Enter' && handleApplySkuRange(rangeFrom, rangeTo)}
                      style={{
                        padding: '0.45rem 0.75rem',
                        fontSize: '0.85rem',
                        borderRadius: '8px',
                        border: '1px solid var(--border-color)',
                        background: 'var(--bg-card)',
                        color: 'var(--text-primary)',
                        width: '135px',
                        outline: 'none'
                      }}
                    />
                  </div>
                  <button
                    type="button"
                    onClick={() => handleApplySkuRange(rangeFrom, rangeTo)}
                    style={{
                      padding: '0.5rem 1.25rem',
                      fontSize: '0.825rem',
                      fontWeight: 700,
                      borderRadius: '8px',
                      background: 'linear-gradient(135deg, #059669 0%, #047857 100%)',
                      color: 'white',
                      border: 'none',
                      boxShadow: '0 2px 8px rgba(5, 150, 105, 0.25)',
                      cursor: 'pointer',
                      whiteSpace: 'nowrap',
                      flexShrink: 0
                    }}
                  >
                    Select Range
                  </button>
                </div>

                <div className="product-picker-grid">
                  {products.map((p, index) => {
                    const idStr = p.id.toString();
                    const isSelected = selectedProductIDs.includes(idStr);
                    const skuDisplay = p.mi_sku || `mi-${(index + 1).toString().padStart(3, '0')}`;
                    return (
                      <div
                        key={p.id}
                        className={`product-picker-item ${isSelected ? 'selected' : ''}`}
                        onClick={() => toggleSelectProductInModal(idStr)}
                        title={`${skuDisplay} - ${p.title}`}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.45rem', overflow: 'hidden' }}>
                          <span style={{
                            fontSize: '0.7rem',
                            fontWeight: 800,
                            color: isSelected ? '#059669' : 'var(--text-tertiary)',
                            background: isSelected ? 'rgba(5, 150, 105, 0.15)' : 'rgba(0,0,0,0.06)',
                            padding: '0.12rem 0.4rem',
                            borderRadius: '5px',
                            flexShrink: 0
                          }}>
                            {skuDisplay}
                          </span>
                          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {p.title}
                          </span>
                        </div>
                        {isSelected && <span style={{ marginLeft: '0.5rem', fontWeight: 800 }}>✓</span>}
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Reviews per Product */}
              <div style={{ marginTop: '0.25rem' }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '0.65rem' }}>
                  Reviews per selected product
                </label>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                  {[1, 2, 3, 5].map(num => (
                    <button
                      key={num}
                      type="button"
                      onClick={() => setReviewsPerProduct(num)}
                      style={{
                        flex: 1,
                        height: '44px',
                        borderRadius: '10px',
                        fontWeight: 800,
                        fontSize: '0.875rem',
                        border: reviewsPerProduct === num ? '2px solid #059669' : '1px solid var(--border-color)',
                        background: reviewsPerProduct === num ? 'rgba(5, 150, 105, 0.1)' : 'var(--bg-input)',
                        color: reviewsPerProduct === num ? '#059669' : 'var(--text-primary)',
                        cursor: 'pointer',
                        transition: 'all 0.15s ease'
                      }}
                    >
                      {num} {num === 1 ? 'Review' : 'Reviews'}
                    </button>
                  ))}
                </div>
              </div>

              {/* Batch Limit Summary Counter */}
              <div style={{
                background: calculatedTotalBatchSize >= 10 ? 'rgba(245, 158, 11, 0.08)' : 'rgba(5, 150, 105, 0.08)',
                border: calculatedTotalBatchSize >= 10 ? '1px solid rgba(245, 158, 11, 0.25)' : '1px solid rgba(5, 150, 105, 0.25)',
                borderRadius: '14px',
                padding: '1rem 1.25rem',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                margin: '0.25rem 0'
              }}>
                <span style={{ fontSize: '0.875rem', fontWeight: 700, color: 'var(--text-primary)' }}>
                  Total Batch Review Count:
                </span>
                <span style={{ fontSize: '1.15rem', fontWeight: 800, color: calculatedTotalBatchSize >= 10 ? '#f59e0b' : '#059669' }}>
                  {calculatedTotalBatchSize} / 10 Max
                </span>
              </div>

              {/* Shop & Email */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>
                <div>
                  <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', display: 'block', marginBottom: '0.45rem' }}>
                    Shop Domain
                  </label>
                  <input
                    type="text"
                    value={shopDomain}
                    onChange={e => setShopDomain(e.target.value)}
                    style={{
                      width: '100%',
                      height: '42px',
                      borderRadius: '10px',
                      fontSize: '0.875rem',
                      padding: '0 0.85rem',
                      border: '1px solid var(--border-color)',
                      background: 'var(--bg-input)',
                      color: 'var(--text-primary)',
                      outline: 'none'
                    }}
                  />
                </div>
                <div>
                  <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', display: 'block', marginBottom: '0.45rem' }}>
                    Base Recipient Email
                  </label>
                  <input
                    type="email"
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    style={{
                      width: '100%',
                      height: '42px',
                      borderRadius: '10px',
                      fontSize: '0.875rem',
                      padding: '0 0.85rem',
                      border: '1px solid var(--border-color)',
                      background: 'var(--bg-input)',
                      color: 'var(--text-primary)',
                      outline: 'none'
                    }}
                  />
                </div>
              </div>

              <label style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-primary)', marginTop: '0.25rem' }}>
                <input
                  type="checkbox"
                  checked={aliasEmail}
                  onChange={e => setAliasEmail(e.target.checked)}
                  style={{ width: '18px', height: '18px', accentColor: '#059669' }}
                />
                Append Indian reviewer name alias (e.g. email+hemanth23@gmail.com)
              </label>

            </div>

            {/* Modal Footer Actions */}
            <div style={{ marginTop: '2.25rem', paddingTop: '1.25rem', borderTop: '1px solid var(--border-color)', display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => setIsGenerateModalOpen(false)}
                style={{ height: '44px', padding: '0 1.25rem', borderRadius: '10px', fontWeight: 700 }}
              >
                Cancel
              </button>

              <button
                type="button"
                className="btn-primary"
                onClick={handleGenerateReviews}
                disabled={isGenerating}
                style={{
                  height: '44px',
                  padding: '0 1.5rem',
                  borderRadius: '10px',
                  fontWeight: 800,
                  background: 'linear-gradient(135deg, #059669 0%, #047857 100%)',
                  boxShadow: '0 4px 14px rgba(5, 150, 105, 0.3)',
                  border: 'none',
                  color: 'white',
                  cursor: isGenerating ? 'not-allowed' : 'pointer'
                }}
              >
                {isGenerating ? 'Generating Batch...' : `Generate ${calculatedTotalBatchSize} Reviews`}
              </button>
            </div>

          </div>
        </div>
      )}

      {/* PUBLISHED REVIEW DETAIL MODAL DRAWER */}
      {selectedPublishedReview && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.5)' }} onClick={() => setSelectedPublishedReview(null)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px', width: '92%', borderRadius: '20px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
              <div>
                <span className="badge-pill" style={{ background: 'rgba(16, 185, 129, 0.15)', color: '#10b981', fontWeight: 800, marginBottom: '0.5rem', display: 'inline-block' }}>
                  PUBLISHED TO JUDGE.ME
                </span>
                <h3 style={{ margin: '0.25rem 0 0 0', fontSize: '1.3rem', fontWeight: 800 }}>
                  {selectedPublishedReview.product_title}
                </h3>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                  Product ID: {selectedPublishedReview.product_id}
                </div>
              </div>
              <button className="btn-secondary" onClick={() => setSelectedPublishedReview(null)} style={{ padding: '0.4rem 0.8rem', borderRadius: '8px' }}>✕</button>
            </div>

            <div style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)', borderRadius: '14px', padding: '1.25rem', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <div style={{ fontWeight: 800, fontSize: '0.95rem', color: 'var(--text-primary)' }}>
                  {selectedPublishedReview.reviewer_name}
                </div>
                <span className="star-rating-badge">{selectedPublishedReview.rating} ★★★★★</span>
              </div>

              <div style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
                "{selectedPublishedReview.title}"
              </div>

              <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.5, margin: 0 }}>
                {selectedPublishedReview.body}
              </p>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
              <div>
                <strong style={{ color: 'var(--text-primary)' }}>Reviewer Email:</strong> {selectedPublishedReview.email}
              </div>
              <div>
                <strong style={{ color: 'var(--text-primary)' }}>Shop Domain:</strong> {selectedPublishedReview.shop_domain}
              </div>
              <div>
                <strong style={{ color: 'var(--text-primary)' }}>Published At:</strong> {new Date(selectedPublishedReview.published_at).toLocaleString()}
              </div>
              <div>
                <strong style={{ color: 'var(--text-primary)' }}>Status Code:</strong> HTTP {selectedPublishedReview.status_code}
              </div>
            </div>

            <div style={{ marginTop: '1.75rem', textAlign: 'right' }}>
              <button className="btn-primary" onClick={() => setSelectedPublishedReview(null)} style={{ padding: '0.5rem 1.25rem', borderRadius: '8px', fontWeight: 800 }}>Close</button>
            </div>
          </div>
        </div>
      )}

      {/* Execution Results Summary Modal */}
      {resultsReport && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.5)' }} onClick={() => setResultsReport(null)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '850px', width: '95%', borderRadius: '20px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <div>
                <h3 style={{ margin: 0, fontSize: '1.3rem', fontWeight: 800 }}>Submission Execution Report</h3>
                <p style={{ margin: '2px 0 0 0', fontSize: '0.85rem', color: 'var(--text-tertiary)' }}>Detailed report of HTTP posts sent to Judge.me API.</p>
              </div>
              <button className="btn-secondary" onClick={() => setResultsReport(null)} style={{ padding: '0.4rem 0.85rem', borderRadius: '8px' }}>Close</button>
            </div>

            <div style={{ maxHeight: '400px', overflowY: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
              <table className="premium-table" style={{ width: '100%' }}>
                <thead>
                  <tr>
                    <th style={{ paddingLeft: '1rem', width: '40px' }}>#</th>
                    <th>Product</th>
                    <th>Reviewer</th>
                    <th>Status</th>
                    <th>Response Body</th>
                  </tr>
                </thead>
                <tbody>
                  {resultsReport.map((res, i) => (
                    <tr key={i}>
                      <td style={{ paddingLeft: '1rem' }}>{res.index}</td>
                      <td style={{ fontWeight: 700, fontSize: '0.85rem' }}>{res.product_title}</td>
                      <td style={{ fontSize: '0.85rem' }}>{res.reviewer_name}</td>
                      <td>
                        <span className="badge-pill" style={{
                          background: res.status === 'SUCCESS' ? 'rgba(16, 185, 129, 0.15)' : res.status === 'DRY_RUN' ? 'rgba(245, 158, 11, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                          color: res.status === 'SUCCESS' ? 'var(--status-active)' : res.status === 'DRY_RUN' ? 'var(--status-warning)' : 'var(--status-danger)',
                          fontWeight: 800,
                          fontSize: '0.75rem',
                          padding: '0.25rem 0.65rem'
                        }}>
                          {res.status} ({res.status_code})
                        </span>
                      </td>
                      <td style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', maxWidth: '280px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {res.response_body}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div style={{ marginTop: '1.5rem', textAlign: 'right' }}>
              <button className="btn-primary" onClick={() => setResultsReport(null)} style={{ padding: '0.5rem 1.25rem', borderRadius: '8px', fontWeight: 800 }}>Done</button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
};
