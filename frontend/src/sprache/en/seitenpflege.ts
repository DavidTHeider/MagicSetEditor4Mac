// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Tidying up',
  hinweis:
    'Over time a library collects rows that no longer point at anything. ' +
    'Here is what that is — and which of it can safely be cleared away.',

  nachsehen: 'Take a look',
  nachsehenLaeuft: 'Looking …',
  sauber: 'Nothing found. The library is in order.',
  gueltigbis: 'This overview is valid until {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Field values without their card',
    setangabenOhneSet: 'Set details without their set',
    suchzeilenOhneKarte: 'Search rows without their card',
    deckplaetzeOhneDeck: 'Deck slots without their deck',
    zuordnungenOhneExemplar: 'Deck assignments without their copy',
    verlaufOhneSet: 'History entries for sets that no longer exist',
    abgelaufeneNachweise: 'Expired approval records',
    abgelaufeneToken: 'Expired access tokens',
    moeglicheDubletten: 'Looks like duplicates',
    offeneRueckkehren: 'Open restores',
  } as Record<string, string>,

  klassen: {
    sicher: 'Can safely be cleared away',
    entscheidung: 'You decide',
    nurInformieren: 'Will not be touched',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'These rows point at nothing and hold nothing. They are preselected.',
    entscheidung:
      'This looks like a leftover — and may be a record. Look at the ' +
      'examples before you choose it.',
    nurInformieren:
      'This is not cleared away. Identical card numbers are not duplicates ' +
      '— condition, language, finish, acquisition and storage place tell ' +
      'them apart — and an open restore is a record.',
  } as Record<string, string>,

  anzahl: '{anzahl} places in the library',
  beispieleTitel: 'For example',

  aufraeumen: 'Clear away the selected',
  aufraeumenLaeuft: 'Clearing …',
  aufraeumenHinweis:
    'The library is backed up first. Afterwards it is checked again.',

  bilanzTitel: 'What happened',
  gesichert: 'Backed up to {ort}',

  ausgaenge: {
    erledigt: 'Cleared away',
    konflikt: 'Changed since the overview — not touched',
    schonWeg: 'Was already gone',
    uebersprungen: 'Skipped',
    gescheitert: 'Did not work',
    bereit: 'Ready',
  } as Record<string, string>,

  bilanzZeile: '{art}: {ausgang} ({anzahl})',
  bilanzStaende: 'Before {vorher} findings, afterwards {nachher}.',
}
