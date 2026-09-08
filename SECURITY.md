# Security

Report vulnerabilities privately:

- Use GitHub private vulnerability reporting if enabled in the Security tab.
- Otherwise, email [mail@thomasnegele.de](mailto:mail@thomasnegele.de).

Include reproduction steps, expected behaviour and actual behaviour. Exclude other
people's personal data. No response time is guaranteed.

## Supported version

Only the current `main` branch is supported. There are no backports or automatic
updates for deployed containers; rebuild to receive changes.

## Access model

The following behaviours are intentional:

- Anyone who knows or guesses a room URL can enter. There are no accounts or passwords.
- Visitors receive participant information and revealed results before taking a seat.
- Any seated participant can reveal or start a new round.
- Restarting the server deletes all rooms and votes.
- Instances do not share state; run one instance.

Hidden-vote disclosure, seat impersonation and server compromise are security issues.
See [PRIVACY.md](PRIVACY.md) for data storage and cookie details.
