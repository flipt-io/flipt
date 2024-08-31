import loadable from '@loadable/component';
import nightwind from 'nightwind/helper';
import formbricks from '@formbricks/js/website';
import { useEffect, useState } from 'react';
import { Helmet } from 'react-helmet';
import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { createHashRouter, redirect, RouterProvider } from 'react-router-dom';
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
import { store } from './store';
const Flags = loadable(() => import('./app/flags/Flags'));
const ConditionalFlagRouter = loadable(
  () => import('./app/flags/ConditionalFlagRouter')
);
const Rules = loadable(() => import('./app/flags/rules/Rules'));
const Analytics = loadable(() => import('./app/flags/analytics/Analytics'));
const Segments = loadable(() => import('./app/segments/Segments'));
const Console = loadable(() => import('./app/console/Console'));
const Login = loadable(() => import('./app/auth/Login'));
const Settings = loadable(() => import('./app/Settings'));
const Onboarding = loadable(() => import('./app/Onboarding'));
const Support = loadable(() => import('./app/Support'));
const Preferences = loadable(() => import('./app/preferences/Preferences'));
const Namespaces = loadable(() => import('./app/namespaces/Namespaces'));
const Tokens = loadable(() => import('./app/tokens/Tokens'));

if (typeof window !== 'undefined') {
  formbricks.init({
    environmentId: import.meta.env.FLIPT_FORMBRICKS_ENVIRONMENT_ID || '',
    apiHost: 'https://app.formbricks.com'
  });
}

const namespacedRoutes = [
  {
    element: <Onboarding firstTime={true} />,
    loader: () => {
      const state = store.getState();
      if (state?.user.completedOnboarding) {
        return redirect('/flags');
      }
      return null;
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
        element: <ConditionalFlagRouter />
      },
      {
        path: 'rules',
        element: <Rules />
      },
      {
        path: 'analytics',
        element: <Analytics />
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
        children: namespacedRoutes
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
        path: 'onboarding',
        element: <Onboarding />
      },
      {
        path: 'support',
        element: <Support />
      },
      ...namespacedRoutes
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

  const { ui } = useSelector(selectConfig);

  const namespace = useSelector(selectCurrentNamespace);

  let title = `Flipt · ${namespace.key}`;
  if (ui.topbar?.label) {
    title = `Flipt/${ui.topbar.label} · ${namespace.key}`;
  }

  return (
    <>
      <Helmet>
        <meta charSet="utf-8" />
        <title>{title}</title>
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
