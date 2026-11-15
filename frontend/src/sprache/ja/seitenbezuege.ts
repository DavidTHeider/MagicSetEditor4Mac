// パッケージページの **取得済みの状態** のテキスト — G04 (#2251)。
// ドイツ語の原文は `de/seitenbezuege.ts`。

export const bezuege = {
  ueberschrift: '取得済みの状態',
  erklaerung:
    '取得したパッケージはひとつのコミットに留まり、ひとりでに古くなることはありません。新しいものがあるか取得元に問い合わせてください。そのときだけネットワークを使います。',
  fragen: '取得元に問い合わせる',
  leer: 'まだパッケージを取得していません。「パッケージを取得 …」で取得すると、取得元がどこまで進んだかがここに表示されます。',
  bezogen: '取得: {stand}',
  geprueft: '検証済み: {stand}',
  ohnePruefung: '検証済みの取得元の一覧にありません',
  pakete: '{zahl} 個のパッケージ',
  paketEins: '1 個のパッケージ',
  keinPaket: 'パッケージなし',
  lage: {
    ungefragt: 'まだ問い合わせていません。',
    aktuell: '取得元の最新の状態です。',
    neuerEins: '新しい状態 {spitze}（{datum}）— 1 コミット先行。',
    neuer: '新しい状態 {spitze}（{datum}）— {vorsprung} コミット先行。',
    abweichend:
      '取得元は先行していません — リセットまたは分岐（先端 {spitze}）。',
    unbekannt: '取得元に問い合わせできませんでした: {grund}',
  },
  ansehenBezeichnung: '更新を確認 … ({name})',
  zurueckBezeichnung: '保管した状態に戻す ({name}, {stand})',
  ansehen: '更新を確認 …',
  aufgehoben: '保管中: {stand}',
  zurueck: '保管した状態に戻す',
  vorschauTitel: '{name} の更新',
  vonZu: '{von} → {zu}: {vorsprung} 先行、{rueckstand} 遅れ',
  paketzeile: '{paket}: 新規 {neu}、変更 {geaendert}、削除 {entfernt}',
  keinePakete: '変更されるテンプレートパッケージはありません。',
  uebrige: 'ほかに {zahl} 個のファイルはどのパッケージにも属しません。',
  sets: '影響を受けるパッケージを指定しているセット: {sets}',
  keineSets: '影響を受けるパッケージを指定しているセットはありません。',
  aufheben: '現在の状態は保管され、元に戻せます。',
  vorigeWeicht: 'すでに保管されている状態 {stand} は置き換えられます。',
  gruende: {
    setVerliert: 'セットが指定しているパッケージからファイルがなくなります。',
    nichtVoraus: '取得元は先行していません — 更新は後退または横道になります。',
    gekuerzt:
      '取得元は 300 を超える変更ファイルを示しています。上の一覧は省略されています。',
  },
  bestaetigen: '理由を確認しました。それでも更新します。',
  ausfuehren: '{zu} に更新',
  verwerfen: '更新しない',
}
