import { type VariantProps, cva } from 'class-variance-authority';
import * as React from 'react';

import { cls } from '~/utils/helpers';

const badgeVariants = cva(
  'inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-hidden focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default:
          'border-transparent bg-primary text-primary-foreground shadow-sm hover:bg-primary/80',
        secondary:
          'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
        muted:
          'border-transparent bg-secondary text-muted-foreground hover:bg-secondary/80',
        destructive:
          'border-transparent bg-destructive text-destructive-foreground shadow hover:bg-destructive/80',
        destructiveoutline: 'border-destructive text-destructive shadow',
        outline: 'text-foreground',
        outlinemuted: 'text-muted-foreground',
        enabled: 'text-destructive-foreground bg-green-500'
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cls(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
