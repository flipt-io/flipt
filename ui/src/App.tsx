import loadable from '@loadable/component';
import nightwind from 'nightwind/helper';
import { useEffect, useState } from 'react';
import { Helmet } from 'react-helmet';
import { useSelector } from 'react-redux';
import { createHashRouter, RouterProvider } from 'react-router-dom';
import ErrorLayout from './app/ErrorLayout';
import Flag from './app/flags/Flag';
import NewFlag from './app/flags/NewFlag';
import Layout from './app/Layout';
import NotFoundLayout from './app/NotFoundLayout';
import { selectTheme } from './app/preferences/preferencesSlice';
import NewSegment from './app/segments/NewSegment';
import Segment from './app/segments/Segment';
import SessionProvider from './components/SessionProvider';
import { Theme } from './types/Preferences';
const Flags = loadable(() => import('./app/flags/Flags'));
const Variants = loadable(() => import('./app/flags/variants/Variants'));
const Rules = loadable(() => import('./app/flags/rules/Rules'));
const Segments = loadable(() => import('./app/segments/Segments'));
const Console = loadable(() => import('./app/console/Console'));
const Login = loadable(() => import('./app/auth/Login'));
const Settings = loadable(() => import('./app/Settings'));
const Support = loadable(() => import('./app/Support'));
const Preferences = loadable(() => import('./app/preferences/Preferences'));
const Namespaces = loadable(() => import('./app/namespaces/Namespaces'));
const Tokens = loadable(() => import('./app/tokens/Tokens'));

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
        element: <Variants />
      },
      {
        path: 'rules',
        element: <Rules />
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
            element: <Preferences />,
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
      {
        path: 'support',
        element: <Support />
      },
      ...namespacesRoutes
    ]
  },
  {
    path: '*',
    element: <NotFoundLayout />
  }
]);

export default function App() {
  const theme = useSelector(selectTheme);
  const [systemPrefersDark, setSystemPrefersDark] = useState(
    window.matchMedia('(prefers-color-scheme: dark)').matches
  );

  useEffect(() => {
    window
      .matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', (e) => setSystemPrefersDark(e.matches));
  }, []);

  useEffect(() => {
    if (theme === Theme.SYSTEM) {
      nightwind.enable(systemPrefersDark);
    } else {
      nightwind.enable(theme === Theme.DARK);
    }
  }, [theme, systemPrefersDark]);

  return (
    <>
      <Helmet>
        <meta charSet="utf-8" />
        <title>Flipt</title>
        <link rel="icon" href="/favicon.svg" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <script dangerouslySetInnerHTML={{ __html: nightwind.init() }} />
      </Helmet>
      <SessionProvider>
        <RouterProvider router={router} />
      </SessionProvider>
    </>
  );
}
