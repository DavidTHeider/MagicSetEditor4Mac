// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: '書き出し',
  erklaerung:
    'このアプリケーションからセットを取り出す5つの方法と、それぞれの用途です。',
  wohin: '書き出し先',
  keinZiel: '書き出し先が選ばれていません',
  nichtGemerkt:
    '出力先は使われますが、記憶されませんでした — 次回は事前選択されません：',
  keinZielErklaerung:
    '「書き出し先」でセットをどの形で書き出すかを選ぶと、ここに設定する項目が表示されます。',
  ohneKarten:
    '画像と印刷シートには表示するものがありません。カードは「カード」で作成・貼り付け・読み込みできます。',
  ziele: {
    setdatei: {
      name: 'セットファイル',
      kurz: 'セットそのもの、再び開くためのものです — Magic Set Editor 2 でも読める `.mse-set` ファイルです。',
      lang: 'ショートカットは ⌘S です。デッキ、所持カード、価格、タグ、個々の値の由来は**含まれません** — それらはデータに残ります。MSE2 は保存時に、知らないものを破棄するためです。',
    },
    bilder: {
      kurz: 'カード1枚につき1つの画像ファイルです。フォーラム、表計算、ゲームプログラムなどで利用できます。',
    },
    druckbogen: {
      kurz: 'カードをトンボ付きで用紙に並べます。',
    },
    vorlage: {
      kurz: '画像ではなくテキストです。リスト、表、他のプログラム向けのファイルなどがあります。',
      lang: 'どのテンプレートを使えるかは、取得したパッケージによって決まります。',
    },
    mtgjson: {
      kurz: 'mtgjson.comのカードフォーマットです。これを読み込むツール向けです。',
      lang: 'このアプリケーションで作業を続けるためのものではありません。MTG JSONに収まらないものは、書き出しの後に一覧で示されます。',
    },
  },
  vorlagenEingrenzen: 'テンプレートを絞り込む',
  weitereOptionen: 'その他のオプション（{zahl}）',
  // **„できあがりの確認"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'できあがりの確認',
    rechnet: '計算しています…',
    hinweis: '下のボタンを押すまで、何も書き込まれません。',
    bogen: {
      jeSeite: '1ページあたり{zahl} · {seiten}',
      masse: '{karten} · 用紙 {breite} × {hoehe} mm · {dpi} dpi',
      letzte: '最後のページ: {karten}',
    },
    vorlage: {
      datei: 'ファイル: {name}（{groesse}）',
      karten: '{karten}に対して評価しました。',
      beigaben:
        '隣に置かれるファイル: {zahl}（フォルダー「{ordner}」、{groesse}）',
      mindestens: '少なくとも{groesse}',
      befunde: '評価中に気づいた点:',
    },
  },
}
