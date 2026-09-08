// Client-side routing for the entry page and room URLs.

export type Route = { name: 'entry' } | { name: 'room'; roomId: string };

const ROOM_PATH = /^\/g\/([^/]+)\/?$/;

function parse(pathname: string): Route {
  const match = ROOM_PATH.exec(pathname);
  if (match) return { name: 'room', roomId: decodeRoomId(match[1]) };
  return { name: 'entry' };
}

/**
 * Preserve undecodable segments so the server rejects them as invalid room IDs.
 * Throwing here would prevent the initial page from rendering.
 */
function decodeRoomId(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

/** The current route, kept in step with the address bar and the back button. */
export const router = $state<{ route: Route }>({ route: parse(window.location.pathname) });

window.addEventListener('popstate', () => {
  router.route = parse(window.location.pathname);
});

/** Navigate without reloading the page. */
export function navigate(path: string): void {
  window.history.pushState({}, '', path);
  router.route = parse(path);
}

/** The path of a room, which is also the invitation link once made absolute. */
export function roomPath(roomId: string): string {
  return `/g/${encodeURIComponent(roomId)}`;
}

/** The full invitation URL for a room, as it should be sent to somebody else. */
export function roomURL(roomId: string): string {
  return new URL(roomPath(roomId), window.location.href).toString();
}
