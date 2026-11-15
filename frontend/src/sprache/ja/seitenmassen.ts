// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'プレビュー',
  vorschauErklaerung:
    'プレビューには、カードごとに何が変わるかが表示されます。「適用」するまで何も書き込まれず、その間に他の誰も書き込んでいない場合にのみ書き込まれます。',
  karte: 'カード',
  alt: '変更前',
  neu: '変更後',
  keinWertBisher: '値なし',
  bleibtGleich: '変更なし — すでにその値です',
  zusammenfassung:
    '{geaendert} 枚のカードが変更され、{gleich} 枚はそのままです。',
  nichtsZuTun:
    '選択したカードはすべて、すでにこの値になっています。変更する項目はありません。',
  konflikt:
    'これらのカードは、プレビューの後で変更されました。何も書き込まれていません。プレビューを取得し直してください：',
  inzwischenGeaendert: 'その後に変更',
  angewendet: '{anzahl} 枚のカードを変更しました。',
} as const
