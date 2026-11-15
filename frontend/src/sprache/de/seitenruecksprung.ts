// Die Texte der **Rückkehr** — Werkstück O7 von Phase 5.2
// (#1431).
//
// Eine eigene Datei, wie `seitenbestand.ts` daneben: `seiten.ts` läge mit
// ihnen über der Zeilengrenze aus §2. Der Katalog bleibt **einer**.

export const ruecksprung = {
  titel: 'Zurückgehen',
  hinweis:
    'Vier Dinge können zurück, und sie kosten Verschiedenes. Zwei davon kann ' +
    'diese Anwendung nicht selbst — dann steht hier, wer es tut.',

  arten: {
    appfassung: 'Die Anwendung',
    schema: 'Der Aufbau des Bestands',
    daten: 'Die Daten',
    pakete: 'Die Vorlagenpakete',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'Ein Mensch tauscht das Bündel im Finder, wenn diese Anwendung nicht ' +
      'mehr läuft. Ein Programm, das sich selbst ersetzt, während es läuft, ' +
      'bräuchte einen zweiten Prozess mit Schreibrecht auf „Programme“.',
    nurVorwaerts:
      'Es gibt keinen Weg rückwärts durch die Migrationen. Der Weg zurück ' +
      'ist eine Sicherung: spiele den Stand von damals zurück.',
    unbekannteArt: 'Diese Art kennt die Anwendung nicht.',
  } as Record<string, string>,

  gehtHier: 'Das kann diese Anwendung selbst.',

  vergleichTitel: 'Was ein Zurückspielen kostet',
  datenstand: 'Wiederhergestellter Stand',
  spalteGegenstand: 'Gegenstand',
  spalteSicherung: 'In der Sicherung',
  spalteJetzt: 'Jetzt',
  spalteVerlust: 'Geht verloren',

  gegenstaende: {
    sets: 'Sets',
    karten: 'Karten',
    decks: 'Decks',
    exemplare: 'Exemplare',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'Was hier verloren geht, kannst du vorher ausfahren: ein Set über „Speichern unter“, die Sammlung über „Ausgeben“.',

  unklarTitel: 'Danach unklar',
  unklarHinweis:
    'Diese Stücke stehen nach dem Zurückspielen wieder in deiner Sammlung. ' +
    'Ob du sie noch hast, weiß niemand — sie können seither abgegeben ' +
    'worden sein. Sie bleiben aus dem Handel, bis du nachgesehen hast; ' +
    'abgezogen wird nichts.',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'Der Vergleich der Sammlungen war nicht zu bekommen — diese Sicherung ' +
    'oder dein Bestand gibt sie nicht her. Sieh deine Sammlung nach dem ' +
    'Zurückspielen von Hand durch.',

  rueckkehrenTitel: 'Verbuchte Rückkehren',
  keineRueckkehren:
    'Noch keine. Sobald du eine Sicherung zurückspielst, steht sie hier — ' +
    'mit dem Stand, auf den sie zurückging, und dem, was sie unklar ließ.',
  rueckkehrZeile: 'Am {zurueckgespielt} zurück auf den Stand vom {datenstand}',
  ausDatei: 'aus {quelle}',
  offen: 'Offen',
  quittiert: 'Am {geklaertAm} nachgesehen',
  betrifft: 'betrifft {menge} Stück',
  quittieren: 'Nachgesehen',
  quittierenLaeuft: 'Wird eingetragen …',
  quittierenHinweis:
    'Damit sagst du, dass du die betroffenen Stücke von Hand geprüft hast. ' +
    'Es bucht nichts ab.',
}
