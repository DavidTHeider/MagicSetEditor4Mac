// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: 'Going back',
  hinweis:
    'Four things can go back, and they cost different amounts. Two of them ' +
    'this application cannot do itself — then it says here who does.',

  arten: {
    appfassung: 'The application',
    schema: 'The layout of the library',
    daten: 'The data',
    pakete: 'The template packages',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'A person swaps the bundle in the Finder once this application is no ' +
      'longer running. A program that replaces itself while running would ' +
      'need a second process with write access to “Applications”.',
    nurVorwaerts:
      'There is no way backwards through the migrations. The way back is a ' +
      'backup: restore the state from back then.',
    unbekannteArt: 'The application does not know this kind.',
  } as Record<string, string>,

  gehtHier: 'This application can do it itself.',

  vergleichTitel: 'What restoring costs',
  datenstand: 'Restored state',
  spalteGegenstand: 'Item',
  spalteSicherung: 'In the backup',
  spalteJetzt: 'Now',
  spalteVerlust: 'Will be lost',

  gegenstaende: {
    sets: 'Sets',
    karten: 'Cards',
    decks: 'Decks',
    exemplare: 'Copies',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'What is lost here you can export first: a set via “Save as”, the collection via “Export”.',

  unklarTitel: 'Unclear afterwards',
  unklarHinweis:
    'These copies will be back in your collection after restoring. Nobody ' +
    'knows whether you still have them — they may have been given away ' +
    'since. They stay out of trading until you have checked; nothing is ' +
    'deducted.',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'The comparison of the collections was not available — this backup or ' +
    'your library does not provide it. Check your collection by hand after ' +
    'restoring.',

  rueckkehrenTitel: 'Recorded restores',
  keineRueckkehren:
    'None yet. As soon as you restore a backup it appears here — with the ' +
    'state it went back to and what it left unclear.',
  rueckkehrZeile: 'On {zurueckgespielt} back to the state of {datenstand}',
  ausDatei: 'from {quelle}',
  offen: 'Open',
  quittiert: 'Checked on {geklaertAm}',
  betrifft: 'affects {menge} copies',
  quittieren: 'Checked',
  quittierenLaeuft: 'Recording …',
  quittierenHinweis:
    'This says you have checked the affected copies by hand. It does not ' +
    'deduct anything.',
}
