## ADDED Requirements

### Requirement: A release build carries its source version

The application SHALL have a three-part `major.minor.patch` version recorded with its source. A production build SHALL carry that recorded version without increasing it. Rebuilding the same source revision SHALL produce the same displayed version.

Before a functional change is merged into `main`, its version SHALL be increased once by the major, minor or patch step chosen by the user. The proposed step SHALL be presented to the user before it is applied. A major step increases major and resets minor and patch to zero; a minor step increases minor and resets patch to zero; a patch step increases patch. Documentation-only changes SHALL NOT require an increase.

#### Scenario: Rebuilding the same revision

- **WHEN** the same source revision is built more than once
- **THEN** each build carries the same `major.minor.patch` version

#### Scenario: Functional change is prepared for merge

- **WHEN** a functional change is prepared for merge into `main`
- **THEN** major, minor or patch is proposed, the user chooses the step, and the recorded version is increased once before merge

#### Scenario: Documentation-only change

- **WHEN** a change affects only documentation
- **THEN** it can be merged without changing the application version
