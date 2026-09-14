export const DECK_OPTIONS = [
  { name: 't-shirt', label: 'T-shirt sizes' },
  { name: 'fibonacci', label: 'Fibonacci' },
] as const;

export type DeckName = (typeof DECK_OPTIONS)[number]['name'];

/**
 * Narrow a name that arrived over the network to a deck this page knows.
 *
 * The frontend and the server ship as one binary, so the page can never be
 * older than the catalogue it is talking to. A name that fails this check is
 * therefore not a newer deck; it is something being wrong, and the page ignores
 * it rather than passing it on.
 */
export function isDeckName(value: string): value is DeckName {
  return DECK_OPTIONS.some((option) => option.name === value);
}

export function deckLabel(name: DeckName): string {
  return DECK_OPTIONS.find((option) => option.name === name)?.label ?? name;
}
