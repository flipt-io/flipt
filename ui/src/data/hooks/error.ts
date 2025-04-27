import { toast } from 'sonner';

import { getErrorMessage } from '~/utils/helpers';

const setError = (msg: any) => {
  if (msg == null) {
    return;
  }
  toast.error(getErrorMessage(msg), {
    style: {
      background: 'var(--color-red-50)'
    }
  });
};

export const useError = () => {
  return { setError };
};
