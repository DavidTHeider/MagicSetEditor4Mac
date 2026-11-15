// Die Texte des **Zufallspacks** — die Lasche, die einen Booster zieht.
//
// Aus `seitenkarten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs
// (§2) — der Grund für die Zeile war der Ladehinweis aus Werkstück C von
// Phase 4.10. Der Schnitt folgt der Frage: dort stehen die Texte der
// Kartenseite, hier die einer ihrer Laschen.

export const pack = {
  ueberschrift: 'Zufallspack',
  erklaerung:
    'Ein Booster aus den Karten dieses Sets, nach den Regeln des Spielpakets. Derselbe Same gibt dieselbe Ziehung — so lässt sich ein Ergebnis wiederherstellen.',
  // **Warum das Laden dauert** — Werkstück C von Phase 4.10
  // (#1325).
  // Gemessen am 2026-10-19 mit *Blue* (3.693 Karten, Mac17,3): der erste
  // Aufruf braucht 16,6 Sekunden (vorher rund 100, #2466), jeder weitere
  // Millisekunden. Sekunden „Wird geladen …“ ohne Grund sehen aus wie ein
  // Fehler (§8).
  laedtGrund:
    'Jede Packart fragt jede Karte, ob sie hineinpasst — das rechnet beim ersten Mal einen Augenblick und steht danach bereit.',
  art: 'Packart',
  same: 'Same',

  // **Was ein Same sein darf** — Werkstück U3 von Phase 5.1
  // (#1414).
  //
  // Die Regel steht in `app.samenAus`: `strconv.ParseUint(text, 10, 64)`.
  // Also Ziffern und höchstens 2⁶⁴−1 — die Zahl steht ausgeschrieben da,
  // weil „2 hoch 64 minus eins“ niemandem sagt, was er tippen darf. Ein
  // **leeres** Feld ist gültig und heißt null.
  sameZahl: 'Der Same ist eine ganze Zahl von 0 bis 18446744073709551615.',
  ziehen: 'Ziehen',
  neuerSame: 'Neuer Same',
  gezogen: 'Gezogen',
  keineArten:
    'Das Spiel dieses Sets bringt keine Packarten mit — oder sein Paket liegt nicht vor.',
  keineKarten: 'Aus dieser Packart kam keine Karte.',
  artMitZahl: '{name} ({zahl} Karten)',
  nurBaustein: 'Baustein — wird von anderen Packarten gezogen',
  vorbehalte: 'Packarten, deren Filter sich nicht ausrechnen ließ',
}
