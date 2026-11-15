// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Pulizia',
  hinweis:
    'Con il tempo un archivio accumula righe che non puntano più a nulla. ' +
    'Qui trovi che cosa sono — e quali di queste si possono eliminare senza rischi.',

  nachsehen: 'Controlla',
  nachsehenLaeuft: 'Controllo in corso …',
  sauber: 'Non è stato trovato nulla. L’archivio è in ordine.',
  gueltigbis: 'Questa panoramica è valida fino a {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Valori dei campi senza la loro carta',
    setangabenOhneSet: 'Dati del set senza il loro set',
    suchzeilenOhneKarte: 'Righe di ricerca senza la loro carta',
    deckplaetzeOhneDeck: 'Posizioni di mazzo senza il loro mazzo',
    zuordnungenOhneExemplar: 'Assegnazioni ai mazzi senza la loro copia',
    verlaufOhneSet: 'Voci della cronologia di set che non esistono più',
    abgelaufeneNachweise: 'Prove di approvazione scadute',
    abgelaufeneToken: 'Token di accesso scaduti',
    moeglicheDubletten: 'Sembrano doppioni',
    offeneRueckkehren: 'Ripristini aperti',
  } as Record<string, string>,

  klassen: {
    sicher: 'Si può eliminare senza rischi',
    entscheidung: 'Decidi tu',
    nurInformieren: 'Non verrà toccato',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'Queste righe non puntano a nulla e non contengono nulla. Sono ' +
      'preselezionate.',
    entscheidung:
      'Sembra un avanzo — e potrebbe essere una registrazione. Guarda gli ' +
      'esempi prima di sceglierlo.',
    nurInformieren:
      'Questo non viene eliminato. Numeri di carta identici non sono ' +
      'doppioni — condizione, lingua, finitura, acquisizione e ubicazione li ' +
      'distinguono — e un ripristino aperto è una registrazione.',
  } as Record<string, string>,

  anzahl: '{anzahl} punti nell’archivio',
  beispieleTitel: 'Per esempio',

  aufraeumen: 'Elimina la selezione',
  aufraeumenLaeuft: 'Pulizia in corso …',
  aufraeumenHinweis:
    'Prima viene eseguito un backup dell’archivio. Poi si controlla di nuovo.',

  bilanzTitel: 'Che cosa è successo',
  gesichert: 'Backup eseguito in {ort}',

  ausgaenge: {
    erledigt: 'Eliminato',
    konflikt: 'Modificato dopo la panoramica — non toccato',
    schonWeg: 'Era già sparito',
    uebersprungen: 'Saltato',
    gescheitert: 'Non ha funzionato',
    bereit: 'Pronto',
  } as Record<string, string>,

  bilanzZeile: '{art}: {ausgang} ({anzahl})',
  bilanzStaende: 'Prima: {vorher} rilevamenti, dopo: {nachher}.',
}
