/**
 * Post-login deep-link restore for the external OIDC/GitHub authorize hop.
 *
 * The login page hands the browser to the IdP and the callback lands back
 * with a full-page reload, so no router state survives. Before navigating
 * away we stash the current in-app location; after the session comes back
 * we consume it once and send the user home to where they were.
 */

// Storage key for the in-app target stashed before an external authorize
// round trip. sessionStorage is tab-scoped, so a login in one tab can never
// redirect another tab.
export const postLoginRedirectKey = 'flipt_post_login_redirect';

// How long a stashed post-login target stays valid. Bounds the damage of an
// abandoned login (user clicks a provider but never completes the flow at
// the IdP): the stale target is ignored after this long.
export const postLoginRedirectMaxAgeMs = 15 * 60 * 1000;

// Only same-app relative paths are valid post-login targets. Absolute URLs
// (including protocol-relative `//host/...`) are rejected so a tampered
// value can never turn the login flow into an open redirect.
export function isSafeInAppRedirect(candidate: unknown): candidate is string {
  return (
    typeof candidate === 'string' &&
    candidate.startsWith('/') &&
    !candidate.startsWith('//') &&
    !candidate.includes('\\')
  );
}

export function savePostLoginRedirect(target: string) {
  // Callers must pass the *router* location (useLocation), not
  // window.location: with the hash router the path lives after the `#`, so a
  // window.location value looks like `/#/flags` and navigates nowhere.
  // consume() tolerates that shape anyway (belt and braces), but save the
  // router path directly.
  try {
    window.sessionStorage.setItem(
      postLoginRedirectKey,
      JSON.stringify({ target, createdAt: Date.now() })
    );
  } catch {
    // storage unavailable (private mode etc.) — login still works, we just
    // fall back to the dashboard afterwards
  }
}

export function clearPostLoginRedirect() {
  try {
    window.sessionStorage.removeItem(postLoginRedirectKey);
  } catch {
    // ignore — nothing to restore from anyway
  }
}

// Reads and clears the stashed target (consume-once, so it can only ever
// redirect a single login). Returns null when nothing usable was stored.
// `/login` normalizes to `/` since landing there while authenticated just
// bounces home anyway.
export function consumePostLoginRedirect(
  now: number = Date.now(),
  maxAgeMs: number = postLoginRedirectMaxAgeMs
): string | null {
  let raw: string | null = null;
  try {
    raw = window.sessionStorage.getItem(postLoginRedirectKey);
    window.sessionStorage.removeItem(postLoginRedirectKey);
  } catch {
    return null;
  }
  if (!raw) {
    return null;
  }
  let parsed: { target?: unknown; createdAt?: unknown };
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof parsed !== 'object' || parsed === null) {
    return null;
  }
  if (
    typeof parsed.createdAt !== 'number' ||
    parsed.createdAt > now ||
    now - parsed.createdAt > maxAgeMs
  ) {
    return null;
  }
  if (!isSafeInAppRedirect(parsed.target)) {
    return null;
  }
  let target: string = parsed.target;
  // Hash-router tolerance: a window.location-style value (`/#/flags`) still
  // resolves to the router path (`/flags`). Without this, navigating to the
  // raw `/#/...` value lands nowhere and the user ends up on the dashboard.
  if (target.startsWith('/#')) {
    target = target.slice(2) || '/';
    if (!isSafeInAppRedirect(target)) {
      return null;
    }
  }
  return target === '/login' ? '/' : target;
}
