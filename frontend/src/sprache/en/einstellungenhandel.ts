// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'Allow trading',
  'richtlinie.handel.markt': 'Allow trading through an outside provider',
  'richtlinie.handel.menge': 'Maximum quantity per line',
  'richtlinie.handel.wert':
    'Maximum offer value in the smallest unit (cents for EUR)',
  'richtlinie.handel.frist': 'Longest hold in minutes',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    'Turns off sharing, deck loans, swaps and the market at once. **A limit, not a preference**: what is off here cannot be turned back on in any profile.',
  'richtlinie.handel.markt':
    'Defaults to **off**. The market reaches an outside provider and money — everything else stays between two people.',
  'richtlinie.handel.menge':
    'How many copies a single line of an offer may name at most. **Empty means no limit.**',
  'richtlinie.handel.wert':
    'What an offer may be worth in total. Applies only where the valuation is **part of the contract** — a mere market view does not limit it. **Empty means no limit.**',
  'richtlinie.handel.frist':
    'How long a commitment may hold copies. **Empty means no limit.**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'Off',
  'richtlinie.handel.an': 'On',
  'richtlinie.handel.markt.aus': 'Off',
  'richtlinie.handel.markt.an': 'On',
}
