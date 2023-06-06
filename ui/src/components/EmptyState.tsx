import { PlusCircleIcon } from '@heroicons/react/24/outline';
import { Icon } from 'types/Icon';
import { classNames } from '~/utils/helpers';

type EmptyStateProps = {
  className?: string;
  text?: string;
  secondaryText?: string;
  disabled?: boolean;
  Icon?: Icon;
  onClick?: () => void;
};

export default function EmptyState(props: EmptyStateProps) {
  const {
    text,
    secondaryText,
    className = '',
    disabled = false,
    Icon = PlusCircleIcon,
    onClick
  } = props;

  return (
    <button
      className={classNames(
        disabled ? 'hover:cursor-not-allowed' : 'hover: cursor-hand',
        `${className} relative block h-full w-full rounded-lg border-2 border-dashed p-12 text-center border-gray-300 hover:border-gray-400 focus:outline-none`
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
