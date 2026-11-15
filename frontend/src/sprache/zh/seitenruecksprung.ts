// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: '回退',
  hinweis:
    '有四样东西可以回退，代价各不相同。其中两样本应用程序无法自己完成——此时这里会说明由谁来做。',

  arten: {
    appfassung: '应用程序',
    schema: '资料库的结构',
    daten: '数据',
    pakete: '模板包',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      '本应用程序不再运行后，由人在访达中替换应用程序包。程序在运行时替换自己，需要另一个对“应用程序”文件夹有写入权限的进程。',
    nurVorwaerts: '迁移无法向后回退。回退的办法是备份：恢复当时的状态。',
    unbekannteArt: '应用程序不认识这一类型。',
  } as Record<string, string>,

  gehtHier: '本应用程序可以自己完成。',

  vergleichTitel: '恢复的代价',
  datenstand: '恢复后的状态',
  spalteGegenstand: '项目',
  spalteSicherung: '备份中',
  spalteJetzt: '现在',
  spalteVerlust: '将丢失',

  gegenstaende: {
    sets: '系列',
    karten: '卡牌',
    decks: '套牌',
    exemplare: '实体卡',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    '此处会丢失的内容，你可以事先导出：系列通过“存储为”，收藏通过“导出”。',

  unklarTitel: '恢复后不明确',
  unklarHinweis:
    '这些实体卡在恢复后会重新出现在你的收藏中。是否还在你手里，没有人知道——它们此后可能已被送出。在你核实之前，它们不参与交易；不会扣除任何数量。',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    '无法获得收藏之间的对比——此备份或你的资料库没有提供。恢复后，请手动核对你的收藏。',

  rueckkehrenTitel: '已记录的恢复',
  keineRueckkehren:
    '还没有。一旦你恢复备份，它就会出现在这里——连同它所回到的状态，以及它留下的不明确项。',
  rueckkehrZeile: '于 {zurueckgespielt} 恢复到 {datenstand} 的状态',
  ausDatei: '来自 {quelle}',
  offen: '待处理',
  quittiert: '已于 {geklaertAm} 核实',
  betrifft: '涉及 {menge} 张',
  quittieren: '已核实',
  quittierenLaeuft: '正在记录…',
  quittierenHinweis: '这表示你已手动核实了受影响的实体卡。它不会扣除任何数量。',
}
