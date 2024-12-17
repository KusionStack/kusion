import React, { lazy, ReactNode, Suspense } from 'react'
import { Navigate, Outlet, useRoutes } from 'react-router-dom'
import { MacCommandOutlined } from '@ant-design/icons'
import Layout from '@/components/layout'
import Loading from '@/components/loading'
import ProjectDetail from '@/pages/projects/projectDetail'

const Projects = lazy(() => import('@/pages/projects'))
const Insights = lazy(() => import('@/pages/insights'))
const Modules = lazy(() => import('@/pages/modules'))
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

const router: RouteObject[] = [
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        key: '/projects',
        path: '/projects',
        title: 'Projects',
        element: <Outlet />,
        icon: <MacCommandOutlined />,
        children: [
          {
            index: true,
            key: '',
            path: '',
            element: lazyLoad(<Projects />)
          },
          {
            key: 'projectDetail',
            path: 'projectDetail/:id',
            element: lazyLoad(<ProjectDetail />)
          },
        ]
      },
      {
        key: '/workspaces',
        path: '/workspaces',
        title: 'Workspaces',
        element: lazyLoad(<Workspaces />),
        icon: <MacCommandOutlined />,
      },
      {
        key: '/modules',
        path: '/modules',
        title: 'modules',
        element: lazyLoad(<Modules />),
        icon: <MacCommandOutlined />,
      },
      {
        key: '/insights',
        path: '/insights',
        title: 'Insights',
        element: lazyLoad(<Insights />),
        icon: <MacCommandOutlined />,
      },
      // {
      //   key: '/search',
      //   path: '/search',
      //   element: <Outlet />,
      //   icon: <SearchOutlined />,
      //   children: [
      //     {
      //       key: 'result',
      //       path: 'result',
      //       element: lazyLoad(<Result />),
      //     },
      //     {
      //       path: '',
      //       element: lazyLoad(<Search />),
      //       children: [
      //         {
      //           // index: true,
      //           key: 'modules',
      //           path: 'modules',
      //           element: lazyLoad(<Modules />),
      //         },
      //         {
      //           key: 'projects',
      //           path: 'projects',
      //           element: lazyLoad(<Projects />),
      //         },
      //         {
      //           key: 'insights',
      //           path: 'insights',
      //           element: lazyLoad(<Insights />),
      //         },
      //         {
      //           key: 'workspaces',
      //           path: 'workspaces',
      //           element: lazyLoad(<Workspaces />),
      //         },
      //         {
      //           path: '',
      //           element: <Navigate replace to="modules" />,
      //         },
      //       ],
      //     },
      //   ],
      // },
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
