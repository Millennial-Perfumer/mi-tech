import React, { useEffect, useState, useMemo } from 'react';
import { useMarketingStore } from '../store/useMarketingStore';
import { Target, Layers, PlayCircle, ArrowLeft, TrendingUp, MousePointer2, Eye } from 'lucide-react';

type ViewLevel = 'campaign' | 'adset' | 'ad';

export const Marketing: React.FC = () => {
  const { campaigns, adsets, ads, summary: accountSummary, isLoading, fetchOverview, fetchAdSets, fetchAds } = useMarketingStore();
  const [level, setLevel] = useState<ViewLevel>('campaign');

  useEffect(() => {
    fetchOverview();
  }, []);

  const handleCampaignClick = (campaign: any) => {
    fetchAdSets(campaign.id);
    setLevel('adset');
  };

  const handleAdsetClick = (adset: any) => {
    fetchAds(adset.id);
    setLevel('ad');
  };

  const goBack = () => {
    if (level === 'ad') setLevel('adset');
    else if (level === 'adset') setLevel('campaign');
  };

  // Calculate cumulative metrics based on current level
  const summary = useMemo(() => {
    // If at account/campaign level and we have a ground truth summary, use it
    if (level === 'campaign' && accountSummary && accountSummary.length > 0) {
      const s = accountSummary[0];
      return {
        spend: parseFloat(s.spend || '0'),
        impressions: parseInt(s.impressions || '0'),
        clicks: parseInt(s.inline_link_clicks || '0'),
      };
    }

    const data = level === 'campaign' ? campaigns : level === 'adset' ? adsets : ads;
    
    return data.reduce((acc, item) => ({
      spend: acc.spend + (parseFloat(item.spend as any) || 0),
      impressions: acc.impressions + (parseInt(item.impressions as any) || 0),
      clicks: acc.clicks + (parseInt((item as any).inline_link_clicks || item.clicks || '0')),
    }), { spend: 0, impressions: 0, clicks: 0 });
  }, [level, campaigns, adsets, ads, accountSummary]);

  const avgCtr = summary.impressions > 0 ? (summary.clicks / summary.impressions) * 100 : 0;

  const InsightCard = ({ insight, onClick, icon: Icon }: any) => (
    <div className="glass-card" onClick={onClick} style={{ padding: '1rem', marginBottom: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.75rem' }}>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <div style={{ color: 'var(--accent-color)', background: 'var(--accent-subtle)', padding: '8px', borderRadius: '10px' }}>
            <Icon size={18} />
          </div>
          <div>
            <h3 style={{ fontSize: '0.95rem', color: '#fff', maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{insight.name}</h3>
            <span style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: insight.status === 'ACTIVE' ? 'var(--accent-color)' : 'var(--text-tertiary)' }}>
              {insight.status}
            </span>
          </div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <p style={{ fontWeight: 700, color: '#fff' }}>₹{insight.spend.toFixed(2)}</p>
          <p style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)' }}>Spent</p>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.5rem', borderTop: '1px solid var(--glass-border)', paddingTop: '0.75rem' }}>
        <div style={{ textAlign: 'center' }}>
          <p style={{ fontSize: '0.85rem', color: '#fff', fontWeight: 600 }}>{insight.impressions.toLocaleString()}</p>
          <p style={{ fontSize: '0.65rem', color: 'var(--text-tertiary)' }}>Impr.</p>
        </div>
        <div style={{ textAlign: 'center' }}>
          <p style={{ fontSize: '0.85rem', color: '#fff', fontWeight: 600 }}>{insight.clicks.toLocaleString()}</p>
          <p style={{ fontSize: '0.65rem', color: 'var(--text-tertiary)' }}>Clicks</p>
        </div>
        <div style={{ textAlign: 'center' }}>
          <p style={{ fontSize: '0.85rem', color: '#fff', fontWeight: 600 }}>{insight.ctr.toFixed(2)}%</p>
          <p style={{ fontSize: '0.65rem', color: 'var(--text-tertiary)' }}>CTR</p>
        </div>
      </div>
    </div>
  );

  return (
    <div>
      <header style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          {level !== 'campaign' && (
            <button onClick={goBack} className="glass-panel" style={{ width: '40px', height: '40px', borderRadius: '10px', background: 'transparent' }}>
              <ArrowLeft size={18} />
            </button>
          )}
          <div>
            <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>
              Marketing Console
            </p>
            <h1>{level === 'campaign' ? 'Meta Intelligence' : level === 'adset' ? 'Ad Sets' : 'Ads'}</h1>
          </div>
        </div>
      </header>

      {/* Cumulative Summary Section */}
      <div className="glass-card" style={{ marginBottom: '2rem', padding: '1.25rem', background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.05), rgba(6, 182, 212, 0.05))' }}>
        <p style={{ fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '0.1em', color: 'var(--text-tertiary)', marginBottom: '1rem' }}>
          {level === 'campaign' ? 'Overall Account Performance' : `Scope: ${level} Level Summary`}
        </p>
        
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--accent-color)', marginBottom: '0.25rem' }}>
              <TrendingUp size={14} />
              <span style={{ fontSize: '0.7rem', fontWeight: 600 }}>TOTAL SPEND</span>
            </div>
            <p style={{ fontSize: '1.25rem', fontWeight: 800 }}>₹{summary.spend.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--secondary-color)', marginBottom: '0.25rem' }}>
              <MousePointer2 size={14} />
              <span style={{ fontSize: '0.7rem', fontWeight: 600 }}>AVG. CTR</span>
            </div>
            <p style={{ fontSize: '1.25rem', fontWeight: 800 }}>{avgCtr.toFixed(2)}%</p>
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
              <Eye size={14} />
              <span style={{ fontSize: '0.7rem', fontWeight: 600 }}>IMPRESSIONS</span>
            </div>
            <p style={{ fontSize: '1.1rem', fontWeight: 700 }}>{summary.impressions.toLocaleString()}</p>
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
              <MousePointer2 size={14} />
              <span style={{ fontSize: '0.7rem', fontWeight: 600 }}>CLICKS</span>
            </div>
            <p style={{ fontSize: '1.1rem', fontWeight: 700 }}>{summary.clicks.toLocaleString()}</p>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem', marginTop: '3rem', opacity: 0.5 }}>
          <div className="animate-spin" style={{ width: '32px', height: '32px', border: '3px solid var(--accent-subtle)', borderTopColor: 'var(--accent-color)', borderRadius: '50%' }}></div>
          <p style={{ fontSize: '0.85rem' }}>Analyzing Meta data...</p>
        </div>
      ) : (
        <div className="drill-down-content">
          <div style={{ marginBottom: '1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
             <h2 style={{ fontSize: '1.1rem' }}>
               {level === 'campaign' ? 'Active Campaigns' : level === 'adset' ? 'Campaign Ad Sets' : 'Ad Set Creatives'}
             </h2>
             <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
               {level === 'campaign' ? campaigns.length : level === 'adset' ? adsets.length : ads.length} Items
             </span>
          </div>
          {level === 'campaign' && campaigns.map(c => <InsightCard key={c.id} insight={c} icon={Target} onClick={() => handleCampaignClick(c)} />)}
          {level === 'adset' && adsets.map(s => <InsightCard key={s.id} insight={s} icon={Layers} onClick={() => handleAdsetClick(s)} />)}
          {level === 'ad' && ads.map(a => <InsightCard key={a.id} insight={a} icon={PlayCircle} />)}
        </div>
      )}
    </div>
  );
};
