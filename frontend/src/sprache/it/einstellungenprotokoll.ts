// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'Quanto scrive il registro',
  'protokoll.datei': 'File di registro',
  'protokoll.system': 'Unified Log di macOS',
  'protokoll.ausgabe': 'Standard error',
  'protokoll.format': 'Formato delle righe',
  'protokoll.grenze': 'Dimensione di un file di registro in MB',
  'protokoll.aufbewahrung': 'Quanti file più vecchi vengono conservati',
  'protokoll.komponenten': 'Dettagliato solo per',
  'protokoll.ort': 'Cartella del file di registro',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Solo errori',
  'protokoll.stufe.warnung': 'Avvisi ed errori',
  'protokoll.stufe.info': 'Eventi (predefinito)',
  'protokoll.stufe.ausfuehrlich': 'Dettagliato — ogni passaggio',
  'protokoll.datei.an': 'Attivato',
  'protokoll.datei.aus': 'Disattivato',
  'protokoll.system.an': 'Attivato',
  'protokoll.system.aus': 'Disattivato',
  'protokoll.ausgabe.aus': 'Disattivato',
  'protokoll.ausgabe.an': 'Attivato',
  'protokoll.format.text': 'Testo leggibile',
  'protokoll.format.json': 'JSON per gli strumenti',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'Che cosa finisce nel registro. **Dettagliato** aiuta a trovare un problema — in tal caso conserva al massimo 1 MB per file e un solo file più vecchio, e la diagnostica indica che è attivo. Ha effetto dal prossimo avvio.',
  'protokoll.datei':
    'Il file `mse.log` che alleghi a una segnalazione di errore. Se non può essere creato, gli altri canali continuano a scrivere e la diagnostica ne indica il motivo.',
  'protokoll.system':
    'Il registro di sistema — da leggere nell’app Console di macOS sotto “MagicSetEditor4Mac”.',
  'protokoll.ausgabe':
    'Per l’avvio da un terminale: ogni riga anche lì dove il terminale la mostra.',
  'protokoll.format':
    '**JSON** è una riga per evento con campi denominati — per gli strumenti che raccolgono i registri.',
  'protokoll.grenze':
    'Da questa dimensione in poi, al prossimo avvio inizia un nuovo file. Vuoto significa 4 MB; sono consentiti da 1 a 64.',
  'protokoll.aufbewahrung':
    'Quanti file più vecchi restano accanto a quello corrente, da `mse.log.1` fino al più vecchio. Vuoto significa uno; sono consentiti da 1 a 9.',
  'protokoll.komponenten':
    'Solo con **dettagliato**: di quali parti vengono registrati i passaggi, ad esempio `app` per le chiamate della finestra. Vuoto significa tutte; eventi ed errori vengono sempre registrati.',
  'protokoll.ort':
    'Una cartella che inizia con / o ~/. Vuoto significa: accanto all’archivio.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
