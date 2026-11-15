// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: 'Tempi di esecuzione e consumo',
  zeitraum: 'Rilevato da {seit} a {bis}.',
  aufwand: 'Costo della rilevazione: {mikros} µs',
  speicher: 'Memoria',
  speicherWert:
    '{fussabdruck} di impronta — di cui {heap} di heap Go, {system} dal sistema',
  speicherOhneFussabdruck:
    '{heap} di heap Go, {system} dal sistema — l’impronta non è misurata',
  web: 'Processi web',
  webWert: '{summe} in totale — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Contenuto web', gpu: 'Grafica', netz: 'Rete' },
  ohneWeb: 'Questo processo non contiene alcuna vista web.',
  webUngemessen: 'Non misurato: {grund}',
  goroutinen: 'Routine concorrenti',
  rechenzeit: 'Tempo CPU dall’avvio',
  datenbank: 'Connessioni all’archivio',
  datenbankWert:
    '{offen} aperte, {inBenutzung} in uso, attesa {gewartet} volte, {dauer} in totale',
  ohneBestand: 'Questo processo non gestisce alcun archivio.',
  zwischenspeicher: 'Cache dei servizi esterni',
  zwischenspeicherWert:
    '{treffer} riscontri, {geholt} recuperati, {veraltet} serviti obsoleti, {fehlgeschlagen} falliti',
  laufzeiten: 'Tempi di esecuzione per interfaccia',
  laufzeitenLeer:
    'Qui non è misurata alcuna chiamata — vengono misurati solo il servizio e il server degli strumenti.',
  ungemessen: 'Non misurato: {liste}.',
  ungemessenNamen: {
    fenster:
      'i binding della finestra (non hanno un punto comune in cui una chiamata termina)',
    kommandozeile: 'la riga di comando (vive per la durata di un comando)',
    speicher: 'l’impronta di questo processo (il sistema non l’ha fornita)',
  },
  spalteArt: 'Tipo',
  spalteAnzahl: 'Chiamate',
  spalteFehler: 'Tasso di errore',
  spalteMittel: 'Media',
  spalteVerteilung: 'Distribuzione',
  arten: { lesend: 'lettura', schreibend: 'scrittura' },
  eimer: 'fino a {grenze} ms',
  verteilung: '{eimer}: {anzahl} chiamate',
  eimerOffen: 'oltre',
  spurAus:
    'La traccia è disattivata. Attivala in Impostazioni › Diagnostica — vale dal prossimo avvio.',
  spurAn:
    'Traccia attiva: {aufgenommen} di {gesehen} chiamate, {abtastung} % campionato. Al massimo {grenze} voci per {minuten} min — {verdraengt} scartate, {abgelaufen} scadute.',
}

export const leistungsprofil = {
  ueberschrift: 'Profilo delle prestazioni',
  erklaerung:
    'Un profilo CPU di questa finestra per il tempo scelto, nel formato di “go tool pprof”. Viene salvato accanto all’archivio; quanti ne restano si imposta in Impostazioni › Diagnostica.',
  sekunden: 'Durata in secondi',
  grenze: 'Un numero intero da 1 a 60.',
  aufzeichnen: 'Registra profilo',
  laeuft: 'Registrazione in corso — la finestra resta utilizzabile.',
  fertig: 'Scritto: {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Cache',
  erklaerung:
    'Ciò che questa applicazione conserva per non recuperarlo o calcolarlo due volte — su disco e in memoria. Niente di tutto ciò sono i tuoi dati: una volta svuotata, ogni cache si riempie di nuovo al prossimo utilizzo.',
  spalteSpeicher: 'Cache',
  spalteOrt: 'Dove',
  spalteQuote: 'Percentuale di riscontri',
  spalteEintraege: 'Voci',
  spalteGroesse: 'Dimensione',
  spalteGrenze: 'Limite',
  keineQuote: 'ancora nessuna richiesta',
  keineGrenze: 'un file che viene sostituito',
  namen: {
    'scryfall.antworten': 'Risposte di Scryfall',
    'scryfall.bilder': 'Immagini delle carte da Scryfall',
    'scryfall.sammeldaten': 'File bulk di Scryfall',
    'render.ebenen': 'Livelli campionati',
    'pakete.bilder': 'Immagini dei pacchetti di modelli',
    'skript.baeume': 'Script analizzati',
    'vorschau.bilder': 'Anteprime delle carte',
    'tisch.texturen': 'Immagini delle carte al tavolo',
  },
  orte: { platte: 'Disco', arbeitsspeicher: 'Memoria' },
  leerenHinweis:
    'Svuotare elimina tutto. Senza rete, le risposte e le immagini di Scryfall tornano solo quando la rete è di nuovo disponibile.',
  leeren: 'Svuota tutto',
  geleert: 'Svuotate — si riempiono di nuovo al prossimo utilizzo.',
}
