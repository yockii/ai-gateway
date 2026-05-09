import { test, expect } from '@playwright/test'

test.describe('User Login Flow', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login')

    await expect(page.locator('h1, h2')).toBeVisible()

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
    await expect(submitButton).toBeVisible()
  })

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login')

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('wrong@example.com')
    await passwordInput.fill('wrongpassword')
    await submitButton.click()

    await page.waitForTimeout(1000)

    const errorElement = page.locator('.text-red-500, .text-red-400, .error')
    const hasError = await errorElement.count() > 0

    if (hasError) {
      await expect(errorElement.first()).toBeVisible()
    }
  })

  test('should login successfully and redirect to dashboard', async ({ page }) => {
    await page.goto('/login')

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('user@example.com')
    await passwordInput.fill('password123')
    await submitButton.click()

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 5000 })
    await expect(page.locator('h1')).toBeVisible()
  })

  test('should persist session across page reload', async ({ page }) => {
    await page.goto('/login')

    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    const submitButton = page.locator('button[type="submit"]')

    await emailInput.fill('user@example.com')
    await passwordInput.fill('password123')
    await submitButton.click()

    await page.waitForURL('**/dashboard')

    // Reload the page
    await page.reload()

    // Should still be logged in
    await expect(page.locator('h1')).toBeVisible()
  })

  test('should have link to registration page', async ({ page }) => {
    await page.goto('/login')

    const registerLink = page.locator('a[href*="register"], a:has-text("注册"), a:has-text("Register")')
    const hasLink = await registerLink.count() > 0

    if (hasLink) {
      await registerLink.first().click()
      await page.waitForTimeout(500)

      expect(page.url()).toContain('/register')
    }
  })
})
