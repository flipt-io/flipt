import { useContext } from 'react';
import { NotificationContext } from '~/components/NotificationProvider';

export const useError = () => {
  const { error, setError, clearError } = useContext(NotificationContext);
  return { error, setError, clearError };
};
