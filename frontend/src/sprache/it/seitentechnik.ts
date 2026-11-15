// The texts of the **technical page** — operating state, effect counter and
// the three tools without a window.
//
// Split out of `seiten.ts` when the file grew past 400 lines with the effect
// counter (work item O9, #1433).

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: 'Stato operativo',
  gesundheitErklaerung:
    'Se questo processo è vivo e se può lavorare. Un servizio esterno che non risponde non lo rende non pronto — la modifica prosegue senza rete.',
  lebt: 'Il processo risponde.',
  bereit: 'Può lavorare.',
  nichtBereit: 'Al momento non può lavorare.',
  grundBestandZu: 'Nessun archivio è aperto.',
  grundWanderungLaeuft: 'L’archivio sta migrando a un nuovo schema.',
  grundBestandUnlesbar: 'L’archivio non risponde.',
  fremdeDienste: 'Servizi esterni',
  erreichbar: 'raggiungibile',
  nichtErreichbar: 'non raggiungibile',
  metriken: 'Effetti dall’avvio',
  metrikenErklaerung:
    'Ciò che questo processo ha prodotto verso l’esterno — per interfaccia, classe di effetto ed esito. Vengono contati solo gli effetti controllati; la semplice lettura non compare qui.',
  metrikenGrenze:
    'Al massimo {grenze} righe — non crescono con il numero di carte, set o profili.',
  metrikenLeer: 'Finora nulla ha avuto effetti verso l’esterno.',
  spalteFlaeche: 'Interfaccia',
  spalteWirkung: 'Effetto',
  spalteAusgang: 'Esito',
  spalteAnzahl: 'Numero',
  flaechen: {
    fenster: 'Finestra',
    kommandozeile: 'Riga di comando',
    mcp: 'MCP',
    dienst: 'Avviare il servizio per script',
  },
  wirkungen: {
    lokalLesen: 'lettura locale',
    lokalSchreiben: 'scrittura locale',
    externSenden: 'invio verso l’esterno',
    externSchreiben: 'modifica all’esterno',
    verbindlich: 'vincolante',
  },
  ausgaenge: { gewaehrt: 'concesso', abgewiesen: 'rifiutato' },
  paket: {
    ueberschrift: 'Pacchetto diagnostico',
    erklaerung:
      'Le stesse informazioni in più parti, come file con checksum — per una segnalazione che richiede più di un testo. Viene creato dove indichi tu e non va da nessuna parte da solo.',
    inhalt:
      'Contiene versione, piattaforma, archivio, pacchetti, contatori del processo, righe del registro e tracce dei crash — nessuna carta, immagine, collezione o credenziale.',
    zusammenstellen: 'Componi',
    sichern: 'Salva pacchetto',
    loeschen: 'Elimina pacchetto',
    groesse: '{bytes} su un massimo di {grenze}.',
    zeitraum: 'Le righe del registro vanno da {von} a {bis}.',
    ausgelassen: 'Omesso — {teil}:',
    teile: {
      fassung: 'Versione e piattaforma',
      bestand: 'Posizione e schema dell’archivio',
      pakete: 'Pacchetti di modelli installati',
      metriken: 'Contatori di questo processo',
      protokoll: 'Righe del registro',
    },
    spalteTeil: 'Parte',
    spalteDatei: 'File',
    spalteZeilen: 'Righe',
    spalteBytes: 'Dimensione',
    spalteMitnehmen: 'Includi',
    mitnehmen: 'Includi {teil} nel pacchetto',
  },
  diagnose: {
    ueberschrift: 'Diagnostica per una segnalazione di errore',
    erklaerung:
      'Quando qualcosa non funziona, questo testo aiuta a cercarne la causa. Viene mostrato qui — condividerlo spetta a te.',
    inhalt:
      'Contiene versione, computer, database, registro e pacchetti di modelli — nessun testo di carte, nessun valore di campo, nessun contenuto di set. Nel dettaglio: versione, tipo di computer, la build da cui proviene il programma, il contratto di binding tra finestra e programma, la posizione del database, lo stato dello schema, il numero di righe orfane, la posizione del registro, i tuoi pacchetti di modelli e le ultime 200 righe del registro. La tua cartella home compare come “~”.',
    anzeigen: 'Mostra la diagnostica',
    kopieren: 'Copia',
    kopiert: 'Copiato',
    kopierfehler:
      'La copia non è riuscita — la finestra non ha il permesso di accedere agli appunti. Seleziona il testo e copialo a mano.',
    sichern: 'Salva come file',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'Cerca nelle righe',
    zeilenGezeigt: '{gezeigt} di {alle} righe',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Tecnica e diagnostica',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Rete e raggiungibilità',

  ueberschrift: 'Strumenti senza finestra',
  erklaerung:
    'Questa applicazione include tre programmi che funzionano senza finestra. mse per la riga di comando, mse-mcp per un modello linguistico, mse-rest per un servizio per script; si trovano nel bundle dell’applicazione, entrano in un percorso di ricerca solo quando li configuri qui sotto, e le righe qui riportano la posizione completa.',
  keineShell:
    'Si copia, non si esegue: ciò che compare qui va in un terminale di tua scelta.',
  ort: 'Si trovano in {ort}.',
  vorhanden: 'Presenti: {namen}.',
  keineWerkzeuge:
    'Nessuno dei tre è presente. È normale finché l’applicazione gira dai sorgenti — la procedura di rilascio li inserisce nel bundle.',
  kopieren: 'Copia comando',
  kopiert: 'Copiato.',
  kopierfehler:
    'La copia non è riuscita — la finestra non ha il permesso di accedere agli appunti. Seleziona il testo e copialo a mano.',
  befehle: {
    hilfe: 'Cosa può fare la riga di comando',
    sets: 'Elenca i set nell’archivio',
    verknuepfen:
      'Rendi mse disponibile nel percorso di ricerca (chiede la tua password)',
    dienst: 'Avvia il servizio sull’interfaccia di loopback',
    mcp: 'Avvia il server MCP — per un modello linguistico, tramite stdio',
    karten: 'Elenca le carte del set aperto',
    statistik: 'Conta il set aperto',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'Indice di ricerca locale',
  suchindexErklaerung:
    'Testi delle carte e passaggi delle regole, ricercabili in locale — a testo completo e tramite vettori, uniti in un unico elenco. Ogni risultato indica da dove proviene e in quale revisione si trovava la fonte.',
  suchindexKeiner:
    'Non è stato creato alcun indice di ricerca. Fino ad allora la ricerca qui non trova nulla.',
  suchindexModell: 'Modello',
  suchindexLaufzeit: 'Runtime',
  suchindexVerfahren: 'Metodo',
  suchindexStaende: 'Revisione per fonte',
  suchindexSemantisch: 'Ricerca semantica',
  suchindexJa: 'sì',
  suchindexNein: 'no',
  suchindexFrage: 'Cosa cerchi?',
  suchindexSuchen: 'Cerca',
  suchindexLeer: 'Nessun risultato.',
  suchindexModus: 'Cercato',
  suchindexModusHybrid: 'a testo completo e tramite vettori',
  suchindexModusLexikalisch: 'solo a testo completo',
  suchindexModusOhneIndex: 'per nulla — nessun indice disponibile',
  suchindexAufbau: 'Ricostruisci',
  suchindexAufbauErklaerung:
    'Crea una nuova generazione accanto a quella in uso e passa a essa solo alla fine. Richiede minuti e un modello di embedding configurato.',
  suchindexModellfeld: 'Modello di embedding',
  suchindexQuellfeld: 'File in blocco',
  suchindexRevisionfeld: 'Revisione della fonte',
  suchindexDelta: 'Leggi solo ciò che è cambiato',
  suchindexAufbauen: 'Crea',
  suchindexAbbrechen: 'Annulla',
  suchindexLaeuft: 'Creazione in corso',
  suchindexGeschrieben: 'sezioni scritte',
  suchindexFertig: 'generazione creata',
  // **Configuration for agents** — A14, #1453.
  agenten: 'Pronto per un agente',
  agentenErklaerung:
    'Claude Code e Codex parlano MCP. I frammenti qui sotto riportano il percorso completo di mse-mcp su questo computer — copia e incolla. Questa applicazione non scrive in nessuno di questi file e non avvia alcun agente.',
  agentenOhneGeheimnis:
    'Non vi compare alcuna chiave di accesso. mse-mcp le legge da sé dal portachiavi — ciò che fosse scritto qui starebbe in chiaro in un file che si condivide.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Configura la riga di comando',
  befehlszeileSatz:
    'Crea collegamenti a mse e mse-mcp in una cartella di tua proprietà — senza diritti di amministratore e senza toccare il profilo della tua shell. Ciò che c’è già e non proviene da qui resta.',
  befehlszeileVerzeichnis: 'Cartella: {verzeichnis}',
  befehlszeileNichtImPfad:
    'Questa cartella non è nel tuo PATH. Aggiungi tu stesso questa riga al profilo della tua shell:',
  befehlszeileStaende: {
    frei: 'non ancora configurato',
    eigen: 'configurato',
    veraltet: 'obsoleto — punta a un programma che non c’è più',
    fremd: 'c’è un altro comando, che rimane',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    'Nascosto da {pfad}: è quello che viene eseguito quando digiti il nome.',
  befehlszeilePruefen: 'Verifica con',
  befehlszeileEinrichten: 'Configura',
  befehlszeileHerausnehmen: 'Rimuovi di nuovo',
  befehlszeileKopierenPfad: 'Copia la riga per il profilo della shell',
  befehlszeileKopierenPruefung: 'Copia il comando di verifica',
  agentenDatei: 'Va in {datei}',
  agentenInhalt: 'Contenuto per {datei}',
  agentenBefehl: 'Oppure tramite la riga di comando dell’agente',
  agentenNachgelesen: 'Verificato: {quelle}',
  agentenKopierenDatei: 'Copia frammento',
  agentenKopierenBefehl: 'Copia il comando dell’agente',

  peerTitel: 'Come questo profilo si identifica',

  peerZeigen: 'Mostra l’identità',
  peerErklaerung:
    'Una coppia di chiavi con cui gli altri riconoscono di parlare davvero con te. La parte privata resta nel portachiavi di questo computer e non va da nessuna parte; all’esterno va solo l’identificativo.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'Appartiene a questo profilo: non è il tuo Account Apple né il nome di questo computer.',
  peerStandOhne: 'Per questo profilo non esiste ancora un’identità.',
  peerStandBereit: 'Pronta.',
  peerStandVerloren: 'La chiave manca.',
  peerStandFremd: 'Portachiavi e archivio si contraddicono.',
  peerKennung: 'Identificativo',
  peerFingerabdruck: 'Impronta',
  peerFingerabdruckHinweis:
    'È ciò che leggi alla controparte — un nome visualizzato non è una prova.',
  peerSeit: 'Dal {zeitpunkt}',
  peerVorher: 'Prima: {kennung}',
  peerAnlegen: 'Crea identità',
  peerRotieren: 'Sostituisci identità',
  peerRotierenFrage:
    'Creare una nuova identità? Chiunque abbia confermato l’impronta attuale dovrà verificarla di nuovo.',
  peerRotierenJa: 'Sostituisci',
  peerRotierenNein: 'Annulla',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'Chi conosce questo profilo',
  kontakteErklaerung:
    'Controparti che questo profilo ha già incontrato. Un incontro da solo non consente nulla: solo il confronto dell’impronta rende qualcuno verificabile, e solo un’autorizzazione per scopo consente qualcosa.',
  kontakteLeer:
    'Non hai ancora incontrato nessuno. Chi si fa vivo comparirà qui in seguito — non confermato.',
  kontakteZeigen: 'Mostra i contatti',
  kontakteAufnehmen: 'Aggiungi contatto',
  kontakteAufnehmenErklaerung:
    'Finché non c’è una connessione, gli identificativi si scambiano a mano. Chi viene inserito qui non è confermato e non può fare nulla.',
  kontakteAufnehmenKennung: 'Identificativo della controparte',
  kontakteAufnehmenName: 'Come la chiami',
  kontakteStandUnbestaetigt: 'Non confermato',
  kontakteStandBestaetigt: 'Confermato',
  kontakteStandGesperrt: 'Bloccato',
  kontakteAbdruck: 'Impronta',
  kontakteNennt: 'Si fa chiamare: {name}',
  kontakteNenntHinweis: 'Un nome scelto da sé non è una prova.',
  kontakteGesehen: 'Visto l’ultima volta: {zeitpunkt}',
  kontakteVergleichen: 'Inserisci l’impronta che hai confrontato',
  kontakteBestaetigen: 'Conferma',
  kontakteSperren: 'Blocca',
  kontakteEntsperren: 'Sblocca',
  kontakteRechte: 'Autorizzato',
  kontakteKeineRechte: 'Nulla di autorizzato.',
  zweckPaarung: 'Fare conoscenza',
  zweckTeilen: 'Trasferire oggetti',
  zweckHandel: 'Scambiare carte',
  zweckPartie: 'Giocare una partita',
  kontakteAdresse: 'Raggiungibile a',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'Un indirizzo nella stessa rete — oppure un invito che la controparte ti ha inviato. Vuoto significa: nessuno.',
  kontakteAdresseMerken: 'Ricorda',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Nota',
  kontakteNotizBeispiel: 'scambia in modo corretto, risponde lentamente …',
  kontakteNotizHinweis:
    'Solo per te. Non viaggia mai — né in una vista né in una condivisione.',
  kontakteNotizMerken: 'Salva la nota',
  freigabenTitel: 'Cosa vede questa controparte',
  freigabenErklaerung:
    'Senza condivisione non vede nulla. Più ambiti valgono uno accanto all’altro — vince il più ampio.',
  freigabenLeer: 'Nessuna vista concessa.',
  freigabenOhneZweck:
    'Questa controparte non può ancora fare nulla. Consentile prima uno scopo.',
  freigabenZweck: 'Per cosa',
  freigabenUmfang: 'Cosa',
  freigabenErteilen: 'Concedi la vista',
  freigabenWiderrufen: 'Revoca',
  freigabenVorschau: 'Guarda cosa vede',
  freigabenVorschauLeer: 'Non vedrebbe nulla.',
  fernTitel: 'Essere raggiungibile da altre reti',
  fernErklaerung:
    'Così una controparte ti raggiunge anche quando non siete nella stessa rete — senza inoltro di porte e senza account. Ricevi un invito e glielo mandi, nel modo che preferisci.',
  fernStill: 'Questo profilo non è raggiungibile da altre reti.',
  fernHinweis:
    'La connessione cerca il percorso diretto e usa un relay solo quando non c’è altro modo. Attraverso il relay passa solo traffico cifrato — nessuna carta, nessun nome.',
  fernStarten: 'Crea invito',
  fernBeenden: 'Non essere più raggiungibile',
  fernWeitergeben:
    'Questo è ciò che serve alla controparte. Chi lo intercetta ottiene una linea, non un’identità — la verifica avviene al momento della connessione, tramite l’impronta.',
  fernStirbt:
    'L’invito vale finché questo profilo è raggiungibile. Se lo termini, diventa nullo — e la prossima volta ne riceverai uno nuovo.',
  fernKopieren: 'Copia invito',
  fernKopiert: 'Copiato.',
  horchenTitel: 'Essere raggiungibile da altri',
  horchenErklaerung:
    'Finché questo profilo è in ascolto, una controparte nota può stabilire ' +
    'una connessione con esso. Non succede mai da solo — nemmeno dopo un ' +
    'riavvio — e il cambio di profilo lo termina.',
  horchenAdresse: 'Indirizzo su cui restare in ascolto',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'La sola porta (:41234) ascolta su ogni interfaccia di questo computer. ' +
    'Con 127.0.0.1 davanti, solo su questo computer stesso.',
  horchenStarten: 'Diventa raggiungibile',
  horchenBeenden: 'Non essere più raggiungibile',
  horchenLaeuft: 'In ascolto su {adresse}',
  horchenNamen: 'Nome (resta valido dopo un cambio di indirizzo)',
  horchenAdressen: 'Indirizzo',
  horchenNurHier:
    'Con questo binding nessuno ti raggiunge dall’esterno: 127.0.0.1 ascolta solo su questo computer. Indica invece solo la porta (:41234), così questo Mac è raggiungibile nella sua rete.',
  horchenNurDiesesNetz:
    'Entrambi valgono in questa rete. Un indirizzo assegnato in prestito può cambiare, un nome di solito resiste — e un indirizzo valido ovunque arriva solo con la connessione da altre reti.',
  horchenStill: 'Questo profilo al momento non è in ascolto.',
  horchenWeitergeben:
    'Questo è ciò che serve alla controparte — insieme all’identificativo di questo profilo.',
  sichtbarTitel: 'Essere trovato nelle vicinanze',
  sichtbarErklaerung:
    'Annunciarsi dice a tutti nella stessa rete che questo profilo ' +
    'esiste — insieme al suo identificativo. Essere raggiungibile ed essere ' +
    'trovato sono due decisioni, e questa è la seconda.',
  sichtbarName: 'Nome con cui annunciarsi',
  sichtbarNameHinweis:
    'Un nome scelto da sé non prova nulla. Il nome host di questo Mac ' +
    'non viene diffuso.',
  sichtbarWerden: 'Annuncia',
  sichtbarBeenden: 'Non annunciare più',
  sichtbarLaeuft: 'Questo profilo si sta annunciando.',
  sichtbarStill: 'Questo profilo non si sta annunciando.',
  sichtbarBrauchtHorchen:
    'Prima diventa raggiungibile — un annuncio senza porta sarebbe ' +
    'un invito a una porta dietro cui non c’è nessuno.',
  nachbarnTitel: 'Chi c’è nelle vicinanze',
  nachbarnErklaerung:
    'Controparti che in questo momento si annunciano nella stessa rete. Ciò ' +
    'che vedi è la loro affermazione — viene verificata solo al momento della connessione.',
  nachbarnSuchen: 'Dai un’occhiata',
  nachbarnSucht: 'Ricerca in corso …',
  nachbarnLeer: 'Al momento nessuno si sta annunciando.',
  nachbarnUnbekannt: 'Sconosciuto',
  nachbarnAufnehmen: 'Aggiungi come contatto',
  nachbarnGemerkt: 'Ricordato come',
  nachbarnSchonBekannt: 'Già nell’elenco',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'Questo peer si trova a un nuovo indirizzo.',
  nachbarnAdresseBisher: 'Ricordato finora',
  nachbarnAdresseKeine: 'nessuno',
  nachbarnAdresseUebernehmen: 'Usa il nuovo indirizzo',
  nachbarnAdresseHinweis:
    'Questo cambia solo l’indirizzo a cui ci si connette. La fiducia nel peer resta quella di prima, e la sua chiave viene verificata a ogni connessione.',
  nachbarnVerworfen:
    '{zahl} risposte non contenevano un ID peer e non sono nell’elenco.',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: 'Dall’ambiente di questo processo',
  netzwegErklaerung:
    'Questi valori provengono dall’ambiente in cui è stato avviato il programma. Sostituiscono i campi qui sotto e nessuna finestra li modifica.',
  netzwegProxyHTTP: 'Proxy per http',
  netzwegProxyHTTPS: 'Proxy per https',
  netzwegProxyAus: 'Nessun proxy per',
  netzwegAusUmgebung: '{name}: {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'L’interfaccia per gli script',
  dienstRESTErklaerung:
    'Un servizio web sullo stesso archivio — per leggere e modificare da uno script. Gira solo finché lo lasci girare e termina con questa finestra.',
  dienstMCPTitel: 'Il server degli strumenti via HTTP',
  dienstMCPErklaerung:
    'Gli stessi strumenti che passano dallo standard input, qui a un indirizzo — raggiungibile anche da un altro dispositivo.',
  dienstRESTAdresse: 'Indirizzo dell’interfaccia per gli script',
  dienstRESTPort: 'Porta dell’interfaccia per gli script',
  dienstMCPAdresse: 'Indirizzo del server degli strumenti',
  dienstMCPPort: 'Porta del server degli strumenti',
  dienstMarke: 'Accesso',
  dienstMarkeKopieren: 'Copia accesso',
  dienstFingerabdruck: 'Impronta',
  dienstFingerabdruckKopieren: 'Copia impronta',
  dienstKopiert: 'Copiato',
  dienstStarten: 'Avvia',
  dienstBeenden: 'Termina',
  dienstErreichbar: 'Raggiungibile a:',
  dienstMitTLS:
    'Cifrato. Il certificato è autoemesso e riporta l’ID di questo profilo — chi si connette ne confronta l’impronta.',
  dienstOhneTLS:
    'Non cifrato, e qui va bene: l’indirizzo vale solo su questo computer.',
  dienstHinweis:
    '127.0.0.1 significa: solo questo computer. 0.0.0.0 significa: ogni dispositivo della rete — allora cifratura e accesso sono obbligatori, ed entrambi si configurano da soli. La porta 0 significa: cercarne una libera.',
}
