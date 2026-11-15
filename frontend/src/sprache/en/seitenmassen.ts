// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'Preview',
  vorschauErklaerung:
    'The preview shows, per card, what would change. Nothing is written until “Apply” — and only if nobody wrote in between.',
  karte: 'Card',
  alt: 'Before',
  neu: 'After',
  keinWertBisher: 'no value',
  bleibtGleich: 'unchanged — already set',
  zusammenfassung:
    '{geaendert} card(s) will change, {gleich} stay as they are.',
  nichtsZuTun:
    'All selected cards already have this value. There is nothing to change.',
  konflikt:
    'These cards have changed since the preview. Nothing was written — fetch the preview again:',
  inzwischenGeaendert: 'changed since',
  angewendet: '{anzahl} card(s) changed.',
} as const
