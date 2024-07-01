import { Menu, Transition } from '@headlessui/react';
import { EllipsisVerticalIcon } from '@heroicons/react/24/outline';
import { forwardRef, Fragment, Ref } from 'react';
import { IEvaluatable } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { ISegment } from '~/types/Segment';
import { cls } from '~/utils/helpers';
import QuickEditRuleForm from './forms/QuickEditRuleForm';

type RuleProps = {
  flag: IFlag;
  rule: IEvaluatable;
  segments: ISegment[];
  onSuccess?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
  readOnly?: boolean;
};

const Rule = forwardRef(
  (
    {
      flag,
      rule,
      segments,
      onSuccess,
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
      className={`${className} bg-white border-violet-300 w-full items-center space-y-2 rounded-md border shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-4 lg:py-2`}
    >
      <div className="bg-white border-gray-200 w-full border-b p-2">
        <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
          <span
            key={rule.id}
            className={cls(
              'text-gray-400 hidden h-4 w-4 justify-start hover:text-violet-300 sm:flex',
              {
                'hover:cursor-not-allowed': readOnly,
                'hover:cursor-move': !readOnly
              }
            )}
            {...rest}
          >
            {rule.rank}
          </span>
          <h3
            className={cls('text-gray-700 text-sm font-normal leading-6', {
              'hover:cursor-not-allowed': readOnly,
              'hover:cursor-move': !readOnly
            })}
            {...rest}
          >
            Rule
          </h3>
          <Menu as="div" className="hidden sm:flex">
            <Menu.Button
              data-testid="rule-menu-button"
              className="text-gray-600 ml-4 block hover:text-gray-900"
            >
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
                <Menu.Items className="bg-white absolute right-0 z-10 mt-2 w-32 origin-top-right rounded-md py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none">
                  <Menu.Item>
                    {({ active }) => (
                      <a
                        href="#"
                        onClick={(e) => {
                          e.preventDefault();
                          onDelete && onDelete();
                        }}
                        className={cls(
                          'text-gray-900 block px-3 py-1 text-sm leading-6',
                          { 'bg-gray-50': active }
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
      </div>
      <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
        <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
          <QuickEditRuleForm
            flag={flag}
            rule={rule}
            segments={segments}
            onSuccess={onSuccess}
          />
        </div>
      </div>
    </li>
  )
);

Rule.displayName = 'Rule';
export default Rule;
