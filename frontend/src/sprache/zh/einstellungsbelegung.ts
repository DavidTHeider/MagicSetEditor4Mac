// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: '⌘0 至 ⌘9 的分配',
  belegungHinweis:
    '按下 ⌘ 加数字键所打开的页面。默认分配为侧边栏中的各个区域；“与 0.6.0 之前相同”会恢复 0.5.0 及更早版本的分配，其对照表见更新日志。',
  belegungFuer: '{kuerzel} 对应的页面',
  belegungFrei: '—',
  belegungVor06: '与 0.6.0 之前相同',
  belegungVorgabe: '恢复默认',
  kuerzel: {
    palette: '打开命令面板',
    suche: '在系列中搜索',
    zurueck: '撤销更改',
    vor: '重做更改',
    sichern: '存储系列',
    sichernUnter: '将系列存储为…',
    oeffnen: '打开系列…',
    neu: '新建系列',
    schliessen: '关闭系列',
    drucken: '前往打印页',
    navigator: '显示或隐藏导航栏',
    inspector: '显示或隐藏检查器',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: '上一页',
    seiteVor: '下一页',
  },
} as const
