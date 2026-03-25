import { useState } from 'react'
import { createBrowserRouter, redirect, Outlet } from 'react-router-dom'
import { Sidebar } from './components/layout/Sidebar'
import { Header } from './components/layout/Header'
import { MobileNav } from './components/layout/MobileNav'
import { AuroraSidebar } from './components/layout/AuroraSidebar'
import { AuroraHeader } from './components/layout/AuroraHeader'
import { useTheme } from './contexts/ThemeContext'

// Auth pages
import LoginPage from './routes/auth/login'
import RegisterPage from './routes/auth/register'

// Dashboard pages
import DashboardHome from './routes/dashboard/index'
import BotsPage from './routes/dashboard/bots/index'
import BotDetailPage from './routes/dashboard/bots/$botId'
import ClientsPage from './routes/dashboard/clients/index'
import ClientDetailPage from './routes/dashboard/clients/$clientId'
import LoyaltyProgramsPage from './routes/dashboard/loyalty/index'
import ProgramDetailPage from './routes/dashboard/loyalty/$programId'
import POSListPage from './routes/dashboard/pos/index'
import POSDetailPage from './routes/dashboard/pos/$posId'
import CampaignsPage from './routes/dashboard/campaigns/index'
import CreateCampaignPage from './routes/dashboard/campaigns/create'
import CampaignDetailPage from './routes/dashboard/campaigns/$campaignId'
import ScenariosPage from './routes/dashboard/campaigns/scenarios'
import SalesAnalyticsPage from './routes/dashboard/analytics/sales'
import LoyaltyAnalyticsPage from './routes/dashboard/analytics/loyalty'
import MailingsAnalyticsPage from './routes/dashboard/analytics/mailings'
import IntegrationsPage from './routes/dashboard/integrations/index'
import IntegrationDetailPage from './routes/dashboard/integrations/$integrationId'
import SegmentsPage from './routes/dashboard/clients/segments'
import PromotionsPage from './routes/dashboard/promotions/index'
import PromoCodesPage from './routes/dashboard/promotions/codes'
import PromotionsArchivePage from './routes/dashboard/promotions/archive'
import OnboardingPage from './routes/dashboard/onboarding/index'
import MenusPage from './routes/dashboard/menus/index'
import MenuDetailPage from './routes/dashboard/menus/$menuId'
import WalletPage from './routes/dashboard/loyalty/wallet'
import MarketplacePage from './routes/dashboard/marketplace/index'
import BillingPage from './routes/dashboard/billing/index'
import InvoicesPage from './routes/dashboard/billing/invoices'
import CampaignTemplatesPage from './routes/dashboard/campaigns/templates'
import PredictionsPage from './routes/dashboard/clients/predictions'

function DashboardLayout() {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const { theme } = useTheme()

  if (theme === 'aurora') {
    return (
      <div className="flex min-h-screen">
        <AuroraSidebar />
        <MobileNav isOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
        <div className="flex-1 flex flex-col min-w-0">
          <AuroraHeader />
          <main className="flex-1 p-6 md:p-10 max-w-7xl">
            <Outlet />
          </main>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <MobileNav isOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <div className="flex-1 flex flex-col min-w-0">
        <Header onMenuToggle={() => setMobileNavOpen(true)} />
        <main className="flex-1 p-6 md:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

async function authLoader() {
  const token = localStorage.getItem('token')
  if (!token) return redirect('/auth/login')

  // Check onboarding status
  try {
    const baseURL = import.meta.env.VITE_API_URL || '/api/v1'
    const response = await fetch(`${baseURL}/onboarding`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (response.ok) {
      const data = await response.json()
      if (!data.onboarding_completed) {
        // Only redirect if not already on onboarding page
        if (!window.location.pathname.includes('/onboarding')) {
          return redirect('/dashboard/onboarding')
        }
      }
    }
  } catch {
    // If onboarding check fails, don't block — just continue to dashboard
  }

  return null
}

export const router = createBrowserRouter(
  [
    {
      path: '/',
      children: [
        { index: true, loader: () => redirect('/dashboard') },
        { path: 'auth/login', element: <LoginPage /> },
        { path: 'auth/register', element: <RegisterPage /> },
        {
          path: 'dashboard',
          element: <DashboardLayout />,
          loader: authLoader,
          children: [
            { index: true, element: <DashboardHome /> },
            { path: 'bots', element: <BotsPage /> },
            { path: 'bots/:botId', element: <BotDetailPage /> },
            { path: 'clients', element: <ClientsPage /> },
            { path: 'clients/segments', element: <SegmentsPage /> },
            { path: 'clients/predictions', element: <PredictionsPage /> },
            { path: 'clients/:clientId', element: <ClientDetailPage /> },
            { path: 'loyalty', element: <LoyaltyProgramsPage /> },
            { path: 'loyalty/wallet', element: <WalletPage /> },
            { path: 'loyalty/:programId', element: <ProgramDetailPage /> },
            { path: 'pos', element: <POSListPage /> },
            { path: 'pos/:posId', element: <POSDetailPage /> },
            { path: 'campaigns', element: <CampaignsPage /> },
            { path: 'campaigns/create', element: <CreateCampaignPage /> },
            { path: 'campaigns/scenarios', element: <ScenariosPage /> },
            { path: 'campaigns/templates', element: <CampaignTemplatesPage /> },
            { path: 'analytics/sales', element: <SalesAnalyticsPage /> },
            { path: 'analytics/loyalty', element: <LoyaltyAnalyticsPage /> },
            { path: 'analytics/mailings', element: <MailingsAnalyticsPage /> },
            { path: 'campaigns/:campaignId', element: <CampaignDetailPage /> },
            { path: 'promotions', element: <PromotionsPage /> },
            { path: 'promotions/codes', element: <PromoCodesPage /> },
            { path: 'promotions/archive', element: <PromotionsArchivePage /> },
            { path: 'integrations', element: <IntegrationsPage /> },
            { path: 'integrations/:integrationId', element: <IntegrationDetailPage /> },
            { path: 'onboarding', element: <OnboardingPage /> },
            { path: 'marketplace', element: <MarketplacePage /> },
            { path: 'menus', element: <MenusPage /> },
            { path: 'menus/:menuId', element: <MenuDetailPage /> },
            { path: 'billing', element: <BillingPage /> },
            { path: 'billing/invoices', element: <InvoicesPage /> },
          ],
        },
      ],
    },
  ],
  { basename: '/revisitr' },
)
