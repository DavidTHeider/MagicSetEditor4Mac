// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'テンプレート',
  erklaerung:
    'カードにできることと見た目は、テンプレートパッケージの中のテキストファイルに記述されています。フィールド、スクリプト、寸法です。ここで変更できます。変更前の状態は、パッケージの隣にバックアップとして残ります。',
  paket: 'パッケージ',
  keinPaket: 'パッケージがまだ選ばれていません。',
  waehleEines:
    '上でパッケージを選ぶと、そのテキストファイル（フィールド、スクリプト、寸法）を表示して変更できます。',
  keineDateien: 'このパッケージには、編集できるファイルがありません。',
  dateien: 'ファイル',
  datei: 'ファイル',
  inhalt: '内容',
  speichern: '保存',
  gespeichert: '保存しました。',
  ungespeichert: '保存されていません。',
  warnung:
    'このファイルは、このパッケージを使うすべてのセットに適用されます。開いているセットだけではありません。',
  konflikt:
    'ファイルは、開いてからの間に変更されました。ディスク上のバージョンを読み込む（あなたのテキストは失われます）か、あなたのバージョンを残してください。その場合、変更内容には、ディスク上のものに対して何を上書きするかが表示されます。',
  neuLaden: 'ディスク上のバージョンを読み込む',
  behalten: '自分のバージョンを残す',
  vorschau: 'プレビュー',
  vorschauErklaerung:
    'このテンプレートを使う最初のセットの最初のカードを、ここに表示されているテキストで描画します。これによって何も保存されません。',
  vorschauLaeuft: '描画中…',
  vorschauKeineKarte:
    'このテンプレートを使っているセットが開かれていません。「セット」で、これを指定しているセットを開くと、変更の効果をここで確認できます。',

  arbeitsplatz: 'テンプレート工房',
  baumBezeichnung: 'パッケージとファイル',
  inspectorBezeichnung: 'プレビューと診断',
  baumSuche: 'パッケージとファイルを検索',
  baumSuchePlatzhalter: 'パッケージまたはファイル…',
  baumLeer: '検索に一致するパッケージもファイルもありません。',
  baumOhnePakete:
    '編集できるパッケージがまだありません。「パッケージ」で作成するか取得すると、ここに表示されます。',
  nurLesbar: '同梱 — 読み取り専用。独自のコピーは編集できます。',
  baumNurMitgeliefert:
    'ここにあるパッケージはすべて同梱のものです。どれか1つの独自のコピーを作成するか、「パッケージ」で新しく作成してください。どちらも編集でき、上に表示されます。',
  dateienLaden: 'ファイルを読み込み中…',
  inspectorLeer:
    '開いているファイルのプレビュー、診断、出所がここに表示されます。',
  waehleImBaum:
    '左のツリーからファイルを選択してください。フィールド、スクリプト、寸法はテキストとしてそこにあり、ここで変更できます。',
  verwerfenFrage:
    '開いているファイルに、保存されていない変更があります。破棄して{datei}を開きますか？',
  verwerfen: '破棄して開く',
  hierBleiben: 'ここにとどまる',
  testkarte: 'テストカード',
  testkarteVorgabe: '最初のセットの最初のカード',
  testkarteOhneNamen: 'カード{nummer}',
  testkarteErklaerung:
    'テンプレートが最初のカードで壊れることはまれです。サンプルセット「Werkstatt-Testkarten」には、長文、特殊レイアウト、裏面、欠けた画像、特殊文字が含まれています。「セット」で開き、そのスタイルをこのパッケージに切り替えてください。',
  veraltet:
    '古い内容です。ファイルの{zeile}行目に検出事項があるため、最後の有効な状態を表示しています。',
  wirdNeuGezeichnet: '再描画中…',
  diagnose: '診断',
  diagnoseFrei: 'このファイルに検出事項はありません。',
  feldartFehlt:
    'このフィールドには「type:」がありません。MSE2はここでパッケージの読み込みを中止しますが、このアプリではテキストフィールドとして扱います。',
  feldartUnbekannt:
    'MSE2はフィールドタイプ「{0}」を認識せず、ここでパッケージの読み込みを中止します。このアプリではテキストとして扱います。認識されるタイプ：{1}。',
  diagnoseStelle: '{zeile}行目、{spalte}列目',
  zurStelle: 'その位置へ移動',
  herkunft: '名前の出所',
  herkunftKeine:
    'キャレットを名前に置くと、その出所と使用箇所がここに表示されます。',
  herkunftUnbekannt:
    '「{name}」はこのパッケージでは定義されていません。名前はライブラリか、組み込まれたパッケージに由来します。',
  herkunftStelle: '{datei}の{zeile}行目の{art}',
  zurDefinition: '定義へ移動',
  verwendungen: '参照箇所',
  verwendungenKeine: 'この名前を挙げているほかの箇所はありません。',
  verwendungLaden: '参照箇所を検索中…',
  symbolarten: {
    'card field': 'カードフィールド',
    'set field': 'セットフィールド',
    'styling field': 'スタイルオプション',
    'extra card field': '追加フィールド',
    script: 'スクリプト名',
  } as Record<string, string>,
  symbolartSonst: 'ここでの定義',
  aenderungen: '保存前の変更',
  aenderungenZeigen: '変更を表示',
  aenderungenVerbergen: '変更を隠す',
  aenderungenKeine: '読み込んだ状態から変更はありません。',
  aenderungenZusammenfassung: '{neu}行が追加、{weg}行が削除されました。',
  aenderungenAbschnitt: '変更{nummer}',
  zeileWeg: '削除',
  zeileNeu: '追加',
  kopieHinweis: 'パッケージの出所',
  kopieren: '独自のコピーを作成',
  kopieMitgeliefert:
    'このパッケージはアプリケーションに同梱されているため、変更できません。独自のコピーを作成してください。作成後はご自身のパッケージに表示され、元のものはそのまま残ります。',
  kopieFremd:
    'このパッケージは外部から取得したものです。ここでの変更はローカルの変更として扱われ、次回の取得時に新しいバージョンと競合します。変更を残したい場合は、独自のコピーを作成してください。',
  kopieOriginal:
    '元のものは残ります。作業しているのはコピーで、{original}自体は変更されません。',
  dateiaktionen: 'パッケージのファイル',
  dateiNeu: '新規ファイル',
  dateiName: '新しいファイルの名前',
  dateiAnlegen: '作成',
  dateiAbbrechen: 'キャンセル',
  dateiHinzufuegen: '画像またはフォントを追加…',
  dateiUmbenennen: '名前を変更',
  dateiNeuerName: '新しい名前',
  dateiUmbenennenJa: '名前を変更',
  dateiLoeschen: '削除…',
  dateiLoeschenFrage:
    '削除しますか？削除後、ファイルはパッケージの隣のバックアップに残ります。',
  dateiLoeschenJa: '削除する',
  dateiErstSichern: '先に保存または破棄してください。',
  beiwerk: '付属ファイル',
  beiwerkHinweis:
    'テキストではありません。パッケージはこのファイルをそのまま使います。名前の変更と削除は、ここでできます。',
  beiwerkZeile: '付属ファイル（{anzahl}）',
  weitereZeigen: 'さらに表示（{gesamt} 件中 {gezeigt} 件）',
} as const
