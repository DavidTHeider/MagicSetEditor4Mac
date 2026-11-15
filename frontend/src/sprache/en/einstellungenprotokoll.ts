// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'How much the log records',
  'protokoll.datei': 'Log file',
  'protokoll.system': 'macOS Unified Log',
  'protokoll.ausgabe': 'Standard error',
  'protokoll.format': 'Line format',
  'protokoll.grenze': 'Size of a log file in MB',
  'protokoll.aufbewahrung': 'How many older files stay',
  'protokoll.komponenten': 'Detailed only for',
  'protokoll.ort': 'Folder of the log file',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Errors only',
  'protokoll.stufe.warnung': 'Warnings and errors',
  'protokoll.stufe.info': 'Events (default)',
  'protokoll.stufe.ausfuehrlich': 'Detailed — every step',
  'protokoll.datei.an': 'On',
  'protokoll.datei.aus': 'Off',
  'protokoll.system.an': 'On',
  'protokoll.system.aus': 'Off',
  'protokoll.ausgabe.aus': 'Off',
  'protokoll.ausgabe.an': 'On',
  'protokoll.format.text': 'Text to read',
  'protokoll.format.json': 'JSON for tools',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'What goes into the log. **Detailed** helps to find a problem — it then keeps at most 1 MB per file and one older file, and the diagnostics say that it is on. Takes effect after the next start.',
  'protokoll.datei':
    'The file `mse.log` you attach to a bug report. If it cannot be created, the other channels keep writing, and the diagnostics name the reason.',
  'protokoll.system':
    'The system log — read it in the macOS Console under “MagicSetEditor4Mac”.',
  'protokoll.ausgabe':
    'For starting from a terminal: every line also where the terminal shows it.',
  'protokoll.format':
    '**JSON** is one line per event with named fields — for tools that collect logs.',
  'protokoll.grenze':
    'From this size a new file begins at the next start. Empty means 4 MB; 1 to 64 are allowed.',
  'protokoll.aufbewahrung':
    'This many older files stay next to the current one, from `mse.log.1` to the oldest. Empty means one; 1 to 9 are allowed.',
  'protokoll.komponenten':
    'Only with **detailed**: whose steps are recorded, for example `app` for the calls of the window. Empty means all; events and errors always come.',
  'protokoll.ort':
    'A folder starting with / or ~/. Empty means next to the database.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
