import { classNames } from '~/utils/helpers';

export type ButtonProps = {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
  title?: string;
  disabled?: boolean;
};

export default function TextButton(props: ButtonProps) {
  const {
    className,
    onClick,
    children,
    type = 'button',
    title,
    disabled = false
  } = props;

  return (
    <button
      type={type}
      onClick={(e) => {
        !disabled && onClick && onClick(e);
      }}
      className={classNames(
        disabled ? 'cursor-not-allowed' : 'cursor-hand',
        `mb-1 inline-flex items-center justify-center rounded-md border-0 text-sm font-medium text-gray-300 focus:outline-none focus:ring-1 focus:ring-offset-1 enabled:text-gray-500 enabled:hover:text-gray-600 ${className}`
      )}
      disabled={disabled}
      title={title}
    >
      {children}
    </button>
  );
}
