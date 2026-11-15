// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: '这笔交易的进度',
  phasen: {
    gegenueber: '与谁交易',
    angebot: '交换什么',
    zusagen: '锁定',
    verbuchen: '记录',
  },
  erledigt: '已完成',
  jetzt: '当前',
  kommt: '待进行',
  schritte: '你现在可以做什么',
  keinSchritt: '眼下这里没有需要你做的事。',
  neu: '新的交换：把双方来回交换的内容整理好，然后发起。',
  abgleich: '你们可以交换什么',
  unterwegs: '在途中的内容',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: '对方',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: '历史',
  verlaufWechsel: '修订 {revision}：{von} → {nach} · {zeit}',
  verlaufNeu: '修订 {revision}：{nach} · {zeit}',
  verlaufLeer: '暂无变更。',
  verlaufOhneHandel: '请选择一笔交易——这里会显示它是如何走到这一步的。',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: '添加一行：{seite}',
  zeileKurz: '+ 行',
  keinPreis: '无价格',
  differenzDir: '⇆ {betrag}，有利于你',
  differenzGegenueber: '⇆ {betrag}，有利于对方',
  ausgeglichen: '⇆ 持平',
  differenzTeil: '{zahl} 项无价格——只计算有价格的。',
  preisHinweis: '价格截至 {datum}，来源：{quelle}。仅供参考，并非硬性规定。',
  keinePreise: '这些卡牌还没有价格。',
  zeileVerwerfen: '取消',
}
