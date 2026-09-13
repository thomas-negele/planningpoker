## ADDED Requirements

### Requirement: A bundled third-party asset ships with its licence

The self-contained rule means every font, icon and image the application references is served from
its own origin, which in turn means the application redistributes those files. Redistribution is
what a licence governs.

Any third-party asset that is committed to this repository and served to a browser — a typeface
above all, but equally an icon set or an image — SHALL be accompanied in the repository by the text
of the licence it is offered under, stored beside the asset it belongs to and named so that the
pairing is obvious. The licence SHALL permit redistribution in exactly this form: bundled into a
self-hosted application and served from its own origin.

An asset whose licence does not permit that, or whose licence text is not in the repository, SHALL
NOT be committed or served. This is not a formality. The SIL Open Font License, under which the
bundled typeface is offered, requires its own text to travel with the font; shipping the font
without it is a licence violation, and it is one that nobody notices, because the application works
perfectly either way.

#### Scenario: The bundled typeface travels with its licence

- **WHEN** the repository is inspected at the location the bundled typeface is stored in
- **THEN** the text of that typeface's licence is stored beside it

#### Scenario: Nothing is served whose licence is missing

- **WHEN** the third-party assets served by the built application are listed
- **THEN** every one of them has its licence text in the repository, and every one of those licences
  permits redistribution as part of a self-hosted application
