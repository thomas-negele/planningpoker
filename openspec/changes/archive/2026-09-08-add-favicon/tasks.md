## 1. Implementation

- [x] 1.1 Add a self-contained SVG card icon in the app colors and verify its reference in the HTML head.
- [x] 1.2 Build the frontend and verify the icon reference and image response for entry and room pages.
- [x] 1.3 Validate the change, sync the specification and archive it.

## Verification

- `npm run build` passed; the SVG is included in the output.
- Vite HTTP requests and the embedded Go asset handler returned 200 for `/`,
  `/g/demo-room` and `/favicon.svg`; both pages reference the icon and the image
  response matches the source SVG with content type `image/svg+xml`.
- `openspec validate add-favicon --strict` passed.
- No separate design document: this change adds one static asset and an HTML reference.
- Docker and visual browser-tab rendering were not tested.
- Main specifications validated: 7 passed. The favicon requirement matches the archived delta.
