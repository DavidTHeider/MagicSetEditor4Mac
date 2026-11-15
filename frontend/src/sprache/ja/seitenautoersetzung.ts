// Die Texte der **Autoersetzung** —
// #1341,
// Vermerk [0060](../../../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
//
// **Eine eigene Datei**, weil `seitenkarten.ts` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Katalog bleibt **einer**: `seitenkarten.ts` setzt sie ein,
// und der Vertragstest gegen die Go-Seite sieht dieselbe Zahl von Katalogen.

export const autoersetzung = {
  ueberschrift: '自動置換',
  erklaerung:
    '入力中に置き換えます。左側の文字列が右側の文字列になります。ゲームパッケージのルールは固定で、変更できるのは、挿入する内容と、ルールを適用するかどうかです。独自のルールは、このゲームのすべてのセットで有効です。',
  schalter: '自動置換を適用',
  muster: '検索',
  ersatz: '置換後',
  aktiv: '適用',
  ganzesWort: '単語全体',
  anlegen: 'ルールを追加',
  loeschen: '削除',
  zuruecksetzen: '独自ルールを破棄',
  zuruecksetzenFrage: '独自ルールをすべて破棄しますか？',
  zuruecksetzenErklaerung:
    'その後は、ゲームに付属するルールだけが適用されます。独自のルールは元に戻りません。',
  zuruecksetzenJa: '独自ルールを破棄',
  ausDemPaket: 'ゲームパッケージから',
}
