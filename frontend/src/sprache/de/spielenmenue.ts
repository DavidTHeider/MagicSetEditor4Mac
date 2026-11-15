// **Die Wörter des Kartenmenüs am Tisch** — PLAY-04.AK-01
// (#1712).
//
// **Ein Katalog, drei Dateien** (§9): `seitenspielen.ts` steht an der
// Zeilengrenze aus §2, `spielenpartien.ts` trägt die Liste davor, und hier
// steht das Menü an der Karte.
export const menueworte = {
  // **Der Name des Menüs nennt die Karte** — WCAG 2.4.6: „Menü“ allein sagt
  // bei zwanzig Karten nicht, welches.
  kartenmenue: 'Handlungen an {name}',

  // **Der Weg zur Leiste**, für alles, was ein Feld braucht: eine Zone, eine
  // Zahl, ein zweites Objekt. Ein Menüeintrag ist eine Handlung und kein
  // Formular.
  kartenmenueWeitere: 'Weitere Handlungen …',
  // **Auf den Friedhof über die gedeutete Geste** — P4 (#2372): was dabei
  // geschieht, entscheidet das Profil der Partie (`objekt.ziehen`).
  kartenmenueFriedhof: 'Auf den Friedhof',
  // **Leben am Sitz** — P4 (#2372), C-20: −/+, mit ⇧ fünf, und den Wert
  // setzen. Gesendet wird immer eine Änderung (`spieler.leben`).
  lebenMinus: '−',
  lebenPlus: '+',
  lebenWeniger: 'Ein Leben weniger — mit ⇧ fünf',
  lebenMehr: 'Ein Leben mehr — mit ⇧ fünf',
  lebenSetzen: 'Leben {stand} — Wert setzen',
  lebenSetzenTitel: 'Leben setzen',
  lebenNeu: 'Neuer Stand',
  lebenUebernehmen: 'Setzen',
  wuenscheWarten: 'Wartet auf den Tisch: {zahl}',
  // **Ein Wunsch, der nicht hinausging, wird gesagt** — P5 (#2373).
  wunschNichtGestellt:
    'Gerade lief ein anderer Vorgang am Tisch — dieser Zug ging nicht hinaus. Bitte noch einmal.',
  // **Die Sitze stehen schon da** — P6 (#2374): der Name am zweiten Sitz,
  // wenn es noch keine Partie gab, der Knopf, solange angelegt wird, und der
  // Anfang mit der Vorbelegung aus einem Deck heraus.
  sitzZweiVorgabe: 'Spieler 2',
  anfangenLaeuft: 'Wird angelegt …',
  schnellSpielen: 'Schnell spielen',
}
