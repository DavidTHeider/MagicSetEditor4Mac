// The **Decks** page — Werkstück 7 of phase 4.6
// (#1293).
export const decks = {
  titel: 'Mazzi',
  erklaerung:
    'Composizioni di carte tue e stampe reali. Un mazzo rimanda alle carte invece di copiarle — una correzione nel set di origine lo raggiunge subito.',

  leer: 'Ancora nessun mazzo. «Crea mazzo», qui sopra, ne inizia uno. Un mazzo appartiene a un gioco; senza gioco non esisterebbe alcun formato con cui verificarlo.',

  neu: 'Crea mazzo',
  name: 'Nome',
  spiel: 'Gioco',
  notiz: 'Nota',
  anlegen: 'Crea',
  ausSetTitel: 'Oppure da un set',
  ausSetErklaerung:
    'Ogni carta del set diventa una voce del mazzo principale — richiamata, non copiata: se la correggi nel set, viene corretta anche nel mazzo.',
  ausSetSet: 'Set',
  ausSetMenge: 'Copie per carta',
  ausSetAnlegen: 'Crea dal set',
  ausSetKeinSet:
    'Non c’è ancora nessun set con carte — un mazzo creato da un set ne richiede uno.',

  hauptdeck: 'Mazzo principale',
  sideboard: 'Sideboard',
  zaehlung: '{haupt} nel mazzo principale, {side} nel sideboard',
  waehlen: 'Apri',
  offenMarke: 'aperto',
  umbenennen: 'Rinomina',
  neuerName: 'Nuovo nome',
  umbenennenTitel: 'Rinomina {name}',
  nameSichern: 'Salva',
  umbenennenAbbrechen: 'Mantieni il nome',
  duplizieren: 'Duplica',
  kopieTitel: 'Duplica {name}',
  kopieName: 'Nome della copia',
  kopieVorschlag: '{name} (copia)',
  kopieAnlegen: 'Crea copia',
  kopieAbbrechen: 'Non duplicare',
  aufWunschliste: 'Aggiungi alla lista dei desideri ciò che manca',
  wunschlisteErgebnis:
    'Nuove nella lista dei desideri: {neu} · già presenti: {schon}',
  loeschen: 'Elimina mazzo',
  handlungenAn: 'Azioni — {name}',
  handlungenKurz: '…',
  deckmenue: 'Mazzo',
  loeschenFrage: 'Eliminare questo mazzo? Le carte restano.',
  loeschenJa: 'Elimina',
  loeschenNein: 'Annulla',

  plaetze: 'Slot',
  plaetzeLeer: 'Ancora nessuna carta nel mazzo.',
  bezug: 'Carta o stampa',
  menge: 'Quantità',
  bereich: 'Sezione',
  setWaehlen: 'Set',
  karteWaehlen: 'Carta',
  karteSuchen: 'Cerca carta',
  karteKeine: 'Nessuna carta corrisponde a «{suche}».',
  setLeer:
    'Nessun set aperto — le carte dei tuoi set si potranno scegliere solo quando ce ne sarà uno.',
  ohneSet: 'Scegli un set …',
  ohneKarte: 'Scegli una carta …',
  setOhneKarten: 'Questo set non ha ancora carte.',
  kurve: 'Curva di mana',
  kurveHinweis:
    'Quante carte per valore di mana. Si conta ciò di cui il valore è noto: per una stampa di Scryfall quello registrato lì, altrimenti quello calcolato dal costo. Le terre non hanno costo e quindi non compaiono — così come una carta tua con il campo del costo vuoto.',
  farben: 'Colori',
  farbeWeiss: 'Bianco',
  farbeBlau: 'Blu',
  farbeSchwarz: 'Nero',
  farbeRot: 'Rosso',
  farbeGruen: 'Verde',
  farbenHinweis:
    'Una carta bicolore conta in entrambi i colori — il totale è quindi maggiore del mazzo.',
  karten: 'Carte',
  gebietFormate: 'Formati',
  formateKnopf: 'Formati …',

  // **Taking a deck from someone else's profile** — Werkstück 12 of Phase
  // 4.12 (#1351).
  uebernahme: 'Prendi un mazzo da un altro profilo',
  auswahlLoesen: 'Mostra tutti i mazzi',
  leiste: {
    bezeichnung: 'Cifre del mazzo',
    sideboard: 'Sideboard {zahl}',
    pruefung: 'Verifica',
    ohneFormat: 'senza formato',
    nichtGeprueft: 'non verificato',
    fehlt: 'Manca nella collezione',
    unbekannt: '{zahl} sconosciute',
    nichtAbgeglichen: 'non confrontato',
  },
  uebernahmeErklaerung:
    'Puoi prelevare da qualsiasi profilo in cui puoi passare. Se il suo proprietario usa un secondo fattore, serve una condivisione lì — basta «lettore». Non vengono trasferiti: gli slot che puntano a una carta propria di quel profilo e il formato — entrambi sono righe di là. Ciò che è rimasto indietro viene poi elencato qui.',
  uebernahmeProfil: 'ID dell’altro profilo',
  uebernahmeDeck: 'ID del mazzo',
  uebernahmeProfilwahl: 'Da quale profilo?',
  uebernahmeProfilWaehlen: 'Scegli un profilo …',
  uebernahmeProfilEintrag: '{nutzer} · {profil}',
  uebernahmeKeinProfil: 'Su questo computer non c’è nessun altro profilo.',
  uebernahmeDeckwahl: 'Quale mazzo?',
  uebernahmeDeckWaehlen: 'Scegli un mazzo …',
  uebernahmeKeinDeck: 'In questo profilo non c’è nessun mazzo.',
  uebernahmeFachleute: 'Indica tu gli ID',
  uebernahmeFachleuteErklaerung:
    'Profilo e mazzo si possono indicare anche tramite i loro ID — come li riportano la riga di comando e il servizio. Il permesso è lo stesso.',
  uebernahmeHolen: 'Prendilo',
  uebernahmeGeschrieben: '«{name}» è ora tra i tuoi mazzi — {karten}.',
  uebernahmeNichtMit: 'Non trasferito:',

  gebietKarten: 'Carte',
  gebietAuswertung: 'Analisi',
  gebietPruefung: 'Verifica',
  gebieteDeck: 'Aree di questo mazzo',
  legen: 'Aggiungi al mazzo',
  nehmen: 'Rimuovi',
  druck: 'Stampa (ID Scryfall)',
  anschlussOhneDeck:
    '«{name}» è in attesa — apri un mazzo o creane uno, poi comparirà nella sua selezione di carte.',
  holenAus: 'Prendi carte da',
  holenSammlung: 'La mia collezione',
  holenSet: 'Un set',
  wegeTitel: 'Prendere e passare',
  mitspielen: 'Gioca con questo mazzo',
  anschlussWartet:
    '«{name}» è arrivata dalla tua collezione — scegli quantità e sezione, poi va nel mazzo.',

  format: 'Formato',
  formatWaehlen: 'Scegli il formato',
  ohneFormat: 'Nessun formato — non verificato',
  formatHinweis:
    'Un mazzo senza formato non è verificato, non sbagliato. Il formato deve appartenere al gioco del mazzo.',

  pruefen: 'Verifica rispetto al formato',
  zulaessig: 'Legale.',
  unzulaessig: 'Non legale.',
  ungeprueftZahl: '{zahl} carte non sono verificate.',
  ungepruefterklaerung:
    'Non verificata non significa non legale: una carta che hai progettato tu non è elencata in nessun formato al mondo. Nemmeno «ristretta» va contro il mazzo — quante volte una carta ristretta può comparire è scritto nelle regole dei tornei, e questo programma non le ha lette.',

  legal: 'legale',
  not_legal: 'non elencata',
  restricted: 'ristretta',
  banned: 'bandita',
  ungeprüft: 'non verificata',

  bestand: 'Cosa possiedi',
  vorhanden: 'Possedute',
  gebraucht: 'Necessarie',
  fehlend: 'Mancanti',
  unbekannt: 'Sconosciuto',
  bestandVollstaendig: 'C’è tutto.',
  bestandFehlt: 'Mancano {zahl} carte.',

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
  gebietZuordnung: 'Assegnazione',
  zuordnungen: 'Fabbisogno e assegnazione',
  zuordnungenLaden: 'Caricamento …',
  zuordnungenLeer:
    'Ancora nessuna carta nel mazzo. Ciò che compare qui deriva dai suoi slot.',
  zuordnungHinweis:
    'Una lista mazzo non riserva nulla: due mazzi con le stesse tre carte sono entrambi completi. Assegnato è ciò che metti da parte per questo mazzo — libero è ciò che resta dopo.',
  zuordnungszahlen:
    '{vorhanden} di {gebraucht} possedute · {frei} libere · {zugeordnet} assegnate · {reserviert} riservate',
  zuordnungPlatz: 'Sezione',
  zuordnungExemplar: 'Esemplare',
  zuordnungExemplarWaehlen: 'Scegli un esemplare',
  zuordnungMenge: 'Quantità',
  zuordnen: 'Assegna',
  zuordnungLoesen: 'Rilascia',
  abweichend: 'altra stampa',
  abweichendErklaerung:
    'Questo esemplare è una stampa diversa da quella dello slot. Scelto deliberatamente — lo stesso nome non è la stessa carta.',
  fehlkarten: 'Carte mancanti come elenco',
  fehlkartenliste: 'Le carte mancanti',
  fehlkartenKeine: 'Non manca nulla.',

  typen: 'Tipi',
  kennzahlen: 'Indicatori',
  laender: 'Terre',
  durchschnitt: 'Valore di mana medio',
  keineZahl: '—',
  quellen: 'Fonti di colore',
  quellenHinweis:
    'Si contano le terre con uno dei cinque tipi base — producono il loro mana senza testo delle regole (CR 305.6). Una terra che nomina il proprio colore nel testo non è elencata qui: interpretarlo sarebbe una congettura.',

  weniger: '−',
  wenigerVon: 'Una in meno',
  mehr: '+',
  mehrVon: 'Una in più',
  insSideboard: 'Nel sideboard',
  insHauptdeck: 'Nel mazzo principale',
  suche: 'Cerca carte',
  deckbereich: 'Il mazzo',
  analyse: 'Analisi',
  analyseLeer:
    'Qui compaiono analisi, verifica e assegnazione del mazzo aperto.',
  stapelTitel: 'Dalla collezione',
  stapelWartet: '{name} dalla collezione, non ancora in un mazzo: {liste}.',
  stapelLegen: 'Metti tutto nel mazzo',
  stapelTeilweise: '{gelegt} di {gesamt} sono nel mazzo, le altre no:',
  stapelOhneDruck:
    '«{name}» non ha un ID Scryfall — un mazzo accetta solo stampe che ne hanno uno.',
  stapelGrund: '«{name}»: {grund}',
  eintauschen: 'Scambia le {anzahl}',
  deckplatz: 'Banco di costruzione del mazzo',
  pruefungOhneFormat:
    'Senza formato non c’è nulla da verificare. Scegline uno per il mazzo — fino ad allora è non verificato, non sbagliato.',
  pruefungLaeuft: 'Verifica in corso …',
  postenUnzulaessig: 'non legale',
  postenUnzulaessigErklaerung:
    'Questa carta non è ammessa nel formato scelto. Cosa esattamente lo dice la scheda «Verifica».',
  keinDeckOffen:
    'Nessun mazzo aperto. Sceglinene uno a sinistra o creane uno nuovo.',

  formate: 'Formati',
  formateLeer:
    'Ancora nessun formato. Qui un formato è un record e non un elenco integrato — si crea, non viene fornito.',
  bauart: 'Struttura',
  bauarten: 'Strutture documentate',
  bauartHinweis:
    'I numeri provengono dalle Comprehensive Rules. Sono un suggerimento; ciascuno si può modificare in seguito.',
  scryfallname: 'Nome Scryfall',
  scryfallnameHinweis:
    'La chiave nelle legalities di Scryfall, ad es. «standard». Vuoto significa che Scryfall non conosce questo formato, e ogni carta resta non verificata.',
  mindestgroesse: 'Almeno',
  hoechstgroesse: 'Al massimo',
  sideboardgroesse: 'Sideboard',
  sideboardOffen: 'Sideboard senza limite',
  hoechstzahlGleich: 'Stessa carta',
  grenzenHinweis:
    'Zero significa nessun limite. Per il sideboard, zero significa nessun sideboard.',
  formatNeu: 'Crea formato',
  formatSichern: 'Salva formato',
  formatLoeschen: 'Elimina formato',
  formatLoeschenFrage: 'Eliminare {name}?',
  formatLoeschenJa: 'Elimina',
  formatLoeschenHinweis:
    'I mazzi che lo usavano restano — dopo sono non verificati.',
  umschlagTitel: 'Condividi come busta',
  umschlagErklaerung:
    'Una busta porta un mazzo su un altro computer — firmata, con checksum e senza account. Ciò che arriva diventa un mazzo nuovo; niente viene sovrascritto. Il nome, il gioco e ogni voce con una stampa **viaggiano insieme**; la nota, la provenienza e le voci con carte proprie **restano qui** — il loro identificativo vale solo in questa libreria.',
  umschlagSchnueren: 'Salva il mazzo aperto',
  umschlagAnnehmen: 'Mazzo da una busta',
  umschlagOhneDeck: 'Apri prima un mazzo.',
  umschlagGeschnuert: 'Salvato.',
  umschlagAngenommen: 'Importato',
  einfuhrTitel: 'Importa una lista mazzo',
  einfuhrMenue: 'Importa una lista mazzo …',
  einfuhrErklaerung:
    'Incolla una lista mazzo — una carta per riga, «4 Lightning Bolt (M11) 149». Le intestazioni la suddividono: ciò che segue «Sideboard» è il sideboard. L’anteprima non scrive nulla; solo l’importazione mette le righe nel mazzo aperto, e non toglie nulla.',
  einfuhrFeld: 'Lista mazzo',
  einfuhrVorschau: 'Anteprima',
  einfuhrUebernehmen: 'Importa lista mazzo',
  einfuhrLeer: 'Incolla prima una lista.',
  einfuhrOffen: 'in sospeso',
  einfuhrBilanz: 'Mazzo principale / sideboard',
  einfuhrGelegt: 'Importato',
  einfuhrStueck: 'copie',
  einfuhrNichtsOffen: 'Nessuna riga è univoca — niente da inserire.',
} as const
