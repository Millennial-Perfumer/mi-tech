import { test, expect } from '@playwright/test';

// Define a helper function to login before tests
async function login(page: any) {
  await page.goto('/');
  await page.getByPlaceholder('admin').fill('admin'); // Assuming admin exists in test DB
  await page.getByPlaceholder('••••••••').fill('password123'); // Assuming password
  await page.locator('button[type="submit"]').click();
}

test('dashboard loads successfully after login', async ({ page }) => {
  // Wait for login to complete and dashboard to appear
  // Mock the login response to avoid needing a real backend running
  await page.route('http://localhost:8080/api/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ token: 'mock_token_123' }),
    });
  });

  // Mock the settings/date-range request
  await page.route('http://localhost:8080/api/settings/date-range', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ success: true, start_date: '2026-01-01', end_date: '2026-12-31' }),
    });
  });

  // Mock dashboard metrics
  await page.route('http://localhost:8080/api/dashboard/metrics*', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        metrics: {
          total_revenue: 15000,
          total_invoices: 10,
          total_gst_collected: 2000,
          cgst_collected: 1000,
          sgst_collected: 1000,
          igst_collected: 0,
          total_orders: 12,
          cancelled_orders: 1,
          fulfilled_orders: 8,
          unfulfilled_orders: 3
        }
      }),
    });
  });

  // Mock orders
  await page.route('http://localhost:8080/api/orders*', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        orders: [],
        total_count: 0
      }),
    });
  });

  // Mock webhook status
  await page.route('http://localhost:8080/api/webhook/status', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        topic: 'orders/create',
        status: 'active',
        last_received: new Date().toISOString()
      }),
    });
  });


  await login(page);

  // Assert we are on the dashboard view
  await expect(page.locator('.page-title')).toContainText('Overview');
  await expect(page.locator('.dashboard-grid')).toBeVisible();

  // Assert some mocked data is visible
  await expect(page.locator('.dashboard-grid .card:has-text("Total Revenue") .card-value')).toContainText('₹15,000');
});

test('can navigate to GST Reports tab', async ({ page }) => {
  await page.route('http://localhost:8080/api/auth/login', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ token: 'mock_token_123' }) });
  });
  await page.route('http://localhost:8080/api/settings/date-range', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, start_date: '2026-01-01', end_date: '2026-12-31' }) });
  });
  await page.route('http://localhost:8080/api/dashboard/metrics*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, metrics: {} }) });
  });
  await page.route('http://localhost:8080/api/orders*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, orders: [], total_count: 0 }) });
  });
  await page.route('http://localhost:8080/api/webhook/status', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) });
  });
  // Mock GST Reports APIs
  await page.route('http://localhost:8080/api/reports/summary*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, summary: { total_revenue: 0, total_taxable_value: 0, total_gst_collected: 0, total_igst: 0, total_cgst: 0, total_sgst: 0 } }) });
  });
  await page.route('http://localhost:8080/api/reports/state-wise*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, data: [] }) });
  });
  await page.route('http://localhost:8080/api/reports/hsn-wise*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, data: [] }) });
  });
  await page.route('http://localhost:8080/api/reports/documents-issued*', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, data: [] }) });
  });


  await login(page);

  // Click on GST Reports tab
  await page.getByText('GST Reports').click();

  // Assert we are on the GST Reports view
  await expect(page.locator('.page-title')).toContainText('GST Reports');
  await expect(page.locator('.gst-reports-container')).toBeVisible();
});
