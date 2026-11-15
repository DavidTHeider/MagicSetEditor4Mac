import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert: '此字段已在别处改为“{wert}”。你的草稿尚未保存。',
    behalten: '保存我的草稿',
    nehmen: '采用新值',
  },
  listenansicht: {
    suche: '搜索',
    suchePlatzhalter: '名称或游戏',
    sortierung: '排序',
    ansicht: '视图',
    liste: '列表',
    kacheln: '平铺',
    zahl: '{gezeigt} / {gesamt}',
    nichtsPasst: '没有与“{suche}”匹配的内容。',
    sucheLeeren: '清除搜索',
    filterZuruecksetzen: '重置搜索和筛选',
    weitere: '再显示 {anzahl} 个',
    wieImBestand: '按存储顺序',
    zuletztBenutzt: '最近使用的在前',
    nachName: '按名称',
    nachKarten: '按卡牌数量',
  },
  speicherstand: {
    schreibt: '正在存储…',
    geschrieben: '每项更改都已存入资料库。',
    fehler: '“{feld}”未存储——原因显示在该字段处。',
  },
  gegenueberTitel: '添加或确认对方',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: '搜索对方 …',
  feldhilfe: '{feld} 的说明',
  mehrZu: '更多关于“{satz}”',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: '取自档案 {quelle}。',
  herkunftAusEinfuhr: '从 {quelle} 导入。',

  pflichtfeldLeer: '{feld} 需要填写。',

  waehrung: {
    beispiel: '例如 EUR',
    keineGestalt: '“{wert}”不是货币代码——货币代码由三个字母组成，如 EUR。',
    grossSchreiben: '“{wert}”——货币代码须大写：{gross}。',
  },

  drucksprachen: {
    beispiel: '例如 en',
    namen: {
      en: '英语',
      es: '西班牙语',
      fr: '法语',
      de: '德语',
      it: '意大利语',
      pt: '葡萄牙语',
      ja: '日语',
      ko: '韩语',
      ru: '俄语',
      zhs: '简体中文',
      zht: '繁体中文',
      he: '希伯来语',
      la: '拉丁语',
      grc: '古希腊语',
      ar: '阿拉伯语',
      sa: '梵语',
      ph: '费瑞克西亚语',
      qya: '昆雅语',
      dw: '矮人语',
      tlh: '克林贡语',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: '以明文显示密码',
    zeigenKurz: '显示',
    verbergen: '再次隐藏密码',
    verbergenKurz: '隐藏',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: '验证码',
    abschicken: '确认',
  },

  aktionen: {
    abbrechen: '取消',
    oeffnen: '打开',
    schliessen: '关闭',
    speichern: '存储',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— 选择样式 —',
  },
  druckwahl: {
    drucke: '版本',
    hinweis: '对等端只能看到这些版本，以及你拥有的数量。',
    suchen: '搜索版本',
    gewaehlt: '已选：{anzahl}',
    nurSuche: '你的收藏共有 {anzahl}。请输入名称搜索并选择。',
    keiner: '没有与“{suche}”匹配的版本。',
    genauer: '匹配的版本多于显示的数量。请缩小搜索范围。',
    leer: '你的收藏中还没有可以共享的版本。请先添加实体卡。',
    fehlt: '请至少选择一个版本——没有选择版本的选择不会共享任何内容。',
    behalten: '每个版本我保留的数量',
    behaltenHinweis:
      '对等端能看到超出该数量的部分：保留 1 张而有 3 张时，看到 2 张。',
    detailBehalten: '你保留 {behalten}',
  },
  suchauswahl: {
    keiner: '没有与“{suche}”匹配的项。',
    treffer: '结果：{anzahl}',
    begrenzt: '显示 {anzahl} 个结果中的 {gezeigt} 个——请缩小搜索范围 …',
  },
  zeitzonenwahl: {
    system: '与系统一致',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: '新建', heisst: '新建一项内容。' },
      bearbeiten: { kurz: '编辑', heisst: '就地更改。' },
      sichern: { kurz: '存储', heisst: '保留当前状态。' },
      zuruecknehmen: { kurz: '撤销', heisst: '撤回上一步。' },
      duplizieren: {
        kurz: '复制',
        heisst: '在旁边再放一份。',
      },
      kopieren: { kurz: '拷贝', heisst: '放入剪贴板。' },
      entfernen: {
        kurz: '移除',
        heisst: '从此视图中移除；在其他位置仍会保留。',
      },
      loeschen: { kurz: '删除', heisst: '彻底删除，所有位置都会删除。' },
    },
  },
  nurlesen: {
    hinweis:
      '你正以阅读者身份查看此档案：可以查看全部内容，但不能更改任何内容。请切换到你自己的档案再进行工作，或请管理此档案的人授予你“协作者”角色。',
    knopf: '只读——在此配置文件中你不能更改任何内容。',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: '工作环境',
    profilVon: '所有者：{eigner}',
    rollen: {
      leser: '阅读者',
      mitarbeiter: '协作者',
      verwalter: '管理者',
    },
    nurLesen: '只读',
    zuDenNutzern: '切换档案…',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: '正在加载…',
    leer: '暂无内容',
    fehler: '无法加载。',
    fehlerMitGrund: '无法加载。{grund}',
    wiederholen: '再次加载',
    neuLaden: '重新载入',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: '在文本中查找',
    schliessen: '关闭查找',
    suchen: '查找',
    ersetzen: '替换为',
    treffer: '{bei} / {ganz}',
    keinTreffer: '无匹配项',
    weiter: '下一个匹配项',
    zurueck: '上一个匹配项',
    einsetzen: '替换',
    alleEinsetzen: '全部替换',
    grossKlein:
      '区分大小写——在 MSE-Script 中，“card”与“Card”是两个不同的名称。',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: '哪些卡牌',
    ganzesSet: '整个系列（{zahl} 张卡牌）',
    ganzesSetOhneZahl: '整个系列',
    teilmenge: '仅所选（{zahl}）',
    ohneNamen: '未命名',
    eingrenzen: '筛选卡牌',
  },

  vorschlaege: {
    bezeichnung: '建议',
    uebernehmen: '按 ↹ 采纳',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: '“{region}”的宽度',
    breiteWert: '{punkte} 点',
    navigatorAuf: '显示导航器',
    navigatorZu: '隐藏导航器',
    inspectorAuf: '显示检查器',
    inspectorZu: '隐藏检查器',
    navigatorEng: '宽度不足以容纳三栏——导航器将在中央区域上方展开',
    inspectorEng: '宽度不足以容纳三栏——检查器将在中央区域上方展开',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: '临时',
    vorlaeufigErklaerung:
      '该印刷版本还没有 Scryfall 标识符。在其确定之前，不会显示价格和图像。',
  },
  kommandozentrale: {
    titel: '命令中心',
    erklaerung:
      '命令是 MCP 服务器的一个工具，在此窗口中调用——得到的结果与通过 MCP 和命令行调用时相同。输入的内容不会经过 shell：你输入的始终只是数据。',
    lesend: '读取',
    schreibend: '更改',
    gefaehrlich: '难以撤销',
    faehigkeit: '功能',
    cli: '命令行',
    rest: '服务',
    parameter: '参数',
    keineParameter: '此命令不需要参数。',
    pflicht: '必填',
    jsonHinweis: '以 JSON 填写——列表用方括号，对象用花括号。',
    befundPflicht: '此字段需要填写。',
    befundZahl: '这不是数字。',
    befundJson: '这不是 JSON。',
    vorschau: '将发送的内容',
    bestaetigen: '我已阅读预览，并要按此执行。',
    ausfuehren: '执行',
    abbrechen: '取消',
    laeuft: '命令正在执行…',
    ergebnis: '结果',
    ausgabe: '命令输出',
    alsText: '文本',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      '尚未执行任何命令。机密信息——令牌、密码——绝不会以明文显示在这里。',
    verlauf: '本次会话的历史',
    wiederholen: '再次载入 {name}',
    gelungen: '成功',
    gescheitert: '失败',
    abgebrochen: '已取消',
    schliessen: '关闭',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: '打开诊断',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: '设置 AI',
  },
  kikasten: {
    titel: '此卡的 AI',
  },
} as const
