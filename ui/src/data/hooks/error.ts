import { useRef } from 'react';
import { toast } from 'sonner';

import { getErrorMessage } from '~/utils/helpers';

export const useError = () => {
  const toastId = useRef(0);
  const clearError = () => {
    toast.dismiss(toastId.current);
  };
  const setError = (msg: any) => {
    clearError();
    if (msg == null) {
      return;
    }
    toastId.current = toast.error(getErrorMessage(msg), {
      style: {
        background: 'var(--destructive)',
        color: 'var(--destructive-foreground)',
        border: '1px solid var(--destructive-border)',
        boxShadow: '0 2px 5px rgba(0, 0, 0, 0.2)'
      }
    }) as number;
  };

  return { setError, clearError };
};
