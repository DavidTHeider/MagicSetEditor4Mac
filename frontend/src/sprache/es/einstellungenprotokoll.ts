// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'Cuánto anota el registro',
  'protokoll.datei': 'Archivo de registro',
  'protokoll.system': 'Registro unificado de macOS',
  'protokoll.ausgabe': 'Salida de error estándar',
  'protokoll.format': 'Formato de las líneas',
  'protokoll.grenze': 'Tamaño de un archivo de registro en MB',
  'protokoll.aufbewahrung': 'Cuántos archivos anteriores se conservan',
  'protokoll.komponenten': 'Detallado solo para',
  'protokoll.ort': 'Carpeta del archivo de registro',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Solo errores',
  'protokoll.stufe.warnung': 'Advertencias y errores',
  'protokoll.stufe.info': 'Eventos (predeterminado)',
  'protokoll.stufe.ausfuehrlich': 'Detallado — cada paso',
  'protokoll.datei.an': 'Activado',
  'protokoll.datei.aus': 'Desactivado',
  'protokoll.system.an': 'Activado',
  'protokoll.system.aus': 'Desactivado',
  'protokoll.ausgabe.aus': 'Desactivado',
  'protokoll.ausgabe.an': 'Activado',
  'protokoll.format.text': 'Texto legible',
  'protokoll.format.json': 'JSON para herramientas',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'Qué se anota en el registro. **Detallado** ayuda a encontrar un problema: entonces conserva como máximo 1 MB por archivo y un archivo anterior, y el diagnóstico indica que está activado. Surte efecto en el próximo inicio.',
  'protokoll.datei':
    'El archivo `mse.log` que adjuntas a un informe de errores. Si no se puede crear, los demás canales siguen escribiendo y el diagnóstico indica el motivo.',
  'protokoll.system':
    'El registro del sistema: se lee en la Consola de macOS, bajo «MagicSetEditor4Mac».',
  'protokoll.ausgabe':
    'Para iniciar desde una terminal: cada línea también donde la terminal la muestra.',
  'protokoll.format':
    '**JSON** es una línea por evento con campos con nombre, para herramientas que recopilan registros.',
  'protokoll.grenze':
    'A partir de este tamaño, en el siguiente inicio empieza un archivo nuevo. Vacío significa 4 MB; se admite de 1 a 64.',
  'protokoll.aufbewahrung':
    'Tantos archivos anteriores se conservan junto al actual, desde `mse.log.1` hasta el más antiguo. Vacío significa uno; se admite de 1 a 9.',
  'protokoll.komponenten':
    'Solo con **detallado**: de qué partes se anotan los pasos, por ejemplo `app` para las llamadas de la ventana. Vacío significa todas; los eventos y los errores se anotan siempre.',
  'protokoll.ort':
    'Una carpeta que empiece por / o ~/. Vacío significa: junto a la base de datos.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
