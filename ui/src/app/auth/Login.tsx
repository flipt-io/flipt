import LoginWrapper from '~/app/auth/LoginWrapper';
import { NotificationProvider } from '~/components/NotificationProvider';
import ErrorNotification from '~/components/notifications/ErrorNotification';

export default function Login() {
  return (
    <NotificationProvider>
      <LoginWrapper />
      <ErrorNotification />
    </NotificationProvider>
  );
}
