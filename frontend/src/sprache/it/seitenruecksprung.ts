// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: 'Tornare indietro',
  hinweis:
    'Quattro cose possono tornare indietro, e costano in modo diverso. Due di ' +
    'esse questa applicazione non può farle da sola — allora qui è indicato chi lo fa.',

  arten: {
    appfassung: 'L’applicazione',
    schema: 'La struttura della libreria',
    daten: 'I dati',
    pakete: 'I pacchetti di modelli',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'Una persona sostituisce il bundle nel Finder quando questa applicazione ' +
      'non è più in esecuzione. Un programma che si sostituisce da solo mentre ' +
      'è in esecuzione avrebbe bisogno di un secondo processo con permesso di ' +
      'scrittura su «Applicazioni».',
    nurVorwaerts:
      'Non esiste una via all’indietro attraverso le migrazioni. La via del ' +
      'ritorno è un backup: ripristina lo stato di allora.',
    unbekannteArt: 'L’applicazione non conosce questo tipo.',
  } as Record<string, string>,

  gehtHier: 'Questa applicazione può farlo da sola.',

  vergleichTitel: 'Cosa costa un ripristino',
  datenstand: 'Stato ripristinato',
  spalteGegenstand: 'Elemento',
  spalteSicherung: 'Nel backup',
  spalteJetzt: 'Ora',
  spalteVerlust: 'Andrà perso',

  gegenstaende: {
    sets: 'Set',
    karten: 'Carte',
    decks: 'Mazzi',
    exemplare: 'Esemplari',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'Ciò che qui andrebbe perso puoi esportarlo prima: un set tramite «Salva con nome», la collezione tramite «Esporta».',

  unklarTitel: 'Incerto dopo',
  unklarHinweis:
    'Questi esemplari torneranno nella tua collezione dopo il ripristino. ' +
    'Nessuno sa se li hai ancora — nel frattempo potrebbero essere stati ' +
    'ceduti. Restano fuori dagli scambi finché non hai controllato; non ' +
    'viene detratto nulla.',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'Il confronto tra le collezioni non era disponibile — questo backup o ' +
    'la tua libreria non lo fornisce. Dopo il ripristino controlla la tua ' +
    'collezione a mano.',

  rueckkehrenTitel: 'Ripristini registrati',
  keineRueckkehren:
    'Ancora nessuno. Non appena ripristini un backup, compare qui — con lo ' +
    'stato a cui è tornato e ciò che ha lasciato incerto.',
  rueckkehrZeile: 'Il {zurueckgespielt} tornato allo stato del {datenstand}',
  ausDatei: 'da {quelle}',
  offen: 'Aperto',
  quittiert: 'Controllato il {geklaertAm}',
  betrifft: 'riguarda {menge} esemplari',
  quittieren: 'Controllato',
  quittierenLaeuft: 'Registrazione in corso …',
  quittierenHinweis:
    'Con questo dichiari di aver controllato a mano gli esemplari ' +
    'interessati. Non detrae nulla.',
}
