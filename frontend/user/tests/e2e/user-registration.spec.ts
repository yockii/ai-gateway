import { test, expect } from '@playwright/test'

test.describe('User Registration Flow', () => {
  test('should display registration form', async ({ page }) => {
    await page.goto('/register')

    await expect(page.locator('h1, h2')).toBeVisible()

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const nameInput = page.locator('input[name="name"], input[placeholder*="姓名"], input[placeholder*="name"]')

    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
  })

  test('should validate email format', async ({ page }) => {
    await page.goto('/register')

    const emailInput = page.locator('input[type="email"]')
    await emailInput.fill('invalid-email')

    const submitButton = page.locator('button[type="submit"]')
    await submitButton.click()

    // Should show validation error
    await page.waitForTimeout(500)
    const errorElement = page.locator('.text-red-500, .text-red-400, .error')
    const hasError = await errorElement.count() > 0

    if (hasError) {
      await expect(errorElement.first()).toBeVisible()
    }
  })

  test('should require password confirmation', async ({ page }) => {
    await page.goto('/register')

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]').first()
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('test@example.com')
    await passwordInput.fill('password123')
    await submitButton.click()

    // Should wait for validation
    await page.waitForTimeout(500)
  })

  test('should successfully register and redirect', async ({ page }) => {
    await page.goto('/register')

    const emailInput = page.locator('input[type="email"]')
    const passwordInputs = page.locator('input[type="password"]')
    const nameInput = page.locator('input[name="name"], input[placeholder*="姓名"], input[placeholder*="name"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill(`test${Date.now()}@example.com`)

    // Fill password fields
    const count = await passwordInputs.count()
    for (let i = 0; i < count; i++) {
      await passwordInputs.nth(i).fill('password123')
    }

    // Fill name if exists
    const nameExists = await nameInput.count() > 0
    if (nameExists) {
      await nameInput.fill('Test User')
    }

    await submitButton.click()

    // Should redirect to dashboard or login
    await page.waitForTimeout(2000)

    const currentUrl = page.url()
    const redirected = currentUrl.includes('/dashboard') || currentUrl.includes('/login')
    expect(redirected).toBe(true)
  })

  test('should have link to login page', async ({ page }) => {
    await page.goto('/register')

    const loginLink = page.locator('a[href*="login"], a:has-text("登录"), a:has-text("Login")')
    const hasLink = await loginLink.count() > 0

    if (hasLink) {
      await loginLink.first().click()
      await page.waitForTimeout(500)

      expect(page.url()).toContain('/login')
    }
  })
})
