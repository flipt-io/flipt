import { Menu, Transition } from '@headlessui/react';
import {
  ArrowDownIcon,
  ArrowsUpDownIcon,
  ArrowUpIcon,
  CheckIcon,
  EllipsisVerticalIcon,
  VariableIcon
} from '@heroicons/react/24/outline';
import { forwardRef, Fragment, Ref } from 'react';
import { Link } from 'react-router-dom';
import { IEvaluatable } from '~/types/Evaluatable';
import { INamespace } from '~/types/Namespace';
import { classNames } from '~/utils/helpers';

type RuleProps = {
  namespace: INamespace;
  totalRules: number;
  rule: IEvaluatable;
  onEdit?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
  readOnly?: boolean;
};

const Rule = forwardRef(
  (
    {
      namespace,
      totalRules,
      rule,
      onEdit,
      onDelete,
      style,
      className,
      readOnly,
      ...rest
    }: RuleProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rule.id}
      ref={ref}
      style={style}
      className={`${className} w-full items-center space-y-2 rounded-md border shadow-md shadow-violet-100 bg-white border-violet-300 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-6 lg:py-2`}
    >
      <div className="flex w-full flex-1 items-center p-4 text-xs lg:p-0">
        <span
          key={rule.id}
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
            'ml-2 hidden h-4 w-4 justify-start text-gray-400 hover:text-violet-300 sm:flex'
          )}
          {...rest}
        >
          {rule.rank === 1 && <ArrowDownIcon />}
          {rule.rank === totalRules && <ArrowUpIcon />}
          {rule.rank !== 1 && rule.rank !== totalRules && <ArrowsUpDownIcon />}
        </span>
        <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
          <span className="flex items-center text-xs font-light text-gray-600 lg:px-6 lg:py-4">
            IF
          </span>
          <Link
            to={`/namespaces/${namespace.key}/segments/${rule.segment.key}`}
          >
            <span className="flex items-center px-6 py-4">
              <div className="min-w-0 flex-auto space-y-2 text-center">
                <p className="flex whitespace-nowrap text-gray-600">
                  Matches Segment:
                </p>
                <p className="truncate bg-violet-50/80 px-3 py-1 font-medium text-gray-900 hover:underline hover:underline-offset-2">
                  {rule.segment.name}
                </p>
              </div>
            </span>
          </Link>
          <span className="flex items-center px-6 py-4 text-xs font-light text-gray-600">
            THEN
          </span>
          {rule.rollouts.length === 0 && (
            <span className="flex items-center">
              <CheckIcon className="hidden h-6 w-6 text-violet-400 lg:mr-4 lg:block" />
              <div className="min-w-0 flex-auto space-y-2 text-center">
                <p className="mt-1 flex text-gray-600">Return Match:</p>
                <p className="bg-violet-50/80 px-3 py-1 font-medium text-gray-900">
                  true
                </p>
              </div>
            </span>
          )}
          {rule.rollouts.length === 1 && (
            <span className="flex items-center">
              <VariableIcon className="hidden h-6 w-6 text-violet-400 lg:mr-4 lg:block" />
              <div className="min-w-0 flex-auto space-y-2 text-center">
                <p className="mt-1 flex text-gray-600">Return Variant:</p>
                <p className="bg-violet-50/80 px-3 py-1 font-medium text-gray-900">
                  {rule.rollouts[0].variant.key}
                </p>
              </div>
            </span>
          )}
          {rule.rollouts.length > 1 && (
            <span className="flex items-center">
              <VariableIcon className="hidden h-6 w-6 text-violet-400 lg:mr-4 lg:block" />
              <div className="min-w-0 flex-auto space-y-2 text-center">
                <p className="mt-1 flex text-gray-600">Return a Distribution</p>
              </div>
            </span>
          )}
        </div>
        <Menu as="div" className="hidden sm:flex">
          <Menu.Button className="ml-4 block text-gray-600 hover:text-gray-900">
            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          {!readOnly && (
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
                {rule.rollouts.length > 1 && (
                  <Menu.Item>
                    {({ active }) => (
                      <a
                        href="#"
                        onClick={(e) => {
                          e.preventDefault();
                          onEdit && onEdit();
                        }}
                        className={classNames(
                          active ? 'bg-gray-50' : '',
                          'block px-3 py-1 leading-6 text-gray-900'
                        )}
                      >
                        Edit
                      </a>
                    )}
                  </Menu.Item>
                )}
                <Menu.Item>
                  {({ active }) => (
                    <a
                      href="#"
                      onClick={(e) => {
                        e.preventDefault();
                        onDelete && onDelete();
                      }}
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
          )}
        </Menu>
      </div>
      {rule.rollouts.length > 1 && (
        <div className="flex w-full items-center justify-center px-6 py-2 text-xs">
          <div className="flex w-fit space-x-6 px-2 text-xs">
            {rule.rollouts.map((rollout) => (
              <div
                key={rollout.variant.key}
                className="inline-flex items-center gap-x-1.5 px-3 py-1 text-xs font-medium text-gray-700"
              >
                <div className="truncate text-gray-900">
                  {rollout.variant.key}
                </div>
                <div className="m-auto whitespace-nowrap text-gray-600">
                  {rollout.distribution.rollout} %
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </li>
  )
);

Rule.displayName = 'Rule';
export default Rule;
