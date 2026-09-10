export const DECK_OPTIONS = [
  { name: 't-shirt', label: 'T-shirt sizes' },
  { name: 'fibonacci', label: 'Fibonacci' },
] as const;

export type DeckName = (typeof DECK_OPTIONS)[number]['name'];

export function deckLabel(name: DeckName): string {
  return DECK_OPTIONS.find((option) => option.name === name)?.label ?? name;
}
