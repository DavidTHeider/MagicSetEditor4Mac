// **Die Vorschau der Massenänderung** — Werkstück U6 von Phase 5.1
// (#1417).
// Eigene Datei, weil `seitenkarten.ts` an der Zeilengrenze steht (§2); der
// Katalog bleibt einer — `seitenkarten.ts` setzt sie in `massen` ein.
export const massenvorschauDe = {
  vorschau: 'Vorschau',
  vorschauErklaerung:
    'Die Vorschau zeigt je Karte, was sich ändern würde. Geschrieben wird erst mit »Anwenden« — und nur, wenn bis dahin niemand dazwischen geschrieben hat.',
  karte: 'Karte',
  alt: 'Bisher',
  neu: 'Danach',
  keinWertBisher: 'kein Wert',
  bleibtGleich: 'bleibt — steht schon so',
  zusammenfassung: '{geaendert} Karte(n) werden geändert, {gleich} bleiben.',
  nichtsZuTun:
    'Alle angehakten Karten stehen schon so. Es gibt nichts zu ändern.',
  konflikt:
    'Diese Karten stehen inzwischen anders da als in der Vorschau. Geschrieben wurde keine einzige — hol die Vorschau neu:',
  inzwischenGeaendert: 'inzwischen geändert',
  angewendet: '{anzahl} Karte(n) geändert.',
} as const
