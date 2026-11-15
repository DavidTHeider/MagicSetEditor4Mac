// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: '张卡牌', viele: '张卡牌' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: '页', viele: '页' },

    // And the three objects of the overview (#1422).
    sets: { eine: '个系列', viele: '个系列' },
    decks: { eine: '副套牌', viele: '副套牌' },
    pakete: { eine: '个软件包', viele: '个软件包' },
    eigenePakete: { eine: '个自有软件包', viele: '个自有软件包' },
    stile: { eine: '种样式', viele: '种样式' },
    positionen: { eine: '项', viele: '项' },
    exemplare: { eine: '张实体卡', viele: '张实体卡' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: '个版本', viele: '个版本' },
  },
  bestandGesichert: '资料库已备份到',
  zahlenbereich: '介于 {kleinstens} 与 {groesstens} 之间的数字。',
  zahlenAb: '不小于 {kleinstens} 的数字。',
  ganzeZahlBereich: '介于 {kleinstens} 与 {groesstens} 之间的整数。',
  ganzeZahlAb: '不小于 {kleinstens} 的整数。',
  nurZiffern: '仅限数字，不含小数点和逗号——金额以最小单位计。',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: '无法打开文件 {pfad}。{grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: '语言',
    de: 'Deutsch',
    en: 'English',
    es: 'Español',
    fr: 'Français',
    it: 'Italiano',
    ja: '日本語',
    ru: 'Русский',
    zh: '简体中文',
  },
  // **Die fünf Arbeitsbereiche, der Einstieg und die Dienstzone** — UX-02
  // (#1674).
  //
  // **`bereiche` und `dienste` sind die zwei Überschriften der Leiste**, nicht
  // Bereiche selbst — sie sagen, was darunter steht.
  bereiche: {
    // **„Home" und nicht „Start"** — so heißt die Seite in dieser Sprache
    // (`en/seitenstart.ts`), und der Bereich trägt ihren Namen.
    start: '主页',
    gestalten: '设计',
    sammeln: '收藏',
    spielen: '对战',
    handeln: '交易',
    ressourcen: '资源',
    programm: '应用程序',

    ueberschrift: '工作领域',
  },

  navigation: {
    bezeichnung: '主导航',
    sprungZumInhalt: '跳到内容',
    seitenleisteEinblenden: '显示边栏',
    seitenleisteAusblenden: '隐藏边栏',
    seitenleisteKuerzel: '⌥⌘S',
    weg: '你的位置',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: '无法绘制此区域',
    bereichSatz:
      '编号 {kennung}。数据库的其余部分未受影响；该错误以此编号记录在日志中。',
    neuLaden: '重新加载区域',
    diagnoseOeffnen: '打开诊断',
    wurzelTitel: '无法绘制窗口',
    wurzelSatz:
      '编号 {kennung}。你的数据库未受影响；该错误以此编号记录在日志中。',
    fensterNeuLaden: '重新加载窗口',
  },
  meldungen: {
    methodeFehlt: '方法 {name} 尚未绑定。',
    vorschauLeer: '预览返回时既没有图像，也没有错误。',
    kontoLeer: '账户返回时既没有用户，也没有错误。',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: '窗口消息',
    setNichtGesichert: '系列“{name}”未存储——{grund}',
    setNichtGesichertOhneName: '系列未存储——{grund}',
    bestandNichtGesichert: '资料库未备份——{grund}',
    verlaufGescheitert: '撤销或重做失败——{grund}',
    darstellungGescheitert: '无法重置外观——{grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: '此版本可以打开你的资料库。',
    unvertraeglich: '此版本无法打开你的资料库。',
    ungeklaert:
      '此版本没有说明能否打开你的资料库——它早于这项检查。切换前请先备份你的资料库，出问题时再恢复。',
    bereit: '切换已准备就绪：已备份并已回读。',
    vorbereiten: '准备切换',
    hindernisse: {
      keinManifest: '该发行版没有提供关于自身的说明。',
      plattform: '它是为另一种电脑构建的。',
      wanderungFehlt: '你的资料库有一项它不认识的更改。',
      summeAnders: '有一项更改在它那里的内容不同。',
      vorDerGrundlinie: '你的资料库早于它的基线——请先用中间的版本打开它。',
      platz: '可用空间不足。',
      vorgaengeLaufen: '仍有任务在运行。',
      keineSicherung: '尚未创建备份。',
      sicherungUnlesbar: '无法回读备份。',
      halter: '另一个程序正在占用资料库——请在切换前退出它。',
    },
  },
  fassungsband: {
    bereich: '新版本',
    neu: '版本 {fassung} 已发布。',
    seite: '更新内容',
    laden: '下载',
    laeuft: '正在下载…',
    fortschritt: '已下载 {gelesen}，共 {ganz}',
    fortschrittOhneLaenge: '已下载 {gelesen}',
    geladen:
      '版本 {fassung} 位于 {pfad}。要安装，请在访达中打开它，并将应用程序拖入“应用程序”文件夹。',
    weglegen: '关闭此消息',
  },
  wiederanlauf: {
    bereich: '上次会话',
    satz: '上次会话（开始于 {zeit}）意外结束。',
    satzOhneZeit: '上次会话意外结束。',
    mitSpur: '有一份崩溃记录；创建诊断包时会包含它。',
    ohneSpur: '没有崩溃记录——应用可能是被外部结束的。',
    ausblenden: '隐藏',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: '系列数据',
      befund: '有问题',
      nichtGeprueft: '未检查',
      nichtGelesen: '未读取',
      geprueft: '已检查 {wann}',
    },
    laschen: '系列的各个分区',

    gruppen: {
      bauen: '制作卡牌',
    },
    schubladen: {
      eigenschaften: {
        knopf: '属性',
        titel: '系列属性',
        laschen: '属性的各个分区',
        hinweis: '适用于此系列中所有卡牌的内容。',
      },
      werkzeuge: {
        knopf: '工具',
        titel: '工具',
        laschen: '系列的工具',
        hinweis: '此系列的统计、随机补充包和控制台。',
      },
      schliessen: '关闭',
    },

    keinSet: '没有打开的系列',
    keinSetErklaerung:
      '在“系列”中选择一个系列、新建一个或导入一个——它的卡牌、统计和打印页就会显示在这里。',
    zuDenSets: '前往系列',
    zuDenPaketen: '前往资源包',
    zuDenKarten: '前往卡牌',
    ohneKartenTitel: '这个系列还没有卡牌',
    ohneKarten: '此标签页使用该系列的卡牌。卡牌在“卡牌”中创建、粘贴或导入。',
    ohneNamen: '已打开的系列',
  },
  palette: {
    bezeichnung: '命令面板',
    eingabe: '前往哪里？',
    keineTreffer: '未找到内容。换个词试试？',
    oeffnen: '打开命令面板',
    seiten: '页面',
    befehle: '命令',
    befehlEingabe: '前往哪里，或执行哪个命令？',
    tun: '执行',
    finden: '查找',
    karteErstellen: '创建卡牌',
    setWaehlen: '新建或打开系列',
    deckSpielen: '用{name}快速对局',
    tauschenMit: '与 {name} 交换卡牌',
    sicherungErstellen: '创建备份',
    schnittstellenEinrichten: '设置接口（命令行、MCP、REST）',
    woKarte: '设计 · {set}',
    woSets: '设计 · 系列',
    woPartie: '对战 · 开始对局',
    woHandel: '交易 · 对方',
    woEinstellungen: '设置 · {gruppe}',
    alsKarte: '卡牌 · {set}',
    alsSet: '系列 · {karten}',
    alsDeck: '套牌 · {spiel}',
    alsGegenueber: '对方 · 已确认',
    fuss: '↑↓ 选择 · ↵ 执行 · esc 关闭',
  },
} as const
