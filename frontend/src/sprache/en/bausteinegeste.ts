// Die Ansagen einer Geste — GESTE (#2358); Ausgangssprache ist Deutsch.
export const geste = {
  gegriffen:
    'Picked up. Target: {ziel} ({stelle} of {von}). Arrow keys choose the target, Return drops, Escape cancels.',
  ziel: 'Target: {ziel} ({stelle} of {von}).',
  abgelegt: 'Dropped on {ziel}.',
  abgebrochen: 'Cancelled. Nothing was moved.',
  keinZiel: 'There is no target for this object right now.',
  zielFehlt: 'That target no longer exists — nothing was moved.',
  veraltet: 'Things have changed in the meantime — nothing was moved.',
} as const
