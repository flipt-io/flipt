import * as Dialog from '@radix-ui/react-dialog';
import { Command } from 'cmdk';
import { SearchIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useLocation, useMatches, useNavigate } from 'react-router';

import { themeChanged } from '~/app/preferences/preferencesSlice';

import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import { Theme } from '~/types/Preferences';
import { RouteMatches } from '~/types/Routes';

import { useAppDispatch } from '~/data/hooks/store';
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
      className="data-[selected=true]:bg-brand/20 data-[selected=true]:text-foreground flex cursor-pointer place-items-center px-4 py-2 data-[selected=true]:rounded-md"
      onSelect={() => {
        item.onSelected();
      }}
    >
      <div className="flex grow flex-col">
        <span className="text-secondary-foreground/90 font-semibold">
          {item.name}
        </span>
        {item.description && (
          <span className="text-muted-foreground dark:text-muted-foreground truncate text-xs dark:data-[selected=true]:text-gray-200">
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
    keywords: ['console', 'debug', 'test', 'Playground']
  }
];

const nonNamespacedRoutes = [
  {
    name: 'Settings: Preferences',
    description: 'Change your preferences',
    route: '/settings',
    keywords: ['settings', 'preferences']
  },
  {
    name: 'Settings: Namespaces',
    description: 'Manage namespaces',
    route: '/settings/namespaces',
    keywords: ['settings', 'namespaces']
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

  const namespaces = useSelector(selectNamespaces).filter(
    (n) => n.key !== currentNamespace?.key
  );

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
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Portal>
        <Dialog.Overlay className="data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 bg-foreground/60 dark:bg-background/80 fixed inset-0 z-20 transition-opacity" />
        <div className="fixed inset-0 z-20 overflow-y-auto p-4 pt-[15vh]">
          <Dialog.Content className="bg-background ring-opacity-5 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 ring-input mx-auto max-w-xl transform rounded-xl p-2 shadow-2xl ring-1 transition-all">
            <Dialog.Title className="sr-only">Flipt Command Menu</Dialog.Title>
            <Dialog.Description className="sr-only">
              Quickly navigate to a page or perform an action.
            </Dialog.Description>
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
                  <SearchIcon
                    className="text-muted-foreground pointer-events-none absolute top-3.5 left-4 h-5 w-5"
                    aria-hidden="true"
                  />
                  <Command.Input
                    className="bg-input/60 h-12 w-full rounded-md border-0 px-4 py-2.5 pr-4 pl-11 focus:ring-0 sm:text-sm"
                    value={search}
                    onValueChange={setSearch}
                  />
                </div>
              </div>

              <Command.List className="flex max-h-96 flex-col overflow-y-auto py-2 text-sm">
                <Command.Empty className="text-muted-foreground mt-4 px-4 text-sm">
                  No results found
                </Command.Empty>

                {page === 'namespaces' && (
                  <>
                    <Command.Item
                      disabled={true}
                      className="text-muted-foreground px-4 py-2.5 font-semibold"
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
                      className="text-muted-foreground px-4 py-2.5 font-semibold"
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

                    {namespaceNavEnabled && namespaces.length > 0 && (
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

                    <Command.Separator className="border-input my-2 border-t" />
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
          </Dialog.Content>
        </div>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
