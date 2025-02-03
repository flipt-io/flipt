import { forwardRef, Ref } from 'react';
import { IFlag } from '~/types/Flag';
import { IRollout, rolloutTypeToLabel } from '~/types/Rollout';
import { ISegment } from '~/types/Segment';
import { cls } from '~/utils/helpers';
import Dropdown from '~/components/Dropdown';
import QuickEditRolloutForm from './forms/QuickEditRolloutForm';

type RolloutProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  style?: React.CSSProperties;
  className?: string;
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
      ...rest
    }: RolloutProps,
    ref: Ref<HTMLLIElement>
  ) => (
    <li
      key={rollout.id}
      ref={ref}
      style={style}
      className={`${className} w-full items-center space-y-2 rounded-md border border-violet-300 bg-background shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-6 lg:py-2`}
    >
      <div className="w-full rounded-t-lg border-b border-gray-200 p-2">
        <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
          <span
            key={rollout.id}
            className={cls(
              'hidden h-4 w-4 justify-start text-gray-400 hover:cursor-move hover:text-violet-300 sm:flex'
            )}
            {...rest}
          >
            {rollout.rank}
          </span>
          <h3
            className={cls(
              'text-sm font-normal leading-6 text-gray-700 hover:cursor-move'
            )}
            {...rest}
          >
            {rolloutTypeToLabel(rollout.type)} Rollout
          </h3>
          <Dropdown
            data-testid="rollout-menu-button"
            label=""
            kind="dots"
            actions={[
              {
                id: 'rollout-edit',
                disabled: false,
                label: 'Edit',
                onClick: () => {
                  onEdit && onEdit();
                }
              },
              {
                id: 'rollout-delete',
                disabled: false,
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
              <p className="text-sm font-light text-gray-600">
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
