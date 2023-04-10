import { useContext } from 'react';
import { NotificationContext } from '~/components/NotificationProvider';

export const useSuccess = () => {
  const { success, setSuccess, clearSuccess } = useContext(NotificationContext);
  return { success, setSuccess, clearSuccess };
};
