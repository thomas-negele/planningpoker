## MODIFIED Requirements

### Requirement: Other present participants expose an accessible throw picker

Hovering another present participant's seat SHALL reveal a compact overlay offering exactly
Paper ball, Paper plane, Flower, Heart and — only when the server accepts it — Pile of poo, in that
order. Moving the pointer away from
both seat and picker SHALL hide the overlay again. The mouse interface SHALL NOT show a separate
persistent trigger, pin or toggle button beneath the choices. Touch activation of that seat's throw
control SHALL open the same picker. Standard keyboard navigation and button activation SHALL remain
available as an accessibility path, without custom shortcuts or global throw key bindings. Each
choice SHALL have an accessible name; the trigger SHALL identify its target and expose its expanded
state. Keyboard focus SHALL be visibly indicated when navigating by keyboard, but keyboard hints,
shortcut labels, extra keyboard controls and keyboard-only focus decoration SHALL NOT appear during
normal mouse/touch use.

Moving from the seat to its picker SHALL keep the picker usable. Escape and activating outside
the picker SHALL close it; closing by keyboard SHALL return focus to its trigger. The picker
SHALL NOT alter seat geometry, obscure essential controls, make full participant names
unreadable, or extend beyond the visible viewport. The viewer's own seat and away seats SHALL not
offer throws. Throw controls SHALL be unavailable while disconnected or before a fresh connection
has confirmed the viewer's seat.

#### Scenario: A pointer can reach every object

- **WHEN** a participant hovers another present seat and moves into its picker
- **THEN** the overlay remains open and each object the server accepts can be selected

#### Scenario: The picker follows the server's switch

- **WHEN** the server has the pile of poo disabled
- **THEN** the picker offers exactly Paper ball, Paper plane, Flower and Heart, with no gap or
  disabled control where the pile of poo would be
- **AND** when the server has it enabled, the picker offers all five

#### Scenario: Pointer departure leaves no persistent control

- **WHEN** a mouse user moves away from another participant and its picker
- **THEN** the picker and its activation control disappear without leaving a pin, toggle or fixed
  throw button visible

#### Scenario: Keyboard and touch provide the same choices

- **WHEN** a participant uses the keyboard or taps the throw trigger on a narrow touchscreen
- **THEN** the same named choices are available and can be activated without hover

#### Scenario: Five choices fit at every seat

- **WHEN** the pile of poo is enabled and the picker opens for a seat at the outermost left or right of the wide table, or for any
  entry of the narrow list on a 320 pixel wide viewport, next to a maximum-length name
- **THEN** all five choices lie fully inside the viewport, and the seat's name, card and the
  estimation controls remain readable and usable

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

The five objects SHALL be recognisable, locally served SVG illustrations that fit the existing
interface. The paper ball SHALL read as irregular crumpled paper rather than a faceted polygon.
The flower SHALL be one loose flower rather than a bouquet, with seed-selected variation in flower
shape and colour. The heart SHALL be a single heart whose colour is selected from the shared seed
out of a fixed set of cheerful colours — red, pink, orange, yellow, green, blue and purple — and
never black, white, grey, brown or a muted shade. The pile of poo SHALL resemble the familiar
emoji: a brown, three-tiered swirl with a tip curling to one side, two eyes and a wide grin, drawn
for this project rather than reproduced from any vendor's emoji artwork. Every received throw SHALL
enter fully from outside the visible left or right edge, chosen randomly per event, and fly toward
the target's actual displayed seat. It SHALL finish just in front of/below the seat without hiding
its name, card or controls, remain there for 2 seconds, and fade out over 0.6 seconds before
removal.

Flight SHALL take approximately 0.6–1.2 seconds with bounded variation in duration, trajectory,
rotation and impact offset. Impact positions SHALL vary visibly across a bounded area in front
of/below the target instead of converging on one apparent magnetic point. Motion SHALL convey
gravity, momentum and a soft landing: the paper ball tumbles, bounces and rolls onward briefly;
the paper plane glides nose-first and skids; the single flower rotates gently, lands softly and
slides a shorter distance; the heart flies on a light arc, lands softly, slides a short distance
and pulses once, briefly growing and returning to its size; and the pile of poo flies heavily,
lands with a small plop, squashes briefly on impact, returns to its shape with a short wobble and
barely slides. Squash, wobble and pulse SHALL be complete before the object starts resting. Neither
the heart nor the pile of poo SHALL leave a stain, splash, trail or other residue. Each object SHALL
decelerate along a seed-derived continuation of its incoming direction before reaching a distinct
final resting position. The flight interval includes this settling and slide. Repeated throws SHALL
not all follow the same path, speed, impact point or resting point. Objects SHALL not collide with
other objects or move participants, cards or the table.

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

#### Scenario: Hearts arrive in varying cheerful colours

- **WHEN** several hearts are thrown in a room
- **THEN** they show different colours from the cheerful set, every participant sees the same
  colour for the same heart, and no heart is black, white, grey, brown or muted

#### Scenario: The pile of poo lands heavily and leaves nothing behind

- **WHEN** a pile of poo is thrown at a participant
- **THEN** it squashes on impact, wobbles back into shape, slides less than the paper ball, and
  after fading leaves no mark at the seat

#### Scenario: The heart pulses once on landing

- **WHEN** a heart lands
- **THEN** it pulses exactly once during its settling and rests at its normal size

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
without travel, bounce, slide, rotation, squash, wobble or pulse, and then use the same rest and
fade durations. Switching the preference while objects are moving SHALL stop their motion. Object
age SHALL follow elapsed time rather than frame count, so returning to a backgrounded tab SHALL not
replay expired animations.

#### Scenario: Reduced motion keeps the reaction visible

- **WHEN** a throw arrives with reduced motion enabled
- **THEN** its object appears directly below the target, rests for 2 seconds and fades over
  0.6 seconds without a flight

#### Scenario: Reduced motion shows the new objects at rest

- **WHEN** a heart or a pile of poo arrives with reduced motion enabled
- **THEN** it appears at its normal size and shape, without squash, wobble or pulse; a heart keeps
  its seed-selected colour

#### Scenario: A sleeping tab does not resume old throws

- **WHEN** a page resumes after its objects' lifetimes have elapsed
- **THEN** those objects are removed instead of completing old flights in front of the user
