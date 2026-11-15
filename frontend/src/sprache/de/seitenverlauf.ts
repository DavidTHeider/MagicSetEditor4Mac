// **Die Verlaufsliste der Kartenseite** — Werkstück U6 von Phase 5.1
// (#1417).
// Eigene Datei, weil `seitenkarten.ts` an der Zeilengrenze steht (§2); der
// Katalog bleibt einer — `seitenkarten.ts` setzt sie unter `verlauf` ein.
export const verlaufDe = {
  ueberschrift: 'Verlauf',
  erklaerung:
    'Was zuletzt geschah — der jüngste Schritt zuerst. Eine Massenänderung ist ein Schritt; ⌘Z nimmt den obersten zurück, der Knopf daneben genau diesen.',
  leer: 'Noch kein Verlauf. Die erste Änderung an einer Karte erscheint hier.',
  wann: 'Wann',
  karte: 'Karte',
  feld: 'Feld',
  bisher: 'Bisher',
  danach: 'Danach',
  wer: 'Wer',
  keinWert: 'kein Wert',
  zurueckgenommen: 'zurückgenommen',
  karten: '{anzahl} Karten',
  zuruecknehmen: 'Diesen Schritt zurücknehmen',
  vergleichen: 'Bilder vergleichen',
  vergleichSchliessen: 'Vergleich schließen',
  vergleichBisher: 'Karte {karte} mit dem Wert von vorher',
  vergleichDanach: 'Karte {karte} mit dem Wert von danach',
  handlungKarte: 'Karte angelegt oder gelöscht',
  // **Wohin ein Kartenschritt ging** (G01, #2248) — der Satz darüber
  // bleibt für einen Schritt, dessen Richtung nicht feststeht.
  handlungKarteAngelegt: 'Karte angelegt',
  handlungKarteGeloescht: 'Karte gelöscht',
  karteVorhanden: 'vorhanden',
  keineKarte: 'keine Karte',
  handlungReihenfolge: 'Reihenfolge',
  handlungSchluesselwort: 'Schlüsselwort',
  urheberFenster: 'Fenster',
  urheberServer: 'MCP',
  urheberKommandozeile: 'Zeile',
  urheberDienst: 'Dienst',
} as const
