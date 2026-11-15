// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: '全部赛事（{anzahl}）',
  titel: '赛事',
  erklaerung:
    '赛事是随时间展开的过程：轮次、对阵、结果。排名由此计算得出，不存放在任何地方——更正一个结果，立刻就能看到新的排名。',

  leer: '还没有赛事。用上方的“创建赛事…”开始一个，或用“导入赛事…”从文件导入。',

  neu: '创建赛事',
  neuOeffnen: '创建赛事…',
  name: '名称',
  spiel: '游戏',
  notiz: '备注',
  einzelspiel: '每场比赛只打一局',
  einzelspielHinweis:
    '此时对局胜率和对手对局胜率没有意义——它们只是照搬比赛的数据。但仍然会计算。',
  anlegen: '创建',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: '四步创建新赛事',
  schrittTurnier: '赛事',
  schrittRegeln: '规则',
  schrittTeilnehmende: '参赛者',
  schrittErsteRunde: '第一轮',
  weiter: '下一步',
  zurueck: '返回',
  verlassen: '稍后继续',
  vorlage: '模板',
  ohneVorlage: '无模板',
  vorlageHinweis:
    '模板带有游戏、赛制、规则和备注。如果需要，也可带上选手姓名，但绝不带套牌：提交的清单只属于一场赛事。',
  mitTeilnehmenden: '沿用选手姓名',
  nameWarum: '必填——以便你在列表中再次找到它。同一游戏的两场赛事名称不同。',
  spielWarum: '必填——它决定哪些赛制和套牌适用。',
  format: '赛制',
  ohneFormat: '无赛制',
  formatWarum: '可选——可从此游戏的赛制中选取。',
  regelnAusVorlage: '规则来自模板。',
  schweizerSystem:
    '对阵遵循瑞士制：积分相同的选手互相对阵，且任何人都不会遇到同一位对手两次。应用会根据选手人数建议轮数。',
  zuWenige: '一轮需要两个人互相对阵——请至少报名两位。',
  ersteRundeHinweis: '第一轮大致如下。在你按下按钮之前，不会进行任何对阵。',
  ersteRundeAuslosen: '为第 1 轮排定对阵',

  waehlen: '打开',
  loeschen: '删除赛事',
  handlungenAn: '操作——{name}',
  handlungenKurz: '…',
  turniermenue: '锦标赛',
  loeschenFrage:
    '要删除这场赛事吗？轮次、对阵和结果会一并删除。选手的套牌会保留。',
  loeschenJa: '删除',
  loeschenNein: '取消',

  teilnehmer: '参赛者',
  teilnehmerLeer: '还没有人报名。',
  anmelden: '报名',
  gebiete: '此赛事的各个部分',
  gebietTeilnehmer: '参赛者',
  gebietRunden: '轮次',
  gebietRangliste: '排名',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: '结束',
  abgeschlossenAm: '已于 {datum} 结束。选手、轮次和结果均已定稿。',
  wiedereroeffnen: '重新开启',
  nichtWiedereroeffnen: '不重新开启',
  wiedereroeffnenFrage: '要重新开启这场赛事吗？',
  wiedereroeffnenErklaerung:
    '之后选手、轮次和结果又可以更改。日志会记录此事已经发生。',
  gebietProtokoll: '日志',
  protokollLeer: '还没有发生任何事。',
  protokollAngelegt: '已创建为“{angaben}”',
  protokollAngemeldet: '{angaben} 已报名',
  protokollAusgestiegen: '{angaben} 已退赛',
  protokollEingestiegen: '{angaben} 已重新报名',
  protokollRunde: '第 {angaben} 轮已排定对阵',
  protokollErgebnis: '对阵 {paarung} 的结果：{angaben}',
  protokollZurueckgenommen: '对阵 {paarung} 的结果已撤回',
  protokollAbgeschlossen: '已结束',
  protokollWiedereroeffnet: '已重新开启',
  protokollEingelesen: '已导入为“{angaben}”',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: '导出…',
  berichtSichern: '存储报告…',
  einlesen: '导入赛事…',
  gesichertUnter: '已存储到 {pfad}。',
  eingelesen: '已导入并打开。',
  einlesebefunde: '导入时注意到：',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: '瑞士制 · 第 {runde} / {von} 轮 · {anzahl} 位选手',
  kopfzeileOhneRunde: '瑞士制 · {anzahl} 位选手',
  kopfzeileOhneEmpfehlung: '瑞士制 · 第 {runde} 轮 · {anzahl} 位选手',
  gebietDecklisten: '套牌清单',
  decklistenLeer: '还没有人提交清单——报名时选择套牌，即提交其清单。',
  deckAufschlagen: '打开套牌',
  partieOeffnen: '打开对局',
  deck: '套牌',
  ohneDeck: '无套牌',
  deckHinweis:
    '如果指定了套牌，会复制其清单而不是引用它：赛事中的套牌清单是一份已提交的文件。',
  deckabzug: '已提交的清单',
  deckabzugLeer: '尚未提交套牌。',
  deckabzugUnbenannt: '卡牌',
  aussteigen: '退赛',
  einsteigen: '重新参赛',
  ausgestiegen: '已退赛',
  ausgestiegenHinweis:
    '他们不再被排入对阵，但仍保留在排名中：其胜率仍会计入对手的胜率。',

  runden: '轮次',
  rundenLeer: '还没有排定任何一轮。',
  runde: '第 {nummer} 轮',
  empfehlung:
    '第 {gelegt} / {empfohlen} 轮——赛事规则建议 {spieler} 位选手进行 {empfohlen} 轮。这是建议，而非规定。',
  vorschlag: '预览下一轮',
  vorschlagTitel: '下一轮大致如下',
  vorschlagHinweis: '仅为预览，尚未创建。只有下方的按钮才会创建。',
  rundeLegen: '创建这一轮',
  bye: '轮空（2–0）',
  gegen: '对',
  offen: '尚未进行',
  spiele: '对局数',
  spieleA: '第一位选手的对局数',
  spieleB: '第二位选手的对局数',
  unentschiedenSpiele: '平局',
  eintragen: '记录',
  zuruecknehmen: '撤销',
  ergebnis: '结果',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: '这次更正会改变排名。',
  folgenPlatz:
    '{name}：第 {vorher} 名 → 第 {nachher} 名，{punkteVorher} → {punkteNachher} 分',
  folgenSpaeter: '第 {runden} 轮是按旧排名排定的，保持不变。',
  folgenEintragen: '仍然录入',
  folgenNicht: '不录入',
  ergebnisZahlen: '对局数按从零开始的整数计数。',
  ergebnisOhneSpiel: '没有一局对局的结果不算结果——还没打的人还没有结果。',
  ergebnisZuViele: '这里一场比赛最多进行 {hoechstens} 局。',

  rangliste: '排名',
  ranglisteLeer: '还没有进行任何比赛。',
  platz: '名次',
  bilanz: '战绩',
  matchpunkte: '积分',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    '按《万智牌赛事规则》的顺序使用四项同分裁定：积分、对手比赛胜率、对局胜率、对手对局胜率。每项百分比的下限为 0.33。',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: '在此进行',
  matchErklaerung:
    '如果这场对阵在本程序中进行，请将它绑定到对局。结果由双方分别上报，只有两边一致时才会录入。',
  matchNichtGebunden: '这场对阵没有关联对局。',
  matchAnsehen: '查看',
  matchPartie: '对局 ID',
  matchSitzA: '第一位选手的座位',
  matchSitzB: '第二位选手的座位',
  matchBinden: '绑定到此对局',
  matchGebunden: '已绑定到对局',
  matchSitz: '座位',
  matchDeckabdruck: '套牌指纹',
  matchKeineMeldung: '还没有人上报。',
  matchMeldungVon: '上报者：',
  matchMelden: '为此座位上报',
  matchAusDerPartie: '根据对局上报结果',
  matchAusDerPartieHilfe: '比分和凭证来自已结束的对局——你只需说明是谁在上报。',
  matchNachweis: '凭证（最后一个链节）',
  matchUebernehmen: '录入结果',
  matchUebernommen: '结果已录入赛事。',
  matchGrenze:
    '凭证说明双方看到的是同一段历史。它不说明这局对局是否正确进行——而且私人对局并不等同于官方赛事认证。',
}
