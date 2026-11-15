// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: '通用',
    bereiche: '区域与功能',
    erscheinungsbild: '外观',
    nutzer: '用户与个人资料',
    fach: '专业功能',
    ki: 'AI',
    verbindungen: '连接',
    dienste: '服务',
    daten: '数据与备份',
    diagnose: '诊断',
  },

  untertitel: {
    allgemein: '语言、启动、快捷键',
    bereiche: '工作区、功能、成熟度',
    erscheinungsbild: '主题、字体、窗口',
    nutzer: '登录、解锁',
    fach: '卡牌、收藏、对局',
    ki: '提供方、模型、费用',
    verbindungen: '网络、对方、命令行、MCP、REST',
    dienste: '卡牌服务、价格、市场',
    daten: '资料库、路径、备份',
    diagnose: '日志、测量、技术信息',
  },

  abschnitte: {
    fenster: '窗口与侧栏',
    spalten: '栏与视图',
    handel: '交易限制',
    protokoll: '日志',
    messung: '测量',
    speicher: '缓存与搜索索引',
    adressen: '地址',
    ordner: '文件夹与文件名',
    eingriffe: '回退、整理与重置',
    mitnehmen: '带走设置',
  },
} as const
