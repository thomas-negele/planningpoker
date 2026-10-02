## MODIFIED Requirements

### Requirement: The browser coordinates throw traffic with the configured message allowance

The server SHALL make the connection's applicable message rate and burst allowance available to
the browser together with the fixed throw policy and the list of objects this server accepts. The
browser SHALL wait for this information and a fresh confirmed seat before enabling throws, and
SHALL offer only the objects on that list. Missing policy information SHALL leave throws
unavailable without affecting game controls.

The central connection SHALL account for every locally sent intent when deciding whether another
throw fits. Excess clicks SHALL be dropped without buffering or automatic retry, and outgoing
socket backlog SHALL suppress new throws. At low configured message rates, the effective throw
rate SHALL decrease as necessary, preserving allowance for ordinary game actions. The browser
SHALL NOT assume the default 10 messages per second or treat a local limit as server authority.

#### Scenario: A low server rate does not turn clicks into disconnects

- **WHEN** the server's message rate is set to 1 per second and a user repeatedly clicks throws
  while making occasional ordinary game actions
- **THEN** the browser suppresses excess throws, leaves message capacity for game actions, and
  clicking alone causes neither repeated message-limit errors nor a disconnection

#### Scenario: A game action takes precedence over an unsent throw

- **WHEN** available message capacity is low and the user votes or reveals while clicking throws
- **THEN** the game intent is sent immediately, its cost is accounted for, and throws that do not
  fit the remaining allowance are discarded without delaying that game intent

#### Scenario: A custom client cannot bypass transport protection

- **WHEN** a client bypasses local controls and floods the socket with throw requests or oversized
  messages
- **THEN** the existing server-side message-rate, message-size and persistent-flood protections
  still apply before unbounded decoding or room work can occur

#### Scenario: The browser learns which objects it may offer

- **WHEN** a browser receives its first snapshot from a server with the pile of poo disabled, and
  another from a server with it enabled
- **THEN** the first policy lists paper ball, paper plane, flower and heart, the second also the
  pile of poo, and each browser offers exactly the objects listed

#### Scenario: The list of objects is not an authority

- **WHEN** a client sends a throw for an object that the server does not accept
- **THEN** the server refuses it as an unknown object, regardless of what the client was told
