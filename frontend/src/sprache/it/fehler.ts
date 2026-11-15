export const fehler = {
  unbekannt:
    'Si è verificato un errore imprevisto. I dettagli si trovano nel registro — in «Impostazioni › Diagnostica» puoi vederne la fine e salvarlo.',
  unbekanntMitVorgang:
    'Si è verificato un errore imprevisto. I dettagli si trovano nel registro — in «Impostazioni › Diagnostica» puoi vederne la fine e salvarlo. Cerca lì {vorgang}: tutte le righe di questa operazione riportano questo identificativo.',
  nichtGefunden:
    'La voce cercata non è stata trovata. Forse l’ha eliminata qualcun altro — ricarica l’elenco.',
  ohnePeeridentitaet:
    'Questo profilo non ha ancora un’identità. Creala nel riquadro «Come si identifica questo profilo» — senza di essa nessuno può dimostrare chi è.',
  peerNichtErreichbar:
    'Al momento la tua controparte non è raggiungibile. È in ascolto, e il suo indirizzo è giusto? Altrimenti riprova più tardi.',
  peerOhneAdresse:
    'Per questa controparte non è registrato alcun indirizzo. Inseriscilo nel suo contatto in «Indirizzo e nota» — senza, questo programma la trova solo nella stessa rete.',
  ohneNetzerlaubnis:
    'Nulla è uscito verso la rete locale. Se macOS ha appena chiesto l’autorizzazione, concedila e riprova. Da qui non si può stabilire con certezza il motivo: o manca l’autorizzazione — Impostazioni di Sistema → Privacy e sicurezza → Rete locale —, oppure la rete non lascia passare il multicast, ad esempio dietro una VPN.',
  ohneNetzroute:
    'Al momento non c’è nessun percorso verso la rete locale. C’è un’interfaccia di rete collegata? In Impostazioni → Connessioni puoi sceglierne una precisa.',
  schluesselbundAblage:
    'La chiave di questo profilo non si è potuta salvare nel portachiavi. Il portachiavi di login è sbloccato? Poi riprova.',
  einladungZuNeu:
    'Questo invito proviene da una versione più recente di questo programma. Chiedi all’altra parte un invito dalla sua versione precedente — oppure procurati tu la versione più recente.',
  fernhorchenLaeuft:
    'Questo profilo è già raggiungibile attraverso altre reti. Interrompilo prima se vuoi un nuovo invito — quello vecchio diventa non valido.',
  anmeldungAbgelaufen:
    'L’accesso non è tornato entro dieci minuti. Riprova quando sei pronto nel browser.',
  umschlagFassung:
    'Questa busta proviene da una versione diversa di questo programma. Chiedi all’altra parte una busta dalla sua versione precedente — oppure procurati tu la versione più recente.',
  umschlagUnterschrift:
    'La firma di questa busta non è valida. È stato modificato durante il tragitto oppure proviene da qualcuno diverso da chi dichiara — non accettarlo e verifica con l’altra parte.',
  umschlagInhalt:
    'Una parte di questa busta non corrisponde al suo manifesto. Chiedi all’altra parte di inviarlo di nuovo — durante il tragitto qualcosa si è danneggiato o è stato modificato.',
  fremderNutzer:
    'Solo questo utente può farlo. Passa prima a questo utente — l’app chiederà il suo secondo fattore.',
  nichtFreigegeben:
    'La controparte non condivide nulla con te per lo scambio. Chiedile di condividerti una vista, oppure scambiate tramite file.',
  zeileNichtVerbuchbar:
    'La riga {zeile} non potrebbe essere registrata dalla parte che la riceve: manca la condizione o la lingua, oppure un dato non esiste. Rimuovila e aggiungila di nuovo con condizione, lingua e finitura.',
  zusageUnpassend:
    'Questa copia non soddisfa nessuna delle righe che dai: condizione, lingua o finitura sono diverse. Scegline una corrispondente o aggiungila alla collezione.',
  erlaubnis:
    'Non hai il ruolo necessario. Chi gestisce questo profilo può assegnartelo.',
  unbekannteHerkunftsart:
    'Il programma non conosce questa origine. Scegli «ripresa da un profilo» oppure «importata».',
  // The command centre (#1423).
  befehlBestaetigen:
    'Questo comando modifica qualcosa che non si annulla facilmente — leggi l’anteprima, poi conferma.',
  befehlArgumente:
    'Gli argomenti non sono un oggetto JSON. Inserisci un oggetto tra parentesi graffe — un campo per ogni parametro.',
  herkunftOhneProfil:
    'Un’origine da un profilo lo indica, un’importazione non ne indica alcuno — indicane uno oppure lascialo vuoto.',
  nurLesend:
    'In questo profilo puoi solo leggere. Appartiene a qualcun altro e il tuo ruolo è «lettore». Passa al tuo profilo per lavorare — oppure chiedi a chi lo gestisce il ruolo «collaboratore».',
  bestandUnbrauchbar:
    'Questo file è danneggiato o non è una libreria di MagicSetEditor4Mac — non può essere aperto. Ripristina un backup; l’ultimo si trova dove lo hai salvato.',
  bestandNeuereFassung:
    'Questa libreria proviene da una versione più recente di MagicSetEditor4Mac. Il file è a posto — non sostituirlo. Procurati la versione più recente oppure ripristina un backup precedente.',
  bestandAeltereFassung:
    'Questa libreria proviene da una versione precedente alla 0.7.0, e questa versione non può più aggiornarla. Il file è a posto — non sostituirlo. Aprila una volta con MagicSetEditor4Mac 0.7.0; dopo, anche questa versione la aprirà.',
  bestandNurLesbar:
    'La libreria si può leggere ma non scrivere — al file o alla sua cartella manca il permesso di scrittura. Il file è a posto — non sostituirlo. Dagli il permesso di scrittura nel Finder, in «Ottieni informazioni», e riprova.',
  bestandEA:
    'Il disco ha segnalato un errore durante la lettura della libreria. Non si sa se il file è integro — non sostituirlo subito. Controlla il disco, ad esempio con Utility Disco, e riprova.',
  kaumPlatz:
    'Non è stato possibile aprire o scrivere la raccolta, e sul disco lo spazio libero è quasi esaurito. Il file è probabilmente integro: non sostituirlo. Libera spazio e riprova; se non basta, controlla il disco, ad esempio con Utility Disco, e il permesso di scrittura.',
  startUngeklaert:
    'Non è stato possibile aprire la libreria e il motivo non è chiaro — è riportato qui sotto. Riprova. Ripristina un backup solo quando sei sicuro che questo file non si può più aprire.',
  sicherungVerworfen:
    'Il backup non ha superato la verifica ed è stato scartato — non è stato salvato nulla. Riprova, oppure salva in un altro percorso.',
  sicherungUnvermerkt:
    'Il backup è scritto e verificato — solo la sua ora non si è potuta registrare, e l’età dell’ultimo backup mostra ancora quello precedente. Conserva il file e salva di nuovo quando la libreria sarà libera.',
  schluesselFehlt:
    'Questa libreria è cifrata e la sua chiave non si trova nel portachiavi di questo Mac. Il file è a posto — non sostituirlo. La chiave è stata creata quando è stata attivata la cifratura e non può essere ricalcolata; senza di essa i dati sono irraggiungibili.',
  belegt:
    'La libreria è occupata in questo momento — probabilmente un’altra finestra o un comando sta scrivendo. Riprova tra un attimo.',
  bestandWirdGetauscht:
    'La libreria è in fase di sostituzione — per il ripristino di un backup, il cambio della cifratura o un nuovo inizio. Attendi un attimo, poi riprova.',
  anwendungEndet:
    'MagicSetEditor4Mac si sta chiudendo — questo passaggio non è stato eseguito. Avvia di nuovo l’applicazione per continuare.',
  vorgangAbgebrochen:
    'L’operazione è stata interrotta prima della fine. Avviala di nuovo se ti serve ancora.',
  konflikt: 'La voce è stata modificata nel frattempo. Ricarica.',
  konfliktgrund:
    'Non è andata: {grund}. Controlla lo stato attuale prima di riprovare.',
  doppelt:
    'Esiste già una voce con questi dati. Modifica uno dei dati — oppure modifica quella esistente.',
  doppelterName: 'Questo nome esiste già qui — scegline un altro.',
  doppeltesExemplar:
    'La collezione contiene già questa stampa — nello stesso stato, lingua e finitura. Aumenta lì la quantità oppure cambia uno dei dati.',
  gesperrt:
    'Questa app non può accedere a questa posizione — è bloccata. Scegline un’altra oppure consenti l’accesso, nei permessi del file o in Impostazioni di Sistema › Privacy e sicurezza › File e cartelle, e riprova.',
  voll: 'Non c’è più spazio sul disco. Libera dello spazio e riprova.',
  netzName:
    'Non è stato possibile risolvere il nome «{wirt}». Controlla la connessione a internet e, se ne è configurato uno, il proxy in «Impostazioni › Connessioni». Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  netzNichtErreichbar:
    'Su «{wirt}» nessuno accetta connessioni. Il servizio è attivo e il suo indirizzo è corretto? Altrimenti riprova più tardi. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  netzKeineAntwort:
    '«{wirt}» non ha risposto entro {sekunden} secondi. Riprova più tardi. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  netzAntwortStockt:
    'La risposta di «{wirt}» è rimasta ferma per {sekunden} secondi. Controlla la connessione e riprova. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  netzAbgerissen:
    'La connessione con «{wirt}» si è interrotta prima che la risposta fosse completa. Riprova; se succede ancora, il problema è nel servizio o in un proxy intermedio. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  netzProxyAnmeldung:
    'Il proxy «{wirt}» richiede un accesso. Indica le sue credenziali tramite HTTPS_PROXY nell’ambiente: in «Impostazioni › Connessioni» il proxy non accetta di proposito alcuna password. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  gedrosselt:
    'L’altra parte segnala troppe richieste. Attendi un momento e poi riprova. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  gedrosseltBis:
    '«{wirt}» al momento non accetta richieste. Riprova dalle {uhrzeit}: fino ad allora questo programma non chiederà di nuovo. Ciò che la connessione ha segnalato si trova in «Impostazioni › Diagnostica» sotto {vorgang}.',
  kontingent: 'Troppe chiamate in poco tempo. Attendi {sekunden} s e riprova.',
  prozessVeraltet:
    'Questo processo proviene da una versione più vecchia della libreria e si sta chiudendo: non è stato scritto nulla. Un servizio installato si riavvia da solo, con il programma per cui è installato; riavvia qualsiasi altro processo con la nuova versione, poi riprova.',
  bestandVomDienstGehalten:
    'Il servizio per script tiene aperta la libreria. Revocalo in Impostazioni › Diagnostica › «Configurare la riga di comando e il server degli strumenti», poi ripristina di nuovo.',
  falschesSpiel:
    'Il mazzo appartiene a un gioco diverso da quello di questo torneo. Scegli un mazzo dello stesso gioco.',
  turnierformatSpiel:
    'Il formato appartiene a un gioco diverso da quello di questo torneo. Scegli un formato dello stesso gioco — oppure nessuno.',
  paketGeaendert:
    'Questo pacchetto differisce da come è stato installato. Conferma la sostituzione — la modifica viene conservata come versione precedente.',
  symbolZuVieleTeile:
    'Questo simbolo ha {anzahl} parti, ne sono ammesse {grenze}. Unisci delle parti oppure scegli un disegno più semplice.',
  symbolZuVielePunkte:
    'Questo simbolo ha {anzahl} punti, ne sono ammessi {grenze}. Semplifica le forme oppure scegli un disegno più semplice.',
  symbolZuTief:
    'I gruppi di questo simbolo sono annidati su {anzahl} livelli, ne sono ammessi {grenze}. Scegli un disegno più piatto oppure separa i gruppi interni.',
  symbolZuGross:
    'Questo file è più grande di {grenze} e quindi non è un file di simbolo. Scegli un file .mse-symbol.',
  symbolGespeichertZuGross:
    'Salvato, questo simbolo supererebbe {grenze}, più di quanto possa contenere un file di simbolo. Semplifica le forme o usa meno punti.',
  bildZuGross:
    'Questa immagine è troppo grande per essere decompressa. Scegline una più piccola o riducila prima.',
  setOhneSymbol:
    'Questo set non ha ancora un simbolo. Disegnane uno e salvalo — poi potrà essere aggiunto.',
  // **Die eigenen Übersetzungen** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '«{kennung}» non è un identificativo di lingua secondo BCP 47. Usane uno come pt-BR o zh-Hant.',
  sprachkennungEingebaut:
    '«{kennung}» è una lingua integrata. Scegli un identificativo tuo, ad esempio {kennung}-x-meine.',
  platzhalterAbweichend:
    'Questa traduzione usa segnaposto diversi da quelli della frase tedesca. Inserisci esattamente questi: {erwartet}.',
  auszeichnungUnsicher:
    'Questa traduzione contiene qualcosa che non è testo, oppure una formattazione che non si chiude. Rimuovi i tag e chiudi ** e `.',
  sprachpaketNeuer:
    'Questa traduzione proviene da un programma più recente ({format}). Procurati la nuova versione.',
  sprachpaketOhneLizenz:
    'Senza licenza nessuna traduzione può essere diffusa. Inseriscine una, ad esempio CC0-1.0.',
  bereichAbgeschaltet:
    'Quest’area è disattivata ({schalter}). Riattivala in Impostazioni › Aree e funzioni: non è andato perso nulla.',
  dienstAus:
    'Il servizio {dienst} è disattivato. Attivalo in Impostazioni › Servizi se vuoi interrogarlo.',
  fremdeEigeneKarten:
    'Le carte proprie del mittente non sono arrivate — il loro identificativo vale solo nella sua libreria: {karten}. Se ti servono, apri il suo set e aggiungile qui.',
  nichtGelegtePlaetze:
    'Non è stato possibile aggiungere qui queste voci: {plaetze}. Il resto del mazzo è stato accettato; controllale e aggiungile a mano.',
  fremdeKartenUndNichtGelegte:
    'Le carte proprie del mittente non sono arrivate — il loro identificativo vale solo nella sua libreria: {karten}. E non è stato possibile aggiungere qui queste voci: {plaetze}. Controlla entrambe e aggiungile a mano.',
  symbolVerwendet:
    'Questi set usano il simbolo: {sets}. Conferma l’eliminazione — conservano la loro copia.',
  symbolOhneVorige:
    'Questo simbolo non ha una versione precedente. Salva prima una nuova versione — poi si potrà ripristinare quella attuale.',
  turnierAbgeschlossen:
    'Questo torneo è chiuso. Scegli «Riapri» per modificare qualcosa.',
  sitzGegangen:
    'Questo posto ha lasciato la partita e non può più agire. Scegli un posto ancora in gioco.',
  fremderWert:
    'Nel frattempo qualcun altro ha modificato questo campo. Controlla il nuovo valore e decidi se mantenere la tua bozza.',
  vorschlagNurGegenseite:
    'Solo l’altra parte può modificare un’offerta aperta. Attendi la sua risposta, oppure annulla lo scambio e offri di nuovo.',
  turnierNichtFertig:
    'Inserisci prima ogni risultato: un torneo si può chiudere quando ha almeno un turno e ogni abbinamento ha un risultato.',
  turniernameLeer: 'Un torneo ha bisogno di un nome — inseriscilo.',
  turnierspielLeer:
    'Un torneo appartiene a un gioco — altrimenti non c’è alcun formato con cui confrontare. Scegline uno.',
  validierungsgrund: 'Non è andata: {grund}. Correggi il dato e riprova.',

  validierung:
    'Il dato inserito non è valido. Il suggerimento accanto al campo ne spiega il motivo — correggilo lì.',

  // **The rule itself is the remedy** (#1350).
  // **The password of a backup** (OPS-06, #1723).
  sicherungskennwortKurz:
    'La password di un backup ha almeno dodici caratteri. Scegline una più lunga — una frase di quattro parole è ottima e facile da ricordare.',
  sicherungskennwortFehlt:
    'Questo backup è cifrato. Inserisci la sua password per vederlo o ripristinarlo.',
  sicherungskennwortFalsch:
    'La password non corrisponde a questo backup. Controllala e inseriscila di nuovo.',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} non è un JSON leggibile: al byte {stelle} il file si interrompe o è malformato. Scegli un altro file.',
  mtgjsonKeinSet:
    '{datei} è JSON, ma non contiene alcun set in formato MTG JSON (differisce al byte {stelle}). Scegli un file con un set, oppure un’intera cartella.',
  sicherungImKlartext:
    'La libreria è cifrata — senza password il backup sarebbe in chiaro. Imposta una password in Impostazioni › Dati e backup oppure scegli lì esplicitamente «esegui il backup senza password».',

  planVerschluesselt:
    'La libreria è cifrata e il piano esegue i backup senza password, quindi il piano non la salva. Esegui un backup manuale con password.',

  paketnameUngueltig:
    'Il nome di un pacchetto è composto da lettere, cifre, trattini e trattini bassi — diventa il nome della cartella. Scegline uno con questi caratteri.',
  ordnerangabe:
    '«{wert}» non è un percorso completo. Inseriscine uno che inizi con / oppure con ~/, senza .. in mezzo.',
  ordnerFehlt:
    'La cartella «{ordner}» non è raggiungibile. Collega l’unità oppure scegli un’altra cartella.',
  ordnerIstDatei:
    '«{ordner}» è un file, non una cartella. Scegli la cartella che lo contiene, oppure un’altra.',
  ordnerGesperrt:
    'Questa app non può scrivere in «{ordner}». Scegli un’altra cartella oppure consenti l’accesso a questa.',
  ordnerVoll:
    'Sull’unità di «{ordner}» restano solo {frei}. Libera spazio oppure scegli un’altra cartella.',
  karteInDeck:
    'Questa carta è usata in {anzahl} mazzo/i — {decks}. Se la elimini, viene tolta anche da lì; conferma se è quello che intendi.',
  pflichtfeld: 'Questo campo non può essere vuoto — inserisci qualcosa.',
  nichtEditierbar:
    'Questo campo non è modificabile — il modello lo mostra soltanto. Modifica il valore da cui deriva.',
  berechnetesFeld:
    'Questo campo deriva da altri e non può essere impostato su molte carte. Le sue parti sono elencate separatamente — scegline una.',
  keinSetMitStil:
    'Nessun set aperto usa questo stile. Apri un set che lo usa — oppure cambia lo stile di quello aperto.',
  vorlageFehlt:
    'Per questa carta non è installato alcun modello. Controlla nella pagina Pacchetti il gioco e lo stile del set — senza di essi una carta non ha campi né aree.',
  sicherungNeuereFassung:
    'Questo backup proviene da una versione più recente di MagicSetEditor4Mac. Contiene {unbekannt}, mentre questa versione arriva fino a {stand}. Procurati la versione più recente — oppure scegli un backup più vecchio.',
  sicherungAeltereFassung:
    'Questo backup proviene da una versione precedente alla 0.7.0 — gli mancano passi dello schema fino a {grundlinie}, e questa versione non può più aggiornarlo. Ripristinalo con MagicSetEditor4Mac 0.7.0; dopo, questa versione aprirà la libreria.',
  sicherungUnbrauchbar:
    'Questo file non è un backup di MagicSetEditor4Mac. Scegline uno scritto da questa applicazione.',
  // The parser's own sentence is German for now — see issue #1178.
  skriptFehler:
    'L’espressione non può essere valutata: {meldung} Correggila nel punto indicato.',

  // **The missing second factor** — piece 7 of phase 4.12 (#1351).
  // **The local session** — ID-02 (#1701).
  kennwortFalsch:
    'Questa password non è corretta. Inseriscila di nuovo — oppure impostane una nuova con il foglio di recupero.',
  kennwortLaenge:
    'Una password ha almeno {mindestens} e al massimo {hoechstens} caratteri. Scegline una di questa lunghezza.',
  sperreOhneRueckweg:
    'Senza password e senza sblocco locale, il blocco non potrebbe essere rimosso. Imposta prima una password.',
  lokaleEntsperrungFehlt:
    'Lo sblocco locale non è configurato per questo utente su questo Mac. Inserisci invece la password.',
  codeFalsch:
    'Questo codice non è corretto — oppure è già stato usato. Controlla il codice sul foglio — ognuno vale una sola volta.',
  zweiterFaktor:
    'Questo utente richiede un secondo fattore. Inserisci il codice dalla tua app di autenticazione oppure conferma con Touch ID.',
  faktorGesperrt: 'Troppi codici errati. Attendi un momento e riprova.',
  sicherungFremdeFassung:
    'Questo backup non corrisponde a questa versione di MagicSetEditor4Mac. Indica {dort}, mentre questa applicazione è {hier}. Ripristinalo con la versione che lo ha scritto.',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: 'il carattere {0} non fa parte del linguaggio',
    einzelnesGleich:
      'un singolo `=`; si assegna con `:=`, si confronta con `==`',
    offeneZeichenkette: 'una stringa senza virgolette di chiusura',
    zuTiefGeschachtelt: 'l’espressione è annidata oltre {0} livelli',
    keinSkript: 'non esiste uno script chiamato {0}',
    skriptkreis: 'questi campi si leggono a vicenda in cerchio: {0}',
    keinVerweis: '`+` non può concatenare un riferimento a un altro script',
    keineZahl: '{0} non è un numero',
    formUnbekannt: 'la forma {0} non può essere valutata',
    nameUnbekannt: 'il nome {0} qui non è noto',
    keineFunktion:
      'qui viene chiamato {0}, il che funziona solo con una funzione',
    ohneRumpf: 'questa funzione non ha un corpo',
    nichtWahrheitswert: '`not` richiede un valore di verità, qui c’è {0}',
    vorzeichenZahl: 'un segno richiede un numero, qui c’è {0}',
    zaehlschleifeZahl:
      'in un ciclo di conteggio, {0} richiede un numero, qui c’è {1}',
    rechnenLinks: '`{0}` calcola con numeri, a sinistra {1}',
    rechnenRechts: '`{0}` calcola con numeri, a destra {1}',
    vergleichLinks: '`{0}` confronta numeri, a sinistra {1}',
    vergleichRechts: '`{0}` confronta numeri, a destra {1}',
    keineListe: 'sort_list ordina un elenco, qui c’è {0}',
    zuTief:
      'la valutazione scende per più di {0} livelli — qualcosa qui richiama se stesso?',
    schleifeSteht:
      'un ciclo di conteggio non avanza a {0} — il numero è troppo grande',
    funktionVerbinden:
      'una funzione può essere combinata solo con una funzione, qui ci sono {0} e {1}',
    fehltNach: 'dopo {0} manca {1}',
    fehltHier: 'qui manca {0}',
    klammerOffen: 'questa {0} non viene mai chiusa — manca {1}',
    ausdruckGehtWeiter: 'qui l’espressione finisce, ma il testo continua',
    zuweisungBrauchtNamen: 'si assegna a un nome',
    durchgereicht: '{0}',
    bedingungWahrheit:
      'una condizione richiede un valore di verità, qui c’è {0}',
    fuerJedenListe: '`for each` scorre un elenco, qui c’è {0}',
    wahrheitLinks: '`{0}` richiede valori di verità, a sinistra c’è {1}',
    wahrheitRechts: '`{0}` richiede valori di verità, a destra c’è {1}',
    bindenFunktion: '`@` collega argomenti a una funzione, qui c’è {0}',
  },
} as const
