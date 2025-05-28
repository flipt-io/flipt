import { Ref, forwardRef } from 'react';

import Dropdown from '~/components/Dropdown';

import { IFlag } from '~/types/Flag';
import { IRollout, rolloutTypeToLabel } from '~/types/Rollout';
import { ISegment } from '~/types/Segment';

import { cls } from '~/utils/helpers';

import QuickEditRolloutForm from './QuickEditRolloutForm';

type RolloutProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
  index?: number;
};

const Rollout = forwardRef(
  (
    {
      flag,
      rollout,
      segments,
      onSuccess,
      onEdit,
      onDelete,
      style,
      className,
      index,
      ...rest
    }: RolloutProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rollout.id}
      ref={ref}
      style={style}
      data-testid={`rollout-${typeof index === 'number' ? index : 0}`}
      className={`${className} w-full items-center space-y-2 rounded-md border bg-background hover:shadow-md hover:shadow-accent sm:flex sm:flex-col lg:px-6 lg:py-2`}
    >
      <div className="w-full rounded-t-lg border-b border-gray-200 dark:border-gray-700 p-2">
        <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
          <span
            key={rollout.id}
            className={cls(
              'hidden h-4 w-4 justify text-muted-foreground hover:cursor-move hover:text-brand sm:flex text-sm'
            )}
            {...rest}
          >
            {typeof index === 'number' ? index + 1 : 1}
          </span>
          <h3
            className={cls(
              'text-sm font-normal hover:cursor-move text-secondary-foreground'
            )}
            {...rest}
          >
            {rolloutTypeToLabel(rollout.type)} Rollout
          </h3>
          <Dropdown
            data-testid="rollout-menu-button"
            size="icon"
            label=""
            kind="dots"
            actions={[
              {
                id: 'rollout-edit',
                label: 'Edit',
                onClick: () => {
                  onEdit && onEdit();
                }
              },
              {
                id: 'rollout-delete',
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
        <div className="flex grow flex-col items-center justify-center sm:ml-2">
          {rollout.description && (
            <div className="flex pb-4 pt-2">
              <p className="text-sm text-muted-foreground/75">
                {rollout.description}
              </p>
            </div>
          )}
          <QuickEditRolloutForm
            flag={flag}
            rollout={rollout}
            segments={segments}
            onSuccess={onSuccess}
          />
        </div>
      </div>
    </li>
  )
);

Rollout.displayName = 'Rollout';
export default Rollout;
