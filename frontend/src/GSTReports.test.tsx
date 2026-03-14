import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GSTReports } from './GSTReports';

describe('GSTReports Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    const fetchWithAuth = vi.fn().mockImplementation(() => new Promise(() => {}));

    render(<GSTReports startDate="2026-01-01" endDate="2026-12-31" fetchWithAuth={fetchWithAuth} />);

    expect(screen.getByText('Loading Reports...')).toBeInTheDocument();
  });

  it('renders dashboard sub-tab and metrics correctly', async () => {
    const mockSummaryResponse = {
      success: true,
      summary: {
        total_revenue: 15000,
        total_taxable_value: 12000,
        total_gst_collected: 3000,
        total_igst: 1000,
        total_cgst: 1000,
        total_sgst: 1000,
        total_orders: 10,
        cancelled_orders: 1,
        fulfilled_orders: 9,
        unfulfilled_orders: 0,
        paid_orders: 10,
        invoices_generated: 10
      }
    };

    const fetchWithAuth = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/reports/summary')) {
        return Promise.resolve({ json: () => Promise.resolve(mockSummaryResponse) });
      }
      return Promise.resolve({ json: () => Promise.resolve({ success: true, data: [] }) });
    });

    render(<GSTReports startDate="2026-01-01" endDate="2026-12-31" fetchWithAuth={fetchWithAuth} />);

    await waitFor(() => {
      expect(screen.queryByText('Loading Reports...')).not.toBeInTheDocument();
    });

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('B2C State-wise')).toBeInTheDocument();

    expect(screen.getByText('₹15,000')).toBeInTheDocument();
    expect(screen.getByText('₹12,000')).toBeInTheDocument();
  });
});
