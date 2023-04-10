import { createContext, useState } from 'react';

type ErrorContext = {
  error: Error | null;
  setError(_error: Error | null): void;
  clearError(): void;
};

type SuccessContext = {
  success: string | null;
  setSuccess(_success: string | null): void;
  clearSuccess(): void;
};

export const NotificationContext = createContext<ErrorContext & SuccessContext>(
  {
    error: null as Error | null,
    setError(_error: Error | null) {},
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
  const [error, setError] = useState<Error | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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
