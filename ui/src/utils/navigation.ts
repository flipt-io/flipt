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

// buildAuthorizeUrl resolves an authorize endpoint against a base origin and
// attaches the caller's location as the `state` query parameter so the server
// can redirect back to it after the OAuth/OIDC callback.
export function buildAuthorizeUrl(
  uri: string,
  state?: string,
  origin?: string
): string {
  const url = new URL(uri, origin ?? window.location.origin);
  if (state) {
    url.searchParams.set('state', state);
  }
  return url.toString();
}

// resolvePostLoginRedirect returns where the login page should send an
// already-authenticated user: the location saved in router state by the auth
// guard, falling back to `/` when it is missing or not a safe relative path.
export function resolvePostLoginRedirect(state: unknown): string {
  if (typeof state === 'string') {
    const to = state.trim();
    if (
      to !== '' &&
      to.startsWith('/') &&
      !to.startsWith('//') &&
      !to.includes('\\') &&
      !to.includes('://')
    ) {
      return to;
    }
  }
  return '/';
}
