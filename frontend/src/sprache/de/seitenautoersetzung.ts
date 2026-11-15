// Die Texte der **Autoersetzung** —
// #1341,
// Vermerk [0060](../../../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
//
// **Eine eigene Datei**, weil `seitenkarten.ts` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Katalog bleibt **einer**: `seitenkarten.ts` setzt sie ein,
// und der Vertragstest gegen die Go-Seite sieht dieselbe Zahl von Katalogen.

export const autoersetzung = {
  ueberschrift: 'Autoersetzung',
  erklaerung:
    'Beim Tippen ersetzt: was links steht, wird zu dem rechts. Die Regeln des Spielpakets stehen fest — ändern lässt sich, was sie einsetzen und ob sie greifen. Eigene Regeln gelten in jedem Set dieses Spiels.',
  schalter: 'Autoersetzung greift',
  muster: 'Suchen',
  ersatz: 'Ersetzen durch',
  aktiv: 'greift',
  ganzesWort: 'ganzes Wort',
  anlegen: 'Regel anlegen',
  loeschen: 'Löschen',
  zuruecksetzen: 'Eigene vergessen',
  // **Gefragt wird davor** — UX-10.AK-04 (#1682): alle eigenen Regeln auf
  // einmal, und keine kommt zurück.
  zuruecksetzenFrage: 'Alle eigenen Regeln vergessen?',
  zuruecksetzenErklaerung:
    'Danach gelten wieder nur die Regeln, die das Spiel mitbringt. Eigene kommen nicht zurück.',
  zuruecksetzenJa: 'Eigene vergessen',
  ausDemPaket: 'aus dem Spielpaket',
}
