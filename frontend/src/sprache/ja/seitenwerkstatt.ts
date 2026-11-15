// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'カードのワークスペース',
  inspectorBezeichnung: 'フィールドとチェック',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'カードを絞り込み',
    platzhalter: '絞り込み — 例：type:creature -goblin',
    leeren: '絞り込みをクリア',
    hilfe:
      '空白は各部分を区切り、「…」は語をまとめ、field:valueはフィールドに対応させ、先頭の-は否定です。大文字と小文字は区別しません。',
    treffer: '{gesamt}枚中{gezeigt}枚',
  },
  zoom: {
    bezeichnung: 'プレビューのサイズ',
    anpassen: '合わせる',
    druck: '印刷サイズ',
    ohneDruck: 'テンプレートに解像度の指定がないため、印刷サイズはありません。',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: '画像をドラッグ',
    hinweis:
      '画像を Finder にドラッグすると、カードを PNG として保存できます。「{weg}」でも同じことができます。',
  },
  overlays: {
    bezeichnung: 'オーバーレイ',
    bildgrenzen: '画像の境界',
    beschnitt: '裁ち落とし',
    diagnose: '診断',
    beschnittHinweis: '線の内側にあるものは、3 mmの裁断でも残ります。',
  },
  probleme: {
    ueberschrift: '問題',
    keine: 'このカードに問題はありません。',
    zumFeld: 'フィールド{feld}へ移動',
    ohneFeld: 'フィールドなし',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: '前のカード',
    naechste: '次のカード',
    stelle: '{stelle} / {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
