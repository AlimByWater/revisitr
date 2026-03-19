import { useState } from 'react'
import { createBrowserRouter, redirect, Outlet } from 'react-router-dom'
import { Sidebar } from './components/layout/Sidebar'
import { Header } from './components/layout/Header'
import { MobileNav } from './components/layout/MobileNav'

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

function DashboardLayout() {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

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

function authLoader() {
  const token = localStorage.getItem('token')
  if (!token) return redirect('/auth/login')
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
            { path: 'clients/:clientId', element: <ClientDetailPage /> },
            { path: 'loyalty', element: <LoyaltyProgramsPage /> },
            { path: 'loyalty/:programId', element: <ProgramDetailPage /> },
            { path: 'pos', element: <POSListPage /> },
            { path: 'pos/:posId', element: <POSDetailPage /> },
            { path: 'campaigns', element: <CampaignsPage /> },
            { path: 'campaigns/create', element: <CreateCampaignPage /> },
            { path: 'campaigns/scenarios', element: <ScenariosPage /> },
            { path: 'analytics/sales', element: <SalesAnalyticsPage /> },
            { path: 'analytics/loyalty', element: <LoyaltyAnalyticsPage /> },
            { path: 'analytics/mailings', element: <MailingsAnalyticsPage /> },
            { path: 'campaigns/:campaignId', element: <CampaignDetailPage /> },
          ],
        },
      ],
    },
  ],
  { basename: '/revisitr' },
)
