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

const NewRule = forwardRef(
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
    }: NewRuleProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rule.id}
      ref={ref}
      style={style}
      className={`${className} w-full items-center space-y-2 rounded-md border px-6 py-2 shadow-md shadow-violet-100 bg-white border-violet-300 hover:shadow-violet-200 sm:flex sm:flex-col`}
    >
      <div className="flex w-full flex-1 items-center text-xs">
        <span
          key={rule.id}
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
            'ml-2 hidden h-4 w-4 justify-start text-gray-300 hover:text-violet-300 sm:flex'
          )}
          {...rest}
        >
          {rule.rank === 1 && <ArrowDownIcon />}
          {rule.rank === totalRules && <ArrowUpIcon />}
          {rule.rank !== 1 && rule.rank !== totalRules && <ArrowsUpDownIcon />}
        </span>
        <div className="ml-2 flex grow items-center justify-around">
          <span className="flex items-center px-6 py-4 text-xs font-light text-gray-600">
            IF
          </span>
          <Link
            to={`/namespaces/${namespace.key}/segments/${rule.segment.key}`}
          >
            <span className="flex items-center px-6 py-4 font-medium">
              <div className="ml-6 min-w-0 flex-auto space-y-1 text-center">
                <p className="flex whitespace-nowrap text-gray-500">
                  Matches Segment
                </p>
                <p className="truncate bg-violet-50/80 px-3 py-1 font-semibold text-gray-900 hover:underline hover:underline-offset-2">
                  {rule.segment.name}
                </p>
              </div>
            </span>
          </Link>
          <span className="flex items-center px-6 py-4 text-xs font-light text-gray-600">
            THEN
          </span>
          {rule.rollouts.length === 0 && (
            <span className="flex items-center font-medium">
              <CheckIcon className="h-6 w-6 text-violet-400" />
              <div className="ml-6 min-w-0 flex-auto space-y-1 text-center">
                <p className="mt-1 flex text-gray-500">Return Match</p>
                <p className="bg-violet-50/80 px-3 py-1 font-semibold text-gray-900">
                  true
                </p>
              </div>
            </span>
          )}
          {rule.rollouts.length === 1 && (
            <span className="flex items-center font-medium">
              <VariableIcon className="h-6 w-6 text-violet-400" />
              <div className="ml-6 min-w-0 flex-auto space-y-1 text-center">
                <p className="mt-1 flex text-gray-500">Return Variant</p>
                <p className="bg-violet-50/80 px-3 py-1 font-semibold text-gray-900">
                  abc
                </p>
              </div>
            </span>
          )}
          {rule.rollouts.length > 1 && (
            <span className="flex items-center font-medium">
              <VariableIcon className="h-6 w-6 text-violet-400" />
              <div className="ml-6 min-w-0 flex-auto space-y-1 text-center">
                <p className="mt-1 flex text-gray-500">Return a Distribution</p>
              </div>
            </span>
          )}
        </div>
        <Menu as="div" className="hidden sm:flex">
          <Menu.Button className="-m-2.5 block p-2.5 text-gray-500 hover:text-gray-900">
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
                className="inline-flex items-center gap-x-1.5 bg-violet-50/60 px-3 py-1 text-xs font-medium text-gray-700"
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

NewRule.displayName = 'NewRule';
export default NewRule;
