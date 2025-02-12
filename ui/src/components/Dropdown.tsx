import { type VariantProps, cva } from 'class-variance-authority';
import { ChevronDown, EllipsisVerticalIcon } from 'lucide-react';
import { Fragment } from 'react';

import { Button } from '~/components/Button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '~/components/DropdownMenu';

import { Icon } from '~/types/Icon';

import { cls } from '~/utils/helpers';

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
  kind?: 'dots';
  'data-testid'?: string;
};

export default function Dropdown(props: DropdownProps) {
  const { label, actions, disabled = false, side, kind } = props;
  let BtnIcon = ChevronDown;
  let variant: 'primary' | 'secondary' | 'soft' | 'link' | 'ghost' =
    'secondary';

  if (kind === 'dots') {
    variant = 'ghost';
    BtnIcon = EllipsisVerticalIcon;
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild disabled={disabled}>
        <Button
          disabled={disabled}
          variant={variant}
          type="button"
          data-testid={props['data-testid']}
        >
          {label}
          <BtnIcon className="ml-1 h-4 w-4" aria-hidden="true" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" side={side || 'bottom'}>
        {actions.map((action, i) => (
          <Fragment key={i}>
            {action.variant === 'destructive' && i != 0 && (
              <DropdownMenuSeparator />
            )}

            <DropdownMenuItem
              onSelect={() => {
                if (!action.disabled) {
                  action.onClick();
                }
              }}
              disabled={action.disabled}
              className={cls(dropdownVariants({ variant: action.variant }))}
            >
              {action.icon && <action.icon aria-hidden="true" />}
              {action.label}
            </DropdownMenuItem>
          </Fragment>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
