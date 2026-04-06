/**
 * Generates a UUID v4
 *
 * @returns A UUID v4 string in the format xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
 *
 * @example
 * const id = uuid(); // "018f2a4b-c2d0-4abc-def012345678"
 */

export function uuid(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }

  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (crypto as any).getRandomValues(new Uint8Array(1))[0] % 16;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
