import { Center, Loader } from '@mantine/core'
import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { useMe } from './lib/api/hooks'
import { AppShell } from './components/AppShell'
import { Login } from './pages/Login'
import { IssuesView } from './pages/IssuesView'
import { MyIssuesView } from './pages/MyIssuesView'
import { BoardView } from './pages/BoardView'
import { CyclesView } from './pages/CyclesView'
import { ProjectsView } from './pages/ProjectsView'
import { ProjectDetailView } from './pages/ProjectDetailView'
import { ViewsView } from './pages/ViewsView'
import { SettingsView } from './pages/SettingsView'

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

function App() {
  return (
    <Routes>
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
          <Route path="/views" element={<ViewsView />} />
          <Route path="/settings" element={<SettingsView />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
