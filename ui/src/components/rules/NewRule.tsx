import { Menu, Transition } from '@headlessui/react';
import {
  ArrowDownIcon,
  ArrowsUpDownIcon,
  ArrowUpIcon,
  CheckIcon,
  ChevronRightIcon,
  EllipsisVerticalIcon,
  UsersIcon,
  VariableIcon
} from '@heroicons/react/24/outline';
import { Fragment } from 'react';
import { IEvaluatable } from '~/types/Evaluatable';
import { INamespace } from '~/types/Namespace';
import { classNames } from '~/utils/helpers';

type NewRuleProps = {
  namespace: INamespace;
  totalRules: number;
  rule: IEvaluatable;
  onEdit?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
  readOnly?: boolean;
};

export default function NewRule(props: NewRuleProps) {
  const {
    namespace,
    totalRules,
    rule,
    onEdit,
    onDelete,
    style,
    className,
    readOnly,
    ...rest
  } = props;
  return (
    <li className="w-full items-center divide-y divide-dotted divide-violet-200 rounded-md border px-6 py-2 shadow-sm shadow-violet-100 bg-white border-violet-300 sm:flex sm:flex-col">
      <div className="flex w-full flex-1 items-center text-xs">
        <span className="ml-2 hidden h-6 w-6 justify-start text-gray-300 sm:flex">
          {rule.rank === 1 && <ArrowDownIcon />}
          {rule.rank === totalRules && <ArrowUpIcon />}
          {rule.rank !== 1 && rule.rank !== totalRules && <ArrowsUpDownIcon />}
        </span>
        <div className="ml-2 flex grow items-center justify-evenly">
          <a href="#" className="group flex">
            <span className="flex items-center px-6 py-4 font-medium">
              <UsersIcon
                className="h-6 w-6 text-violet-400"
                aria-hidden="true"
              />
              <div className="ml-6 min-w-0 flex-auto text-center">
                <p className="mt-1 flex whitespace-nowrap text-gray-500">
                  Matches Segment
                </p>
                <p className="truncate font-semibold text-gray-900">
                  {rule.segment.name}
                </p>
              </div>
            </span>
          </a>
          <ChevronRightIcon className="h-6 w-6 text-gray-300" />
          {rule.rollouts.length === 0 && (
            <a href="#" className="group flex">
              <span className="flex items-center font-medium">
                <CheckIcon className="h-6 w-6 text-violet-400" />
                <div className="ml-6 min-w-0 flex-auto text-center">
                  <p className="mt-1 flex text-gray-500">Return Match</p>
                  <p className="font-semibold text-gray-900">true</p>
                </div>
              </span>
            </a>
          )}
          {rule.rollouts.length === 1 && (
            <a href="#" className="group flex">
              <span className="flex items-center font-medium">
                <VariableIcon className="h-6 w-6 text-violet-400" />
                <div className="ml-6 min-w-0 flex-auto text-center">
                  <p className="mt-1 flex text-gray-500">Return Variant</p>
                  <p className="font-semibold text-gray-900">abc</p>
                </div>
              </span>
            </a>
          )}
          {rule.rollouts.length > 1 && (
            <a href="#" className="group flex">
              <span className="flex items-center font-medium">
                <VariableIcon className="h-6 w-6 text-violet-400" />
                <div className="ml-6 min-w-0 flex-auto text-center">
                  <p className="mt-1 flex text-gray-500">
                    Return a Distribution
                  </p>
                </div>
              </span>
            </a>
          )}
        </div>
        <Menu as="div" className="hidden sm:flex">
          <Menu.Button className="-m-2.5 block p-2.5 text-gray-500 hover:text-gray-900">
            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className="absolute right-0 z-10 mt-2 w-32 origin-top-right rounded-md py-2 shadow-lg ring-1 ring-gray-900/5 bg-white focus:outline-none">
              <Menu.Item>
                {({ active }) => (
                  <a
                    href="#"
                    className={classNames(
                      active ? 'bg-gray-50' : '',
                      'block px-3 py-1 leading-6 text-gray-900'
                    )}
                  >
                    Edit
                  </a>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <a
                    href="#"
                    className={classNames(
                      active ? 'bg-gray-50' : '',
                      'block px-3 py-1 leading-6 text-gray-900'
                    )}
                  >
                    Delete
                  </a>
                )}
              </Menu.Item>
            </Menu.Items>
          </Transition>
        </Menu>
      </div>
      {rule.rollouts.length > 1 && (
        <div className="flex w-full items-center justify-center px-6 py-2 pt-4 text-xs">
          <div className="flex w-fit space-x-6 px-2 text-xs">
            {rule.rollouts.map((rollout) => (
              <span className="inline-flex items-center gap-x-1.5 rounded-full px-3 py-1 text-xs font-medium text-gray-700 bg-violet-100">
                <div className="truncate text-gray-900">
                  {rollout.variant.key}
                </div>
                <div className="m-auto whitespace-nowrap text-gray-600">
                  {rollout.distribution.rollout} %
                </div>
              </span>
            ))}
          </div>
        </div>
      )}
    </li>
  );
}
