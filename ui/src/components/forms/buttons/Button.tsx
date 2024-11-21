import { IconProp } from '@fortawesome/fontawesome-svg-core';
import { faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { cls } from '~/utils/helpers';

export type ButtonProps = {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: 'button' | 'submit' | 'reset';
  variant?: 'primary' | 'secondary' | 'soft' | 'link';
  className?: string;
  title?: string;
  disabled?: boolean;
};

export default function Button(props: ButtonProps) {
  const {
    className,
    onClick,
    children,
    type = 'button',
    variant = 'secondary',
    title,
    disabled = false
  } = props;

  return (
    <button
      type={type}
      onClick={(e) => {
        !disabled && onClick && onClick(e);
      }}
      className={cls(
        'cursor-hand inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium focus:outline-none focus:ring-1 focus:ring-offset-1',
        className,
        {
          'cursor-not-allowed': disabled,
          'border border-transparent bg-violet-400 text-white shadow-sm enabled:bg-violet-600 enabled:hover:bg-violet-500 enabled:focus:ring-violet-600':
            variant === 'primary',
          'border border-violet-300 bg-white text-gray-500 shadow-sm enabled:hover:bg-gray-50 enabled:focus:ring-gray-500':
            variant === 'secondary',
          'border-violet-300 text-violet-600 enabled:hover:bg-violet-100 enabled:focus:ring-violet-500':
            variant === 'soft',
          'enabled:cursor-hand enabled:cursor mb-1 inline-flex items-center justify-center border-0 px-0 py-0 text-sm font-medium text-gray-300 focus:outline-none focus:ring-0 enabled:text-gray-500 enabled:hover:text-gray-600 disabled:cursor-not-allowed':
            variant === 'link'
        }
      )}
      disabled={disabled}
      title={title}
    >
      {children}
    </button>
  );
}

export const ButtonWithPlus = (props: ButtonProps) => {
  return (
    <Button {...props}>
      <FontAwesomeIcon
        icon={faPlus}
        className="-ml-1.5 mr-1.5 h-4 w-4 text-gray-200"
        aria-hidden="true"
      />
      {props.children}
    </Button>
  );
};

export const TextButton = (props: ButtonProps) => {
  return <Button {...props} variant="link" />;
};

export const ButtonIcon = ({
  icon,
  onClick,
  disabled = false
}: {
  icon: IconProp;
  onClick: () => void;
  disabled: boolean;
}) => (
  <button
    type="button"
    className={cls('p-1 text-gray-300 hover:text-gray-500', {
      'hover:text-gray-400': disabled
    })}
    onClick={onClick}
    title={disabled ? 'Not allowed in Read-Only mode' : undefined}
    disabled={disabled}
  >
    <FontAwesomeIcon icon={icon} className="h-4 w-4" aria-hidden="true" />
  </button>
);
