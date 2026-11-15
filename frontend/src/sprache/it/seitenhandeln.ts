// The texts of the **trading** area — Werkstück T9 von Phase 5.6
// (#1474).
//
// **Ein eigener Katalog und keine Zeilen mehr in `seitentechnik.ts`.** Der
// Abgleich, die Post und die Verbuchung standen dort, weil es den Bereich
// noch nicht gab — ihre eigenen Kommentare sagen es. Jetzt gibt es ihn, und
// ein Text über das Handeln unter „Technik" wäre eine Fundstelle, die
// niemand sucht.

import { handelsliste } from './handelsliste'
import { handelsassistent } from './handelsassistent'

export const handeln = {
  liste: handelsliste,
  assistent: handelsassistent,
  titel: 'Scambi',
  erklaerung:
    'Scambia con una persona che hai confermato. Tutto viene calcolato su questo computer; ciò che ne esce lo decidi tu.',
  gebiete: 'Aree degli scambi',
  gebietPartner: 'Con chi',
  gebietMarkt: 'Quanto vale',
  neuerTausch: 'Nuovo scambio …',
  neuerTauschTitel: 'Nuovo scambio',
  partnerTitel: 'Controparti confermate',
  partnerErklaerung:
    'Scambi solo con una controparte con cui hai confrontato un codice. Scegline una e lo scambio si apre; per invitare e confermare usa il riquadro qui sotto.',
  partnerLeer:
    'Ancora nessuna controparte confermata. Scambiatevi i codici d’invito nel riquadro qui sotto e confrontate le cifre.',
  partnerWaehlen: 'Apri',
  partnerWaehlenName: 'Apri {name}',
  partnerOffen: 'Aperto',
  partnerOffenName: '{name} aperto',
  partnerKeiner:
    'Scegli uno scambio a sinistra, oppure avviane uno con “Nuovo scambio …”.',
  marktTitel: 'Quanto vale la tua collezione',
  marktErklaerung:
    'Un’osservazione di mercato non è una promessa: dice cosa è stato richiesto altrove, non cosa otterrai. Da dove viene ogni numero e quanto è vecchio è indicato accanto — e dove non c’è nulla, non viene mostrato nulla.',
  vorgangTitel: 'Uno scambio dall’apertura alla consegna',
  vorgangErklaerung:
    'Qui si svolge il contratto: aprire, offrire, accordarsi, preparare, consegnare. **Qui non si sposta alcun inventario** — la registrazione avviene in «Registrazione», e questo programma lo chiede ancora una volta.',
  wahlTitel: 'Quale scambio',
  wahlNeu: 'Un nuovo scambio',
  wahlNeuZusatz: 'Il programma assegna il suo identificativo all’apertura.',
  wahlEintrag: '{zustand} · revisione {revision} · {positionen}',
  wahlLeer: 'Con questa controparte non c’è ancora nessuno scambio.',
  wahlFachleute: 'Apri uno scambio tramite il suo identificativo',
  wahlFachleuteErklaerung:
    'Ogni scambio ha un **identificativo**. Il programma lo assegna all’apertura e ogni messaggio che lo riguarda lo porta con sé. Se uno scambio non compare ancora qui sopra — ad esempio perché il suo primo messaggio è ancora in viaggio —, lo apri qui con quell’identificativo.',
  wahlKennung: 'Identificativo dello scambio',
  wahlOeffnen: 'Apri',
  wahlGewaehlt: 'È stato scelto lo scambio {kennung}.',
  vorgangOhneKennung: 'Scegli uno scambio qui sopra — oppure aprine uno nuovo.',
  vorgangAnsehen: 'Mostra lo stato',
  vorgangFachleute: 'Mostra lo stato con versione e impronta',
  vorgangEroeffnen: 'Apri',
  vorgangVorschlagen: 'Proponi di nuovo',
  vorgangRevision: 'Versione',
  vorgangAbdruck: 'Impronta',
  vorgangParteien: 'Parti',
  vorgangZustimmungen: 'Finora hanno accettato',
  vorgangKeine: 'Ancora nessuno.',
  vorgangDu: 'tu',
  vorgangEnde: 'Qui non si va oltre.',
  vorgangArt: 'Tipo',
  arten: {
    teilen: 'Condivisione',
    deckleihe: 'Prestito di mazzi',
    tausch: 'Baratto',
    markt: 'Mercato',
  },
  artOhneBestand:
    'Questo scambio non sposta inventario — viene registrato solo un baratto.',
  vorgangVereinbarung:
    'Siete d’accordo. Questo è un **accordo**: le carte passano di mano solo alla consegna, e vengono conteggiate solo quando entrambi hanno registrato.',
  vorgangSchritte: 'Passaggi',
  schrittAnbieten: 'Offri',
  schrittGegenangebot: 'Controfferta',
  schrittZustimmen: 'Accetta',
  schrittVorbereiten: 'Prepara',
  schrittUebergeben: 'Conferma la consegna',
  schrittAblehnen: 'Rifiuta',
  schrittAbbrechen: 'Annulla',
  schrittAblaufen: 'Scaduto',
  schrittKlaeren: 'Segnala per chiarimento',
  positionTitel: 'Che cosa passa di mano',
  positionVon: 'Da chi',
  positionVonIch: 'Da te',
  positionVonGegenueber: 'Dalla tua controparte',
  positionPrinting: 'Stampa',
  kennungVonHand: 'Inserisci l’identificativo a mano',
  positionMenge: 'Quantità',
  positionGuete: 'Condizione',
  positionSprache: 'Lingua',
  positionFinish: 'Finitura',
  positionHinzufuegen: 'Aggiungi una riga',
  anschlussWartet:
    'La carta “{name}” è arrivata dalla tua collezione — controlla quantità e lato, poi è nell’offerta.',
  wunschTitel: 'Lista dei desideri da un mazzo',
  wunschWartet: '{name} da un mazzo, non ancora inserito: {liste}.',
  wunschEintragen: 'Inserisci come desiderio dalla controparte',
  anschlussOhnePartner:
    '“{name}” è in attesa — scegli una controparte da “Nuovo scambio …”, poi finisce nella sua riga.',
  preise: {
    holen: 'Recupera i prezzi',
    abbrechen: 'Annulla il recupero dei prezzi',
    laeuft: 'Recupero dei prezzi … {fertig} di {gesamt}',
    fertig: 'Prezzi recuperati per {zahl} stampe.',
    teils: '{zahl} stampe su {gesamt} senza prezzo: {meldung}',
    abgebrochen:
      'Annullato dopo {fertig} di {gesamt}: ciò che è stato recuperato resta.',
  },
  markttisch: {
    neuTitel: 'Nuovo scambio con',
    neuMit: 'Nuovo scambio con {name}',
    meinFach: 'Il mio scomparto',
    sammlung: 'La mia collezione',
    wunschliste: 'Lista dei desideri',
    seinFach: 'Scomparto di {name}',
    freigegeben: 'Cosa {name} condivide per lo scambio, chiesto all’apertura.',
    meinLeer: 'Qui non c’è ancora niente. Le copie arrivano dalla collezione.',
    wunschLeer:
      'Nessun desiderio. Una carta diventa desiderio nella collezione.',
    seinLeer: 'Niente di condiviso.',
    neuFragen: 'Chiedi di nuovo',
    hinweis:
      'Trascina, oppure Spazio, frecce, Invio. Selezione multipla con ⌘.',
    laedt: 'Richiesta …',
    karte: '{zustand} · {sprache} · {menge}×',
    ansicht: 'Vista',
    alsTisch: 'Come tavolo',
    alsListe: 'Come elenco',
    fachSuchen: 'Cerca nello scomparto',
    fachZaehlung: '{gezeigt} di {gesamt}',
    fachMehr: 'Mostra altro',
    fachKeinTreffer: 'Nessuna carta in questo scomparto si chiama così.',
  },
  zug: {
    veraltet:
      'Nel frattempo lo scambio è cambiato. Guarda il nuovo stato e decidi di nuovo.',
    vorbereitenUebergeben: 'Prepara e consegna',
    uebergebenVerbuchen: 'Consegna e registra',
    anbietenZustimmen: 'Offri e accetta',
    buchungHinweis:
      '«Registra» comunica poi la riconciliazione e conferma la tua chiusura. La conferma di {name} la inserisci a parte.',
    festhaltenOffen:
      'Non è chiaro quale copia parte. Sceglila in «Riserve» e premi di nuovo.',
    gescheitert: '{schritt} non è riuscito: {meldung}',
    lesen:
      'È registrato, ma dopo non è stato possibile leggere lo stato: {meldung}',
    ausgang:
      'È registrato da te, ma non è arrivato nella posta in uscita: {meldung}',
    unterwegs: 'È registrato da te; {name} non lo ha ancora: {meldung}',
    hinaus: 'Inviato a {name}.',
  },
  sicht: {
    titel: 'Cosa condivide la controparte',
    anfragen: 'Chiedi la vista condivisa',
    laeuft: 'Richiesta tramite la connessione …',
    zeilen: '{zahl} righe condivise, di cui {wuensche} cercate.',
    keinVorschlag:
      'Niente combacia: nessun desiderio qui incontra un’offerta là.',
    gibst: 'Dai',
    bekommst: 'Ricevi',
    zeile: '{menge} × {name}',
    uebernehmen: 'Accetta la proposta',
    uebernommen: 'Accettato: {zahl} righe sono nell’offerta.',
  },
  druck: {
    eigen: 'stampa propria · senza immagine',
    ohneNamen: 'Stampa senza nome',
    unbekannt: 'Stampa sconosciuta',
    bezug: '{code} #{nummer}',
    laedt: '…',
  },
  positionEntfernen: 'Rimuovi la riga',
  positionEntfernenName: 'Rimuovi la riga {nummer}',
  positionLeer:
    'Ancora nessuna riga. Un’offerta senza posizione non dice di che cosa si tratta.',
  positionVorbereitung:
    'Quale delle tue copie parte lo indichi in «Riserve»: un’offerta indica una **stampa**, non una riga del tuo inventario.',
  zusagenTitel: 'Quali copie partono',
  zusagenErklaerung:
    'Un’offerta indica una **stampa**, non una riga del tuo inventario. Indica qui **quale** copia parte — altrimenti nulla dice alla registrazione da dove detrarre.',
  zusagenWahl: 'Quale delle tue copie parte',
  zusagenPosition: '{menge} × {name}',
  zusagenZeile: '{lagerort} · {zustand} · {sprache} · {finish} · {stueck}',
  zusagenOhneLagerort: 'nessuna ubicazione',
  zusagenKeinExemplar: 'Nella tua collezione non c’è nulla di tutto ciò.',
  zusagenKeinePassende:
    'Nessuna riga della tua collezione corrisponde a condizione, lingua e finitura di questa riga. Aggiungila alla collezione, oppure aggiungi di nuovo la riga allo scambio con dati corrispondenti.',
  zusagenFachleute:
    'Riserva una riga della collezione tramite il suo identificativo',
  zusagenFachleuteErklaerung:
    'Una **riga della collezione** è una riga della tua collezione e la sua **versione** conta ogni modifica. Entrambe compaiono già nella scelta qui sopra — qui le inserisci a mano, ad esempio per una copia che lì non compare.',
  zusagenExemplar: 'Quale riga della collezione',
  zusagenMenge: 'Quante',
  zusagenFassung: 'Versione vista',
  zusagenFesthalten: 'Riserva',
  zusagenNachsehen: 'Mostra le riserve',
  zusagenOhneVorgang:
    'Scegli uno scambio qui sopra; qui vedrai poi cosa dai in cambio.',
  zusagenLeer: 'Per questo scambio non è riservato nulla.',
  zusagenOffen: 'riservata',
  zusagenUebergeben: 'consegnata',
  ausgangTitel: 'Che cosa parte verso la tua controparte',
  ausgangErklaerung:
    'Lo stato viene firmato e messo nella **posta in uscita**. **Invia ora** apre una linea verso la tua controparte e porta fuori ciò che è in attesa — resta comunque lì finché la tua controparte non lo conferma.',
  ausgangSchicken: 'Metti nella posta in uscita',
  ausgangNachricht: 'Identificativo del messaggio',
  ausgangWegraeumen: 'È arrivato — rimuovilo',
  ausgangWegraeumenName: 'È arrivato — rimuovilo: {nachricht}',
  ausgangFachleute: 'Rimuovi un messaggio tramite il suo identificativo',
  ausgangFachleuteErklaerung:
    'Ogni messaggio nella posta in uscita ha un **identificativo**. Ciò che dopo un’interruzione è ancora in sospeso compare qui sopra con un pulsante per rimuoverlo; qui ne rimuovi uno che lì non compare.',
  ausgangBestaetigen: 'Rimuovi come arrivato',
  ausgangOhneVorgang: 'Scegli qui sopra uno scambio che deve partire.',
  ausgangOhneNachricht: 'Inserisci prima l’identificativo del messaggio.',
  ausgangWiederaufnahme: 'Controlla dopo una disconnessione',
  ausgangZustellen: 'Invia ora',
  ausgangOhneGegenueber: 'Scegli prima a chi inviare.',
  ausgangGetragen: '{zahl} inviati.',
  ausgangKlaerung:
    'Ci sono due cronologie. Non si può calcolare quale valga — parlatene tra voi.',
  ausgangNachzuliefern: 'Deve ancora partire',
  ausgangLeer: 'Nella posta in uscita non c’è nulla.',
  ausgangFolge: 'Sequenza',
  tauschTitel: 'Che cosa potreste scambiare',
  tauschErklaerung:
    'Incolla qui ciò che la tua controparte ti ha inviato — i suoi doppioni e i suoi desideri, così come il suo programma li ha rilasciati. Il confronto avviene su questo computer; non ne esce nulla.',
  tauschStand: 'Che cosa ha inviato la tua controparte',
  tauschAbgleichen: 'Confronta',
  tauschAusUmschlag: 'Leggi da una busta',
  tauschUnlesbar: 'Non è un’istantanea rilasciata.',
  tauschLeer: 'Non c’è nulla che corrisponda.',
  tauschGebe: 'Tu dai',
  tauschBekomme: 'Tu ricevi',
  tauschVorschlag: 'Una proposta',
  tauschOhneVorschlag:
    'Una proposta richiede qualcosa da entrambe le parti — uno scambio in cui uno non dà nulla è un regalo.',
  tauschBefunde: 'Che cosa non è stato possibile confrontare',
  postLeer: 'Niente in viaggio.',
  postOffen: 'In viaggio',
  postVersuche: 'Tentativi',
  buchungTitel: 'Registra uno scambio',
  buchungErklaerung:
    'Quando le carte hanno davvero cambiato proprietario, registra qui lo scambio. La tua collezione cambia e il giornale tiene traccia di che cosa è andato dove.',
  buchungVerbuchen: 'Registra',
  buchungAbgleichen: 'Segnala la riconciliazione',
  buchungEigeneStimme: 'Conferma la conclusione',
  buchungBestaetigen: 'Confermato dall’altra parte',
  buchungNochNicht:
    'Per ora non c’è nulla da registrare: si registra quando la consegna è confermata.',
  buchungZustand: 'Stato',
  buchungStandEntwurf: 'Bozza',
  buchungStandAngeboten: 'Offerto',
  buchungStandVerhandlung: 'In trattativa',
  buchungStandAngenommen: 'Concordato',
  buchungStandVorbereitet: 'Preparato',
  buchungStandUebergabeOffen: 'Consegna confermata',
  buchungStandLokalVerbucht: 'Registrato da parte tua',
  buchungStandAbgleichOffen: 'In riconciliazione',
  buchungStandBeidseitigVerbucht: 'Registrato da entrambe le parti',
  buchungStandAbgelehnt: 'Rifiutato',
  buchungStandAbgebrochen: 'Annullato',
  buchungStandAbgelaufen: 'Scaduto',
  buchungStandKlaerung: 'Da chiarire',
  buchungBeleg: 'Ricevuta',
  buchungEinseitig: 'Solo tu hai firmato.',
  buchungBeidseitig: 'Hanno firmato entrambe le parti.',
  buchungOhneGegenbeleg:
    'Dall’altra parte non c’è ancora una ricevuta verificata.',
  buchungGegenbeleg: 'Ricevuta verificata da {wer}, arrivata {wann}.',
  buchungGegenbelegAbgewiesen:
    'È arrivata una ricevuta da {wer} {wann}, ma non corrisponde alla tua — {mangel}',
  buchungGegenbelegRat:
    'Allinea lo scambio con {wer}; una ricevuta corrispondente sostituisce questo avviso.',
  gegenbelegMaengel: {
    unterschrift: 'la firma non è del mittente.',
    vorgang: 'appartiene a un altro scambio.',
    revision: 'appartiene a un altro stato di questo scambio.',
    parteien: 'nomina altre parti.',
    zeilen: 'le carte registrate non coincidono.',
  },
  buchungJournal: 'Che cosa si è mosso',
  buchungJournalLeer: 'Per questo scambio non c’è ancora nulla nel giornale.',
  kontoTitel: 'Il tuo account su un marketplace',
  kontoErklaerung:
    'Ciò che **tu** offri lì e ciò che **tu** hai venduto o comprato. Viene solo letto: non viene creata alcuna inserzione, nulla viene messo nel carrello, nulla viene acquistato e nulla spedito. Il percorso parte **disattivato** — in Impostazioni → Servizi, oltre che sul marketplace stesso.',
  kontoNachsehen: 'Dai un’occhiata',
  kontoAngebote: 'Inserzioni',
  kontoBestellungen: 'Ordini',
  kontoBewegt: 'ha spostato inventario',
  kontoBewegtNicht: 'non ha spostato inventario',
} as const
