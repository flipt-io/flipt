import { useContext } from 'react';
import { SessionContext } from '~/components/SessionProvider';

export const useSession = () => {
  return useContext(SessionContext);
};
