// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: '概览',
    mitProblemen: '{anzahl} 个有问题',
    geprueftVon: '已检查 {geprueft}/{gesamt}',
    mitBefund: '{anzahl} 个有发现',
    nichtGeprueft: '未完全检查',
    problemeGeprueft: '问题检查于{wann} · {geprueft}/{gesamt} 张卡牌',
    problemeOffen: '问题：{gesamt} 个系列中有 {offen} 个未检查',
    pruefen: '立即检查',
    setsMitProblemen: '有问题的{sets}',
    kartenMitBefund: '{sets}中有发现的{karten}',
    aeltesterPreis: '最旧价格 {wann}',
    ohnePreis: '{anzahl} 个无价格',
    mitFehlenden: '{anzahl} 个缺少卡牌',
    decksMitFehlenden: '缺少卡牌的{decks}',
    alter: {
      gerade: '刚刚',
      minuten: '{minuten} 分钟前',
      stunden: '{stunden} 小时前',
      tage: '{tage} 天前',
      tag: '1 天前',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: '需要留意',
    nichts: '没有——已检查 {sets}、{decks} 和 {pakete}。',
    durchgesehen: '已检查 {sets}、{decks} 和 {pakete}。',
    arten: {
      setVorbehalte: '{anzahl} 项未解读',
      deckFehlend: '{anzahl} 张不在收藏中',
      paketBefund: '脚本中有 {anzahl} 项发现',
    },
    paketeMitBefund: '脚本有问题的{pakete}',
    gruendeZeigen: '显示原因',
    gruendeVerbergen: '隐藏原因',
    setBefunde: '有发现的{karten}',
    befunde: '未检查：',
    alle: '显示全部 {anzahl} 项',
    weniger: '显示较少',
    gruende: {
      vorbehalt: '{0}，第 {1} 行',
    },
    amOrt: '{ort}：{satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: '资料库旁仅剩 {frei} 可用空间',
        sicherungAlt: '上次备份是在 {tage} 天前',
        dienstBeendet: '登录时启动的服务以代码 {ausgang} 结束',
        webprozessNeu: '网页视图已重新加载 {anzahl} 次',
      },
      rat: {
        platzKnapp:
          '可用空间低于约 2 GB 或 5 % 时，每次备份都会很紧张——请在“数据与备份”中清理。',
        sicherungAlt: '已超过 14 天——请在“数据与备份”中备份资料库。',
        dienstBeendet: '出错后 launchd 会重新启动它——结束的原因请查看诊断。',
        webprozessNeu:
          'WebKit 结束了它的进程；打开的表单中未保存的内容可能丢失——请在诊断中检查。',
      },
    },
  },
  auftraege: {
    titel: '后台任务',
    staende: {
      laeuft: '运行中',
      fertig: '已完成',
      gescheitert: '失败',
      abgebrochen: '已取消',
    },
    arten: {
      preise: '获取收藏的价格',
      sammeldaten: '从 Scryfall 载入批量数据',
      fassung: '获取新版本',
      bogen: '排布打印页',
      set: '读取系列',
      mtgjson: '读取 MTG JSON',
      vorlage: '通过模板导出',
      installer: '读取安装包',
      weitergabe: '作为安装包分享',
      probleme: '检查系列中的问题',
    },
    abbrechen: '取消“{art}”',
    dauer: '{sekunden} 秒',
    laufend: { eine: '个任务正在运行', viele: '个任务正在运行' },
  },
  weiterarbeiten: {
    titel: '继续工作',
    weiterAn: '继续处理{art}“{name}”',
    zuletztWann: '上次{vorgang}：{zeit}',
    angeheftet: '已固定',
    angeheftetLeer: '没有已固定的内容——固定一行即可将它保留在这里。',
    leer: '你最近的工作会显示在这里。目前还是空的——第一步是创建系列或获取模板资源包。',
    handlungSets: '创建或打开系列',
    handlungPakete: '获取资源包',
    anheften: '固定',
    loesen: '取消固定',
    anheftenVon: '固定“{name}”',
    loesenVon: '取消固定“{name}”',
    arten: { set: '系列', deck: '套牌', paket: '资源包' },
    vorgaenge: {
      geoeffnet: '打开',
      eingefuehrt: '导入',
      ausgegeben: '导出',
    },
  },
  titel: '主页',
  ueberschrift: 'MagicSetEditor4Mac',
  was: '设计、审阅和打印万智牌及其他游戏的卡牌——使用 Magic Set Editor 2 所用的模板。',
  ersteSchritte: '第一步',
  einfuehrung: {
    willkommen: '三步即可完成第一张卡牌。你随时可以在设置中重新打开这个窗口。',
    wiederzeigen: '启动时再次显示',
    schliessen: '开始吧',
    bezeichnung: '简介',
  },
  schritte: {
    setOeffnen: {
      titel: '打开系列',
      erklaerung: '打开已有的 .mse-set 文件——或从 MTG JSON 导入系列。',
      handlung: '前往系列',
    },
    kartenBearbeiten: {
      titel: '编辑卡牌',
      erklaerung: '查看并更改已打开系列的卡牌；预览会即时显示每一处更改。',
      handlung: '前往卡牌',
      ersteinSet: '这需要先打开一个系列。',
      stattdessen: '请先打开一个系列',
    },
    paketeHolen: {
      titel: '获取模板',
      erklaerung:
        '这里已有八个模板资源包。想要更多游戏或样式？可以在这里获取。',
      handlung: '前往资源包',
    },
  },
}
