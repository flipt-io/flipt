import { Ref, forwardRef } from 'react';

import Dropdown from '~/components/Dropdown';

import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { ISegment } from '~/types/Segment';

import { cls } from '~/utils/helpers';

import QuickEditRuleForm from './QuickEditRuleForm';

type RuleProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  onSuccess?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
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
      ...rest
    }: RuleProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rule.id}
      ref={ref}
      style={style}
      data-testid={`rule-${rule.rank}`}
      className={`${className} w-full items-center space-y-2 rounded-md border border-violet-300 bg-background shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-4 lg:py-2`}
    >
      <div className="w-full rounded-t-lg border-b border-gray-200 p-2">
        <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
          <span
            key={rule.id}
            className={cls(
              'hidden h-4 w-4 justify-start text-gray-400 hover:cursor-move hover:text-violet-300 sm:flex'
            )}
            {...rest}
          >
            {rule.rank + 1}
          </span>
          <h3
            className={cls(
              'text-sm font-normal leading-6 text-gray-700 hover:cursor-move'
            )}
            {...rest}
          >
            Rule
          </h3>
          <Dropdown
            data-testid="rule-menu-button"
            label=""
            kind="dots"
            actions={[
              {
                id: 'rule-delete',
                label: 'Delete',
                variant: 'destructive',
                onClick: () => {
                  onDelete && onDelete();
                }
              }
            ]}
          />
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
