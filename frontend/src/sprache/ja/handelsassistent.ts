// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: 'このトレードの進み具合',
  phasen: {
    gegenueber: '相手',
    angebot: '交換内容',
    zusagen: '確保',
    verbuchen: '記帳',
  },
  erledigt: '完了',
  jetzt: '現在',
  kommt: 'これから',
  schritte: '今できること',
  keinSchritt: '今のところ、ここであなたがすることはありません。',
  neu: '新しいトレード：やり取りするカードをまとめて、開始します。',
  abgleich: 'トレードできるもの',
  unterwegs: '途中のもの',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: '相手',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: '履歴',
  verlaufWechsel: 'リビジョン {revision}：{von} → {nach} · {zeit}',
  verlaufNeu: 'リビジョン {revision}：{nach} · {zeit}',
  verlaufLeer: 'まだ変更はありません。',
  verlaufOhneHandel: 'トレードを選ぶと、ここにその経緯が表示されます。',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: '行を追加：{seite}',
  zeileKurz: '+ 行',
  keinPreis: '価格なし',
  differenzDir: '⇆ あなたが {betrag} 有利',
  differenzGegenueber: '⇆ 相手が {betrag} 有利',
  ausgeglichen: '⇆ 均衡',
  differenzTeil:
    '{zahl} 件は価格なし — 計算に含めるのは価格があるものだけです。',
  preisHinweis:
    '{datum} 時点の価格（出典：{quelle}）。目安であり、決まりではありません。',
  keinePreise: 'これらのカードの価格はまだありません。',
  zeileVerwerfen: 'キャンセル',
}
