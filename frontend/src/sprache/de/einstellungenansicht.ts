// **Wie eine Liste aussieht** — UX-11.AK-04
// (#1683).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungengestaltung.ts`.
// **„Als Kacheln" und nicht „Kacheln"**: das bloße Wort steht im selben
// Verbund schon für das Fensterbild, und dort heißt es etwas anderes
// (`sammelbegriffe.test.ts`).

// ansichtsnamen sind die Beschriftungen der drei Einstellungen.
export const ansichtsnamen = {
  'arbeitsplatz.ansicht.sets': 'Ansicht der Sets',
  'arbeitsplatz.ansicht.decks': 'Ansicht der Decks',
  'arbeitsplatz.ansicht.sammlung': 'Ansicht der Sammlung',
  'ausgabe.ziele': 'Zuletzt gezeigtes Ausgabeziel je Set',
}

// ansichtswerte sind die Ansichten je Bereich — und „wie beim Nutzer" für
// ein Profil, das nichts Eigenes will.
export const ansichtswerte = {
  'arbeitsplatz.ansicht.sets.erben': 'Wie beim Nutzer',
  'arbeitsplatz.ansicht.sets.kacheln': 'Als Kacheln',
  'arbeitsplatz.ansicht.sets.liste': 'Als Liste',
  'arbeitsplatz.ansicht.decks.erben': 'Wie beim Nutzer',
  'arbeitsplatz.ansicht.decks.liste': 'Als Liste',
  'arbeitsplatz.ansicht.decks.kacheln': 'Als Kacheln',
  'arbeitsplatz.ansicht.sammlung.erben': 'Wie beim Nutzer',
  'arbeitsplatz.ansicht.sammlung.tabelle': 'Als Tabelle',
  'arbeitsplatz.ansicht.sammlung.galerie': 'Als Galerie',
  'arbeitsplatz.ansicht.sammlung.gruppen': 'In Gruppen',
}
