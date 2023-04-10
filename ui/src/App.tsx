import loadable from '@loadable/component';
import { createHashRouter, RouterProvider } from 'react-router-dom';
import { SWRConfig } from 'swr';
import ErrorLayout from './app/ErrorLayout';
import EditFlag from './app/flags/EditFlag';
import Evaluation from './app/flags/Evaluation';
import Flag from './app/flags/Flag';
import NewFlag from './app/flags/NewFlag';
import Layout from './app/Layout';
import NotFoundLayout from './app/NotFoundLayout';
import NewSegment from './app/segments/NewSegment';
import Segment from './app/segments/Segment';
import SessionProvider from './components/SessionProvider';
import { request } from './data/api';

const Flags = loadable(() => import('./app/flags/Flags'));
const Segments = loadable(() => import('./app/segments/Segments'));
const Console = loadable(() => import('./app/console/Console'));
const Login = loadable(() => import('./app/auth/Login'));
const Settings = loadable(() => import('./app/settings/Settings'));
const Namespaces = loadable(
  () => import('./app/settings/namespaces/Namespaces')
);
const Tokens = loadable(() => import('./app/settings/tokens/Tokens'));

const namespacesRoutes = [
  {
    element: <Flags />,
    handle: {
      namespaced: true
    },
    index: true
  },
  {
    path: 'flags',
    element: <Flags />,
    handle: {
      namespaced: true
    }
  },
  {
    path: 'flags/new',
    element: <NewFlag />
  },
  {
    path: 'flags/:flagKey',
    element: <Flag />,
    children: [
      {
        path: '',
        element: <EditFlag />
      },
      {
        path: 'evaluation',
        element: <Evaluation />
      }
    ]
  },
  {
    path: 'segments',
    element: <Segments />,
    handle: {
      namespaced: true
    }
  },
  {
    path: 'segments/new',
    element: <NewSegment />
  },
  {
    path: 'segments/:segmentKey',
    element: <Segment />
  },
  {
    path: 'console',
    element: <Console />,
    handle: {
      namespaced: true
    }
  }
];

const router = createHashRouter([
  {
    path: '/login',
    element: <Login />,
    errorElement: <ErrorLayout />
  },
  {
    path: '/',
    element: <Layout />,
    errorElement: <ErrorLayout />,
    children: [
      {
        path: 'namespaces/:namespaceKey',
        children: namespacesRoutes
      },
      {
        path: 'settings',
        element: <Settings />,
        children: [
          {
            element: <Namespaces />,
            index: true
          },
          {
            path: 'namespaces',
            element: <Namespaces />
          },
          {
            path: 'tokens',
            element: <Tokens />
          }
        ]
      },
      ...namespacesRoutes
    ]
  },
  {
    path: '*',
    element: <NotFoundLayout />
  }
]);

const fetcher = async (uri: String) => {
  return request('GET', '/api/v1' + uri);
};

export default function App() {
  return (
    <SWRConfig
      value={{
        fetcher
      }}
    >
      <SessionProvider>
        <RouterProvider router={router} />
      </SessionProvider>
    </SWRConfig>
  );
}
