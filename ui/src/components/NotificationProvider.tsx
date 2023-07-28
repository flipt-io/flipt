import { createContext, useState } from 'react';
import { getErrorMessage } from '~/utils/helpers';

type ErrorContext = {
  error: string | null;
  setError(_error: unknown): void;
  clearError(): void;
};

type SuccessContext = {
  success: string | null;
  setSuccess(_success: string | null): void;
  clearSuccess(): void;
};

export const NotificationContext = createContext<ErrorContext & SuccessContext>(
  {
    error: null as string | null,
    setError(_error: unknown) {},
    clearError() {},
    success: null as string | null,
    setSuccess(_success: string | null) {},
    clearSuccess() {}
  }
);

export function NotificationProvider({
  children
}: {
  children: React.ReactNode;
}) {
  // eslint-disable-next-line @typescript-eslint/naming-convention
  const [error, setError_] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const setError = (error: unknown) => {
    if (error === null) {
      setError_(null);
      return;
    }
    setError_(getErrorMessage(error));
  };

  return (
    <NotificationContext.Provider
      value={{
        error,
        setError,
        clearError: () => setError(null),
        success,
        setSuccess,
        clearSuccess: () => setSuccess(null)
      }}
    >
      {children}
    </NotificationContext.Provider>
  );
}
