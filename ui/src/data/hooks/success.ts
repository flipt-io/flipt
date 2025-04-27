import { toast } from 'sonner';

const setSuccess = (msg: string) => {
  toast.success(msg, {
    style: {
      background: 'var(--color-green-100)'
    }
  });
};
export const useSuccess = () => {
  return { setSuccess };
};
