import {
  CodeBracketIcon,
  Cog6ToothIcon,
  FlagIcon,
  QuestionMarkCircleIcon,
  UsersIcon
} from '@heroicons/react/24/outline';
import { Separator } from '@radix-ui/react-separator';
import { useSelector } from 'react-redux';
import { NavLink, useMatches } from 'react-router';

import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { RouteMatches } from '~/types/Routes';

import { cls } from '~/utils/helpers';

import NamespaceListbox from './namespaces/NamespaceListbox';

type Icon = (
  props: React.PropsWithoutRef<React.SVGProps<SVGSVGElement>>
) => any;

type NavItemProps = {
  to: string;
  name: string;
  Icon: Icon;
  onClick?: () => void;
};

function NavItem(props: NavItemProps) {
  const { to, name, Icon, onClick } = props;

  return (
    <NavLink
      key={name}
      to={to}
      aria-label={name}
      className={({ isActive }) =>
        cls('flex items-center rounded-md p-2 text-sm  text-white', {
          'bg-gray-800 dark:bg-gray-300 md:bg-gray-50 md:text-gray-700 md:dark:bg-gray-300':
            isActive,
          'hover:bg-gray-700 dark:hover:bg-gray-300 dark:hover:text-gray-900 md:text-gray-600 md:hover:bg-gray-50 md:hover:text-gray-700':
            !isActive
        })
      }
      onClick={onClick}
    >
      <Icon
        className="mr-3 h-6 w-6 shrink-0 text-white md:text-gray-500"
        aria-hidden="true"
      />
      {name}
    </NavLink>
  );
}

type NavProps = {
  className?: string;
  sidebarOpen?: boolean;
  setSidebarOpen?: (open: boolean) => void;
};

export default function Nav(props: NavProps) {
  const { className, sidebarOpen, setSidebarOpen } = props;

  const matches = useMatches();
  let path = '';

  const namespace = useSelector(selectCurrentNamespace);
  path = `/namespaces/${namespace?.key}`;

  // if the current route is namespaced, we want to allow the namespace nav to be selectable
  const namespaceNavEnabled = matches.some((m) => {
    let r = m.handle as RouteMatches;
    return r?.namespaced;
  });

  const navigation = [
    {
      name: 'Flags',
      to: `${path}/flags`,
      Icon: FlagIcon
    },
    {
      name: 'Segments',
      to: `${path}/segments`,
      Icon: UsersIcon
    },
    {
      name: 'Console',
      to: `${path}/console`,
      Icon: CodeBracketIcon
    }
  ];

  const secondaryNavigation = [
    {
      name: 'Settings',
      to: 'settings',
      Icon: Cog6ToothIcon
    },
    {
      name: 'Support',
      to: 'support',
      Icon: QuestionMarkCircleIcon
    }
  ];

  return (
    <nav
      className={`${className} flex grow flex-col overflow-y-auto`}
      aria-label="Sidebar"
    >
      <div className="flex shrink-0 flex-col px-2">
        <NamespaceListbox className="w-full" disabled={!namespaceNavEnabled} />
      </div>
      <Separator
        decorative
        className="h-px m-2 bg-gray-400 opacity-40 dark:bg-gray-600"
      />
      <div className="flex grow flex-col space-y-1 px-2">
        {navigation.map((item) => (
          <NavItem
            key={item.name}
            {...item}
            onClick={() => {
              if (sidebarOpen && setSidebarOpen) {
                setSidebarOpen(false);
              }
            }}
          />
        ))}
      </div>
      <div className="shrink-0 space-y-1 px-2">
        {secondaryNavigation.map((item) => (
          <NavItem
            key={item.name}
            {...item}
            onClick={() => {
              if (sidebarOpen && setSidebarOpen) {
                setSidebarOpen(false);
              }
            }}
          />
        ))}
        <div className="flex space-x-1 px-3 pt-2 text-xs text-gray-400">
          <span className="shrink-0">Command Mode:</span>
          <div className="shrink-0">
            <kbd className="text-gray-400">ctrl</kbd> +{' '}
            <kbd className="text-gray-400">k</kbd>
          </div>
        </div>
      </div>
    </nav>
  );
}
