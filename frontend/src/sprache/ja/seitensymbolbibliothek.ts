// Die Texte der **Symbolbibliothek** — DOM-02
// (#1697).

export const symbolbibliothek = {
  ueberschrift: 'シンボルライブラリ',
  erklaerung:
    'このプロファイルのシンボルで、どのセットにも配置できます。配置されるのはコピーです。あとでライブラリからシンボルを削除しても、セットは保存時にそのコピーを保持します。',

  name: 'セットシンボルの名前',
  aufnehmen: 'セットシンボルを追加',
  einlesen: 'シンボルファイルを読み込む…',
  suche: 'シンボルを検索',

  leer: 'まだシンボルがありません。このセットのシンボルを追加するか、.mse-symbol ファイルを読み込んでください。そのあと、他のどのセットにも配置できます。',
  keinTreffer:
    '検索とタグの絞り込みに一致するシンボルがありません。どちらかを変更してください。',
  liste: 'ライブラリのシンボル',
  vorschau: '{name}のプレビュー',
  verwendet: '使用中のセット: {sets}',
  unverwendet: '使用しているセットはありません。',

  einsetzen: 'このセットに配置',
  eingesetzt: '{name}がこのセットのシンボルになりました。',
  duplizieren: '複製',
  kopieName: '{name} (コピー)',
  umbenennen: '名前を変更',
  neuerName: '{name}の新しい名前',
  namenUebernehmen: '名前を適用',
  umbenennenAbbrechen: 'キャンセル',
  ausgeben: '書き出す…',
  ausgegeben: '{pfad} に書き出しました',
  eingelesen: '{name}がライブラリに追加されました。',
  aufgenommen: '{name}がライブラリに追加されました。',

  loeschen: '削除',
  loeschenFrage: '{name}をライブラリから削除しますか？',
  loeschenFolgen:
    '次のセットがこれを使用していますが、それぞれのコピーは残ります: {sets}。失われるのはライブラリの項目だけです。同じシンボルを使用しているという記録も、一緒に失われます。',
  loeschenOhneFolgen: '使用しているセットはありません。削除は取り消せません。',
  loeschenJa: '削除',
  loeschenNein: 'キャンセル',

  schriften: 'シンボルフォント',
  schriftenErklaerung:
    'シンボルフォントは、このコンピュータのパッケージです。スタイルがテキストフィールドに配置する、マナ・シンボルなどの文字が含まれます。フォントは「パッケージ」でインストールし、ここでは表示のみです。',
  keineSchriften:
    'シンボルフォントがインストールされていません。「パッケージ」で取得できます。',
  schriftMitFassung: '{name} · バージョン {fassung}',
  schriftenListe: 'インストール済みのシンボルフォント',
  schriftenZeigen: 'インストール済みのシンボルフォント（{zahl}）',

  bearbeiten: '描画を編集',
  bearbeitungsueberschrift: 'ライブラリの{name}',
  bearbeitungserklaerung:
    'ライブラリのシンボルを編集しています。保存すると新しいバージョンが保存され、現在のバージョンは1つ前のバージョンとして残ります。各セットは、適用するまで自分のコピーを保持します。',
  zumSetsymbol: 'セットシンボルに戻る',
  inBibliothekSichern: 'ライブラリに保存',
  inBibliothekGesichert: '{name}の新しいバージョンを保存しました。',
  veraltet: '次のセットは、まだ1つ前のバージョンを使用しています: {sets}。',
  keineVeraltet: '1つ前のバージョンを使用しているセットはありません。',
  nachziehen: 'これらのセットに適用',
  nachgezogen: '{sets}に適用しました。これらのカードは再描画されます。',
  zuruecknehmen: '1つ前のバージョンに戻す',
  zurueckgeholt: '{name}を1つ前のバージョンに戻しました。',

  fassungen: 'このセットのバリエーションで表示',
  fassungenZu: 'バリエーションを閉じる',
  fassungsliste: '{name}のバリエーション',
  keineFassungen:
    'このセットのスタイルには、セットシンボルのバリエーションが指定されていません。バリエーションのあるスタイルを選ぶと、その中でシンボルを確認できます。',
  fassungsbild: '{fassung}の{name}、{groesse}ポイント',
}
