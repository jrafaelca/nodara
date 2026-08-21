import { AppSidebar } from '@/components/app-sidebar'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { useAuth } from '@/lib/auth.jsx'

export function DashboardLayout({ children }) {
  const { user, logout } = useAuth()
  return (
    <SidebarProvider>
      <AppSidebar user={user} onLogout={logout} />
      <SidebarInset>{children}</SidebarInset>
    </SidebarProvider>
  )
}

export default DashboardLayout
