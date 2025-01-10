import React, { lazy, ReactNode, Suspense } from 'react'
import { Navigate, Outlet, useRoutes } from 'react-router-dom'
import Layout from '@/components/layout'
import Loading from '@/components/loading'

import ProjectDetail from '@/pages/projects/projectDetail'
import WorkspacesDetail from '@/pages/workspaces/workspaceDetail'
import Backends from '@/pages/backends'

const Projects = lazy(() => import('@/pages/projects'))
const Insights = lazy(() => import('@/pages/insights'))
const Modules = lazy(() => import('@/pages/modules'))
const Sources = lazy(() => import('@/pages/sources'))
const Workspaces = lazy(() => import('@/pages/workspaces'))

const NotFound = lazy(() => import('@/pages/notfound'))

const lazyLoad = (children: ReactNode): ReactNode => {
  return <Suspense fallback={<Loading />}>{children}</Suspense>
}

export interface RouteObject {
  key?: string
  path?: string
  title?: string
  icon?: React.ReactNode
  element: React.ReactNode
  children?: RouteObject[]
  index?: any
}

export const router: RouteObject[] = [
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        key: '/projects',
        path: '/projects',
        title: 'Projects',
        element: <Outlet />,
        children: [
          {
            index: true,
            key: '',
            path: '',
            element: lazyLoad(<Projects />)
          },
          {
            key: 'project',
            path: 'detail/:projectId',
            element: lazyLoad(<ProjectDetail />)
          },
        ]
      },
      {
        key: '/workspaces',
        path: '/workspaces',
        title: 'Workspaces',
        element: <Outlet />,
        children: [
          {
            index: true,
            key: '',
            path: '',
            element: lazyLoad(<Workspaces />),
          },
          {
            key: 'workspaceDetail',
            path: 'detail/:workspaceId',
            element: lazyLoad(<WorkspacesDetail />)
          },
        ]
      },
      {
        key: '/modules',
        path: '/modules',
        title: 'modules',
        element: lazyLoad(<Modules />),
      },
      {
        key: '/sources',
        path: '/sources',
        title: 'sources',
        element: lazyLoad(<Sources />),
      },
      {
        key: '/insights',
        path: '/insights',
        title: 'Insights',
        element: lazyLoad(<Insights />),
      },
      {
        key: '/backends',
        path: '/backends',
        title: 'backends',
        element: lazyLoad(<Backends />),
      },
      {
        path: '/',
        element: <Navigate replace to="/projects" />,
      },
      {
        path: '*',
        title: '',
        element: <NotFound />,
      },
    ],
  },
]

const WrappedRoutes = () => {
  return useRoutes(router)
}
export default WrappedRoutes
