// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: '显示所有系列',
  erklaerung:
    '每个已打开的系列都存放在资料库中。从这里可以进入它的卡牌——或将它写回文件。',
  titel: '系列',
  leer: '尚未打开任何系列。打开一个 .mse-set 文件即可开始。',
  setOeffnen: '打开系列…',
  neuTitel: '新系列',
  neuKopf: '新系列',
  einlesenTitel: '从别处读取系列',
  neuHinweis: '一个空系列——由游戏和样式决定其卡牌的外观。样式之后可以更改。',
  neuName: '系列名称',
  neuStil: '新系列的样式',
  neuAnlegen: '创建系列',
  neuSchritte: '新建系列的步骤',
  neuSchrittArt: '类型',
  neuSchrittAussehen: '外观',
  neuSchrittName: '名称',
  neuArtFrage: '这些卡牌用于什么？资源包名称由我们替你查找。',
  neuArtZusatz: '· 有 {stile} 可用',
  neuEigenes: '完全自己设计——创建游戏资源包',
  neuWeiter: '下一步',
  neuSchrittZurueck: '返回',
  neuUeberspringen: '跳过',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    '尚未安装游戏资源包。没有它就没有卡牌字段——资源包可在“资源包”页面获取。',
  neuOhneStile:
    '尚未为 {spiel} 安装样式资源包。没有样式就无法创建系列——资源包可在“资源包”页面获取。',
  neuZuDenPaketen: '前往资源包',
  neuZurueck: '按 ⌘[ 可回到这里；你已输入的内容仍然保留。',
  beispielsets: '随附的试用示例',
  geoeffnet: '已打开',
  schonOffen: '这个文件已经打开——就在这里。',
  teilerfolg:
    '文件没有读取到末尾。尽管如此，数据库中已有 {anzahl} 个系列——它们在下方列表中。',
  beimOeffnenBefunde: '打开时有无法读取的内容。系列依然存在，但缺少以下位置：',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: '读入之前：{name}',
    umfang: '{karten} · 游戏 {spiel} · 样式 {stil}',
    umfangOhneStil: '{karten} · 游戏 {spiel}',
    schonOffen:
      '这个文件已经在你的资料库中——“打开”会显示已有的系列，而不是再添加一个。',
    befunde: '以下位置无法读取；系列将不含这些内容：',
    bedarf:
      '缺少资源包。仍然可以读入——在资源包就位之前，卡牌会用占位内容绘制：',
    einlesen: '读入',
    aufschlagen: '打开',
    verwerfen: '不读入',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: '删除…',
  loeschenTitel: '删除此系列',
  loeschenWarnung:
    '{name} 及其 {karten} 将从资料库中消失。没有废纸篓，也无法撤销。保存在文件中的内容会保留——被删除的是资料库中的副本。',
  loeschenSammlung:
    '你的收藏会保留：一件藏品属于某个印刷版本，而不属于某个系列。',
  loeschenFeld: '输入系列名称以确认',
  loeschenPasstNicht: '名称与 {name} 尚不匹配。',
  loeschenJa: '永久删除',
  loeschenNein: '取消',
  loeschenBefunde: '系列已删除。以下残留内容未能清理：',

  spalten: {
    name: '名称',
    spiel: '游戏',
    karten: '卡牌',
    herkunft: '来源',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    '来自 mtgjson.com 的卡牌定义——单个文件或整个文件夹。它们对应哪些卡牌字段，取决于游戏。',
  mtgjsonSpiel: '游戏',
  mtgjsonDatei: '打开文件…',
  mtgjsonVerzeichnis: '打开文件夹…',
  mtgjsonKeinSpiel:
    '没有任何游戏为 MTG JSON 字段提供对应关系。没有它，就无法读入任何内容。',
  scryfallsatzUeberschrift: '来自 Scryfall 的系列',
  scryfallsatzErklaerung:
    '由系列代码决定导入哪些卡牌——带文字，不带图片。名称、法术力费用、类别、规则文字、风味文字、力量、防御力、稀有度和画师都会一并导入；图片需按卡牌逐张获取。样式之后再选。',
  scryfallsatzKeinSpiel: '没有带对应关系的游戏，就无法导入任何系列。',
  scryfallsatzSpiel: '用于对应关系的游戏',
  scryfallsatzCode: 'Scryfall 上的系列代码',
  scryfallsatzPlatzhalter: '例如 lea',
  scryfallsatzEinlesen: '导入系列',
  vorlageAusfuhr: '通过导出模板…',
  vorlageWaehlen: '导出模板',
  vorlageAusfuehren: '运行',
  vorlageSchliessen: '关闭',
  vorlageKeine: '尚未为此游戏安装导出模板。模板来自已下载的资源包。',
  vorlageGeschrieben: '已写入',
  vorlageBefunde: '运行时出现了一些情况——文件仍然已写出：',
  vorlageGescheitert: '此模板无法完整运行。缺少的内容见下方：',
  mtgjsonSichern: '存储为 MTG JSON…',
  mtgjsonGesichert: '张卡牌已存储到',
  mtgjsonVerluste: '导出时无法表达的内容。文件已生成，但缺少以下细节：',
  mtgjsonSpalte: '导出',
  speichern: '存储',
  speichernUnter: '存储为…',
  speicherSpalte: '存储',
  gespeichert: '已存储到',
  beimSpeichernBefunde:
    '存储时未能一并带上的内容。系列文件已写出，但缺少与它放在一起的以下文件：',
  oeffnen: '显示此系列的卡牌',
  handlungenAn: '操作——{name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: '未能理解（{anzahl}）',
    ueberschrift: '未解读',
    erklaerung: '导入时未能理解这些行。它们会原样保留，存储时不会丢失。',
    keine: '无——此系列已完整读取。',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: '存储为信封',
  umschlagGeschnuert: '已存储。信封已签名——收到它的人可以核实它来自这里。',
  umschlagTitel: '接收信封',
  umschlagErklaerung:
    '信封是别人存储的系列。它会先被检查，之后才被接收——在此之前，数据库不会有任何变化。',
  umschlagWaehlen: '检查信封…',
  umschlagUrheber: '签名者',
  umschlagAbdruckHinweis:
    '指纹说明的是签名者是谁——而不是你是否应该信任对方。如果你认识此人，请与对方核对指纹。',
  umschlagInhalt: '内容',
  umschlagUebernehmen: '接收',
  umschlagVerwerfen: '算了',
  umschlagUebernommen:
    '已接收。该系列是新建的——有自己的标识符，不会覆盖任何内容。',
}
