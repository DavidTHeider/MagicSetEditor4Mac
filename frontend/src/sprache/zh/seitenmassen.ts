// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: '预览',
  vorschauErklaerung:
    '预览会逐张显示哪些内容将被更改。只有点按“应用”后才会写入——而且仅当期间没有其他人写入过。',
  karte: '卡牌',
  alt: '之前',
  neu: '之后',
  keinWertBisher: '无值',
  bleibtGleich: '保持不变——已是此值',
  zusammenfassung: '{geaendert} 张卡牌将被更改，{gleich} 张保持不变。',
  nichtsZuTun: '所有选中的卡牌已是此值。无需更改。',
  konflikt: '这些卡牌自预览以来已发生变化。没有写入任何内容——请重新获取预览：',
  inzwischenGeaendert: '此后已更改',
  angewendet: '已更改 {anzahl} 张卡牌。',
} as const
