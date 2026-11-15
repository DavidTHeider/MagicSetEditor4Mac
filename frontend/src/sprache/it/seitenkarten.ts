// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Aggiungi carta',
  nachOben: 'Sposta su',
  nachUnten: 'Sposta giù',
  duplizieren: 'Duplica carta',
  loeschen: 'Elimina carta',
  loeschenFrage: 'Eliminare questa carta? Con ⌘Z puoi annullare.',
  loeschenJa: 'Elimina',
  loeschenNein: 'Annulla',
  loeschenTrotzdem: 'Elimina comunque',
  symbole: 'Inserisci simboli',
  hergeleitet: 'Derivato',
  hergeleitetWert: 'Derivato: {wert}',
  bildWaehlen: 'Scegli immagine',
  stilfelder: 'Campi dello stile',
  stilfelderHilfe:
    'Questi campi arrivano con lo stile della carta. Il loro valore vale solo per questo stile.',
  symbolZeichnen: 'Disegna simbolo',
  kartenskripteTitel: 'Carte da uno script',
  kartenskripteErklaerung:
    'Azioni pronte del pacchetto di gioco — un clic aggiunge più carte in una volta. ' +
    'Un solo annulla le rimuove tutte.',
  kartenskriptEine: 'Una carta aggiunta.',
  kartenskriptMehrere: '{anzahl} carte aggiunte.',
  bildEinsetzen: 'Inserisci immagine',
  pack: {
    ueberschrift: 'Bustina casuale',
    erklaerung:
      'Una bustina con le carte di questo set, secondo le regole del pacchetto di gioco. Lo stesso seme dà la stessa estrazione — è così che un risultato si può riprodurre.',
    laedtGrund:
      'Ogni tipo di bustina chiede a ogni carta se è adatta — la prima volta ci vuole un momento, poi è pronto.',
    art: 'Tipo di bustina',
    same: 'Seme',
    sameZahl: 'Il seme è un numero intero da 0 a 18446744073709551615.',
    ziehen: 'Estrai',
    neuerSame: 'Nuovo seme',
    gezogen: 'Estratte',
    keineArten:
      'Il gioco di questo set non prevede tipi di bustina — oppure il suo pacchetto non è installato.',
    keineKarten: 'Questo tipo di bustina non ha prodotto alcuna carta.',
    artMitZahl: '{name} ({zahl} carte)',
    nurBaustein: 'Elemento di base — estratto da altri tipi di bustina',
    vorbehalte: 'Tipi di bustina il cui filtro non ha potuto essere valutato',
  },
  symbol: {
    ueberschrift: 'Simbolo',
    erklaerung:
      'Il simbolo di questo set — disegnato, non caricato. Compare subito nell’anteprima ed entra nel file quando il set viene salvato. Ogni forma si può posizionare e spostare con la tastiera.',

    grundform: 'Aggiungi una forma di base',

    verlauf: 'Cronologia del disegno',
    nachzeichnen: 'Ricalca un’immagine',
    groesser: 'Più grande',
    nachzeichnenHilfe:
      'Un PNG, JPEG o GIF con una silhouette netta diventa forme con punti di Bézier che puoi continuare a modificare qui.',
    zuruecknehmen: 'Annulla gesto',
    wiederherstellen: 'Ripristina gesto',
    kreis: 'Cerchio',
    rechteck: 'Rettangolo',
    dreieck: 'Triangolo',
    stern: 'Stella',

    teile: 'Parti',
    keineTeile:
      'Ancora nulla disegnato. Una forma di base è l’inizio — ogni punto si può poi modificare.',
    teilLoeschen: 'Elimina parte',
    teilMitPunkten: '{name} — {zahl} punti',
    teilMitKopien: '{name} — simmetria, {zahl} copie',
    gruppieren: 'Raggruppa',
    gruppierenHilfe:
      'Due o più parti selezionate diventano un gruppo. Tieni premuto ⌘ per selezionarne più di una.',
    aufloesen: 'Separa',
    nachOben: 'Sposta su',
    nachUnten: 'Sposta giù',
    nachObenFuer: 'Sposta {name} su',
    nachUntenFuer: 'Sposta {name} giù',
    teilLoeschenFuer: 'Elimina {name}',
    teilMitKindern: '{name} — {zahl} parti',
    teilKind: '{name}, in un gruppo',
    wahlHilfe:
      'Tieni premuto ⌘ per selezionare più parti. L’ordine è quello di disegno: ciò che sta per primo si trova sopra.',

    zeichnung: 'Area di disegno',
    punkt: 'Punto {nummer} di {name}',
    griffDavor: 'Maniglia prima del punto {nummer}',
    griffDanach: 'Maniglia dopo il punto {nummer}',
    punktHinzufuegen: 'Inserisci punto',
    punktLoeschen: 'Elimina punto',

    tastatur: 'Con la tastiera',
    tastaturhilfe:
      'Tab passa da un punto all’altro. I tasti freccia spostano di un centesimo, con Maiusc di un decimo. Invio inserisce un nuovo punto dopo quello selezionato, Backspace lo elimina.',

    kombination: 'Combina',
    ueberdecken: 'Sovrapponi',
    vereinen: 'Unisci',
    abziehen: 'Sottrai',
    schneiden: 'Interseca',
    unterschied: 'Differenza',
    umranden: 'Contorna',

    symmetrie: 'Simmetria',
    symmetrieAnlegen: 'Aggiungi simmetria',
    symmetrieAufloesen: 'Dissolvi simmetria',
    symmetrieErklaerung:
      'Una simmetria moltiplica ciò che contiene. Le copie compaiono in tenue sul disegno; si modifica sempre e solo l’originale.',
    symmetrieHinweis: 'Seleziona una parte per trasformarla in una simmetria.',
    symmetrieart: 'Tipo',
    drehung: 'Rotazione',
    achse: 'Riflessione',
    kopien: 'Copie',
    kopienHinweis:
      'Una riflessione richiede coppie — un numero dispari lascia una copia da sola, e MSE arrotonda per difetto.',
    mitteX: 'Centro, orizzontale',
    mitteY: 'Centro, verticale',
    griffX: 'Maniglia, orizzontale',
    griffY: 'Maniglia, verticale',
    mitte: 'Centro della simmetria',
    griff: 'Maniglia della simmetria',

    sichern: 'Inserisci nel set',
    gesichert:
      'Memorizzato come {name}. Entra nel file quando il set viene salvato.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Scelte',
    liste: 'Scegli da {liste}',
    keine: 'nessuna di queste',
    gruppe: '{name} (gruppo)',
  },
  regeln: {
    ueberschrift: 'Chiarimenti da Scryfall',
    holen: 'Scarica i chiarimenti',
    keine: 'Per questa carta non ci sono chiarimenti.',
    offiziell: 'ufficiale',
    anmerkung: 'Nota di Scryfall',
    anmerkungErklaerung:
      'Le note provengono dal team di Scryfall e spesso spiegano un formato non ufficiale. Non sono regole.',
    quelle: 'Vedi questa carta su Scryfall',
    alter: 'Stato: vecchio di {stunden} ore',
    manaPruefen: 'Controlla il costo di mana',
    manaGut: 'Il costo di mana è a posto.',
    manaSchlecht: 'Sconosciuti: {symbole}',
    manaLeer: 'Questa carta non ha costo di mana.',
  },
  abgleich: {
    ueberschrift: 'Dati della carta da Scryfall',
    erklaerung: 'Mostra cosa cambierebbe. Viene scritto solo ciò che spunti.',
    holen: 'Scarica la proposta',
    uebernehmen: 'Applica le voci spuntate',
    nichtsZuTun: 'Niente da cambiare — la carta corrisponde a Scryfall.',
    alt: 'attuale',
    neu: 'nuovo',
    leer: '(vuoto)',
    alter: 'Stato: vecchio di {stunden} ore',
    frisch: 'Stato: appena scaricato',
    quelle: 'Vedi questa carta su Scryfall',
  },
  scryfall: {
    ueberschrift: 'Immagine da Scryfall',
    erklaerung:
      'Il nome della carta sceglie l’immagine. L’artista viene scritto anche nella carta — Scryfall richiede che il nome stia accanto all’illustrazione.',
    name: 'Nome della carta su Scryfall',
    art: 'Tipo di immagine',
    artKlein: 'Piccola (anteprima)',
    artNormal: 'Carta intera',
    artAusschnitt: 'Solo l’illustrazione',
    holen: 'Scarica immagine',
    ausschnittHinweis:
      'Il ritaglio mostra solo l’illustrazione. Viene rifiutato se Scryfall non indica l’artista.',
    quelle: 'Vedi questa carta su Scryfall',
  },
  scryfallkarte: {
    name: 'Nome della carta su Scryfall',
    platzhalter: 'ad es. Lightning Bolt',
    mitBild: 'con illustrazione',
    anlegen: 'Crea da Scryfall',
  },
  zuschnitt: {
    ueberschrift: 'Scegli il ritaglio',
    erklaerung:
      'Il riquadro indica quale parte dell’immagine finisce sulla carta. Trascinalo per spostarlo, trascina un angolo per ridimensionarlo.',
    quelle: 'Immagine di origine con il ritaglio scelto',
    vorschau: 'Anteprima alla dimensione del campo',
    links: 'Sinistra',
    oben: 'In alto',
    breite: 'Larghezza',
    hoehe: 'Altezza',
    seitenverhaeltnis: 'Mantieni le proporzioni del campo',
    uebernehmen: 'Inserisci',
    abbrechen: 'Annulla',
    masse: '{breite} × {hoehe} pixel',
  },
  feldWaehlen: 'Modifica questo campo',
  leerTitel: 'Ancora nessuna carta',
  listeBezeichnung: 'Carte del set',
  mehr: 'Altro',
  gruppeAnlage: 'Carta',
  gruppeAnsicht: 'Vista',
  ausScryfallAufklapp: 'Crea da Scryfall …',
  ausScryfallErklaerung:
    'Invece di una carta vuota: scrivi un nome o cerca su Scryfall — i campi verranno da lì.',
  dieseKarte: 'Questa carta',
  kartenmenue: 'Carta',
  auswahlLoesen: 'Mostra tutte le carte',
  bezugAufklapp: 'Collega a una stampa …',
  dieseKarteLeer: 'Apri una carta per modificarla.',
  keineKarteTitel: 'Nessuna carta selezionata',
  titel: 'Carte',
  leer: 'Creane una o incolla carte copiate.',
  ohneName: 'Senza nome',
  ohneFelder: 'Questa carta non ha alcun valore nei campi.',
  nummer: 'Numero',
  vorschau: 'Anteprima della carta',
  keineKarte:
    'Scegline una dall’elenco — qui compariranno i suoi campi e la sua anteprima.',
  bearbeiten: 'Modifica',

  notizen: {
    ueberschrift: 'Note sulla carta',
    platzhalter: 'Solo per te — non compare mai sulla carta.',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'Cosa sa l’applicazione al riguardo',
  tiefe: 'Quanto vedi',
  tiefeNormal: 'Normale',
  tiefeErweitert: 'Avanzato',
  tiefeExperte: 'Esperto',
  tiefeErweitertTitel: 'Ciò che riempie uno script',
  tiefeErweitertHinweis:
    'Il pacchetto di gioco calcola questi campi — dicono perché qualcosa si trova sulla carta. Non si possono modificare.',
  befundVerdeckt: 'Un rilievo si trova sotto “{tiefe}”.',
  befundSpringen: 'Vai lì',
  buchfuehrungHinweis:
    'Queste voci non appartengono alla carta ma alla sua gestione — quale modello usa e cosa è impostato. Modificale solo se sai perché.',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'Stato di lavoro di questa carta',
  arbeitsstandHinweis:
    'Nota e contrassegni — appartengono al tuo lavoro sulla carta, non alla carta. Un file che condividi non li include.',

  marken: {
    ueberschrift: 'Contrassegni',
    erklaerung:
      'Parole per il tuo lavoro — stanno accanto alla carta, non al suo interno. Un file che condividi non li include.',
    platzhalter: 'ad es. finito',
    setzen: 'Aggiungi contrassegno',
    entfernen: 'Rimuovi il contrassegno {marke}',
    keine: 'Ancora nessun contrassegno su questa carta.',
    schonGesetzt:
      'La carta porta già «{marke}», aggiunta altrove nel frattempo.',
    filter: 'Filtra per contrassegno',
    alle: 'Tutti',
  },

  scryfallsuche: {
    frage: 'Cerca su Scryfall',
    platzhalter: 'ad es. Lightning Bolt — oppure t:creature c:r cmc<=3',
    erklaerung:
      'Basta il nome di una carta. Se conosci il linguaggio di ricerca di Scryfall, puoi usarlo qui.',
    suchen: 'Cerca',
    nichtsGefunden: 'Nessun risultato. Prova un’altra parola?',
    mehr: 'Scryfall ha più risultati. Restringi la ricerca — non viene scaricato nulla in automatico.',
    alter: 'Stato: vecchio di {stunden} ore',
    frisch: 'Stato: appena scaricato',
    befund: 'Rilievo: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Collegata a',
    keiner:
      'Questa carta non è collegata a un’edizione. La sincronizzazione con Scryfall ne sceglierà una qualsiasi — cerca qui sopra e collegala.',
    druck: 'da {setcode} {nummer}',
    gesetzt: '— collegata il {wann}',
    setzen: 'Collega a {name} ({setcode} {nummer})',
    loesen: 'Scollega',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Copia ({zahl})',
    ausschneiden: 'Taglia ({zahl})',
    einfuegen: 'Incolla',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'Colonne dell’elenco',
    jeSpiel: 'La scelta vale per tutti i set dello stesso gioco.',
    filtern: 'Trova colonna',
    gezeigt: '{gezeigt} di {alle} colonne',
    ausgeblendet:
      'Altre {anzahl} sono nascoste — cercale con il campo qui sopra.',
    genauer: 'Ne corrispondono altre — affina la ricerca per vederle.',
    nachOben: 'sposta più a sinistra',
    nachUnten: 'sposta più a destra',
    sortieren: 'Ordina per',
    garNicht: 'Ordine del set',
    aufsteigend: 'Crescente',
    absteigend: 'Decrescente',
    ohneVorlage:
      'Il pacchetto di gioco non può essere letto. Le colonne provengono dalle carte stesse — senza intestazioni e senza l’ordine del pacchetto.',
    spalteAusblenden: 'Nascondi colonna',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— scegli un campo —',
    keinWert: '— scegli un valore —',
    ueberschrift: 'Modifica in blocco',
    erklaerung:
      'Imposta un campo su tutte le carte spuntate — in una sola operazione. Se qualcosa va storto, non ne viene modificata nemmeno una.',
    auswaehlen: 'Spunta per la modifica in blocco',
    mehrereWaehlen: 'Sceglierne più',
    ausgewaehlt: 'spuntate',
    keineAuswahl:
      'Nessuna carta spuntata. Spunta più carte nell’elenco per modificarle insieme.',
    feld: 'Campo',
    wert: 'Nuovo valore',
    anwenden: 'Applica',
    betroffen:
      'La modifica è fallita per queste carte. Nessuna carta è stata modificata:',
    gescheitert: 'non riuscita',
    ohneWert: 'ancora nessun valore',
  },
  stilTitel: 'Stile',
  stilWaehlen: 'Stile di questo set',
  stilLeer:
    'Nessuno stile installato. Senza stile l’anteprima ripiega su un sostituto.',
  stilFrage: 'Passare a questo stile?',
  stilSpielwechsel: 'Il nuovo stile appartiene a un altro gioco.',
  stilUngedeutet:
    'Il nuovo gioco non conosce questi campi. I loro valori restano nell’archivio — ma non verranno più disegnati:',
  stilVerloreneOptionen:
    'Il nuovo stile non ha queste opzioni di stile. I loro valori vanno persi:',
  stilUebernehmen: 'Cambia stile',
  stilAbbrechen: 'Mantieni lo stile',
  stilKeiner:
    'Il set indica uno stile che non è installato — scegline uno tra quelli installati.',
  stiloptionen: 'Opzioni di stile',
  stiloptionenSichtbar: '{zahl} di {gesamt}',
  stiloptionenWeitere: 'Altre opzioni ({zahl})',
  marktTitel: 'Dove trovare questa carta',
  marktErklaerung:
    'Questi collegamenti vengono creati sul tuo computer — dal nome della carta o da un identificativo del fornitore. Non esce nulla finché non fai clic, nel browser che preferisci. Non serve alcuna chiave; se lì ci sia qualcosa in vendita lo decide il fornitore.',
  marktGenau: 'porta esattamente a questa carta',
  marktSuche: 'porta a una ricerca',
  marktNurAnbieter: 'porta solo al fornitore',
  umschlagAnnehmen: 'Carta da una busta',
  umschlagAngenommen: 'La carta è ora in questo set.',
}
