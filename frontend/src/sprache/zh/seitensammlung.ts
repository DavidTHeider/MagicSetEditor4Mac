// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: '收藏',
  erklaerung:
    '你实际拥有的东西——每个印刷版本、品相、语言和表面处理各占一行。这与你的系列无关：设计出来的卡牌只是草稿，一份副本则是一张实实在在的纸板。',

  aufnehmenOeffnen: '添加…',
  einfuhrOeffnen: '导入与导出…',
  einfuhrTitel: '导入与导出',
  leerTitel: '还没有记录任何内容',
  leer: '一份副本需要其印刷版本的 Scryfall ID——即卡牌页面地址中的 UUID。整份列表可以一次导入。',

  druck: '印刷版本',
  menge: '数量',
  zustand: '品相',
  sprache: '语言',
  ausfuehrung: '表面处理',
  foil: '闪卡',
  nonfoil: '非闪卡',
  etched: '蚀刻闪',
  erwerbsdatum: '获得日期',
  erwerbspreis: '购入价格',
  erwerbswaehrung: '货币',
  preisMehrdeutig: '有两种读法：{erste} 或 {zweite}。请写得明确无歧义。',
  notiz: '备注',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: '卡牌名称',
  setcode: '系列',
  nummer: '编号',
  sicht: '收藏视图',
  suche: '搜索',
  suchePlatzhalter: '名称、系列、ID 或备注',
  sortierung: '排序方式',
  sortierBezeichnung: '卡牌',
  sortierMenge: '数量',
  sortierErwerbsdatum: '获得日期',
  ausfuehrungAlle: '所有表面处理',
  nurFoil: '仅闪卡',
  nurNonfoil: '仅非闪卡',
  nurEtched: '仅蚀刻闪',

  navigator: '视图与工具',
  inspector: '副本',
  exemplarmenue: '副本',
  inDecks: '所在套牌',

  inspectorLeer:
    '在列表中选择一份副本——数量、品相、获得方式和价格会显示在这里，也在这里修改。',
  inspectorAusserhalb: '这份副本已不在当前视图中。更改仍会作用于这份副本。',
  arbeitsplatz: '收藏工作区',
  listeBezeichnung: '副本',
  spalteWahl: '选择',
  spalteBezeichnung: '名称',
  spalteMenge: '数量',
  spalteZustand: '品相',
  spalteSprache: '语言',
  spalteSet: '系列',
  spalteLagerort: '存放位置',
  gruppeOhne: '未指定',
  gruppeSumme: '{anzahl} 份副本 · 共 {menge} 张',
  kachelWaehlen: '选择 {bezeichnung}',

  ansicht: '视图',
  ansichtTabelle: '表格',
  ansichtGalerie: '图库',
  ansichtGruppen: '分组',
  gruppierung: '分组依据',
  gruppierungKeine: '不分组',
  gruppierungSet: '系列',
  gruppierungLagerort: '存放位置',
  gruppierungZustand: '品相',
  gruppierungSprache: '语言',
  ohneTreffer: '没有副本符合搜索和筛选条件。',
  ohneTrefferZuruecksetzen: '重置搜索和筛选',
  gezeigt: '{gezeigt} / {gesamt}',
  gekappt: '已绘制 {gezeichnet} / {passend}',
  mehrZeigen: '显示更多',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: '粘贴列表',
  listeErklaerung:
    '每行一张卡牌，例如“4 Lightning Bolt (M11) 149”。空行、“Deck”、“Sideboard”以及以 # 或 // 开头的行会被跳过。',
  listeFeld: '卡牌列表',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: '匹配',
  listeEinlesen: '添加',
  listeOffen: '{gesamt} 行中有 {offen} 行需要决定——请选择你拥有的印刷版本。',
  listeAlleEindeutig: '全部 {gesamt} 行都没有歧义。',
  listeDruckWahl: '哪个印刷版本？',
  listeOhneWahl: '— 尚未决定 —',
  listeDrucke: '{drucke} 个印刷版本',
  listeBericht: '新增 {angelegt} 行，增加 {erhoeht} 行，共添加 {stueck} 张。',
  listeNichtsGewaehlt: '没有决定任何一行——请至少选择一个印刷版本。',
  listeAusgelassen: '已跳过 {ausgelassen} 行已有的行。',
  listeKeinDruck: '没有叫这个名字的卡牌——请检查拼写。',
  listeZuViele:
    '叫这个名字的印刷版本超过 {drucke} 个——请在括号中注明系列，例如“(M11)”，即可确定一个。',
  listeOhneKarte: '{anzahl} 行找不到卡牌，将被跳过。',
  listeOhneKarteEins: '有一行找不到卡牌，将被跳过。',
  vorschauTitel: '导入之前',
  vorschauErklaerung:
    '有些行与你的收藏中已有的行匹配——相同的印刷版本、品相、语言、表面处理和用途。请选择如何处理它们；在此之前不会写入任何内容。',
  vorschauZaehlung:
    '{neu} 行新增 · {zusammen} 行与已有行匹配 · {wiederholt} 行重复了本次导入中的另一行',
  vorschauZusammen: '{bezeichnung}，{zustand}：{vorher} → {nachher}',
  vorschauModus: '已有的行',
  vorschauAddieren: '数量相加',
  vorschauAuslassen: '跳过已有的行',
  vorschauEinlesen: '导入',
  vorschauAbbrechen: '不导入',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: '导出与备份',
  ausfuhrErklaerung:
    '“CSV”包含每一项信息，并且可以完整地读回。“列表”是可粘贴到构筑套牌工具中的形式——只包含数量、卡牌、印刷版本和表面处理。',
  ausfuhrAlsCSV: '导出为 CSV',
  ausfuhrAlsListe: '导出为列表',
  ausfuhrErgebnis: '输出',
  ausfuhrZurueck: '读回 CSV',
  ausfuhrLeer: '尚未导出任何内容。',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: '存放位置',
  wunsch: '想要，尚未拥有',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: '整理',
  ordnungLagerorte: '存放位置',
  ordnungDubletten: '重复项',
  ordnungWuensche: '愿望清单',
  ordnungBehalten: '保留',
  ordnungLeerLagerorte: '还没有任何藏品——没有藏品，就没有存放位置。',
  ordnungLeerDubletten: '没有任何印刷版超过 {behalten} 张。',
  ordnungLeerWuensche:
    '愿望清单是空的。把藏品标为“想要，尚未拥有”即可加入，或在套牌中使用“将缺少的卡牌加入愿望清单”。',
  ordnungStueck: '{zeilen} 行共 {stueck} 张',
  ordnungOhneOrt: '未指定',
  ordnungUeberzaehlig: '多出 {ueberzaehlig} 张',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: '你拥有其中 {stueck} 张。',
  besitzNein: '你没有这个印刷版本。',
  besitzGewuenscht: '（愿望清单中有 {gewuenscht} 张）',

  // **Die Massenvorgänge** (#1292).
  masseTitel: '批量操作',
  masseWaehlen: '选择',
  masseGewaehlt: '已选择 {gewaehlt} / {gesamt}',
  masseLagerortSetzen: '将存放位置设为',
  masseZustandSetzen: '将品相设为',
  masseAendern: '更改所选项',
  masseLoeschen: '删除所选项',
  masseLoeschenFrage: '要删除 {anzahl} 份副本吗？',
  masseLoeschenErklaerung:
    '删掉的内容需要重新输入：数量、品相、存放位置和购入价格在其他任何地方都没有。收藏不像系列那样有历史记录。',
  masseLoeschenJa: '删除',
  masseBericht: '已更改 {geaendert} 项。',
  masseZuruecknehmen: '撤销',
  masseZuruecknehmenVon: '撤销 {anzahl} 行的存放位置',
  masseZurueckgenommen: '已撤销 {anzahl} 项。',
  massePreise: '获取全部价格',
  massePreisschub: '已获取 {geholt} 项，剩余 {offen} 项',
  massePreiseFehlgeschlagen: '{anzahl} 项失败',
  massePreiseSteht: '这一批没有任何进展——再试一次结果也一样。',
  massePreiseFertig: '所有价格都已获取。',

  aufnehmen: '添加',
  aenderungSichern: '保存',
  loeschen: '删除',
  loeschenFrage: '要删除这份副本吗？',
  loeschenErklaerung:
    '数量、品相、存放位置和购入价格在其他任何地方都没有——收藏不像系列那样有历史记录。',
  loeschenJa: '删除',
  abbrechen: '取消',
  bearbeiten: '编辑',
  zumTausch: '用于交换',
  insDeck: '加入套牌',
  auswahlTitel: '选择',
  auswahlInsDeck: '{anzahl} → 套牌',

  neuesTitel: '添加一份副本',
  weitereAngaben: '更多信息',

  dubletteFrage:
    '收藏中已有 {bezeichnung}——{menge} 张，品相、语言和表面处理都相同。',
  dubletteErklaerung:
    '“增加”会加上新的数量；已有行的其他信息保持不变。取消不会更改任何内容——你的输入仍会保留。',
  dubletteErhoehen: '增加到 {summe}',
  dubletteZurZeile: '前往已有的行',
  dubletteAbbrechen: '取消',

  zustaende: {
    mint: '全新',
    nearMint: '近全新',
    excellent: '极佳',
    good: '良好',
    lightPlayed: '轻微磨损',
    played: '有磨损',
    poor: '较差',
  },

  preiseTitel: '价格',
  preiseHolen: '获取价格',
  marktUndVerlauf: '报价与历史',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: '交易市场上的报价',
  marktErklaerung:
    '个别卖家对这个印刷版本的要价——价格、数量、品相和语言。查询会通过网络发出。不会购买任何东西。',
  marktHolen: '获取报价',
  marktHolenBei: '从 {quelle} 获取报价',
  marktLaeuft: '正在获取……',
  marktLeer: '该印刷版本目前在那里没有在售的商品。',
  marktZugeordnet:
    '该印刷版本现已关联到卡牌版本 {fassung}——下次查询会立即找到它。',
  marktStueck: '可购 {menge} 张',
  marktBuendel: '每捆 {buendel} 张',
  marktImUrlaub: '目前不发货',
  marktGradiert: '已评级',
  marktSigniert: '已签名',
  marktVeraendert: '已改动',
  marktBefunde: '并非每个报价都能解读：',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde: '并非每个价格都能解读。缺少这些价格序列：',
  preiseLaufen: '正在获取……',
  keinePreise: '此印刷版本没有价格记录。',
  preisAlter: '{stunden} 小时前 · {quelle}',
  preisFrisch: '刚刚 · {quelle}',
  preisVeraltet: '已过期，{tage} 天前 · {quelle}',
  preisQuellstand: '来源日期 {tag}',

  vorbehalt:
    '价格是 Scryfall 及其合作方给出的估价。对任何数值都不作保证；以商店中的价格为准。',

  verlauf: '历史',
  keinVerlauf: '此序列只有一个时点的数据——再次获取价格后，才会形成历史。',

  bulkTitel: '获取整个来源的价格',
  leiste: {
    bezeichnung: '收藏数据',
    exemplare: '卡牌',
    eintraege: '{zahl} 条',
    wert: '价值',
    keinPreis: '无价格',
    ohnePreis: '未获取价格',
    nichtGelesen: '未读取',
  },
  drucksuche: {
    karte: '卡牌',
    hinweis: '名称，可附系列和编号 — 例如 Lightning Bolt (M11) 149',
    suchen: '搜索',
    drucke: '版本',
    keinTreffer: '未找到此名称的卡牌。',
    weitere: '还有更多版本 — 加上系列和编号可缩小选择。',
    treffer: '{name} · {set} ({code}) · 编号 {nummer}',
  },
  bulkErklaerung:
    'MTGJSON 不是逐张卡牌回答，而是给出一个文件。“今天”带来当天的数据，“九十天”带来历史——并且超过一个 GB。',
  bulkQuelle: '来源',
  bulkUmfang: '范围',
  bulkHeute: '今天',
  bulkNeunzigTage: '九十天',
  bulkHolen: '获取',
  bulkHolenBei: '从 {quelle} 获取价格',
  bulkLaeuft: '正在获取……',
  bulkBilanz: '已存入 {drucke} 个印刷版本的 {staende} 条价格。',
  bulkOhneDrucke:
    '此收藏中没有任何印刷版本带有 MTGJSON 标识符——未发出任何请求。',
  uebersichtTitel: '概览',
  uebersichtLeer: '尚未获取任何价格。',
  ohnePreis: '无价格：{zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} 份副本',
  exemplare: '{zahl} 份副本',

  keineUmrechnung: '每种货币一个总额。不做换算——那需要汇率，而本程序没有汇率。',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: '来自其他来源的价格历史',
  verlaufErklaerung:
    '获取一段时间内的序列并存储。查询会通过网络发出并消耗配额；品相按来源给出的写法显示。',
  verlaufHolen: '获取历史',
  verlaufHolenBei: '从 {quelle} 获取历史',
  verlaufLaeuft: '正在获取……',
  verlaufBilanz: '已存入 {staende} 个价格点——卡牌版本 {fassung}。',
} as const
