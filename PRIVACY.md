# Data and cookies

Technical description of the application. This is not an installation-specific privacy policy.

## Server storage

The application has no database and keeps game state in memory:

| Data | Retained until |
| --- | --- |
| Room ID, trimmed display names, participant IDs and seat tokens | Room deletion |
| Votes | New round or room deletion |

By default, a room is deleted five minutes after its last connection closes, plus up
to 30 seconds for cleanup. The grace period is configurable. Connected rooms do not
expire through inactivity. Restarting the process deletes all game state.

## Cookies

### `pp_seat_<room>`

- Random credential identifying a browser's seat in one room.
- Set by the server; `HttpOnly`, `SameSite=Lax`, scoped to `/ws/<room>`.
- `Secure` on HTTPS, detected through TLS or the proxy's `X-Forwarded-Proto` header.
- Session cookie with no explicit expiry. Browser session restoration can retain it
  across restarts; see [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie#expiresdate).
- Contains no display name.

It restores the existing seat after reconnecting. Losing it requires taking a new
seat; the previous participant remains marked away until the room is deleted.

### `pp_name`

- Optional display name, used to prefill the name field.
- Set only when “Remember my name on this device” is selected and the join or rename
  form is submitted. Unchecked on a fresh browser.
- Retained for one year; readable by page scripts, `SameSite=Lax`, path `/`, `Secure` on HTTPS.
- Unticking the choice deletes it immediately, including when the rename dialog is cancelled.
- Sent with application requests; the server does not read or use it.

Without this cookie, an existing seat token can still return the browser to its seat.

## Room access

Anyone who knows or guesses a room URL can enter. Visitors receive participant
information and revealed results before taking a seat. Custom room names may be easy
to guess. Avoid confidential information in names and room IDs.

## Network and logs

The application loads no third-party fonts, scripts, analytics or other remote
resources at runtime. Production CSP restricts resource loading and connections to
the application origin. Development omits this policy for Vite hot reloading.

The server logs startup, shutdown and errors. It does not log display names, seat
tokens or message contents. An unexpected error while reaching a room can log its ID.
Docker retains stdout/stderr according to the rotation settings in Compose.

Proxy and hosting logs may contain IP addresses, timestamps and request paths,
including room IDs. Their contents and retention depend on the installation.

## Operator information

For an instance used by others, assess applicable requirements and provide the
installation's controller and contact details, legal basis, handling of data-subject
requests, hosting arrangements, applicable processing agreements and log retention.
