// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'Modelli',
  erklaerung:
    'Ciò che una carta sa fare e come appare si trova in file di testo dentro il pacchetto di modelli — campi, script, misure. Qui si possono modificare. Lo stato precedente resta come backup accanto al pacchetto.',
  paket: 'Pacchetto',
  keinPaket: 'Nessun pacchetto ancora scelto.',
  waehleEines:
    'Scegli un pacchetto qui sopra per vedere e modificare i suoi file di testo — campi, script, misure.',
  keineDateien: 'Questo pacchetto non ha file da modificare.',
  dateien: 'File',
  datei: 'File',
  inhalt: 'Contenuto',
  speichern: 'Salva',
  gespeichert: 'Salvato.',
  ungespeichert: 'Non salvato.',
  warnung:
    'Questo file vale per ogni set che usa questo pacchetto — non solo per quello aperto.',
  konflikt:
    'Il file è cambiato dopo che lo hai aperto. Carica la sua versione — il tuo testo va perso — oppure tieni la tua: le modifiche mostreranno allora cosa sovrascrivi nella sua.',
  neuLaden: 'Carica la sua versione',
  behalten: 'Tieni la mia',
  vorschau: 'Anteprima',
  vorschauErklaerung:
    'La prima carta del primo set che usa questo modello — disegnata con il testo mostrato qui. Nulla viene salvato in questo modo.',
  vorschauLaeuft: 'Disegno in corso …',
  vorschauKeineKarte:
    'Nessun set aperto usa questo modello. Apri in «Set» uno che lo nomina — così qui vedrai cosa comporta una modifica.',

  arbeitsplatz: 'Officina dei modelli',
  baumBezeichnung: 'Pacchetti e file',
  inspectorBezeichnung: 'Anteprima e diagnosi',
  baumSuche: 'Cerca pacchetti e file',
  baumSuchePlatzhalter: 'Pacchetto o file …',
  baumLeer: 'Nessun pacchetto e nessun file corrisponde alla ricerca.',
  baumOhnePakete:
    'Ancora nessun pacchetto da modificare. Creane uno in «Pacchetti» o scaricane uno — poi comparirà qui.',
  nurLesbar: 'Incluso — sola lettura. Una copia tua si può modificare.',
  baumNurMitgeliefert:
    'Tutti i pacchetti qui sono inclusi. Crea una tua copia di uno, oppure uno nuovo in «Pacchetti» — entrambi si possono modificare e compaiono in alto.',
  dateienLaden: 'Lettura dei file …',
  inspectorLeer:
    'Qui compaiono anteprima, diagnostica e origine del file aperto.',
  waehleImBaum:
    'Scegli un file nell’albero a sinistra — campi, script e misure vi si trovano come testo e si possono modificare qui.',
  verwerfenFrage:
    'Il file aperto ha modifiche non salvate. Scartarle e aprire {datei}?',
  verwerfen: 'Scarta e apri',
  hierBleiben: 'Resta qui',
  testkarte: 'Carta di prova',
  testkarteVorgabe: 'Prima carta del primo set',
  testkarteOhneNamen: 'Carta {nummer}',
  testkarteErklaerung:
    'Un modello raramente si rompe sulla prima carta. Il set di esempio «Werkstatt-Testkarten» porta con sé testo lungo, un layout speciale, un secondo lato, un’immagine mancante e caratteri speciali — aprilo in «Set» e passa il suo stile a questo pacchetto.',
  veraltet:
    'Non aggiornato — il file ha una segnalazione alla riga {zeile}; viene mostrato l’ultimo stato valido.',
  wirdNeuGezeichnet: 'Ridisegno in corso …',
  diagnose: 'Diagnosi',
  diagnoseFrei: 'Nessuna segnalazione in questo file.',
  feldartFehlt:
    'A questo campo manca «type:». MSE2 interrompe qui la lettura del pacchetto; questa app lo tratta come campo di testo.',
  feldartUnbekannt:
    'MSE2 non conosce il tipo di campo «{0}» e interrompe qui la lettura del pacchetto. Questa app tratta il campo come testo. Tipi noti: {1}.',
  diagnoseStelle: 'Riga {zeile}, colonna {spalte}',
  zurStelle: 'Vai alla posizione',
  herkunft: 'Da dove viene un nome',
  herkunftKeine:
    'Posiziona il cursore su un nome — qui compare da dove viene e chi lo usa.',
  herkunftUnbekannt:
    '«{name}» non è definito da questo pacchetto — il nome proviene dalla libreria o da un pacchetto incluso.',
  herkunftStelle: '{art} in {datei}, riga {zeile}',
  zurDefinition: 'Vai alla definizione',
  verwendungen: 'Riferimenti',
  verwendungenKeine: 'Nessun altro punto nomina questo.',
  verwendungLaden: 'Ricerca dei riferimenti …',
  symbolarten: {
    'card field': 'Campo della carta',
    'set field': 'Campo del set',
    'styling field': 'Opzione di stile',
    'extra card field': 'Campo aggiuntivo',
    script: 'Nome dello script',
  } as Record<string, string>,
  symbolartSonst: 'Definito qui',
  aenderungen: 'Modifiche prima del salvataggio',
  aenderungenZeigen: 'Mostra modifiche',
  aenderungenVerbergen: 'Nascondi modifiche',
  aenderungenKeine: 'Nessuna modifica rispetto allo stato letto.',
  aenderungenZusammenfassung: 'Righe aggiunte: {neu}, rimosse: {weg}.',
  aenderungenAbschnitt: 'Modifica {nummer}',
  zeileWeg: 'rimossa',
  zeileNeu: 'aggiunta',
  kopieHinweis: 'Origine del pacchetto',
  kopieren: 'Crea copia personale',
  kopieMitgeliefert:
    'Questo pacchetto è fornito con l’app e non può essere modificato. Crea una tua copia — poi comparirà tra i tuoi pacchetti, e l’originale resta com’è.',
  kopieFremd:
    'Questo pacchetto proviene dall’esterno. Una modifica qui conta come modifica locale e al prossimo download entra in conflitto con la nuova versione. Crea una tua copia se deve restare.',
  kopieOriginal:
    'L’originale resta. Stai lavorando su una copia — {original} stesso è invariato.',
  dateiaktionen: 'File del pacchetto',
  dateiNeu: 'Nuovo file',
  dateiName: 'Nome del nuovo file',
  dateiAnlegen: 'Crea',
  dateiAbbrechen: 'Annulla',
  dateiHinzufuegen: 'Aggiungi immagine o font …',
  dateiUmbenennen: 'Rinomina',
  dateiNeuerName: 'Nuovo nome',
  dateiUmbenennenJa: 'Rinomina',
  dateiLoeschen: 'Elimina …',
  dateiLoeschenFrage:
    'Eliminare? Il file si troverà poi nel backup accanto al pacchetto.',
  dateiLoeschenJa: 'Sì, elimina',
  dateiErstSichern: 'Prima salva o scarta.',
  beiwerk: 'Risorsa',
  beiwerkHinweis:
    'Non è testo — il pacchetto usa questo file così com’è. Qui si può rinominare ed eliminare.',
  beiwerkZeile: 'Risorse ({anzahl})',
  weitereZeigen: 'Mostra altri ({gezeigt} di {gesamt})',
} as const
