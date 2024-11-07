import { cls } from '~/utils/helpers';

export type ButtonProps = {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: 'button' | 'submit' | 'reset';
  variant?: 'primary' | 'secondary' | 'soft';
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
        'cursor-hand mb-1 inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium focus:outline-none focus:ring-1 focus:ring-offset-1',
        className,
        {
          'cursor-not-allowed': disabled,
          'border border-transparent bg-violet-400 text-white shadow-sm enabled:bg-violet-600 enabled:hover:bg-violet-500 enabled:focus:ring-violet-600':
            variant === 'primary',
          'border border-violet-300 bg-white text-gray-500 shadow-sm enabled:hover:bg-gray-50 enabled:focus:ring-gray-500':
            variant === 'secondary',
          'border-violet-300 text-violet-600 enabled:hover:bg-violet-100 enabled:focus:ring-violet-500':
            variant === 'soft'
        }
      )}
      disabled={disabled}
      title={title}
    >
      {children}
    </button>
  );
}
