// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: '卡牌工作区',
  inspectorBezeichnung: '字段与检查',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: '筛选卡牌',
    platzhalter: '筛选——例如 type:creature -goblin',
    leeren: '清除筛选',
    hilfe:
      '空白分隔各部分，“…”将它们合在一起，field:value 绑定到某个字段，开头的 - 表示取反。不区分大小写。',
    treffer: '{gezeigt} / {gesamt} 张卡牌',
  },
  zoom: {
    bezeichnung: '预览大小',
    anpassen: '适合窗口',
    druck: '打印尺寸',
    ohneDruck: '模板没有给出分辨率，因此没有打印尺寸。',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: '拖动图片',
    hinweis: '将图片拖到访达，即可把卡牌存为 PNG。“{weg}”也能做到。',
  },
  overlays: {
    bezeichnung: '叠加层',
    bildgrenzen: '图像边界',
    beschnitt: '出血',
    diagnose: '诊断',
    beschnittHinweis: '线内的内容不会被 3 mm 的裁切裁掉。',
  },
  probleme: {
    ueberschrift: '问题',
    keine: '这张卡牌没有问题。',
    zumFeld: '转到字段 {feld}',
    ohneFeld: '无字段',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: '上一张卡牌',
    naechste: '下一张卡牌',
    stelle: '第 {stelle} 张，共 {gesamt} 张',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
