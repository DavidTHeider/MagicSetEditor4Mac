// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': '日志记录的详细程度',
  'protokoll.datei': '日志文件',
  'protokoll.system': 'macOS 统一日志',
  'protokoll.ausgabe': '标准错误输出',
  'protokoll.format': '行格式',
  'protokoll.grenze': '单个日志文件的大小（MB）',
  'protokoll.aufbewahrung': '保留多少个较旧的文件',
  'protokoll.komponenten': '仅对以下部分详细记录',
  'protokoll.ort': '日志文件所在的文件夹',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': '仅错误',
  'protokoll.stufe.warnung': '警告和错误',
  'protokoll.stufe.info': '事件（默认）',
  'protokoll.stufe.ausfuehrlich': '详细 — 记录每一步',
  'protokoll.datei.an': '开启',
  'protokoll.datei.aus': '关闭',
  'protokoll.system.an': '开启',
  'protokoll.system.aus': '关闭',
  'protokoll.ausgabe.aus': '关闭',
  'protokoll.ausgabe.an': '开启',
  'protokoll.format.text': '便于阅读的文本',
  'protokoll.format.json': '供工具使用的 JSON',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    '哪些内容会记入日志。**详细**有助于查找错误——此时每个文件最多保留 1 MB 和一个较旧的文件，并且诊断信息会说明它已开启。下次启动后生效。',
  'protokoll.datei':
    '你附在错误报告中的文件 `mse.log`。如果无法创建，其他渠道会继续写入，诊断信息会指出原因。',
  'protokoll.system':
    '系统日志——可在 macOS 的“控制台”中查看，位于“MagicSetEditor4Mac”之下。',
  'protokoll.ausgabe': '用于从终端启动：每一行也会显示在终端显示它的地方。',
  'protokoll.format':
    '**JSON** 每个事件一行，带有具名字段——适用于收集日志的工具。',
  'protokoll.grenze':
    '达到此大小后，下次启动时会开始一个新文件。留空表示 4 MB；允许 1 到 64。',
  'protokoll.aufbewahrung':
    '当前文件旁会保留这么多个较旧的文件，从 `mse.log.1` 到最旧的一个。留空表示一个；允许 1 到 9。',
  'protokoll.komponenten':
    '仅在**详细**时生效：记录哪些部分的步骤，例如 `app` 表示窗口的调用。留空表示全部；事件和错误始终会记录。',
  'protokoll.ort': '以 / 或 ~/ 开头的文件夹。留空表示放在数据库旁边。',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
