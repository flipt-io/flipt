import { PlusCircleIcon } from '@heroicons/react/24/outline';
import { Icon } from 'types/Icon';

import { cls } from '~/utils/helpers';

type EmptyStateProps = {
  text?: string;
  secondaryText?: string;
  disabled?: boolean;
  Icon?: Icon;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
};

export default function EmptyState(props: EmptyStateProps) {
  const {
    text,
    secondaryText,
    disabled = false,
    Icon = PlusCircleIcon,
    onClick
  } = props;

  return (
    <button
      className={cls(
        'hover:cursor-hand relative block h-full w-full rounded-lg border-2 border-dashed p-12 text-center selection:border-gray-300 hover:border-gray-400 focus:outline-none',
        {
          'hover:cursor-not-allowed': disabled
        }
      )}
      disabled={!onClick || disabled}
      onClick={onClick}
    >
      {Icon && onClick && (
        <Icon
          className="mx-auto h-12 w-12 text-violet-200"
          aria-hidden="true"
        />
      )}
      {text && (
        <span className="mt-2 block text-sm font-medium text-gray-900">
          {text}
        </span>
      )}
      {secondaryText && (
        <span className="mt-2 block text-sm text-gray-400">
          {secondaryText}
        </span>
      )}
    </button>
  );
}
