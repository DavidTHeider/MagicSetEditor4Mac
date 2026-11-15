// **Die Wörter des Zuges** — Werkstück G9 von Phase 5.5
// (#1592),
// herausgelöst mit PLAY-05.AK-03.
//
// **Ein Katalog, fünf Dateien** (§9): `seitenspielen.ts` ging mit dem fünften
// Baustein über die Zeilengrenze aus §2. Der Schnitt folgt der Frage —
// Zugstruktur (CR 500) steht hier, der Tisch daneben.
export const zugworte = {
  // **Wer am Zug ist** (G9, #1592).
  zugLeiste: 'Der Zug',
  zug: 'Zug',
  zugAktiv: 'Am Zug',
  zugPassen: 'Passen',
  zuruecknehmenKnopf: 'Letzten Zug zurücknehmen',

  // **Die Abkürzung** — CR 732: sie reicht durch, bis etwas geschieht.
  zugAbkuerzen: 'Durchreichen',

  // **Den Zug vorzeitig beenden** — CR 724.
  zugBeenden: 'Zug beenden',
  zugDuBistDran: 'Du bist dran.',
  // **Mitten im Beginn** — CR 103.8: der erste Zug kommt erst danach.
  zugWartetAufBeginn: 'Der Zug wartet, bis jeder seine Starthand behalten hat.',
  zugWartetAuf: 'Es wartet auf',
  zugNiemandDran: 'Hier hat niemand Priorität — es geht von selbst weiter.',
  zugAbwurf: 'Wirf noch ab',
  phaseAnfang: 'Anfangsphase',
  phaseHauptVorKampf: 'Erste Hauptphase',
  phaseKampf: 'Kampfphase',
  phaseHauptNachKampf: 'Zweite Hauptphase',
  phaseEnde: 'Endphase',
  schrittEnttappen: 'Enttappen',
  schrittVersorgung: 'Versorgung',
  schrittZiehen: 'Ziehen',
  schrittKampfbeginn: 'Kampfbeginn',
  schrittAngreifer: 'Angreifer',
  schrittBlocker: 'Blocker',
  schrittKampfschaden: 'Kampfschaden',
  schrittKampfende: 'Kampfende',
  schrittEnde: 'Endschritt',
  schrittAufraeumen: 'Aufräumen',

  // **Der Kampf in Zeilen** — Artboard *Tisch*, CR 508.1 und CR 509.1.
  kampfTitel: 'Kampf',
  kampfGegen: '{wer} → {gegen}',
  kampfUngeblockt: '{zahl} ungeblockt → {ziel}',
  kampfWeitereUngeblockt: '{zahl} weitere ungeblockt → {ziel}',

  // **Aufgeben** — CR 104.3a; seit P1b (#2370, C-06) mit Rückfrage. Hier,
  // weil es ein Befehl des Zugs ist (`zug.aufgeben`), und weil
  // `seitenspielen.ts` an der Zeilengrenze aus §2 stand.
  aufgeben: 'Aufgeben',
  // **Aufgeben fragt nach** — P1b (#2370), C-06; CR 104.3a.
  aufgebenFrage:
    'Aufgeben? Das Spiel endet sofort für diesen Sitz, und seine Karten gehen mit.',
  aufgebenJa: 'Ja, aufgeben',
  aufgebenNein: 'Weiterspielen',
  aufgebenHinweis:
    'Aufgeben geht jederzeit, und es ist sofort endgültig. Du verlässt die Partie, und deine Karten gehen mit.',
}
