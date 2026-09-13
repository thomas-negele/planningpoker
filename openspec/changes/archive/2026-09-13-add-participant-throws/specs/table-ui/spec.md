## ADDED Requirements

### Requirement: Other present participants expose an accessible throw picker

Hovering another present participant's seat SHALL reveal a compact overlay offering exactly
Paper ball, Paper plane and Flower. Moving the pointer away from both seat and picker SHALL hide
the overlay again. The mouse interface SHALL NOT show a separate persistent trigger, pin or toggle
button beneath the three choices. Touch activation of that seat's throw control SHALL open the
same picker. Standard keyboard navigation and button activation SHALL remain available as an
accessibility path, without custom shortcuts or global throw key bindings. Each choice SHALL have
an accessible name; the trigger SHALL identify its target and expose its expanded state. Keyboard
focus SHALL be visibly indicated when navigating by keyboard, but keyboard hints, shortcut labels,
extra keyboard controls and keyboard-only focus decoration SHALL NOT appear during normal
mouse/touch use.

Moving from the seat to its picker SHALL keep the picker usable. Escape and activating outside
the picker SHALL close it; closing by keyboard SHALL return focus to its trigger. The picker
SHALL NOT alter seat geometry, obscure essential controls, or make full participant names
unreadable. The viewer's own seat and away seats SHALL not offer throws. Throw controls SHALL
be unavailable while disconnected or before a fresh connection has confirmed the viewer's seat.

#### Scenario: A pointer can reach every object

- **WHEN** a participant hovers another present seat and moves into its picker
- **THEN** the overlay remains open and each of the three objects can be selected

#### Scenario: Pointer departure leaves no persistent control

- **WHEN** a mouse user moves away from another participant and its picker
- **THEN** the picker and its activation control disappear without leaving a pin, toggle or fixed
  throw button visible

#### Scenario: Keyboard and touch provide the same choices

- **WHEN** a participant uses the keyboard or taps the throw trigger on a narrow touchscreen
- **THEN** the same three named choices are available and can be activated without hover

#### Scenario: Dismissal preserves keyboard orientation

- **WHEN** a keyboard user dismisses an open picker with Escape
- **THEN** the picker closes and focus returns to the trigger for that participant

#### Scenario: Accessible keyboard use adds no pointer-mode interface

- **WHEN** a participant operates the picker with a mouse or touchscreen
- **THEN** no keyboard instructions, shortcut labels or extra keyboard controls are displayed
- **AND** switching to standard keyboard navigation reveals focus and permits the same actions
  without a custom throw shortcut

#### Scenario: A changed target or connection is no longer actionable

- **WHEN** an open picker's target becomes away, or the viewer loses their connection or seat
- **THEN** the picker closes and cannot submit a throw until the relevant conditions are valid

### Requirement: Thrown objects fly in from offscreen and settle at the target seat

The three objects SHALL be recognisable, locally served SVG illustrations that fit the existing
interface. The paper ball SHALL read as irregular crumpled paper rather than a faceted polygon.
The flower SHALL be one loose flower rather than a bouquet, with seed-selected variation in flower
shape and colour. Every received throw SHALL enter fully from outside the visible left or right edge,
chosen randomly per event, and fly toward the target's actual displayed seat. It SHALL finish
just in front of/below the seat without hiding its name, card or controls, remain there for
2 seconds, and fade out over 0.6 seconds before removal.

Flight SHALL take approximately 0.6–1.2 seconds with bounded variation in duration, trajectory,
rotation and impact offset. Impact positions SHALL vary visibly across a bounded area in front
of/below the target instead of converging on one apparent magnetic point. Motion SHALL convey
gravity, momentum and a soft landing: the paper ball tumbles, bounces and rolls onward briefly;
the paper plane glides nose-first and skids; and the single flower rotates gently, lands softly and
slides a shorter distance. Each object SHALL decelerate along a seed-derived continuation of its
incoming direction before reaching a distinct final resting position. The flight interval includes
this settling and slide. Repeated throws SHALL not all follow the same path, speed, impact point or
resting point. Objects SHALL not collide with other objects or move participants, cards or the table.

Animation SHALL use the local layout on both the wide table and narrow participant list.
Scrolling, resizing and participants changing position SHALL not leave resting objects attached
to the wrong seat. Objects SHALL not intercept input or introduce page scrollbars. A target
outside the viewport SHALL not cause automatic scrolling or a misplaced visible effect.

#### Scenario: A throw begins outside the visible page

- **WHEN** a throw starts from either selected side
- **THEN** the entire object initially lies beyond that viewport edge before flying into view

#### Scenario: Repeated throws vary without missing their target

- **WHEN** the same object is thrown repeatedly at one participant
- **THEN** flights show differing paths and speeds within the stated interval, hit visibly varied
  points, and slide or roll a varied short distance before resting in front of/below that seat

#### Scenario: Landed objects disappear on schedule

- **WHEN** an object has settled
- **THEN** it rests for 2 seconds, fades over 0.6 seconds, and is removed without moving the layout

#### Scenario: A changing layout keeps the target meaningful

- **WHEN** the viewport changes between the wide table and narrow list, scrolls, or seats move
- **THEN** retained resting objects follow the correct target, an obsolete flight is safely
  retargeted or discarded, and no effect appears at a stale seat position

#### Scenario: Game controls remain usable under repeated throws

- **WHEN** several objects are flying or resting near one participant
- **THEN** the name and card remain readable, estimation controls still receive input, and the
  page gains no scrollbars from objects entering offscreen

### Requirement: Reduced motion and inactive pages do not replay flights

When reduced motion is requested, a throw SHALL appear directly at its final resting position,
without travel, bounce, slide or rotation, and then use the same rest and fade durations. Switching the preference
while objects are moving SHALL stop their motion. Object age SHALL follow elapsed time rather
than frame count, so returning to a backgrounded tab SHALL not replay expired animations.

#### Scenario: Reduced motion keeps the reaction visible

- **WHEN** a throw arrives with reduced motion enabled
- **THEN** its object appears directly below the target, rests for 2 seconds and fades over
  0.6 seconds without a flight

#### Scenario: A sleeping tab does not resume old throws

- **WHEN** a page resumes after its objects' lifetimes have elapsed
- **THEN** those objects are removed instead of completing old flights in front of the user
