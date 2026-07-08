import { isSafeRedirectUrl } from './helpers';

export function redirectAfterLogout(
  redirect: (next: string, hard: boolean) => void,
  response: { nextUri?: string },
  issuer?: string
) {
  const nextUri = response.nextUri;
  if (nextUri && typeof nextUri === 'string' && isSafeRedirectUrl(nextUri)) {
    redirect(nextUri, true);
  } else if (issuer) {
    redirect(`//${issuer}`, true);
  } else {
    redirect('/login', false);
  }
}
