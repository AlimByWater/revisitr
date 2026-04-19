import { Suspense, lazy, useState, type ComponentType, type LazyExoticComponent } from 'react'
import { createBrowserRouter, redirect, Outlet } from 'react-router-dom'
import { Sidebar } from './components/layout/Sidebar'
import { Footer } from './components/layout/Footer'
import { Header } from './components/layout/Header'
import { MobileNav } from './components/layout/MobileNav'
import { AuroraSidebar } from './components/layout/AuroraSidebar'
import { AuroraHeader } from './components/layout/AuroraHeader'
import { useTheme } from './contexts/ThemeContext'

const LoginPage = lazy(() => import('./routes/auth/login'))
const RegisterPage = lazy(() => import('./routes/auth/register'))
const ForgotPasswordPage = lazy(() => import('./routes/auth/forgot-password'))

const DashboardHome = lazy(() => import('./routes/dashboard/index'))
const BotsPage = lazy(() => import('./routes/dashboard/bots/index'))
const BotDetailPage = lazy(() => import('./routes/dashboard/bots/$botId'))
const CreateBotPage = lazy(() => import('./routes/dashboard/bots/create'))
const ClientsPage = lazy(() => import('./routes/dashboard/clients/index'))
const ClientDetailPage = lazy(() => import('./routes/dashboard/clients/$clientId'))
const LoyaltyProgramsPage = lazy(() => import('./routes/dashboard/loyalty/index'))
const ProgramDetailPage = lazy(() => import('./routes/dashboard/loyalty/$programId'))
const POSListPage = lazy(() => import('./routes/dashboard/pos/index'))
const POSDetailPage = lazy(() => import('./routes/dashboard/pos/$posId'))
const CampaignsPage = lazy(() => import('./routes/dashboard/campaigns/index'))
const CreateCampaignPage = lazy(() => import('./routes/dashboard/campaigns/create'))
const CampaignDetailPage = lazy(() => import('./routes/dashboard/campaigns/$campaignId'))
const ScenarioDetailPage = lazy(() => import('./routes/dashboard/campaigns/scenario.$scenarioId'))
const SalesAnalyticsPage = lazy(() => import('./routes/dashboard/analytics/sales'))
const LoyaltyAnalyticsPage = lazy(() => import('./routes/dashboard/analytics/loyalty'))
const MailingsAnalyticsPage = lazy(() => import('./routes/dashboard/analytics/mailings'))
const IntegrationsPage = lazy(() => import('./routes/dashboard/integrations/index'))
const IntegrationDetailPage = lazy(() => import('./routes/dashboard/integrations/$integrationId'))
const SegmentsPage = lazy(() => import('./routes/dashboard/clients/segments'))
const PromotionsPage = lazy(() => import('./routes/dashboard/promotions/index'))
const PromoCodesPage = lazy(() => import('./routes/dashboard/promotions/codes'))
const PromotionsArchivePage = lazy(() => import('./routes/dashboard/promotions/archive'))
const OnboardingPage = lazy(() => import('./routes/dashboard/onboarding/index'))
const MenusPage = lazy(() => import('./routes/dashboard/menus/index'))
const MenuDetailPage = lazy(() => import('./routes/dashboard/menus/$menuId'))
const WalletPage = lazy(() => import('./routes/dashboard/loyalty/wallet'))
const MarketplacePage = lazy(() => import('./routes/dashboard/marketplace/index'))
const BillingPage = lazy(() => import('./routes/dashboard/billing/index'))
const InvoicesPage = lazy(() => import('./routes/dashboard/billing/invoices'))
const CreatePromotionPage = lazy(() => import('./routes/dashboard/promotions/create'))
const PredictionsPage = lazy(() => import('./routes/dashboard/clients/predictions'))
const RFMDashboardPage = lazy(() => import('./routes/dashboard/rfm/index'))
const RFMOnboardingPage = lazy(() => import('./routes/dashboard/rfm/onboarding'))
const RFMTemplatePage = lazy(() => import('./routes/dashboard/rfm/template'))
const SegmentDetailPage = lazy(() => import('./routes/dashboard/rfm/segments/$segment'))
const AccountPage = lazy(() => import('./routes/dashboard/account/index'))
const CustomSegmentsPage = lazy(() => import('./routes/dashboard/clients/custom-segments'))
const EmojiPacksPage = lazy(() => import('./routes/dashboard/emoji-packs/index'))

