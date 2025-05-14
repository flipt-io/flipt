import { VariantProps } from 'class-variance-authority';
import { LucideIcon, PlusIcon } from 'lucide-react';

import { Button, buttonVariants } from './ui/button';

type ButtonProps = React.ComponentProps<'button'> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

export { Button };

export const ButtonWithPlus = (props: ButtonProps) => {
  return (
    <Button {...props}>
      <PlusIcon className="-ml-1.5 mr-1.5 h-4 w-4" aria-hidden="true" />
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
