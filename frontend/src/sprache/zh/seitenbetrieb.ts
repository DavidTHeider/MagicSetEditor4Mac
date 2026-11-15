// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: '运行时间与资源消耗',
  zeitraum: '统计自 {seit} 至 {bis}。',
  aufwand: '统计开销：{mikros} µs',
  speicher: '内存',
  speicherWert: '占用 {fussabdruck}——其中 Go 堆 {heap}，系统分配 {system}',
  speicherOhneFussabdruck: 'Go 堆 {heap}，系统分配 {system}——占用未测量',
  web: '网页进程',
  webWert: '合计 {summe}——{teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: '网页内容', gpu: '图形', netz: '网络' },
  ohneWeb: '此进程没有网页视图。',
  webUngemessen: '未测量：{grund}',
  goroutinen: '并发例程',
  rechenzeit: '启动以来的 CPU 时间',
  datenbank: '与资料库的连接',
  datenbankWert:
    '{offen} 个打开，{inBenutzung} 个使用中，累计等待 {gewartet} 次，共 {dauer}',
  ohneBestand: '此进程不持有资料库。',
  zwischenspeicher: '外部服务的缓存',
  zwischenspeicherWert:
    '{treffer} 次命中，{geholt} 次获取，{veraltet} 次返回过期数据，{fehlgeschlagen} 次失败',
  laufzeiten: '各接入方式的运行时间',
  laufzeitenLeer: '这里没有测得任何调用——只测量服务和工具服务器。',
  ungemessen: '未测量：{liste}。',
  ungemessenNamen: {
    fenster: '窗口的绑定（它们没有一个统一的位置来结束一次调用）',
    kommandozeile: '命令行（它只存活一条命令的时间）',
    speicher: '此进程的占用（系统未提供）',
  },
  spalteArt: '类型',
  spalteAnzahl: '调用',
  spalteFehler: '错误率',
  spalteMittel: '平均',
  spalteVerteilung: '分布',
  arten: { lesend: '读取', schreibend: '写入' },
  eimer: '不超过 {grenze} ms',
  verteilung: '{eimer}：{anzahl} 次调用',
  eimerOffen: '以上',
  spurAus: '跟踪已关闭。请在“设置”›“诊断”中开启——下次启动后生效。',
  spurAn:
    '跟踪已开启：在 {gesehen} 次调用中记录了 {aufgenommen} 次，采样率 {abtastung} %。{minuten} 分钟内最多保留 {grenze} 条记录——{verdraengt} 条被挤出，{abgelaufen} 条已过期。',
}

export const leistungsprofil = {
  ueberschrift: '性能剖析',
  erklaerung:
    '在所选时间内对此窗口进行的 CPU 剖析，格式与“go tool pprof”相同。之后它会存放在资料库旁边；保留多少份，可在“设置”›“诊断”中设定。',
  sekunden: '时长（秒）',
  grenze: '1 到 60 之间的整数。',
  aufzeichnen: '记录剖析',
  laeuft: '正在记录——窗口仍可使用。',
  fertig: '已写入：{datei}（{groesse}）',
}

export const zwischenspeicher = {
  ueberschrift: '缓存',
  erklaerung:
    '本应用保留下来的内容，以免重复获取或计算——存放在磁盘和内存中。这些都不是你的数据：清空后，每个缓存会在下次使用时自动重新填充。',
  spalteSpeicher: '缓存',
  spalteOrt: '位置',
  spalteQuote: '命中率',
  spalteEintraege: '条目',
  spalteGroesse: '大小',
  spalteGrenze: '上限',
  keineQuote: '尚无查询',
  keineGrenze: '一个会被替换的文件',
  namen: {
    'scryfall.antworten': 'Scryfall 的响应',
    'scryfall.bilder': 'Scryfall 的卡牌图像',
    'scryfall.sammeldaten': 'Scryfall 批量文件',
    'render.ebenen': '已采样的图层',
    'pakete.bilder': '模板包图像',
    'skript.baeume': '已解析的脚本',
    'vorschau.bilder': '卡牌预览',
    'tisch.texturen': '牌桌上的卡牌图像',
  },
  orte: { platte: '磁盘', arbeitsspeicher: '内存' },
  leerenHinweis:
    '清空会丢弃全部内容。没有网络时，Scryfall 的响应和图像要等网络恢复后才会再次出现。',
  leeren: '全部清空',
  geleert: '已清空——它们会在下次使用时重新填充。',
}
