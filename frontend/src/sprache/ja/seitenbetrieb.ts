// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: '実行時間とリソース使用量',
  zeitraum: '{seit} から {bis} までに収集。',
  aufwand: '収集の所要時間：{mikros} µs',
  speicher: 'メモリ',
  speicherWert:
    'フットプリント {fussabdruck}（うち Go ヒープ {heap}、システムから {system}）',
  speicherOhneFussabdruck:
    'Go ヒープ {heap}、システムから {system} — フットプリントは未計測です',
  web: 'Web プロセス',
  webWert: '合計 {summe}（{teile}）',
  webTeil: '{art} {groesse}',
  webArten: {
    webinhalt: 'Web コンテンツ',
    gpu: 'グラフィックス',
    netz: 'ネットワーク',
  },
  ohneWeb: 'このプロセスは Web 表示を持っていません。',
  webUngemessen: '未計測: {grund}',
  goroutinen: '並行処理',
  rechenzeit: '起動後の CPU 時間',
  datenbank: 'ストアへの接続',
  datenbankWert:
    '{offen} 件オープン、{inBenutzung} 件使用中、待機 {gewartet} 回、合計 {dauer}',
  ohneBestand: 'このプロセスはストアを保持していません。',
  zwischenspeicher: '外部サービスのキャッシュ',
  zwischenspeicherWert:
    '{treffer} 件ヒット、{geholt} 件取得、{veraltet} 件は古いまま提供、{fehlgeschlagen} 件失敗',
  laufzeiten: 'インターフェイスごとの実行時間',
  laufzeitenLeer:
    'ここでは呼び出しを計測していません。計測対象はサービスとツールサーバのみです。',
  ungemessen: '計測していないもの：{liste}。',
  ungemessenNamen: {
    fenster:
      'ウインドウのバインディング（呼び出しが終了する共通の場所がないため）',
    kommandozeile: 'コマンドライン（1つのコマンドの間だけ動作するため）',
    speicher: 'このプロセスのフットプリント（システムが提供しませんでした）',
  },
  spalteArt: '種類',
  spalteAnzahl: '呼び出し数',
  spalteFehler: 'エラー率',
  spalteMittel: '平均',
  spalteVerteilung: '分布',
  arten: { lesend: '読み取り', schreibend: '書き込み' },
  eimer: '{grenze} ms まで',
  verteilung: '{eimer}：{anzahl} 回の呼び出し',
  eimerOffen: 'それ以上',
  spurAus:
    'トレースはオフです。「設定」＞「診断」でオンにできます。次回の起動から有効になります。',
  spurAn:
    'トレースはオン：{gesehen} 件の呼び出しのうち {aufgenommen} 件を記録（サンプリング率 {abtastung} %）。最大 {grenze} 件のエントリを {minuten} 分間保持 — {verdraengt} 件は押し出され、{abgelaufen} 件は期限切れです。',
}

export const leistungsprofil = {
  ueberschrift: 'パフォーマンスプロファイル',
  erklaerung:
    '選択した時間にわたるこのウインドウの CPU プロファイルを、「go tool pprof」の形式で記録します。ストアの横に保存され、残す個数は「設定」＞「診断」で指定します。',
  sekunden: '所要時間（秒）',
  grenze: '1〜60 の整数。',
  aufzeichnen: 'プロファイルを記録',
  laeuft: '記録中 — ウインドウは引き続き操作できます。',
  fertig: '書き込みました：{datei}（{groesse}）',
}

export const zwischenspeicher = {
  ueberschrift: 'キャッシュ',
  erklaerung:
    'このアプリケーションが、同じものを二度取得したり計算したりしないように保持しているデータです。ディスク上とメモリ上にあります。いずれもあなたのデータではありません。空にしても、次に使うときに自動的に再び満たされます。',
  spalteSpeicher: 'キャッシュ',
  spalteOrt: '場所',
  spalteQuote: 'ヒット率',
  spalteEintraege: 'エントリ',
  spalteGroesse: 'サイズ',
  spalteGrenze: '上限',
  keineQuote: '照会はまだありません',
  keineGrenze: '置き換えられる1つのファイル',
  namen: {
    'scryfall.antworten': 'Scryfall の応答',
    'scryfall.bilder': 'Scryfall のカード画像',
    'scryfall.sammeldaten': 'Scryfall の一括データファイル',
    'render.ebenen': 'サンプリングしたレイヤー',
    'pakete.bilder': 'テンプレートパッケージの画像',
    'skript.baeume': '構文解析済みのスクリプト',
    'vorschau.bilder': 'カードのプレビュー',
    'tisch.texturen': 'テーブル上のカード画像',
  },
  orte: { platte: 'ディスク', arbeitsspeicher: 'メモリ' },
  leerenHinweis:
    '空にすると、すべて破棄されます。ネットワークに接続していない間は、ネットワークが復旧するまで Scryfall の応答と画像は戻りません。',
  leeren: 'すべて空にする',
  geleert: '空にしました。次に使うときに再び満たされます。',
}
