import { Center, Loader } from '@mantine/core'
import type { ReactNode } from 'react'
import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom'
import { useMe, useSetupStatus } from './lib/api/hooks'
import { AppShell } from './components/AppShell'
import { Login } from './pages/Login'
import { Setup } from './pages/Setup'
import { IssuesView } from './pages/IssuesView'
import { MyIssuesView } from './pages/MyIssuesView'
import { BoardView } from './pages/BoardView'
import { CyclesView } from './pages/CyclesView'
import { ProjectsView } from './pages/ProjectsView'
import { ProjectDetailView } from './pages/ProjectDetailView'
import { WorkLogView } from './pages/WorkLogView'
import { ViewsView } from './pages/ViewsView'
import { SettingsView } from './pages/SettingsView'
import { TemplatesView } from './pages/TemplatesView'

function RequireAuth() {
  const { data: user, isLoading, isError } = useMe()

  if (isLoading) {
    return (
      <Center h="100vh" bg="dark.7">
        <Loader color="accent" />
      </Center>
    )
  }
  if (isError || !user) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}

// RequireSetup gates every route on first-run setup state. Before any user
// exists, /setup is the only reachable route; once setup is complete the
// backend permanently rejects it, so this redirects away from /setup too
// (there's no other way back to it).
function RequireSetup({ children }: { children: ReactNode }) {
  const location = useLocation()
  const { data, isLoading, isError } = useSetupStatus()

  if (isLoading) {
    return (
      <Center h="100vh" bg="dark.7">
        <Loader color="accent" />
      </Center>
    )
  }

  const setupRequired = !isError && data?.required === true
  if (setupRequired && location.pathname !== '/setup') {
    return <Navigate to="/setup" replace />
  }
  if (!setupRequired && location.pathname === '/setup') {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function App() {
  return (
    <RequireSetup>
      <Routes>
        <Route path="/setup" element={<Setup />} />
        <Route path="/login" element={<Login />} />
        <Route element={<RequireAuth />}>
          <Route element={<AppShell />}>
            <Route index element={<Navigate to="/issues" replace />} />
            <Route path="/issues" element={<IssuesView />} />
            <Route path="/my-issues" element={<MyIssuesView />} />
            <Route path="/board" element={<BoardView />} />
            <Route path="/cycles" element={<CyclesView />} />
            <Route path="/projects" element={<ProjectsView />} />
            <Route path="/projects/:id" element={<ProjectDetailView />} />
            <Route path="/worklog" element={<WorkLogView />} />
            <Route path="/views" element={<ViewsView />} />
            <Route path="/templates" element={<TemplatesView />} />
            <Route path="/settings" element={<SettingsView />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </RequireSetup>
  )
}

export default App
