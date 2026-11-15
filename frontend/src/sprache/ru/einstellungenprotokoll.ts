// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'Сколько записывает журнал',
  'protokoll.datei': 'Файл журнала',
  'protokoll.system': 'Unified Log macOS',
  'protokoll.ausgabe': 'Стандартный вывод ошибок',
  'protokoll.format': 'Формат строк',
  'protokoll.grenze': 'Размер файла журнала в МБ',
  'protokoll.aufbewahrung': 'Сколько старых файлов сохраняется',
  'protokoll.komponenten': 'Подробно только для',
  'protokoll.ort': 'Папка файла журнала',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Только ошибки',
  'protokoll.stufe.warnung': 'Предупреждения и ошибки',
  'protokoll.stufe.info': 'События (по умолчанию)',
  'protokoll.stufe.ausfuehrlich': 'Подробно — каждый шаг',
  'protokoll.datei.an': 'Вкл.',
  'protokoll.datei.aus': 'Выкл.',
  'protokoll.system.an': 'Вкл.',
  'protokoll.system.aus': 'Выкл.',
  'protokoll.ausgabe.aus': 'Выкл.',
  'protokoll.ausgabe.an': 'Вкл.',
  'protokoll.format.text': 'Текст для чтения',
  'protokoll.format.json': 'JSON для инструментов',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'Что попадает в журнал. **Подробно** помогает найти ошибку — тогда журнал хранит не более 1 МБ на файл и один старый файл, а диагностика сообщает, что режим включён. Действует со следующего запуска.',
  'protokoll.datei':
    'Файл `mse.log`, который ты прикладываешь к отчёту об ошибке. Если его не удаётся создать, остальные каналы продолжают запись, а диагностика называет причину.',
  'protokoll.system':
    'Журнал системы — его можно прочитать в программе «Консоль» в macOS под именем «MagicSetEditor4Mac».',
  'protokoll.ausgabe':
    'Для запуска из терминала: каждая строка выводится также там, где её показывает терминал.',
  'protokoll.format':
    '**JSON** — одна строка на событие с именованными полями — для инструментов, собирающих журналы.',
  'protokoll.grenze':
    'По достижении этого размера при следующем запуске начинается новый файл. Пусто — 4 МБ; допустимо от 1 до 64.',
  'protokoll.aufbewahrung':
    'Столько старых файлов остаётся рядом с текущим — от `mse.log.1` до самого старого. Пусто — один; допустимо от 1 до 9.',
  'protokoll.komponenten':
    'Только при **подробном** режиме: шаги каких частей записываются, например `app` для вызовов окна. Пусто — все; события и ошибки записываются всегда.',
  'protokoll.ort':
    'Папка, путь к которой начинается с / или ~/. Пусто — рядом с базой данных.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
