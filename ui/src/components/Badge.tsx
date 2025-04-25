import { type VariantProps, cva } from 'class-variance-authority';
import * as React from 'react';

import { cls } from '~/utils/helpers';

// Common styles that are shared across variants
const commonTransparentBorder = 'border-transparent';
const commonHoverBg = {
  primary: 'hover:bg-primary/80',
  secondary: 'hover:bg-secondary/80',
  destructive: 'hover:bg-destructive/80'
};

const badgeVariants = cva(
  'inline-flex items-center rounded-md border px-2 py-1 text-xs font-semibold transition-colors focus:outline-hidden focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        // Base variant
        default: [
          commonTransparentBorder,
          'bg-primary text-primary-foreground shadow-sm',
          commonHoverBg.primary
        ].join(' '),

        // Secondary variants - share background
        secondary: [
          commonTransparentBorder,
          'bg-secondary text-secondary-foreground',
          commonHoverBg.secondary
        ].join(' '),

        // Destructive variant
        destructive: [
          commonTransparentBorder,
          'bg-destructive text-destructive-foreground shadow',
          commonHoverBg.destructive
        ].join(' '),

        // Outline variants - no background
        outline: 'text-foreground',
        outlinemuted: 'text-muted-foreground'
      },
      state: {
        none: '',
        success: 'text-green-600',
        error: 'text-red-600',
        muted: 'text-muted-foreground'
      }
    },
    defaultVariants: {
      variant: 'default',
      state: 'none'
    },
    compoundVariants: [
      {
        variant: 'secondary',
        state: ['success', 'error', 'muted'],
        class: [
          commonTransparentBorder,
          'bg-secondary',
          commonHoverBg.secondary
        ].join(' ')
      }
    ]
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  state?: 'none' | 'success' | 'error' | 'muted';
}

function Badge({ className, variant, state, ...props }: BadgeProps) {
  return (
    <div
      className={cls(badgeVariants({ variant, state }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
