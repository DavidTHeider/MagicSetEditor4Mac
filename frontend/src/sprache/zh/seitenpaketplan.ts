// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: '软件包工作区',
  navigator: '搜索与筛选',
  inspector: '软件包',
  zentrum: '软件包',

  herkunft: '来源',
  herkunftAlle: '所有来源',
  herkunftsarten: {
    mitgeliefert: '随附',
    bezogen: '已获取',
    uebernommen: '已导入',
    eigen: '自建',
  },

  spalteStand: '状态',
  standEingebunden: '已安装',
  standGeaendert: '本地已更改',

  inspectorLeer:
    '在列表中选择一个软件包——这里会显示它的来源、版本、依赖项，以及更改会带来什么影响。',
  kopf: '{art} · 版本 {version}',
  kopfOhneFassung: '{art} · 无版本',
  eingebunden: '已安装',
  keinSet: '此档案中没有系列使用它。',
  einSet: '有一个系列使用它。',
  vieleSets: '有 {anzahl} 个系列使用它。',
  setsUnbekannt: '哪些系列需要它，没有记录在任何系列头部中。',
  geaendert:
    '自安装以来已在本地更改。替换前会先询问；所做的更改会作为上一版本保留。',
  woher: '来源',
  woherMitgeliefert: '随应用程序一起提供。',
  woherBezogen: '从 {herkunft} 获取。',
  woherUebernommen: '从 {herkunft} 导入。',
  woherEigen: '在此创建，或手动放入软件包文件夹。',
  pruefsumme: '校验和',
  lizenz: '许可证',
  lizenzUnbekannt: '未注明——MSE 软件包没有许可证字段；其内容属于各自的权利人',
  ort: '位置',
  haengtAb: '依赖于',
  gebrauchtVon: '被以下项目需要',
  nichts: '无',

  einbinden: '安装软件包…',
  beziehen: '获取软件包…',
  beziehenTitel: '获取软件包',
  loeschen: '删除…',
  zurueck: '恢复上一版本…',
  weitergeben: '作为安装程序分享…',
  weitergegeben: '已存储为安装程序：{datei}',

  titel: {
    loeschen: '删除 {paket}',
    einbinden: '安装 {paket}',
    ersetzen: '替换 {paket}',
    zurueck: '将 {paket} 重置为上一版本',
  },
  fassung: '版本 {von} → {zu}',
  ohneFassung: '无版本',
  sets: '受影响的系列',
  keineSets: '此档案中没有系列受影响。',
  konflikte: '之后将不再满足',
  keineKonflikte: '没有依赖项会失效。',
  planGeaendert: '该软件包自安装以来已在本地更改。',
  ueberschreiben: '替换本地更改——它会作为上一版本保留',
  nichtMoeglich: {
    loeschen:
      '此软件包不在受管理的文件夹中。这里只会删除本应用程序安装的内容。',
    einbinden: '此软件包无法在这里安装。',
    ersetzen: '此软件包无法在这里替换。',
    zurueck: '没有上一版本。在此软件包上安装新版本时，会生成一个上一版本。',
  },
  bestaetigen: {
    loeschen: '删除',
    einbinden: '安装',
    ersetzen: '替换',
    zurueck: '重置',
  },
  verwerfen: '不做更改',
  bedarf: {
    ueberschrift: '此系列缺少的软件包',
    erklaerung:
      '没有它们，预览会绘制备用方案，或省略该软件包提供的内容。请从文件夹安装该软件包——或在软件包中搜索它，以获取它或从 MSE2 导入。',
    suchen: '在软件包中搜索',
  },
  quelle: '来自 {quelle}',
  keineRechte:
    '软件包带有模板和 MSE-Script。脚本无法访问文件、网络或程序，软件包所需的其他内容也不会一并获取——它们会作为冲突列在上方。',
}
