import { PlusCircle } from 'lucide-react';
import { Icon } from 'types/Icon';
import { cls } from '~/utils/helpers';
type EmptyStateProps = {
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
    disabled = false,
    Icon = PlusCircle,
    onClick
  } = props;

  return (
    <button
      className={cls(
        'hover:cursor-hand selection:border-input hover:border-sidebar/50 relative block h-full w-full rounded-lg border-2 border-dashed p-12 text-center focus:outline-hidden',
        {
          'hover:cursor-not-allowed': disabled
        }
      )}
      disabled={!onClick || disabled}
      onClick={onClick}
    >
      {Icon && onClick && (
        <Icon className="text-brand/75 mx-auto h-8 w-8" aria-hidden="true" />
      )}
      {text && <span className="text-muted-foreground mt-2">{text}</span>}
      {secondaryText && (
        <span className="text-muted-foreground/80 mt-2 block text-sm">
          {secondaryText}
        </span>
      )}
    </button>
  );
}
