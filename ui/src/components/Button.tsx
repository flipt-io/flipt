import { Slot } from '@radix-ui/react-slot';
import { LucideIcon, PlusIcon } from 'lucide-react';
import React from 'react';

import { cls } from '~/utils/helpers';

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children?: React.ReactNode;
  variant?:
    | 'primary'
    | 'secondary'
    | 'soft'
    | 'link'
    | 'ghost'
    | 'destructive'
    | 'outline';
  className?: string;
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      children,
      type = 'button',
      variant = 'secondary',
      asChild = false,
      ...props
    },
    ref
  ) => {
    const Comp = asChild ? Slot : 'button';
    const baseStyles = cls(
      'cursor-hand inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium',
      className
    );

    const variantStyles = cls({
      'border border-transparent bg-violet-600 text-white shadow-sm enabled:hover:bg-violet-500':
        variant === 'primary',
      'border border-violet-300 dark:border-violet-700 bg-background text-gray-500 dark:text-gray-300 shadow-xs enabled:hover:bg-gray-50 dark:enabled:hover:bg-gray-800':
        variant === 'secondary',
      'border-violet-300 dark:border-violet-700 text-violet-600 dark:text-violet-400 enabled:hover:bg-violet-100 dark:enabled:hover:bg-violet-900/40':
        variant === 'soft',
      'border-none bg-transparent px-2 py-0 text-sm font-medium text-gray-300 enabled:text-gray-500 dark:enabled:text-gray-300 enabled:hover:text-gray-600 dark:enabled:hover:text-gray-100 disabled:cursor-not-allowed':
        variant === 'link',
      'bg-transparent text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800':
        variant === 'ghost',
      'border border-transparent bg-red-600 text-white shadow-sm enabled:hover:bg-red-500':
        variant === 'destructive',
      'bg-gray-50 ': variant === 'outline'
    });

    return (
      <Comp
        type={type}
        className={cls(baseStyles, variantStyles)}
        ref={ref}
        {...props}
      >
        {children}
      </Comp>
    );
  }
);

Button.displayName = 'Button';

export { Button };

export const ButtonWithPlus = (props: ButtonProps) => {
  return (
    <Button {...props}>
      <PlusIcon
        className="-ml-1.5 mr-1.5 h-4 w-4 text-white"
        aria-hidden="true"
      />
      {props.children}
    </Button>
  );
};

export const TextButton = (props: ButtonProps) => {
  return <Button {...props} variant="link" />;
};

export interface IconButtonProps extends ButtonProps {
  icon: LucideIcon;
}

export const IconButton = ({ icon, ...props }: IconButtonProps) => {
  const Icon = icon;
  return (
    <Button {...props} variant="ghost">
      <Icon className="h-4 w-4" aria-hidden="true" />
    </Button>
  );
};
