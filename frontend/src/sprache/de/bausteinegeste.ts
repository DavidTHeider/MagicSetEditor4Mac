// **Die Ansagen einer Geste** — GESTE (#2358). Ein eigener Teil des Katalogs
// `bausteine`, wie `bausteinelauf.ts` (§2, §9). Sie gehen in eine
// Live-Region: ein Satz je Schritt, ohne Zeichen wie ↩, die ein
// Vorleseprogramm nicht spricht.
export const geste = {
  gegriffen:
    'Gegriffen. Ziel: {ziel} ({stelle} von {von}). Pfeiltasten wählen das Ziel, Eingabe legt ab, Escape bricht ab.',
  ziel: 'Ziel: {ziel} ({stelle} von {von}).',
  abgelegt: 'Abgelegt: {ziel}.',
  abgebrochen: 'Abgebrochen. Nichts wurde bewegt.',
  keinZiel: 'Für dieses Objekt gibt es gerade kein Ziel.',
  zielFehlt: 'Das Ziel gibt es nicht mehr — nichts wurde bewegt.',
  veraltet: 'Inzwischen hat sich der Stand geändert — nichts wurde bewegt.',
} as const
