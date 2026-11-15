// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': '允许交易',
  'richtlinie.handel.markt': '允许通过外部提供商交易',
  'richtlinie.handel.menge': '每行最大数量',
  'richtlinie.handel.wert': '单个报价的最高价值，以最小单位计（欧元为分）',
  'richtlinie.handel.frist': '最长承诺期限（分钟）',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    '同时关闭分享、套牌借用、交换和市场。**这是一道限制，不是偏好**：此处关闭的内容，在任何配置文件中都无法再次开启。',
  'richtlinie.handel.markt':
    '默认为**关闭**。市场会接触外部提供商和资金——其余一切都只发生在两个人之间。',
  'richtlinie.handel.menge':
    '一个报价的单独一行最多可以写多少张。**留空表示不设上限。**',
  'richtlinie.handel.wert':
    '一个报价合计最多可值多少。仅当估价属于**合同内容**时才生效——单纯的市场行情视图不受其限制。**留空表示不设上限。**',
  'richtlinie.handel.frist':
    '一项承诺最长可以占用副本多久。**留空表示不设上限。**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': '关闭',
  'richtlinie.handel.an': '开启',
  'richtlinie.handel.markt.aus': '关闭',
  'richtlinie.handel.markt.an': '开启',
}
