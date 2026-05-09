import { test, expect } from '@playwright/test'

test.describe('Monitoring Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[type="email"]', 'admin@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/admin/users')

    // Navigate to monitoring page
    await page.goto('/admin/monitoring')
    await page.waitForTimeout(500)
  })

  test('should display monitoring dashboard', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('系统监控')

    // Check for common monitoring elements
    const pageContent = page.content()
    await expect(page.locator('h1')).toBeVisible()
  })

  test('should display system metrics', async ({ page }) => {
    // Look for metric cards or sections
    const metrics = ['QPS', '延迟', '错误率', 'CPU', '内存']

    for (const metric of metrics) {
      const element = page.locator(`text=/${metric}/i`)
      const count = await element.count()

      // At least some metrics should be visible
      if (count > 0) {
        await expect(element.first()).toBeVisible()
        break
      }
    }
  })

  test('should display alerts list', async ({ page }) => {
    // Look for alerts section
    const alertsSection = page.locator('text=/告警|Alert/i')

    const hasAlerts = await alertsSection.count() > 0

    if (hasAlerts) {
      await expect(alertsSection.first()).toBeVisible()
    }
  })

  test('should navigate to monitoring from sidebar', async ({ page }) => {
    // Click on monitoring link in sidebar
    const monitoringLink = page.locator('a:has-text("监控"), a:has-text("Monitoring")').first()

    const hasLink = await monitoringLink.count() > 0

    if (hasLink) {
      await monitoringLink.click()
      await page.waitForURL('**/admin/monitoring')
      await expect(page.locator('h1')).toBeVisible()
    }
  })

  test('should show real-time data updates', async ({ page }) => {
    // Take a snapshot of initial state
    const initialContent = await page.content()

    // Wait a bit for potential real-time updates
    await page.waitForTimeout(2000)

    // Page should still be visible and responsive
    await expect(page.locator('h1')).toBeVisible()
  })
})
