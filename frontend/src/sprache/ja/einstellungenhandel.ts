// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'トレードを許可',
  'richtlinie.handel.markt': '外部プロバイダを通じたトレードを許可',
  'richtlinie.handel.menge': '1行あたりの最大数量',
  'richtlinie.handel.wert': 'オファーの最大額（最小単位。EURの場合はセント）',
  'richtlinie.handel.frist': '最長の保持時間（分）',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    '共有、デッキの貸し出し、交換、マーケットをまとめてオフにします。**これは好みではなく制限です**：ここでオフにしたものは、どのプロファイルでも再びオンにできません。',
  'richtlinie.handel.markt':
    '初期設定は**オフ**です。マーケットは外部のプロバイダとお金にかかわります。それ以外はすべて、2人の間だけで完結します。',
  'richtlinie.handel.menge':
    'オファーの1行に記載できる枚数の上限です。**空欄の場合は制限なしです。**',
  'richtlinie.handel.wert':
    'オファー全体の価値の上限です。評価が**契約内容の一部**である場合にのみ適用されます。単なるマーケット表示では制限されません。**空欄の場合は制限なしです。**',
  'richtlinie.handel.frist':
    '約束によってカードを確保しておける期間です。**空欄の場合は制限なしです。**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'オフ',
  'richtlinie.handel.an': 'オン',
  'richtlinie.handel.markt.aus': 'オフ',
  'richtlinie.handel.markt.an': 'オン',
}
