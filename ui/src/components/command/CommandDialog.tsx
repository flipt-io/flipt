import { Dialog } from '@headlessui/react';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { Command } from 'cmdk';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useLocation, useMatches, useNavigate } from 'react-router';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import { themeChanged } from '~/app/preferences/preferencesSlice';
import { useAppDispatch } from '~/data/hooks/store';
import { Theme } from '~/types/Preferences';
import { RouteMatches } from '~/types/Routes';
import { addNamespaceToPath } from '~/utils/helpers';

interface Item {
  name: string;
  description?: string;
  onSelected: () => void;
  keywords: string[];
}

interface CommandItemProps {
  item: Item;
}

function CommandItem(props: CommandItemProps) {
  const { item } = props;

  return (
    <Command.Item
      value={item.name + ' ' + item.keywords?.join(' ')}
      key={item.name}
      className="flex cursor-pointer place-items-center px-4 py-2 data-selected:bg-violet-200 dark:data-selected:bg-gray-100"
      onSelect={() => {
        item.onSelected();
      }}
    >
      <div className="flex grow flex-col">
        <span className="font-semibold">{item.name}</span>
        {item.description && (
          <span className="truncate text-xs text-gray-500">
            {item.description}
          </span>
        )}
      </div>
    </Command.Item>
  );
}

const namespacedRoutes = [
  {
    name: 'Flags',
    description: 'Manage feature flags',
    route: '/flags',
    keywords: ['flags', 'feature']
  },
  {
    name: 'Segments',
    description: 'Manage segments',
    route: '/segments',
    keywords: ['segments']
  },
  {
    name: 'Console',
    description: 'Debug and test flags and segments',
    route: '/console',
    keywords: ['console', 'debug', 'test']
  }
];

const nonNamespacedRoutes = [
  {
    name: 'Settings: General',
    description: 'General settings',
    route: '/settings',
    keywords: ['settings', 'general']
  },
  {
    name: 'Settings: Namespaces',
    description: 'Manage namespaces',
    route: '/settings/namespaces',
    keywords: ['settings', 'namespaces']
  },
  {
    name: 'Settings: API Tokens',
    description: 'Manage API tokens',
    route: '/settings/tokens',
    keywords: ['settings', 'api', 'tokens']
  },
  {
    name: 'Support',
    description: 'Get help and support',
    route: '/support',
    keywords: ['support', 'help']
  }
];