function RouteFallback() {
  return (
    <div className="flex min-h-[240px] items-center justify-center">
      <div className="h-10 w-10 rounded-full border-2 border-neutral-300 border-t-accent animate-spin" />
    </div>
  )
}

function lazyElement(Component: LazyExoticComponent<ComponentType<any>>) {
  return (
    <Suspense fallback={<RouteFallback />}>
      <Component />
    </Suspense>
  )
}

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

export function authLoader() {
  if (typeof window === 'undefined') {
    return null
  }

  if (!window.localStorage.getItem('token')) {
    return redirect('/auth/login')
  }

  return null
}

export const router = createBrowserRouter(
  [
    {
      path: '/',
      children: [
        { index: true, loader: () => redirect('/dashboard') },
        { path: 'auth/login', element: lazyElement(LoginPage) },
        { path: 'auth/register', element: lazyElement(RegisterPage) },
        { path: 'auth/forgot-password', element: lazyElement(ForgotPasswordPage) },
        {
          path: 'dashboard',
          element: <DashboardLayout />,
          loader: authLoader,
          children: [
            { index: true, element: lazyElement(DashboardHome) },
            { path: 'bots', element: lazyElement(BotsPage) },
            { path: 'bots/create', element: lazyElement(CreateBotPage) },
            { path: 'bots/:botId', element: lazyElement(BotDetailPage) },
            { path: 'clients', element: lazyElement(ClientsPage) },
            { path: 'clients/segments', element: lazyElement(SegmentsPage) },
            { path: 'clients/custom-segments', element: lazyElement(CustomSegmentsPage) },
            { path: 'clients/predictions', element: lazyElement(PredictionsPage) },
            { path: 'clients/:clientId', element: lazyElement(ClientDetailPage) },
            { path: 'loyalty', element: lazyElement(LoyaltyProgramsPage) },
            { path: 'loyalty/wallet', element: lazyElement(WalletPage) },
            { path: 'loyalty/:programId', element: lazyElement(ProgramDetailPage) },
            { path: 'pos', element: lazyElement(POSListPage) },
            { path: 'pos/:posId', element: lazyElement(POSDetailPage) },
            { path: 'campaigns', element: lazyElement(CampaignsPage) },
            { path: 'campaigns/create', element: lazyElement(CreateCampaignPage) },
            { path: 'campaigns/:campaignId', element: lazyElement(CampaignDetailPage) },
            { path: 'campaigns/scenario/:scenarioId', element: lazyElement(ScenarioDetailPage) },
            { path: 'analytics/sales', element: lazyElement(SalesAnalyticsPage) },
            { path: 'analytics/loyalty', element: lazyElement(LoyaltyAnalyticsPage) },
            { path: 'analytics/mailings', element: lazyElement(MailingsAnalyticsPage) },
            { path: 'rfm', element: lazyElement(RFMDashboardPage) },
            { path: 'rfm/onboarding', element: lazyElement(RFMOnboardingPage) },
            { path: 'rfm/template', element: lazyElement(RFMTemplatePage) },
            { path: 'rfm/segments/:segment', element: lazyElement(SegmentDetailPage) },
            { path: 'promotions/create', element: lazyElement(CreatePromotionPage) },
            { path: 'promotions', element: lazyElement(PromotionsPage) },
            { path: 'promotions/codes', element: lazyElement(PromoCodesPage) },
            { path: 'promotions/archive', element: lazyElement(PromotionsArchivePage) },
            { path: 'integrations', element: lazyElement(IntegrationsPage) },
            { path: 'integrations/:integrationId', element: lazyElement(IntegrationDetailPage) },
            { path: 'onboarding', element: lazyElement(OnboardingPage) },
            { path: 'marketplace', element: lazyElement(MarketplacePage) },
            { path: 'menus', element: lazyElement(MenusPage) },
            { path: 'menus/:menuId', element: lazyElement(MenuDetailPage) },
            { path: 'billing', element: lazyElement(BillingPage) },
            { path: 'billing/invoices', element: lazyElement(InvoicesPage) },
            { path: 'account', element: lazyElement(AccountPage) },
            { path: 'emoji-packs', element: lazyElement(EmojiPacksPage) },
          ],
        },
      ],
    },
  ],
  { basename: '/revisitr' },
)
