import { useRef } from 'react';
import { toast } from 'sonner';

type NotificationOptions = {
  description?: string;
  duration?: number;
};

export const useNotification = () => {
  const toastId = useRef(0);

  const clearNotification = () => {
    toast.dismiss(toastId.current);
  };

  const setNotification = (msg: string, options?: NotificationOptions) => {
    clearNotification();
    toastId.current = toast(msg, {
      description: options?.description,
      duration: options?.duration || 3000,
      style: {
        background: 'var(--card)',
        color: 'var(--card-foreground)',
        border: '1px solid var(--border)',
        boxShadow: '0 2px 5px rgba(0, 0, 0, 0.2)'
      }
    }) as number;
  };

  return { setNotification, clearNotification };
};
