// **Laufzeiten, Verbrauch und das Leistungsprofil** — OPS-05
// (#1722).
// Eine eigene Datei, weil `seitentechnik.ts` an der Zeilengrenze steht (§2);
// sie wird dort eingesetzt und kommt als **ein** Katalog an.

export const betrieb = {
  ueberschrift: 'Laufzeiten und Verbrauch',
  zeitraum: 'Erhoben von {seit} bis {bis}.',
  aufwand: 'Aufwand der Erhebung: {mikros} µs',
  speicher: 'Arbeitsspeicher',
  speicherWert:
    '{fussabdruck} Fußabdruck — davon {heap} Go-Heap, {system} vom System',
  speicherOhneFussabdruck:
    '{heap} Go-Heap, {system} vom System — der Fußabdruck ist nicht gemessen',
  web: 'Webprozesse',
  webWert: '{summe} zusammen — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Webinhalt', gpu: 'Grafik', netz: 'Netz' },
  ohneWeb: 'Dieser Prozess hält keine Webansicht.',
  webUngemessen: 'Nicht gemessen: {grund}',
  goroutinen: 'Nebenläufige Abläufe',
  rechenzeit: 'Rechenzeit seit dem Start',
  datenbank: 'Verbindungen zum Bestand',
  datenbankWert:
    '{offen} offen, {inBenutzung} in Benutzung, {gewartet}-mal gewartet, zusammen {dauer}',
  ohneBestand: 'Dieser Prozess hält keinen Bestand.',
  zwischenspeicher: 'Zwischenspeicher fremder Dienste',
  zwischenspeicherWert:
    '{treffer} Treffer, {geholt} geholt, {veraltet} veraltet ausgeliefert, {fehlgeschlagen} fehlgeschlagen',
  laufzeiten: 'Laufzeiten je Fläche',
  laufzeitenLeer:
    'Hier sind keine Rufe gemessen — gemessen werden nur der Dienst und der Werkzeugserver.',
  ungemessen: 'Nicht gemessen: {liste}.',
  ungemessenNamen: {
    fenster:
      'die Bindungen des Fensters (sie haben keine gemeinsame Stelle, an der ein Ruf endet)',
    kommandozeile: 'die Kommandozeile (sie lebt einen Befehl lang)',
    speicher: 'der Fußabdruck dieses Prozesses (das System gab ihn nicht her)',
  },
  spalteArt: 'Art',
  spalteAnzahl: 'Rufe',
  spalteFehler: 'Fehlerrate',
  spalteMittel: 'Mittel',
  spalteVerteilung: 'Verteilung',
  arten: { lesend: 'lesend', schreibend: 'schreibend' },
  eimer: 'bis {grenze} ms',
  verteilung: '{eimer}: {anzahl} Rufe',
  eimerOffen: 'darüber',
  spurAus:
    'Die Ablaufspur ist aus. Einschalten unter Einstellungen › Diagnose — sie gilt ab dem nächsten Start.',
  spurAn:
    'Ablaufspur an: {aufgenommen} von {gesehen} Rufen, {abtastung} % abgetastet. Höchstens {grenze} Einträge für {minuten} min — {verdraengt} verdrängt, {abgelaufen} abgelaufen.',
}

export const leistungsprofil = {
  ueberschrift: 'Leistungsprofil',
  erklaerung:
    'Ein CPU-Profil dieses Fensters über die gewählte Zeit, im Format von »go tool pprof«. Es liegt danach neben dem Bestand; wie viele liegen bleiben, steht unter Einstellungen › Diagnose.',
  sekunden: 'Dauer in Sekunden',
  grenze: 'Eine ganze Zahl von 1 bis 60.',
  aufzeichnen: 'Profil aufzeichnen',
  laeuft: 'Zeichnet auf — das Fenster bleibt bedienbar.',
  fertig: 'Geschrieben: {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Zwischenspeicher',
  erklaerung:
    'Was diese Anwendung aufhebt, um es nicht zweimal zu holen oder zu rechnen — auf der Platte und im Arbeitsspeicher. Nichts davon sind deine Daten: geleert, füllt sich jeder Speicher beim nächsten Gebrauch von selbst.',
  spalteSpeicher: 'Speicher',
  spalteOrt: 'Ort',
  spalteQuote: 'Trefferquote',
  spalteEintraege: 'Einträge',
  spalteGroesse: 'Größe',
  spalteGrenze: 'Grenze',
  keineQuote: 'noch keine Abfrage',
  keineGrenze: 'eine Datei, die ersetzt wird',
  namen: {
    'scryfall.antworten': 'Antworten von Scryfall',
    'scryfall.bilder': 'Kartenbilder von Scryfall',
    'scryfall.sammeldaten': 'Sammeldatei von Scryfall',
    'render.ebenen': 'Abgetastete Ebenen',
    'pakete.bilder': 'Bilder der Vorlagenpakete',
    'skript.baeume': 'Zerteilte Skripte',
    'vorschau.bilder': 'Kartenvorschauen',
    'tisch.texturen': 'Kartenbilder am Tisch',
  },
  orte: { platte: 'Platte', arbeitsspeicher: 'Arbeitsspeicher' },
  leerenHinweis:
    'Leeren wirft alles weg. Ohne Netz kommen die Antworten und Bilder von Scryfall erst wieder, wenn es zurück ist.',
  leeren: 'Alle leeren',
  geleert: 'Geleert — sie füllen sich beim nächsten Gebrauch.',
}
