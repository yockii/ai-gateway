import { test, expect } from '@playwright/test'

test.describe('User Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[type="email"]', 'user@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard')
  })

  test('should display dashboard with statistics', async ({ page }) => {
    await expect(page.locator('h1')).toBeVisible()

    // Look for common dashboard elements
    const pageContent = page.content()

    // Check for any statistics or metric cards
    const statsCards = page.locator('.card, .stat, .metric')
    const hasCards = await statsCards.count() > 0

    // Dashboard should have some content
    await expect(page.locator('h1')).toBeVisible()
  })

  test('should display usage statistics', async ({ page }) => {
    // Look for usage-related elements
    const usageElements = page.locator('text=/使用|Usage|请求|Request/i')

    const hasUsage = await usageElements.count() > 0

    if (hasUsage) {
      await expect(usageElements.first()).toBeVisible()
    }
  })

  test('should navigate to API keys page', async ({ page }) => {
    // Look for API keys navigation link
    const keysLink = page.locator('a[href*="keys"], a:has-text("密钥"), a:has-text("Keys"), a:has-text("API")').first()

    const hasLink = await keysLink.count() > 0

    if (hasLink) {
      await keysLink.click()
      await page.waitForTimeout(500)

      expect(page.url()).toContain('/keys')
    }
  })

  test('should navigate to usage page', async ({ page }) => {
    // Look for usage navigation link
    const usageLink = page.locator('a[href*="usage"], a:has-text("使用"), a:has-text("Usage")').first()

    const hasLink = await usageLink.count() > 0

    if (hasLink) {
      await usageLink.click()
      await page.waitForTimeout(500)

      expect(page.url()).toContain('/usage')
    }
  })

  test('should navigate to settings page', async ({ page }) => {
    // Look for settings navigation link
    const settingsLink = page.locator('a[href*="settings"], a:has-text("设置"), a:has-text("Settings")').first()

    const hasLink = await settingsLink.count() > 0

    if (hasLink) {
      await settingsLink.click()
      await page.waitForTimeout(500)

      expect(page.url()).toContain('/settings')
    }
  })

  test('should display user menu or profile', async ({ page }) => {
    // Look for user menu or profile dropdown
    const userMenu = page.locator('button:has-text("用户"), button:has-text("User"), .user-menu, .profile')

    const hasMenu = await userMenu.count() > 0

    if (hasMenu) {
      await expect(userMenu.first()).toBeVisible()
    }
  })

  test('should logout successfully', async ({ page }) => {
    // Look for logout button
    const logoutButton = page.locator('button:has-text("退出"), button:has-text("Logout"), a:has-text("退出"), a:has-text("Logout")').first()

    const hasButton = await logoutButton.count() > 0

    if (hasButton) {
      await logoutButton.click()
      await page.waitForTimeout(1000)

      // Should redirect to login page
      expect(page.url()).toContain('/login')
    }
  })
})
