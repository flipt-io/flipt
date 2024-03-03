import { Dialog } from '@headlessui/react';
import { ArrowRightIcon } from '@heroicons/react/24/outline';
import { Command } from 'cmdk';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';

interface Item {
  name: string;
  description: string;
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
      className="flex cursor-pointer place-items-center px-4 py-2 data-[selected]:bg-violet-200 data-[selected]:dark:bg-gray-100"
      onSelect={() => {
        item.onSelected();
      }}
    >
      <div className="flex flex-grow flex-col">
        <span className="font-semibold">{item.name}</span>
        <span className="text-gray-500 truncate text-xs">
          {item.description}
        </span>
      </div>
    </Command.Item>
  );
}

export default function CommandMenu() {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const currentNamespace = useSelector(selectCurrentNamespace);

  const namespacedItems = [
    {
      name: 'Flags',
      description: 'Manage feature flags',
      route: `/flags`,
      keywords: ['flags', 'flags', 'feature']
    },
    {
      name: 'Segments',
      description: 'Manage segments',
      route: `/segments`,
      keywords: ['segments', 'segment']
    },
    {
      name: 'Console',
      description: 'Debug and test flags and segments',
      route: `/console`,
      keywords: ['console', 'debug', 'test']
    }
  ];

  const items = [
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

  // Toggle the menu when ⌘K is pressed
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <Dialog
      open={open}
      onClose={setOpen}
      className="fixed inset-0 z-10 overflow-y-auto p-4 pt-[15vh]"
    >
      <Dialog.Overlay className="fixed inset-0 backdrop-blur-[2px]" />
      <Dialog.Panel className="bg-white mx-auto max-w-xl transform rounded-xl p-2 shadow-2xl ring-1 ring-black ring-opacity-5 transition-all">
        <Command
          loop
          className="text-black relative mx-auto flex max-w-2xl flex-col rounded-lg"
        >
          <div className="border-slate-500 flex items-center text-lg font-medium">
            <Command.Input className="text-gray-900 bg-gray-100 w-full rounded-md border-0 px-4 py-2.5 focus:ring-0 sm:text-sm" />
          </div>

          <Command.List className="flex max-h-96 flex-col overflow-y-auto py-2 text-sm">
            <Command.Empty className="text-gray-700 mt-4 px-4 text-sm">
              No results found
            </Command.Empty>

            {namespacedItems.map((item) => (
              <CommandItem
                item={{
                  onSelected: () => {
                    setOpen(false);
                    navigate(
                      `/namespaces/${currentNamespace.key + item.route}`
                    );
                  },
                  ...item
                }}
              />
            ))}
            <Command.Separator className="border-gray-200 my-2 border-t" />
            {items.map((item) => (
              <CommandItem
                item={{
                  onSelected: () => {
                    setOpen(false);
                    navigate(item.route);
                  },
                  ...item
                }}
              />
            ))}
          </Command.List>
        </Command>
      </Dialog.Panel>
    </Dialog>
  );
}
