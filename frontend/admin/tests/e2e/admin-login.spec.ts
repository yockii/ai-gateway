import { test, expect } from '@playwright/test'

test.describe('Admin Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('should display login form', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('AI Gateway')
    await expect(page.locator('p')).toContainText('管理后台登录')

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
    await expect(submitButton).toBeVisible()
  })

  test('should show validation error for empty fields', async ({ page }) => {
    const submitButton = page.locator('button[type="submit"]')

    // Try to submit with empty fields
    await submitButton.click()

    // Button should be disabled or show error
    await expect(page.locator('.text-red-400')).toBeVisible()
  })

  test('should show error for invalid credentials', async ({ page }) => {
    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('wrong@example.com')
    await passwordInput.fill('wrongpassword')
    await submitButton.click()

    // Should show error message
    await expect(page.locator('.text-red-400')).toBeVisible()
  })

  test('should redirect to dashboard after successful login', async ({ page }) => {
    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('admin@example.com')
    await passwordInput.fill('password123')
    await submitButton.click()

    // Should navigate to users page (default after login)
    await page.waitForURL('**/admin/users')
    await expect(page.locator('h1')).toContainText('用户管理')
  })

  test('should persist session across page reload', async ({ page }) => {
    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('admin@example.com')
    await passwordInput.fill('password123')
    await submitButton.click()

    await page.waitForURL('**/admin/users')

    // Reload the page
    await page.reload()

    // Should still be logged in
    await expect(page.locator('h1')).toContainText('用户管理')
  })
})
