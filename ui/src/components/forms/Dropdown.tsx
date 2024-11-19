import { Icon } from '~/types/Icon';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator
} from '~/components/ui/dropdown-menu';

import { ChevronDown } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '~/lib/utils';

const dropdownVariants = cva('', {
  variants: {
    variant: {
      default: 'text-secondary-foreground',
      destructive: 'text-destructive focus:bg-destructive focus:text-white'
    }
  },
  defaultVariants: {
    variant: 'default'
  }
});

interface DropdownAction extends VariantProps<typeof dropdownVariants> {
  id: string;
  label: string;
  icon?: Icon;
  onClick: () => void;
  disabled?: boolean;
}

type DropdownProps = {
  label: string;
  actions: DropdownAction[];
  disabled?: boolean;
  side?: 'top' | 'bottom';
};

export default function Dropdown(props: DropdownProps) {
  const { label, actions, disabled, side } = props;
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" disabled={disabled}>
          {label} <ChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" side={side || 'bottom'}>
        {actions.map((action) => (
          <>
            {action.variant === 'destructive' && <DropdownMenuSeparator />}

            <DropdownMenuItem
              itemID={action.id}
              onClick={action.onClick}
              disabled={action.disabled}
              className={cn(dropdownVariants({ variant: action.variant }))}
            >
              {action.icon && <action.icon />}
              {action.label}
            </DropdownMenuItem>
          </>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
