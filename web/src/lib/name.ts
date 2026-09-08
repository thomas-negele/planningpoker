// Optional display-name storage. Seat credentials are separate HttpOnly cookies.

const NAME_COOKIE = 'pp_name';

/**
 * Frontend input limit corresponding to game.MaxNameLength. The server validates
 * names independently; HTML maxlength counts UTF-16 units rather than code points.
 */
export const MAX_NAME_LENGTH = 15;

// Remember an opted-in name for one year; the form displays this duration.
const MAX_AGE_SECONDS = 365 * 24 * 60 * 60;

/** Duration label shown next to the storage choice. */
export const REMEMBERED_FOR = 'a year';

/** Shared advisory text for joining and renaming; it adds no name-validation rule. */
export const NAME_VISIBILITY_HINT =
  'A first name or nickname is enough — everyone with the link to this room can see it.';

export function rememberedName(): string {
  const prefix = `${NAME_COOKIE}=`;
  for (const entry of document.cookie.split(';')) {
    const trimmed = entry.trim();
    if (trimmed.startsWith(prefix)) {
      try {
        return decodeURIComponent(trimmed.slice(prefix.length));
      } catch {
        // A cookie somebody edited by hand. Treat it as absent rather than failing.
        return '';
      }
    }
  }
  return '';
}

/** Store a name only after explicit opt-in at the join or rename form. */
export function rememberName(name: string): void {
  const value = encodeURIComponent(name);
  // Apply Secure when the page is served over HTTPS.
  document.cookie = `${NAME_COOKIE}=${value}; path=/; max-age=${MAX_AGE_SECONDS}; SameSite=Lax${secureAttribute()}`;
}

/** Delete the name cookie using the same path and name with a zero max-age. */
export function forgetName(): void {
  document.cookie = `${NAME_COOKIE}=; path=/; max-age=0; SameSite=Lax${secureAttribute()}`;
}

/** Reports whether a name is currently stored on this device. */
export function nameIsRemembered(): boolean {
  return rememberedName() !== '';
}

function secureAttribute(): string {
  return window.location.protocol === 'https:' ? '; Secure' : '';
}
