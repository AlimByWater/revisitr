import { useState } from 'react'
import { createBrowserRouter, redirect, Outlet } from 'react-router-dom'
import { Sidebar } from './components/layout/Sidebar'
import { Footer } from './components/layout/Footer'
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
import CreateBotPage from './routes/dashboard/bots/create'
import ClientsPage from './routes/dashboard/clients/index'
import ClientDetailPage from './routes/dashboard/clients/$clientId'
import LoyaltyProgramsPage from './routes/dashboard/loyalty/index'
import ProgramDetailPage from './routes/dashboard/loyalty/$programId'
import POSListPage from './routes/dashboard/pos/index'
import POSDetailPage from './routes/dashboard/pos/$posId'
import CampaignsPage from './routes/dashboard/campaigns/index'
import CreateCampaignPage from './routes/dashboard/campaigns/create'
import CampaignDetailPage from './routes/dashboard/campaigns/$campaignId'
import ScenarioDetailPage from './routes/dashboard/campaigns/scenario.$scenarioId'
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
import CreatePromotionPage from './routes/dashboard/promotions/create'
import PredictionsPage from './routes/dashboard/clients/predictions'
import RFMDashboardPage from './routes/dashboard/rfm/index'
import RFMOnboardingPage from './routes/dashboard/rfm/onboarding'
import RFMTemplatePage from './routes/dashboard/rfm/template'
import SegmentDetailPage from './routes/dashboard/rfm/segments/$segment'
import AccountPage from './routes/dashboard/account/index'
import CustomSegmentsPage from './routes/dashboard/clients/custom-segments'

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
    <div className="min-h-screen">
      <Header onMenuToggle={() => setMobileNavOpen(true)} />
      <MobileNav isOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <div className="flex items-start px-4 sm:px-8 lg:px-16 py-4 sm:py-6 gap-4 sm:gap-6">
        <Sidebar />
        <main className="flex-1 min-w-0">
          <Outlet />
        </main>
      </div>
      <Footer />
    </div>
  )
}

function authLoader() {
  // TODO: restore auth check after visual redesign
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
            { path: 'bots/create', element: <CreateBotPage /> },
            { path: 'bots/:botId', element: <BotDetailPage /> },
            { path: 'clients', element: <ClientsPage /> },
            { path: 'clients/segments', element: <SegmentsPage /> },
            { path: 'clients/custom-segments', element: <CustomSegmentsPage /> },
            { path: 'clients/predictions', element: <PredictionsPage /> },
            { path: 'clients/:clientId', element: <ClientDetailPage /> },
            { path: 'loyalty', element: <LoyaltyProgramsPage /> },
            { path: 'loyalty/wallet', element: <WalletPage /> },
            { path: 'loyalty/:programId', element: <ProgramDetailPage /> },
            { path: 'pos', element: <POSListPage /> },
            { path: 'pos/:posId', element: <POSDetailPage /> },
            { path: 'campaigns', element: <CampaignsPage /> },
            { path: 'campaigns/create', element: <CreateCampaignPage /> },
            { path: 'campaigns/:campaignId', element: <CampaignDetailPage /> },
            { path: 'campaigns/scenario/:scenarioId', element: <ScenarioDetailPage /> },
            { path: 'analytics/sales', element: <SalesAnalyticsPage /> },
            { path: 'analytics/loyalty', element: <LoyaltyAnalyticsPage /> },
            { path: 'analytics/mailings', element: <MailingsAnalyticsPage /> },
            { path: 'rfm', element: <RFMDashboardPage /> },
            { path: 'rfm/onboarding', element: <RFMOnboardingPage /> },
            { path: 'rfm/template', element: <RFMTemplatePage /> },
            { path: 'rfm/segments/:segment', element: <SegmentDetailPage /> },
            { path: 'promotions/create', element: <CreatePromotionPage /> },
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
            { path: 'account', element: <AccountPage /> },
          ],
        },
      ],
    },
  ],
  { basename: '/revisitr' },
)
