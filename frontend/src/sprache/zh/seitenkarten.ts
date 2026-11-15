// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: '添加卡牌',
  nachOben: '上移',
  nachUnten: '下移',
  duplizieren: '复制卡牌',
  loeschen: '删除卡牌',
  loeschenFrage: '要删除这张卡牌吗？按 ⌘Z 可撤销。',
  loeschenJa: '删除',
  loeschenNein: '取消',
  loeschenTrotzdem: '仍然删除',
  symbole: '插入符号',
  hergeleitet: '派生',
  hergeleitetWert: '派生：{wert}',
  bildWaehlen: '选取图片',
  stilfelder: '样式字段',
  stilfelderHilfe: '这些字段随卡牌的样式而来。其值仅适用于此样式。',
  symbolZeichnen: '绘制符号',
  kartenskripteTitel: '通过脚本生成卡牌',
  kartenskripteErklaerung:
    '游戏资源包中现成的操作——点击一次即可同时添加多张卡牌。' +
    '一次撤销即可全部收回。',
  kartenskriptEine: '已添加 1 张卡牌。',
  kartenskriptMehrere: '已添加 {anzahl} 张卡牌。',
  bildEinsetzen: '插入图片',
  pack: {
    ueberschrift: '随机补充包',
    erklaerung:
      '按照游戏资源包的规则，从此系列的卡牌中抽取一包补充包。相同的种子会得到相同的结果——这样结果就可以重现。',
    laedtGrund:
      '每种补充包类型都会逐张询问卡牌是否适用——首次需要稍等片刻，之后即可直接使用。',
    art: '补充包类型',
    same: '种子',
    sameZahl: '种子是 0 到 18446744073709551615 之间的整数。',
    ziehen: '抽取',
    neuerSame: '新种子',
    gezogen: '已抽取',
    keineArten: '此系列所属的游戏没有提供补充包类型——或者其资源包未安装。',
    keineKarten: '此补充包类型没有产生任何卡牌。',
    artMitZahl: '{name}（{zahl} 张卡牌）',
    nurBaustein: '构建块——由其他补充包类型抽取',
    vorbehalte: '筛选条件无法求值的补充包类型',
  },
  symbol: {
    ueberschrift: '符号',
    erklaerung:
      '此系列的符号——是绘制的，而非载入的。它会立即出现在预览中，并在存储系列时写入文件。每个形状都可以用键盘放置和移动。',

    grundform: '添加基本形状',

    verlauf: '绘制历史',
    nachzeichnen: '描摹图片',
    groesser: '放大',
    nachzeichnenHilfe:
      '轮廓清晰的 PNG、JPEG 或 GIF 图片会变成带贝塞尔点的形状，你可以继续在此编辑。',
    zuruecknehmen: '撤销手势',
    wiederherstellen: '重做手势',
    kreis: '圆形',
    rechteck: '矩形',
    dreieck: '三角形',
    stern: '星形',

    teile: '部件',
    keineTeile: '尚未绘制任何内容。从基本形状开始——之后每个点都可以修改。',
    teilLoeschen: '删除部件',
    teilMitPunkten: '{name}——{zahl} 个点',
    teilMitKopien: '{name}——对称，{zahl} 份副本',
    gruppieren: '编组',
    gruppierenHilfe: '选中两个或更多部件即可编为一组。按住 ⌘ 可选择多个。',
    aufloesen: '取消编组',
    nachOben: '上移',
    nachUnten: '下移',
    nachObenFuer: '上移 {name}',
    nachUntenFuer: '下移 {name}',
    teilLoeschenFuer: '删除 {name}',
    teilMitKindern: '{name}——{zahl} 个部件',
    teilKind: '{name}，位于编组内',
    wahlHilfe: '按住 ⌘ 可选择多个部件。顺序即绘制顺序：排在前面的位于最上层。',

    zeichnung: '绘制区域',
    punkt: '{name} 的第 {nummer} 个点',
    griffDavor: '第 {nummer} 个点之前的控制柄',
    griffDanach: '第 {nummer} 个点之后的控制柄',
    punktHinzufuegen: '插入点',
    punktLoeschen: '删除点',

    tastatur: '使用键盘',
    tastaturhilfe:
      '按 Tab 键在各点之间移动。方向键每次移动百分之一，配合 Shift 则移动十分之一。按回车会在所选点之后插入新点，按 Backspace 会将其删除。',

    kombination: '组合',
    ueberdecken: '叠加',
    vereinen: '合并',
    abziehen: '减去',
    schneiden: '相交',
    unterschied: '差集',
    umranden: '描边',

    symmetrie: '对称',
    symmetrieAnlegen: '添加对称',
    symmetrieAufloesen: '取消对称',
    symmetrieErklaerung:
      '对称会复制其中的所有内容。副本以淡色显示在画布上；只有原件可以被编辑。',
    symmetrieHinweis: '选择一个部件，即可将其变为对称。',
    symmetrieart: '类型',
    drehung: '旋转',
    achse: '镜像',
    kopien: '副本',
    kopienHinweis:
      '镜像需要成对——数量为奇数时会有一份副本落单，MSE 会向下取整。',
    mitteX: '中心，水平',
    mitteY: '中心，垂直',
    griffX: '控制柄，水平',
    griffY: '控制柄，垂直',
    mitte: '对称中心',
    griff: '对称控制柄',

    sichern: '放入系列',
    gesichert: '已存储为 {name}。存储系列时它会写入文件。',
  },
  konsole,

  wortlisten: {
    ueberschrift: '可选项',
    liste: '从 {liste} 中选择',
    keine: '以上皆非',
    gruppe: '{name}（分组）',
  },
  regeln: {
    ueberschrift: '来自 Scryfall 的裁决',
    holen: '获取裁决',
    keine: '这张卡牌没有裁决。',
    offiziell: '官方',
    anmerkung: 'Scryfall 备注',
    anmerkungErklaerung:
      '备注来自 Scryfall 团队，通常用来说明非官方赛制。它们不是规则。',
    quelle: '在 Scryfall 上查看这张卡牌',
    alter: '截至：{stunden} 小时前',
    manaPruefen: '检查法术力费用',
    manaGut: '法术力费用没有问题。',
    manaSchlecht: '未知：{symbole}',
    manaLeer: '这张卡牌没有法术力费用。',
  },
  abgleich: {
    ueberschrift: '来自 Scryfall 的卡牌数据',
    erklaerung: '显示将会发生的更改。只有勾选的内容才会写入。',
    holen: '获取建议',
    uebernehmen: '应用已勾选项',
    nichtsZuTun: '无需更改——这张卡牌与 Scryfall 一致。',
    alt: '当前',
    neu: '新',
    leer: '（空）',
    alter: '状态：{stunden} 小时前',
    frisch: '状态：刚刚获取',
    quelle: '在 Scryfall 上查看这张卡牌',
  },
  scryfall: {
    ueberschrift: '来自 Scryfall 的图片',
    erklaerung:
      '由卡牌名称决定图片。画师也会一并写入卡牌——Scryfall 要求在插图旁标明名字。',
    name: 'Scryfall 上的卡牌名称',
    art: '图片类型',
    artKlein: '小图（预览）',
    artNormal: '完整卡牌',
    artAusschnitt: '仅插图',
    holen: '获取图片',
    ausschnittHinweis:
      '插图裁切只显示插图本身。如果 Scryfall 没有给出画师，则会被拒绝。',
    quelle: '在 Scryfall 上查看这张卡牌',
  },
  scryfallkarte: {
    name: 'Scryfall 上的卡牌名称',
    platzhalter: '例如 Lightning Bolt',
    mitBild: '含插图',
    anlegen: '从 Scryfall 创建',
  },
  zuschnitt: {
    ueberschrift: '选择裁切范围',
    erklaerung:
      '边框决定图片的哪一部分会放到卡牌上。拖动可移动边框，拖动角点可调整大小。',
    quelle: '带有所选裁切范围的源图片',
    vorschau: '按字段大小预览',
    links: '左',
    oben: '上',
    breite: '宽度',
    hoehe: '高度',
    seitenverhaeltnis: '保持字段的宽高比',
    uebernehmen: '插入',
    abbrechen: '取消',
    masse: '{breite} × {hoehe} 像素',
  },
  feldWaehlen: '编辑此字段',
  leerTitel: '还没有卡牌',
  listeBezeichnung: '系列中的卡牌',
  mehr: '更多',
  gruppeAnlage: '卡牌',
  gruppeAnsicht: '视图',
  ausScryfallAufklapp: '从 Scryfall 创建…',
  ausScryfallErklaerung:
    '不创建空白卡牌：输入名称或在 Scryfall 上搜索——字段随后由那里提供。',
  dieseKarte: '这张卡牌',
  kartenmenue: '卡牌',
  auswahlLoesen: '显示所有卡牌',
  bezugAufklapp: '关联到印刷版…',
  dieseKarteLeer: '打开一张卡牌即可编辑。',
  keineKarteTitel: '未选择卡牌',
  titel: '卡牌',
  leer: '创建一张，或粘贴复制的卡牌。',
  ohneName: '未命名',
  ohneFelder: '这张卡牌没有任何字段值。',
  nummer: '编号',
  vorschau: '卡牌预览',
  keineKarte: '从列表中选一张——它的字段和预览会显示在这里。',
  bearbeiten: '编辑',

  notizen: {
    ueberschrift: '卡牌备注',
    platzhalter: '仅供你自己查看——绝不会出现在卡牌上。',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: '应用对它的了解',
  tiefe: '显示程度',
  tiefeNormal: '常规',
  tiefeErweitert: '高级',
  tiefeExperte: '专家',
  tiefeErweitertTitel: '由脚本填入的内容',
  tiefeErweitertHinweis:
    '这些由游戏资源包计算得出——它们说明了卡牌上某项内容的来由。无法编辑。',
  befundVerdeckt: '“{tiefe}”下有一项发现。',
  befundSpringen: '前往查看',
  buchfuehrungHinweis:
    '这些条目不属于卡牌本身，而属于它的记录信息——它使用哪个模板、设置了什么。只有明白原因时才修改。',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: '这张卡牌的进度',
  arbeitsstandHinweis:
    '备注和标记——它们属于你对这张卡牌的工作，而不属于卡牌本身。传给别人的文件不会包含它们。',

  marken: {
    ueberschrift: '标记',
    erklaerung:
      '供你自己工作使用的词语——它们放在卡牌旁边，而不在卡牌里面。传给别人的文件不会包含它们。',
    platzhalter: '例如 完成',
    setzen: '添加标记',
    entfernen: '移除标记 {marke}',
    keine: '这张卡牌还没有标记。',
    schonGesetzt: '这张卡牌已在别处加上了“{marke}”。',
    filter: '按标记筛选',
    alle: '全部',
  },

  scryfallsuche: {
    frage: '在 Scryfall 上搜索',
    platzhalter: '例如 Lightning Bolt——或 t:creature c:r cmc<=3',
    erklaerung:
      '输入卡牌名称即可。如果你了解 Scryfall 的搜索语法，也可以在此使用。',
    suchen: '搜索',
    nichtsGefunden: '没有找到。换个词试试？',
    mehr: 'Scryfall 上还有更多结果。请缩小搜索范围——不会自动获取任何内容。',
    alter: '状态：{stunden} 小时前',
    frisch: '状态：刚刚获取',
    befund: '发现：{text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: '关联到',
    keiner:
      '这张卡牌没有关联到具体的印刷版本。与 Scryfall 同步时会任选一个——请在上方搜索并关联。',
    druck: '来自 {setcode} {nummer}',
    gesetzt: '——关联于 {wann}',
    setzen: '关联到 {name}（{setcode} {nummer}）',
    loesen: '取消关联',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: '拷贝（{zahl}）',
    ausschneiden: '剪切（{zahl}）',
    einfuegen: '粘贴',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: '列表列',
    jeSpiel: '所作选择适用于同一游戏的所有系列。',
    filtern: '查找列',
    gezeigt: '显示 {gezeigt} / {alle} 列',
    ausgeblendet: '另有 {anzahl} 列已隐藏 — 请用上方的输入框查找。',
    genauer: '还有更多匹配 — 请缩小搜索范围以查看。',
    nachOben: '再向左移',
    nachUnten: '再向右移',
    sortieren: '排序方式',
    garNicht: '系列中的顺序',
    aufsteigend: '升序',
    absteigend: '降序',
    ohneVorlage:
      '无法读取游戏资源包。各列取自卡牌本身——没有标题，也不遵循资源包的顺序。',
    spalteAusblenden: '隐藏列',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— 选择字段 —',
    keinWert: '— 选择值 —',
    ueberschrift: '批量更改',
    erklaerung:
      '为每张已勾选的卡牌设置同一个字段——作为一次操作完成。如果出现任何问题，则一张都不会更改。',
    auswaehlen: '勾选以进行批量更改',
    mehrereWaehlen: '多选',
    ausgewaehlt: '已勾选',
    keineAuswahl: '没有勾选任何卡牌。在列表中勾选多张卡牌即可一起更改。',
    feld: '字段',
    wert: '新值',
    anwenden: '应用',
    betroffen: '这些卡牌上的更改失败。没有任何一张卡牌被更改：',
    gescheitert: '失败',
    ohneWert: '尚无值',
  },
  stilTitel: '样式',
  stilWaehlen: '此系列的样式',
  stilLeer: '尚未安装样式。没有样式时，预览会退回到占位显示。',
  stilFrage: '要切换到此样式吗？',
  stilSpielwechsel: '新样式属于另一个游戏。',
  stilUngedeutet:
    '新游戏不认识这些字段。它们的值会保留在数据库中——但不会再被绘制：',
  stilVerloreneOptionen: '新样式没有这些样式选项。它们的值将会丢失：',
  stilUebernehmen: '切换样式',
  stilAbbrechen: '保持样式',
  stilKeiner: '此系列指定的样式尚未安装——请从已安装的样式中选择一个。',
  stiloptionen: '样式选项',
  stiloptionenSichtbar: '{zahl} / {gesamt}',
  stiloptionenWeitere: '更多选项（{zahl}）',
  marktTitel: '在哪里找到这张卡牌',
  marktErklaerung:
    '这些链接由你的电脑生成——依据卡牌名称或提供方的标识符。在你点击之前不会发出任何请求，并且在你选择的浏览器中打开。无需密钥；那里是否有售取决于提供方。',
  marktGenau: '直达这张卡牌',
  marktSuche: '前往搜索页面',
  marktNurAnbieter: '仅前往提供方',
  umschlagAnnehmen: '来自信封的卡牌',
  umschlagAngenommen: '这张卡牌现已在此系列中。',
}
