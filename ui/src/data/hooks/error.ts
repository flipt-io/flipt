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
        background: 'var(--color-red-50)'
      }
    }) as number;
  };

  return { setError, clearError };
};
