import { test, expect } from '@playwright/test';

test('login page has title and login form', async ({ page }) => {
  await page.goto('/');

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/mi-tech/);

  // Expect the login form to be visible.
  await expect(page.locator('form')).toBeVisible();

  // Expect the username and password inputs to be visible.
  await expect(page.getByPlaceholder('admin')).toBeVisible();
  await expect(page.getByPlaceholder('••••••••')).toBeVisible();

  // Expect the submit button to be visible.
  await expect(page.locator('button[type="submit"]')).toBeVisible();
});

test('login form shows error on invalid credentials', async ({ page }) => {
  await page.goto('/');

  // Fill in the username and password inputs.
  await page.getByPlaceholder('admin').fill('wronguser');
  await page.getByPlaceholder('••••••••').fill('wrongpass');

  // Click the submit button.
  await page.locator('button[type="submit"]').click();

  // Expect an error message to be visible.
  // The error message is rendered conditionally inside the form after the inputs
  await expect(page.locator('form > div').nth(2)).toBeVisible(); // This selects the error div
});