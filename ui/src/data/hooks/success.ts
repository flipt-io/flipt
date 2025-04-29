import { useRef } from 'react';
import { toast } from 'sonner';

export const useSuccess = () => {
  const toastId = useRef(0);
  const setSuccess = (msg: string) => {
    toast.dismiss(toastId.current);
    toastId.current = toast.success(msg, {
      style: {
        background: 'var(--color-green-100)',
        color: 'var(--color-green-500)'
      }
    }) as number;
  };
  return { setSuccess };
};
