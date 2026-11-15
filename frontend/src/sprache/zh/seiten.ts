// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: '文件信息',
    kopfHinweis:
      'Magic Set Editor 存储此系列时记下的信息。这些信息会原样保留，不会被更改。',
    mseFassung: '写入所用的 MSE 版本',
    spielfassung: '游戏资源包版本',
    stilfassung: '样式表版本',
    versionskontrolle: '版本控制',
    apprenticecode: 'Apprentice 代码',
    titel: '系列信息',
    leerTitel: '没有系列字段',
    leer: '此系列的游戏资源包没有声明任何 `set field:` 块。请选择包含这些块的游戏。',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: '图片',
    erklaerung:
      '每张卡牌都会单独绘制并写出。运行过程就在此窗口中进行：会显示进度，并且可随时停止。',
    ordnerWaehlen: '选取文件夹…',
    keinOrdner: '尚未选取文件夹。',
    gemerkt: '上次选择——所有系列使用同一个文件夹。',
    nichtGemerkt: '此文件夹用于本次输出，但未被记住：',
    vorlage: '名称模板',
    konflikt: '文件已存在时',
    konfliktHilfe:
      '两张同名卡牌会生成相同的文件名。“重新编号”是 Magic Set Editor 2 的默认做法：在一次运行中依次递增，并覆盖之前运行留下的文件。',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: '保留已有文件',
      ueberschreiben: '覆盖',
      nummerieren: '编号',
      nummerierenneu: '重新编号',
    },
    vorlageHilfe:
      '{nummer} 会变为三位数的卡牌编号，{name} 会变为卡牌名称。名称相同的会在末尾附加编号。',
    format: '格式',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: '质量',
    gueteFast: '快速——用于快速查看',
    gueteMedium: '按 Magic Set Editor 2 的规则',
    gueteHigh: '高——细线不闪烁',
    gueteUltra: '最高——用于全分辨率打印',
    breite: '宽度（像素）',
    breiteHilfe:
      '不使用 dpi：没有任何模板给出卡牌的毫米宽度。63 毫米宽的卡牌为 750 像素时即 300 dpi。',
    ausgeben: '写出图片',
    abbrechen: '停止',
    laeuft: '张已写出，共',
    fertig: '张图片已写入',
    abgebrochen: '已停止。',
    uebersprungen:
      '已跳过 {anzahl} 张卡牌——文件已存在，且选择了“保留已有文件”：',
    befunde: '绘制时未能完整。图片已生成，但缺少以下部分：',
  },
  druckbogen: {
    titel: '打印页',
    erklaerung:
      '卡牌保持原有尺寸：63 × 88 毫米，与裁切后的大小完全一致。裁切标记位于纸张边缘，而不在卡牌上。',
    papier: '纸张',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: '出血（毫米）',
    anschnittHilfe:
      '出血是指图像超出裁切线的部分——可避免裁切后的卡牌露出白边。卡牌之间会留出两倍出血的间距。',
    aufloesung: '分辨率',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: '裁切标记',
    legen: '排布打印页…',
    fertig: '排在',
    seiten: '页：',
    befunde: '未能全部排下。打印页已生成，但以下位置需要检查：',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: '来自 Scryfall 的批量数据',
      erklaerung:
        '包含全部卡牌数据的文件。一旦下载完成，Scryfall 相关功能无需联网即可使用——数据最多滞后一天。',
      fehlt: '尚未获取。',
      vorhanden: '已存在，{mb} MB。',
      angeboten: 'Scryfall 提供 {mb} MB，截至 {stand}。',
      geholt: '获取于 {zeit}。',
      holen: '获取批量数据',
      laedt: '正在获取…',
    },

    titel: '资源包',
    bezugUeberschrift: '获取模板',
    bezugErklaerung:
      'MSE 模板存放在代码库中。获取的是某个固定状态——发行版或提交——绝不是分支：分支会移动，同一条目明天可能得到不同的资源包。',
    bezugAdresse: '代码库地址',
    bezugStand: '状态（发行版或提交）',
    bezugUnterverzeichnis: '子目录（资源包根目录）',
    bezugName: '在资源包目录中的名称',
    bezugPruefen: '查看来源…',
    bezugStarten: '获取',
    bezugAbbrechen: '不获取',
    bezugGeholt: '已获取到',
    bezugPruefsumme: '校验和',
    bezugUebergangen: '未解包——可执行文件不会一并解出：',
    bezugUnvollstaendig: '缺少地址和状态。两者缺一不可，否则无从获取。',
    bezugHerkunft: '将获取 {quelle}，状态 {stand}。',
    bezugLizenz:
      '这些内容并非来自 MagicSetEditor4Mac，也不适用我们的许可。卡牌边框、符号和字体可能受到保护；使用它们的责任由你自行承担。可执行文件不会被解包。',
    mse2Ueberschrift: '从 MSE2 安装中导入',
    mse2Erklaerung:
      '使用 MSE2 的用户，模板已经在自己的电脑上。下面列出了常见位置——在你选定之前，不会读取其中任何一处。',
    mse2Uebernehmen: '从此处导入',
    mse2VonHier: '从此处导入——{ort}',
    mse2Waehlen: '选取文件夹…',
    mse2KeinVorschlag:
      '在常见位置未找到 MSE2 安装。请自行选取包含模板的文件夹——在 MSE2 中它叫“data”。',
    installerErklaerung:
      '也可以使用 .mse-installer 文件——MSE2 分享模板的方式。已安装的内容保持不变。',
    installerWaehlen: '读取安装包…',
    mse2Ergebnis: '已导入 {anzahl} 个模板。',
    mse2Uebergangen: '已存在，因此被跳过——未覆盖任何内容：',
    abhaengigUeberschrift: '依赖项',
    abhaengigAllesGut: '所有依赖项均已满足。',
    abhaengigFehlt:
      '{paket} 需要 {braucht}——尚未安装。请获取它，或从 MSE2 安装中导入。',
    abhaengigZuAlt:
      '{paket} 需要 {braucht} 的 {verlangt} 版本；已安装的是 {da}。没有较新的版本，模板会悄悄绘制出不同的内容。',
    abhaengigZyklus: '{paket} 与 {braucht} 相互依赖，形成循环。',
    abhaengigUnlesbar:
      '{paket} 需要 {braucht}：版本无法比较——要求 {verlangt}，已安装 {da}。',
    quellenErklaerung:
      '四个合集，每个只需点击一次——地址和修订已预先填好；在你同意之前不会获取任何内容。模板所用的字体会一并获取；缺少它们，任何卡牌都无法正确渲染。这四个合集都不带许可：内容归其权利人所有，而非我们。',
    quellenGroesse: '约 {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15: '常用的 M15 边框。渲染出一张卡牌的最快途径。',
      fullNonMagic: '适用于万智牌以外游戏的模板。',
      m15: '公开的 M15 模板——324 个资源包。对大多数人来说是合适的起点。',
      fullMagic:
        '几乎所有公开的万智牌模板——526 个资源包。此合集是本项目的衡量标准。',
    },
    katalogUeberschrift: '已安装的资源包',
    katalogLeer: '尚未安装任何资源包。没有游戏和样式时，预览会退回到占位显示。',
    katalogName: '名称',
    katalogArt: '类型',
    katalogVersion: '版本',
    katalogHandlung: '操作',
    katalogBefunde: '读取资源包目录时：',
    anlegen: '创建资源包',
    anlegenArt: '新资源包的类型',
    anlegenArtSpiel: '游戏（.mse-game）',
    anlegenArtStil: '样式（.mse-style）',
    anlegenArtInclude: '包含文件（.mse-include）',
    anlegenArtExport: '导出模板（.mse-export-template）',
    anlegenName: '简称',
    anlegenNameHinweis: '字母、数字、连字符和下划线——它将成为目录名',
    anlegenVollname: '全称',
    anlegenVollnameHinweis: '留空表示与简称相同',
    anlegenSpiel: '新资源包所属的游戏',
    anlegenSpielHinweis: '样式用于排布其游戏的字段，导出模板则用于写出这些字段',
    anlegenSpielLeer: '请选择',
    anlegenKnopf: '创建',
    anlegenGelungen: '已创建：{name}',
    loeschen: '删除',
    loeschenAn: '删除——{name}',
    handlungenAn: '操作——{name}',
    handlungenKurz: '…',
    katalogSuche: '搜索',
    auswahlLoesen: '显示所有包',
    leiste: {
      bezeichnung: '包数据',
      installiert: '已安装',
      befund: '有问题',
      nichtGelesen: '未读取',
      auswahl: '有问题的{pakete}',
    },
    katalogSucheHinweis: '名称、类型、来源或分组',
    katalogArtAlle: '所有类型',
    katalogOrdnung: '排序方式',
    katalogOrdnungGruppen: '分组',
    katalogOrdnungName: '名称',
    katalogOrdnungVersion: '版本',
    katalogOrdnungHerkunft: '来源',
    katalogGefunden: '{gefunden} / {gesamt} 个资源包',
    katalogNichtsGefunden:
      '没有资源包符合此搜索。请换一种写法，或清除类型筛选。',
    katalogLeerTitel: '还没有资源包',
    katalogNichtsTitel: '没有匹配项',
    katalogOhneGruppe: '未分组',
    katalogGruppeAn: '分组 {name}',
    katalogGruppe: '分组',
    katalogOhneVersion: '无版本',
    katalogZustandFehlt: '缺少依赖项',
    katalogZustandZuAlt: '依赖项版本过旧',
    katalogZustandUnlesbar: '版本无法比较',
    katalogZustandZyklus: '依赖项循环',
    katalogEinVorbehalt: '1 项发现',
    katalogVorbehalte: '{anzahl} 项发现',
    umschlagTitel: '来自信封的模板',
    umschlagErklaerung:
      '信封中装的是模板的**文件**；你需要选择带扩展名的名称，例如 `mygame.mse-game` 或 `mystyle.mse-style`。之后模板就以该名称存放在此处。如果该资源包已存在，不会覆盖任何内容：你会收到拒绝提示，并需另选名称。',
    umschlagName: '资源包的名称',
    umschlagOhneNamen: '请先填写资源包的名称。',
    umschlagAnnehmen: '接收模板',
    umschlagAngenommen: '模板现已存放在你的资源包文件夹中。',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: '替换',
    erklaerung:
      '替换此系列文本字段中的文字。先预览，再应用——一次撤销即可还原整个操作。',
    suchtext: '查找',
    ersatz: '替换为',
    grossKlein: '区分大小写',
    ganzesWort: '仅匹配整词',
    vorschauen: '预览更改',
    ersetzen: '立即替换（{treffer}）',
    nichtsGefunden: '没有找到。换个词试试？',
    zusammenfassung: '匹配：{treffer} · 受影响的卡牌：{karten}',
    getan: '已替换——匹配：{treffer} · 卡牌：{karten}',
    ruecknahme: '按 ⌘Z 可撤销整个操作。',
    stellen: '受影响的位置',
    kartenfeld: '卡牌 {nummer} · {feld}',
    weitere: '其余位置未列出——它们同样会被替换。',
  },

  statistik: {
    worueber: '统计依据',
    belegungWahl: '各字段填写率',
    gruppeSpielpaket: '按游戏资源包',
    gruppeFelder: '按字段值',
    verteilungVon: '“{feld}”的分布——占带有该字段的卡牌的比例',
    nachDimension: '按“{dimension}”统计',
    titel: '统计',
    karten: '卡牌',
    felder: '不同字段',
    leere: '无任何值的卡牌',
    belegung: '各字段的填写率——占系列中所有卡牌的比例',
    belegungLeer: '此系列尚未包含任何字段值。',
    verteilungLeer: '此字段在该系列中没有任何值。',
    anteil: '占比',
    dimensionenErklaerung:
      '游戏资源包会自行说明可以统计什么——颜色类别、法术力值、稀有度。' +
      '每个维度都要逐张卡牌计算其值，' +
      '因此需要稍等片刻。',
    dimensionLeer: '没有卡牌在此维度上带有值。',
    gruppeHinweis: '选择一根柱条，即可在卡牌列表中查看它的卡牌。',
    gruppeGewaehlt: '已选：“{gruppe}” · 卡牌：{karten}',
    gruppeZeigen: '显示卡牌',
    gruppeAufheben: '取消选择',
    gruppeLeer: '没有卡牌带有此值。',
    filterBezeichnung: '来自统计的筛选',
    filterKopf: '来自统计：“{quelle}”=“{gruppe}”',
    filterZurStatistik: '返回统计',
    filterAufheben: '取消筛选',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: '样式',
    vorschauTitel: '预览',
    vorschauLeer: '此系列还没有卡牌——请先创建一张，才能看到样式。',
    eigeneFarbe: '自定颜色',
    vorgabe: '默认',
  },
  suche: {
    titel: '搜索',
    eingabe: '要找什么？',
    recherche: '研究…',
    feld: '仅在此字段中',
    alleFelder: '所有字段',
    set: '仅在此系列中',
    alleSets: '所有系列',
    leer: '尚未搜索。输入内容即可查找卡牌。',
    nichts: '没有找到。换个词试试？',
    gekuerzt: '命中的结果多于此处所示。请缩小搜索范围以查看。',
    treffer: '命中',
  },
  ausgeben,
  hilfe: {
    titel: '帮助',
    erklaerung: '此应用中所用术语的含义——以及首次启动时该做什么。',
    glossar: '术语',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: '本页内容',
    seitenuebersicht: '各页面的作用',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
