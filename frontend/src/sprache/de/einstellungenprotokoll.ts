// **Die Einstellungen des Protokolls** — OPS-04
// (#1721).
// Eine eigene Datei, weil `seiteneinstellungen.ts` an der Zeilengrenze
// steht (§2); sie wird dort eingesetzt und kommt als **ein** Katalog an.
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'Wie viel das Protokoll aufschreibt',
  'protokoll.datei': 'Protokolldatei',
  'protokoll.system': 'Unified Log von macOS',
  'protokoll.ausgabe': 'Standardfehlerausgabe',
  'protokoll.format': 'Form der Zeilen',
  'protokoll.grenze': 'Größe einer Protokolldatei in MB',
  'protokoll.aufbewahrung': 'Wie viele ältere Dateien bleiben',
  'protokoll.komponenten': 'Ausführlich nur für',
  'protokoll.ort': 'Ordner der Protokolldatei',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Nur Fehler',
  'protokoll.stufe.warnung': 'Warnungen und Fehler',
  'protokoll.stufe.info': 'Ereignisse (Vorgabe)',
  'protokoll.stufe.ausfuehrlich': 'Ausführlich — jeder Schritt',
  'protokoll.datei.an': 'An',
  'protokoll.datei.aus': 'Aus',
  'protokoll.system.an': 'An',
  'protokoll.system.aus': 'Aus',
  'protokoll.ausgabe.aus': 'Aus',
  'protokoll.ausgabe.an': 'An',
  'protokoll.format.text': 'Text zum Lesen',
  'protokoll.format.json': 'JSON für Werkzeuge',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'Was ins Protokoll kommt. **Ausführlich** hilft, einen Fehler zu finden — es hält dann höchstens 1 MB je Datei und eine ältere Datei, und die Diagnose sagt, dass es an ist. Gilt ab dem nächsten Start.',
  'protokoll.datei':
    'Die Datei `mse.log`, die du einem Fehlerbericht beilegst. Lässt sie sich nicht anlegen, schreiben die anderen Kanäle weiter, und die Diagnose nennt den Grund.',
  'protokoll.system':
    'Das Protokoll des Systems — zu lesen in der Konsole von macOS unter »MagicSetEditor4Mac«.',
  'protokoll.ausgabe':
    'Für den Start aus einem Terminal: jede Zeile auch dort, wo das Terminal sie zeigt.',
  'protokoll.format':
    '**JSON** ist eine Zeile je Ereignis mit benannten Feldern — für Werkzeuge, die Protokolle sammeln.',
  'protokoll.grenze':
    'Ab dieser Größe beginnt beim nächsten Start eine neue Datei. Leer heißt 4 MB; erlaubt sind 1 bis 64.',
  'protokoll.aufbewahrung':
    'So viele ältere Dateien bleiben neben der laufenden, von `mse.log.1` bis zur ältesten. Leer heißt eine; erlaubt sind 1 bis 9.',
  'protokoll.komponenten':
    'Nur bei **ausführlich**: die Schritte welcher Teile aufgeschrieben werden, etwa `app` für die Rufe des Fensters. Leer heißt alle; Ereignisse und Fehler kommen immer.',
  'protokoll.ort':
    'Ein Ordner mit / oder ~/ vorn. Leer heißt: neben dem Bestand.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
