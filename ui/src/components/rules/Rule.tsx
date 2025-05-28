import { Ref, forwardRef } from 'react';

import Dropdown from '~/components/Dropdown';

import { IRule } from '~/types/Rule';
import { ISegment } from '~/types/Segment';
import { IVariant } from '~/types/Variant';

import { cls } from '~/utils/helpers';

import QuickEditRuleForm from './QuickEditRuleForm';

type RuleProps = {
  rule: IRule;
  segments: ISegment[];
  onSuccess?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
  index?: number;
  variants: IVariant[];
};

const Rule = forwardRef(
  (
    {
      rule,
      segments,
      onSuccess,
      onDelete,
      style,
      className,
      index,
      variants,
      ...rest
    }: RuleProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rule.id}
      ref={ref}
      style={style}
      data-testid={`rule-${typeof index === 'number' ? index : 0}`}
      className={`${className} w-full items-center space-y-2 rounded-md border bg-background hover:shadow-md hover:shadow-accent sm:flex sm:flex-col lg:px-4 lg:py-2`}
    >
      <div className="w-full rounded-t-lg border-b p-2">
        <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
          <span
            key={rule.id}
            className={cls(
              'hidden h-4 w-4 justify text-muted-foreground hover:cursor-move hover:text-brand sm:flex text-sm'
            )}
            {...rest}
          >
            {typeof index === 'number' ? index + 1 : 1}
          </span>
          <h3
            className={cls(
              'text-sm font-normal text-secondary-foreground hover:cursor-move'
            )}
            {...rest}
          >
            Rule
          </h3>
          <Dropdown
            data-testid="rule-menu-button"
            size="icon"
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
            rule={rule}
            variants={variants}
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
