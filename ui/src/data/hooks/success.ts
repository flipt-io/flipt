import { useRef } from 'react';
import { toast } from 'sonner';

export const useSuccess = () => {
  const toastId = useRef(0);
  const setSuccess = (msg: string) => {
    toast.dismiss(toastId.current);
    toastId.current = toast.success(msg, {
      style: {
        background: 'var(--success)',
        color: 'var(--success-foreground)',
        border: '1px solid var(--success-border)',
        boxShadow: '0 2px 5px rgba(0, 0, 0, 0.2)'
      }
    }) as number;
  };
  return { setSuccess };
};
