import { Loader2 } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cls } from '~/utils/helpers';
const loadingVariants = cva('flex items-center justify-center w-full', {
  variants: {
    size: {
      sm: 'h-4 w-4',
      default: 'h-8 w-8',
      lg: 'h-12 w-12'
    },
    variant: {
      default: '',
      fullscreen: 'h-screen',
      start: 'justify-start'
    }
  },
  defaultVariants: {
    size: 'default',
    variant: 'default'
  }
});

export interface LoadingProps extends VariantProps<typeof loadingVariants> {}

function Loading({ size, variant, ...props }: LoadingProps) {
  return (
    <div
      className={loadingVariants({ variant, size })}
      {...props}
      title="loading"
    >
      <Loader2
        className={cls(
          'text-muted-foreground animate-spin',
          loadingVariants({ size })
        )}
      />
    </div>
  );
}

export { Loading, loadingVariants };
