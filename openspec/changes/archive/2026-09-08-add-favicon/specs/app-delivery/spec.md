## ADDED Requirements

### Requirement: Application provides a browser tab icon

The application SHALL reference a card-themed favicon from its HTML document and serve
the icon from the application origin, in development and production.

#### Scenario: Entry and room pages reference the icon

- **WHEN** a visitor opens the entry page or a room URL
- **THEN** the HTML document references the same application favicon
- **AND** requesting that icon returns image content from the application origin