export default function CommandMenu() {
  const currentNamespace = useSelector(selectCurrentNamespace);

  const dispatch = useAppDispatch();
  const namespaces = useSelector(selectNamespaces);

  const matches = useMatches();
  // if the current route is namespaced, we want to allow the namespace nav to be selectable
  const namespaceNavEnabled = matches.some((m) => {
    let r = m.handle as RouteMatches;
    return r?.namespaced;
  });

  const location = useLocation();
  const navigate = useNavigate();

  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [pages, setPages] = useState(['']);
  const page = pages[pages.length - 1];

  // Toggle the menu when ⌘K is pressed
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      } else if (e.key === 'Escape') {
        e.preventDefault();
        setSearch('');
        setOpen(false);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <Dialog
      open={open}
      onClose={setOpen}
      className="fixed inset-0 z-20 overflow-y-auto p-4 pt-[15vh]"
    >
      <Dialog.Overlay className="fixed inset-0 bg-gray-500/75" />
      <Dialog.Panel className="bg-background mx-auto max-w-xl transform rounded-xl p-2 shadow-2xl ring-1 ring-black/5 transition-all">
        <Command
          loop
          className="text-foreground relative mx-auto flex max-w-2xl flex-col rounded-lg"
          onKeyDown={(e) => {
            if ((e.key === 'Escape' || e.key === 'Backspace') && !search) {
              e.preventDefault();
              setPages((pages) => pages.slice(0, -1));
            }
          }}
        >
          <div className="flex items-center border-slate-500 text-lg font-medium">
            <div className="relative w-full">
              <MagnifyingGlassIcon
                className="pointer-events-none absolute top-3.5 left-4 h-5 w-5 text-gray-400"
                aria-hidden="true"
              />
              <Command.Input
                className="h-12 w-full rounded-md border-0 bg-gray-100 px-4 py-2.5 pr-4 pl-11 text-gray-900 focus:ring-0 sm:text-sm"
                value={search}
                onValueChange={setSearch}
              />
            </div>
          </div>

          <Command.List className="flex max-h-96 flex-col overflow-y-auto py-2 text-sm">
            <Command.Empty className="mt-4 px-4 text-sm text-gray-700">
              No results found
            </Command.Empty>

            {page === 'namespaces' && (
              <>
                <Command.Item
                  disabled={true}
                  className="px-4 py-2.5 font-semibold text-gray-600"
                >
                  Switch Namespace
                </Command.Item>
                {namespaces.map((namespace) => (
                  <CommandItem
                    key={namespace.key}
                    item={{
                      name: namespace.key,
                      description: namespace.description,
                      onSelected: () => {
                        setOpen(false);
                        dispatch(currentNamespaceChanged(namespace));
                        // navigate to the current location.path with the new namespace prependend
                        // e.g. /namespaces/default/segments -> /namespaces/namespaceKey/segments
                        const newPath = addNamespaceToPath(
                          location.pathname,
                          namespace.key
                        );
                        navigate(newPath);
                        setSearch('');
                        setPages((pages) => pages.slice(0, -1));
                      },
                      keywords: [namespace.key]
                    }}
                  />
                ))}
              </>
            )}

            {page === 'theme' && (
              <>
                <Command.Item
                  disabled={true}
                  className="px-4 py-2.5 font-semibold text-gray-600"
                >
                  Change Theme
                </Command.Item>
                <CommandItem
                  item={{
                    name: 'Light',
                    onSelected: () => {
                      setOpen(false);
                      setSearch('');
                      dispatch(themeChanged(Theme.LIGHT));
                      setPages((pages) => pages.slice(0, -1));
                    },
                    keywords: ['themes', 'light']
                  }}
                />
                <CommandItem
                  item={{
                    name: 'Dark',
                    onSelected: () => {
                      setOpen(false);
                      setSearch('');
                      dispatch(themeChanged(Theme.DARK));
                      setPages((pages) => pages.slice(0, -1));
                    },
                    keywords: ['themes', 'dark']
                  }}
                />
                <CommandItem
                  item={{
                    name: 'System',
                    onSelected: () => {
                      setOpen(false);
                      setSearch('');
                      dispatch(themeChanged(Theme.SYSTEM));
                      setPages((pages) => pages.slice(0, -1));
                    },
                    keywords: ['themes', 'system']
                  }}
                />
              </>
            )}

            {!page && (
              <>
                {namespacedRoutes.map((item) => (
                  <CommandItem
                    key={item.name}
                    item={{
                      onSelected: () => {
                        setOpen(false);
                        setSearch('');
                        navigate(
                          `/namespaces/${currentNamespace.key + item.route}`
                        );
                      },
                      ...item
                    }}
                  />
                ))}

                {namespaceNavEnabled && namespaces.length > 1 && (
                  <CommandItem
                    item={{
                      name: 'Switch Namespaces',
                      description: 'Switch to a different namespace',
                      onSelected: () => {
                        setSearch('');
                        setPages([...pages, 'namespaces'] as string[]);
                      },
                      keywords: ['namespaces', 'switch']
                    }}
                  />
                )}

                <Command.Separator className="my-2 border-t border-gray-200" />
                {nonNamespacedRoutes.map((item) => (
                  <CommandItem
                    key={item.name}
                    item={{
                      onSelected: () => {
                        setOpen(false);
                        setSearch('');
                        navigate(item.route);
                      },
                      ...item
                    }}
                  />
                ))}

                <CommandItem
                  item={{
                    name: 'Preferences: Change Theme',
                    description: 'Set the application theme',
                    onSelected: () => {
                      setSearch('');
                      setPages([...pages, 'theme'] as string[]);
                    },
                    keywords: ['preferences', 'theme']
                  }}
                />
              </>
            )}
          </Command.List>
        </Command>
      </Dialog.Panel>
    </Dialog>
  );
}
