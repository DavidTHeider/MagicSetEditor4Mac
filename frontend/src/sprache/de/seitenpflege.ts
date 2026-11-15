// Die Texte der **Datenpflege** — Werkstück O8 von Phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Aufräumen',
  hinweis:
    'Ein Bestand sammelt mit der Zeit Zeilen an, die auf nichts mehr zeigen. ' +
    'Hier steht, was das ist — und was davon sich gefahrlos wegräumen lässt.',

  nachsehen: 'Nachsehen',
  nachsehenLaeuft: 'Sieht nach …',
  sauber: 'Nichts gefunden. Der Bestand ist in Ordnung.',
  gueltigbis: 'Diese Übersicht gilt bis {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Feldwerte ohne ihre Karte',
    setangabenOhneSet: 'Setangaben ohne ihr Set',
    suchzeilenOhneKarte: 'Suchzeilen ohne ihre Karte',
    deckplaetzeOhneDeck: 'Deckplätze ohne ihr Deck',
    zuordnungenOhneExemplar: 'Deckzuordnungen ohne ihr Exemplar',
    verlaufOhneSet: 'Verlaufseinträge zu Sets, die es nicht mehr gibt',
    abgelaufeneNachweise: 'Abgelaufene Freigabenachweise',
    abgelaufeneToken: 'Abgelaufene Zugriffstoken',
    moeglicheDubletten: 'Sieht nach Dubletten aus',
    offeneRueckkehren: 'Offene Rückkehren',
  } as Record<string, string>,

  klassen: {
    sicher: 'Lässt sich gefahrlos wegräumen',
    entscheidung: 'Du entscheidest',
    nurInformieren: 'Wird nicht angefasst',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'Diese Zeilen zeigen auf nichts und halten nichts fest. Sie sind ' +
      'vorausgewählt.',
    entscheidung:
      'Das sieht nach einem Rest aus — und kann ein Nachweis sein. Sieh dir ' +
      'die Beispiele an, bevor du es wählst.',
    nurInformieren:
      'Das wird nicht weggeräumt. Gleiche Kartennummern sind keine ' +
      'Dubletten — Zustand, Sprache, Ausführung, Erwerb und Lagerort ' +
      'unterscheiden sie —, und eine offene Rückkehr ist ein Nachweis.',
  } as Record<string, string>,

  // **„Stellen“ und nicht „Zeilen“** — das fand die Sichtprüfung: bei den
  // möglichen Dubletten zählt die Zahl **Kartennummern**, die mehrfach
  // vorkommen, und nicht Zeilen. Ein Wort, das für beides stimmt, ist besser
  // als eines, das für die Hälfte falsch ist.
  anzahl: '{anzahl} Stellen im Bestand',
  beispieleTitel: 'Zum Beispiel',

  aufraeumen: 'Gewähltes aufräumen',
  aufraeumenLaeuft: 'Räumt auf …',
  aufraeumenHinweis:
    'Vorher wird der Bestand gesichert. Danach wird noch einmal nachgesehen.',

  bilanzTitel: 'Was geschah',
  gesichert: 'Gesichert nach {ort}',

  ausgaenge: {
    erledigt: 'Weggeräumt',
    konflikt: 'Hat sich seit der Übersicht geändert — nicht angefasst',
    schonWeg: 'War schon weg',
    uebersprungen: 'Übersprungen',
    gescheitert: 'Ging nicht',
    bereit: 'Bereit',
  } as Record<string, string>,

  bilanzZeile: '{art}: {ausgang} ({anzahl})',
  bilanzStaende: 'Vorher {vorher} Befunde, nachher {nachher}.',
}
