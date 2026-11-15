// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'スクリプトコンソール',
  erklaerung:
    '開いているカードのコンテキストで、MSEスクリプトの式を評価します。環境には、描画時と同じ名前が用意されています。「card」、「set」、「stylesheet」、そしてゲームパッケージの関数です。何も変更しません。',
  ausdruck: '式',
  platzhalter: 'card.name',
  auswerten: '評価',
  ergebnis: '結果',
  art: '型: {art}',
  keineKarte: 'そのためには、カードを開いておく必要があります。',
  zuDenKarten: 'カードへ移動',
  vorbehalte: '環境の構築時の注意事項',
  stelle: '(行 {zeile}、列 {spalte})',
  namen: '使用できる名前',
  namenErklaerung:
    'このカードで式が参照できる名前です。ライブラリと、テンプレートが追加するものが含まれます。クリックすると名前が挿入されます。',
  beispiele: '例',
  beispieleImmer: 'どこでも動作',
  beispieleImmerErklaerung:
    'クリックすると式が挿入されます。これらは、言語そのもの以外は何も必要としません。',
  beispieleVorlage: 'テンプレートが必要',
  beispieleVorlageErklaerung:
    '「card」と「set」は、スタイルパッケージが環境に追加します。このカードにスタイルパッケージがインストールされていない場合は、その旨が結果に表示されます。上の例は、それでも動作します。',
  beispielListe: {
    rechnung: '計算',
    grossSchreiben: '文字列を大文字にする',
    laenge: '文字数を調べる',
    ausschnitt: '文字列の一部を取り出す',
    bedingung: '条件',
    name: 'このカードの名前',
    grossName: 'その名前を大文字にしたもの',
    regeltext: 'ルールテキストの長さ',
    setangabe: 'セットのタイトル',
  },
  vervollstaendigung: '候補',
  uebernehmen: '採用',
  namenFiltern: '名前を絞り込む',
  namenGezeigt:
    '{ganz}件中{gezeigt}件を表示しています。入力して絞り込んでください。',
  namenLeer: '一致する名前はありません。',
  kuerzel: '⌘↩ で評価します。',
  abbrechen: 'キャンセル',
  laeuft: '実行中…',
  dauer: '{ms} ms',
  abgebrochen:
    'キャンセルしました。結果が出る前に実行が終了しました。式はフィールドに残っています。',
  baum: '結果をツリーで表示',
  verlauf: 'このセッションの履歴',
  verlaufErklaerung:
    'このセッション内の直近20件の式です。クリックすると式が読み込まれます。トークンやパスワードに見えるものは伏せ字になり、読み込めません。',
  verlaufLeer: 'まだ何も評価していません。',
  verlaufLoeschen: '履歴を消去',
  ausgangGelungen: '成功',
  ausgangGescheitert: '失敗',
  ausgangAbgebrochen: 'キャンセル',
}
