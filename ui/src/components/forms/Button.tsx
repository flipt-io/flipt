import { classNames } from '~/utils/helpers';

type ButtonProps = {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: 'button' | 'submit' | 'reset';
  primary?: boolean;
  className?: string;
  disabled?: boolean;
};

export default function Button(props: ButtonProps) {
  const {
    className,
    onClick,
    children,
    type = 'button',
    primary = false,
    disabled = false
  } = props;

  return (
    <button
      type={type}
      onClick={onClick}
      className={classNames(
        primary
          ? 'border-transparent bg-violet-300 text-white enabled:bg-violet-400 enabled:hover:bg-violet-600 enabled:focus:ring-violet-500'
          : 'border-violet-300 bg-white text-gray-500 enabled:hover:bg-gray-50 enabled:focus:ring-gray-500',
        `mb-1 inline-flex items-center justify-center rounded-md border px-4 py-2 text-sm font-medium shadow-sm focus:outline-none focus:ring-1 focus:ring-offset-1 ${className}`
      )}
      disabled={disabled}
    >
      {children}
    </button>
  );
}
