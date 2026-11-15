// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'Collezione',
  erklaerung:
    'Ciò che possiedi davvero — una riga per ogni stampa, condizione, lingua e finitura. Non ha nulla a che fare con i tuoi set: una carta progettata è una bozza, una copia è un pezzo di cartoncino.',

  aufnehmenOeffnen: 'Aggiungi …',
  einfuhrOeffnen: 'Importa ed esporta …',
  einfuhrTitel: 'Importa ed esporta',
  leerTitel: 'Ancora nulla registrato',
  leer: 'Una copia richiede l’ID Scryfall della sua stampa — l’UUID preso dall’indirizzo della pagina di una carta. Un’intera lista si importa in un colpo solo.',

  druck: 'Stampa',
  menge: 'Quantità',
  zustand: 'Condizione',
  sprache: 'Lingua',
  ausfuehrung: 'Finitura',
  foil: 'Foil',
  nonfoil: 'Nonfoil',
  etched: 'Etched',
  erwerbsdatum: 'Data di acquisto',
  erwerbspreis: 'Pagato',
  erwerbswaehrung: 'Valuta',
  preisMehrdeutig:
    'Due interpretazioni: {erste} oppure {zweite}. Scrivilo in modo univoco.',
  notiz: 'Nota',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Nome della carta',
  setcode: 'Set',
  nummer: 'Numero',
  sicht: 'Vista della collezione',
  suche: 'Cerca',
  suchePlatzhalter: 'nome, set, ID o nota',
  sortierung: 'Ordina per',
  sortierBezeichnung: 'Carta',
  sortierMenge: 'Quantità',
  sortierErwerbsdatum: 'Data di acquisto',
  ausfuehrungAlle: 'Tutte le finiture',
  nurFoil: 'Solo foil',
  nurNonfoil: 'Solo non-foil',
  nurEtched: 'Solo etched',

  navigator: 'Vista e strumenti',
  inspector: 'Copia',
  exemplarmenue: 'Esemplare',
  inDecks: 'In questi mazzi',

  inspectorLeer:
    'Scegli una copia nell’elenco — qui compaiono quantità, condizione, acquisto e prezzo, ed è qui che la modifichi.',
  inspectorAusserhalb:
    'Questa copia non è più nella vista attuale. Una modifica riguarda comunque proprio questa copia.',
  arbeitsplatz: 'Area di lavoro della collezione',
  listeBezeichnung: 'Copie',
  spalteWahl: 'Scelta',
  spalteBezeichnung: 'Nome',
  spalteMenge: 'Quantità',
  spalteZustand: 'Condizione',
  spalteSprache: 'Lingua',
  spalteSet: 'Set',
  spalteLagerort: 'Ubicazione',
  gruppeOhne: 'Non specificato',
  gruppeSumme: '{anzahl} copie · {menge} pezzi',
  kachelWaehlen: 'Seleziona {bezeichnung}',

  ansicht: 'Vista',
  ansichtTabelle: 'Tabella',
  ansichtGalerie: 'Galleria',
  ansichtGruppen: 'Gruppi',
  gruppierung: 'Raggruppa per',
  gruppierungKeine: 'Nessun raggruppamento',
  gruppierungSet: 'Set',
  gruppierungLagerort: 'Ubicazione',
  gruppierungZustand: 'Condizione',
  gruppierungSprache: 'Lingua',
  ohneTreffer: 'Nessuna copia corrisponde a ricerca e filtro.',
  ohneTrefferZuruecksetzen: 'Reimposta ricerca e filtro',
  gezeigt: '{gezeigt} di {gesamt}',
  gekappt: '{gezeichnet} di {passend} disegnate',
  mehrZeigen: 'Mostra altro',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Incolla un elenco',
  listeErklaerung:
    'Una carta per riga, ad esempio “4 Lightning Bolt (M11) 149”. Righe vuote, “Deck”, “Sideboard” e righe che iniziano con # o // vengono saltate.',
  listeFeld: 'Elenco di carte',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Abbina',
  listeEinlesen: 'Aggiungi',
  listeOffen:
    '{offen} righe su {gesamt} richiedono una decisione — scegli la stampa che possiedi.',
  listeAlleEindeutig: 'Tutte le {gesamt} righe sono univoche.',
  listeDruckWahl: 'Quale stampa?',
  listeOhneWahl: '— non ancora deciso —',
  listeDrucke: '{drucke} stampe',
  listeBericht:
    '{angelegt} nuove, {erhoeht} aumentate, {stueck} copie aggiunte.',
  listeNichtsGewaehlt: 'Nessuna riga decisa — scegli almeno una stampa.',
  listeAusgelassen: '{ausgelassen} righe esistenti saltate.',
  listeKeinDruck: 'Nessuna carta si chiama così: controlla l’ortografia.',
  listeZuViele:
    'Più di {drucke} stampe si chiamano così: indica il set tra parentesi, ad esempio «(M11)», e ne resta una.',
  listeOhneKarte: '{anzahl} righe non trovano alcuna carta e vengono saltate.',
  listeOhneKarteEins: 'Una riga non trova alcuna carta e viene saltata.',
  vorschauTitel: 'Prima dell’importazione',
  vorschauErklaerung:
    'Alcune righe corrispondono a righe che la tua collezione ha già — stessa stampa, condizione, lingua, finitura e scopo. Scegli che cosa ne fai; fino ad allora non viene scritto nulla.',
  vorschauZaehlung:
    '{neu} nuove · {zusammen} corrispondono a righe esistenti · {wiederholt} ripetono una riga di questa importazione',
  vorschauZusammen: '{bezeichnung}, {zustand}: {vorher} → {nachher}',
  vorschauModus: 'Righe esistenti',
  vorschauAddieren: 'Somma le quantità',
  vorschauAuslassen: 'Salta le righe esistenti',
  vorschauEinlesen: 'Importa',
  vorschauAbbrechen: 'Non importare',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Esporta e salva',
  ausfuhrErklaerung:
    '“CSV” contiene ogni dato e si può rileggere per intero. “Elenco” è la forma incollabile per un deck builder — contiene solo quantità, carta, stampa e finitura.',
  ausfuhrAlsCSV: 'Come CSV',
  ausfuhrAlsListe: 'Come elenco',
  ausfuhrErgebnis: 'Risultato',
  ausfuhrZurueck: 'Rileggi il CSV',
  ausfuhrLeer: 'Ancora nulla esportato.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Ubicazione',
  wunsch: 'Desiderata, non posseduta',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Ordine',
  ordnungLagerorte: 'Ubicazioni',
  ordnungDubletten: 'Doppioni',
  ordnungWuensche: 'Lista dei desideri',
  ordnungBehalten: 'Tieni',
  ordnungLeerLagerorte:
    'Ancora nessun esemplare: senza esemplare non c’è collocazione.',
  ordnungLeerDubletten: 'Nessuna stampa è presente più di {behalten} volte.',
  ordnungLeerWuensche:
    'Niente nella lista dei desideri. Un esemplare ci arriva come «Desiderata, non posseduta», oppure da un mazzo con «Aggiungi alla lista dei desideri ciò che manca».',
  ordnungStueck: '{stueck} copie in {zeilen} righe',
  ordnungOhneOrt: 'non specificata',
  ordnungUeberzaehlig: '{ueberzaehlig} in eccesso',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'Ne possiedi {stueck}.',
  besitzNein: 'Non possiedi questa stampa.',
  besitzGewuenscht: '({gewuenscht} nella lista dei desideri)',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Azioni di massa',
  masseWaehlen: 'Seleziona',
  masseGewaehlt: '{gewaehlt} di {gesamt} selezionate',
  masseLagerortSetzen: 'Imposta l’ubicazione su',
  masseZustandSetzen: 'Imposta la condizione su',
  masseAendern: 'Modifica la selezione',
  masseLoeschen: 'Elimina la selezione',
  masseLoeschenFrage: 'Eliminare {anzahl} copie?',
  masseLoeschenErklaerung:
    'Ciò che sparisce qui lo riscrivi a mano: quantità, condizione, ubicazione e prezzo d’acquisto non sono da nessun’altra parte. La collezione non ha una cronologia come un set.',
  masseLoeschenJa: 'Elimina',
  masseBericht: '{geaendert} modificate.',
  masseZuruecknehmen: 'Annulla',
  masseZuruecknehmenVon: 'Annulla l’ubicazione di {anzahl} righe',
  masseZurueckgenommen: '{anzahl} annullate.',
  massePreise: 'Recupera i prezzi per tutte',
  massePreisschub: '{geholt} recuperati, {offen} rimanenti',
  massePreiseFehlgeschlagen: '{anzahl} falliti',
  massePreiseSteht:
    'Questo blocco non ha spostato nulla — un altro tentativo darebbe lo stesso risultato.',
  massePreiseFertig: 'Tutti i prezzi sono arrivati.',

  aufnehmen: 'Aggiungi',
  aenderungSichern: 'Salva',
  loeschen: 'Elimina',
  loeschenFrage: 'Eliminare questa copia?',
  loeschenErklaerung:
    'Quantità, condizione, ubicazione e prezzo d’acquisto non sono da nessun’altra parte — la collezione non ha una cronologia come un set.',
  loeschenJa: 'Elimina',
  abbrechen: 'Annulla',
  bearbeiten: 'Modifica',
  zumTausch: 'Allo scambio',
  insDeck: 'Nel mazzo',
  auswahlTitel: 'Selezione',
  auswahlInsDeck: '{anzahl} → mazzo',

  neuesTitel: 'Aggiungi una copia',
  weitereAngaben: 'Altri dati',

  dubletteFrage:
    '{bezeichnung} è già nella collezione — {menge} nella stessa condizione, lingua e finitura.',
  dubletteErklaerung:
    'Aumentare aggiunge la nuova quantità; gli altri dati della riga esistente restano. Annulla non cambia nulla — quello che hai inserito rimane.',
  dubletteErhoehen: 'Aumenta a {summe}',
  dubletteZurZeile: 'Vai alla riga esistente',
  dubletteAbbrechen: 'Annulla',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  preiseTitel: 'Prezzi',
  preiseHolen: 'Recupera i prezzi',
  marktUndVerlauf: 'Offerte e storico',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'Offerte sul marketplace',
  marktErklaerung:
    'Ciò che i singoli venditori chiedono per questa stampa — prezzo, quantità, condizione e lingua. La ricerca passa dalla rete. Non viene acquistato nulla.',
  marktHolen: 'Recupera le offerte',
  marktHolenBei: 'Recupera le offerte da {quelle}',
  marktLaeuft: 'Recupero in corso …',
  marktLeer: 'Per questa stampa lì al momento non c’è nulla in vendita.',
  marktZugeordnet:
    'Questa stampa ora è collegata alla versione di carta {fassung} — la prossima ricerca la troverà subito.',
  marktStueck: '{menge} disponibili',
  marktBuendel: 'in lotti da {buendel}',
  marktImUrlaub: 'al momento non spedisce',
  marktGradiert: 'gradata',
  marktSigniert: 'firmata',
  marktVeraendert: 'alterata',
  marktBefunde: 'Non è stato possibile leggere ogni offerta:',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde:
    'Non è stato possibile leggere ogni prezzo. Mancano queste serie:',
  preiseLaufen: 'Recupero in corso …',
  keinePreise: 'Per questa stampa non è registrato alcun prezzo.',
  preisAlter: '{stunden} h fa · {quelle}',
  preisFrisch: 'proprio ora · {quelle}',
  preisVeraltet: 'obsoleto, di {tage} giorni · {quelle}',
  preisQuellstand: 'dati della fonte del {tag}',

  vorbehalt:
    'I prezzi sono stime di Scryfall e dei suoi partner. Per nessuno di essi viene data alcuna garanzia; fa fede il prezzo in negozio.',

  verlauf: 'Cronologia',
  keinVerlauf:
    'Per questa serie esiste una sola rilevazione — una cronologia si forma quando recuperi di nuovo i prezzi.',

  bulkTitel: 'Recupera i prezzi di un’intera fonte',
  leiste: {
    bezeichnung: 'Cifre della collezione',
    exemplare: 'Esemplari',
    eintraege: '{zahl} voci',
    wert: 'Valore',
    keinPreis: 'nessun prezzo',
    ohnePreis: 'Senza prezzo',
    nichtGelesen: 'non letto',
  },
  drucksuche: {
    karte: 'Carta',
    hinweis:
      'Nome, eventualmente con set e numero — es. Lightning Bolt (M11) 149',
    suchen: 'Cerca',
    drucke: 'Stampe',
    keinTreffer: 'Nessuna carta con questo nome.',
    weitere: 'Ci sono altre stampe — set e numero restringono la scelta.',
    treffer: '{name} · {set} ({code}) · N. {nummer}',
  },
  bulkErklaerung:
    'MTGJSON non risponde per singola carta ma con un file. “Oggi” porta il giorno corrente, “Novanta giorni” la cronologia — e più di un gigabyte.',
  bulkQuelle: 'Fonte',
  bulkUmfang: 'Ambito',
  bulkHeute: 'Oggi',
  bulkNeunzigTage: 'Novanta giorni',
  bulkHolen: 'Recupera',
  bulkHolenBei: 'Recupera i prezzi da {quelle}',
  bulkLaeuft: 'Recupero in corso …',
  bulkBilanz: '{staende} prezzi per {drucke} stampe salvati.',
  bulkOhneDrucke:
    'Nessuna stampa di questa collezione ha un identificativo MTGJSON — non è stato richiesto nulla.',
  uebersichtTitel: 'Panoramica',
  uebersichtLeer: 'Ancora nessun prezzo recuperato.',
  ohnePreis: 'senza prezzo: {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} copia',
  exemplare: '{zahl} copie',

  keineUmrechnung:
    'Una somma per valuta. Nessuna conversione — servirebbe un tasso di cambio, e questo programma non ce l’ha.',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: 'Cronologia dei prezzi da un’altra fonte',
  verlaufErklaerung:
    'Recupera una serie nel tempo e la salva. La ricerca passa dalla rete e consuma quota; la condizione è mostrata come la chiama la fonte.',
  verlaufHolen: 'Recupera la cronologia',
  verlaufHolenBei: 'Recupera la cronologia da {quelle}',
  verlaufLaeuft: 'Recupero in corso …',
  verlaufBilanz:
    '{staende} punti di prezzo salvati — versione di carta {fassung}.',
} as const
