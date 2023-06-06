import {
  ArrowLongRightIcon,
  ArrowsUpDownIcon,
  Bars2Icon,
  VariableIcon
} from '@heroicons/react/24/outline';
import { forwardRef, Ref } from 'react';
import { Link } from 'react-router-dom';
import { IEvaluatable } from '~/types/Evaluatable';
import { INamespace } from '~/types/Namespace';
import { classNames } from '~/utils/helpers';

type RuleProps = {
  namespace: INamespace;
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
      className={`${className} flex rounded-md border p-6 bg-white border-gray-200 hover:shadow hover:shadow-violet-100 hover:border-violet-200`}
    >
      <div
        className={classNames(
          readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
          'flex items-center justify-start text-center'
        )}
        {...rest}
      >
        <ArrowsUpDownIcon className="hidden h-6 w-6 text-gray-400 lg:flex" />
      </div>

      <div className="flex grow flex-col items-center space-y-3 text-center lg:flex-row lg:justify-around lg:space-y-0">
        <div className="flex">
          <div>
            <p
              className={classNames(
                readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
                'text-sm leading-tight text-gray-500'
              )}
              {...rest}
            >
              <span className="text-gray-900">IF</span> Match Segment
            </p>
            <p className="mt-1 truncate text-sm text-gray-500">
              <Link
                to={`/namespaces/${namespace.key}/segments/${rule.segment.key}`}
                className="text-violet-500"
              >
                {rule.segment.name}
              </Link>
            </p>
          </div>
        </div>

        <ArrowLongRightIcon
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
            'hidden h-6 w-6 text-violet-300 lg:flex'
          )}
          {...rest}
        />

        <div
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
            'flex'
          )}
          {...rest}
        >
          <div>
            <p className="text-sm leading-tight text-gray-500">
              <span className="text-gray-900">THEN</span> Return
            </p>
            <p className="mt-1 truncate text-sm text-gray-500">Variant(s)</p>
          </div>
        </div>

        <div
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move'
          )}
          {...rest}
        >
          {rule.rollouts.length == 1 && (
            <Bars2Icon className="hidden h-6 w-6 text-violet-300 lg:flex" />
          )}
          {rule.rollouts.length > 1 && (
            <VariableIcon className="hidden h-6 w-6 text-violet-300 lg:flex" />
          )}
        </div>

        <div
          className={classNames(
            readOnly ? 'hover:cursor-not-allowed' : 'hover:cursor-move',
            'flex flex-col lg:flex-row'
          )}
          {...rest}
        >
          <div className="flex flex-col divide-y divide-dotted divide-violet-200 text-sm">
            {rule.rollouts.map((rollout) => (
              <div
                key={rollout.variant.key}
                className="flex justify-end space-x-5 py-2"
              >
                <div className="truncate text-gray-500">
                  {rollout.variant.key}
                </div>
                <div className="m-auto whitespace-nowrap text-xs text-gray-500">
                  {rollout.distribution.rollout} %
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-end text-center">
        {!readOnly && rule.rollouts.length > 1 && (
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              onEdit && onEdit();
            }}
            className="text-xs text-violet-600 hover:text-violet-900"
          >
            Edit&nbsp;|&nbsp;
          </a>
        )}
        {!readOnly && (
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              onDelete && onDelete();
            }}
            className="text-xs text-violet-600 hover:text-violet-900"
          >
            Delete
          </a>
        )}
      </div>
    </li>
  )
);

Rule.displayName = 'Rule';
export default Rule;
