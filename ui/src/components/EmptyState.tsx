import { PlusCircleIcon } from '@heroicons/react/24/outline';
import { Icon } from 'types/Icon';

type EmptyStateProps = {
  className?: string;
  text?: string;
  secondaryText?: string;
  Icon?: Icon;
  onClick?: () => void;
};

export default function EmptyState(props: EmptyStateProps) {
  const {
    text,
    secondaryText,
    className = '',
    Icon = PlusCircleIcon,
    onClick
  } = props;

  return (
    <button
      className={`${className} relative block h-full w-full rounded-lg border-2 border-dashed border-gray-300 p-12 text-center hover:border-gray-400 focus:outline-none`}
      disabled={!onClick}
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
