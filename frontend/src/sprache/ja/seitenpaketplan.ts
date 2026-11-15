// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: 'パッケージのワークスペース',
  navigator: '検索とフィルタ',
  inspector: 'パッケージ',
  zentrum: 'パッケージ',

  herkunft: '出所',
  herkunftAlle: 'すべての出所',
  herkunftsarten: {
    mitgeliefert: '同梱',
    bezogen: '取得済み',
    uebernommen: '引き継ぎ',
    eigen: '自作',
  },

  spalteStand: '状態',
  standEingebunden: '組み込み済み',
  standGeaendert: 'ローカルで変更',

  inspectorLeer:
    'リストからパッケージを選択してください。出所、バージョン、依存関係、変更した場合の影響がここに表示されます。',
  kopf: '{art} · バージョン{version}',
  kopfOhneFassung: '{art} · バージョンなし',
  eingebunden: '組み込み済み',
  keinSet: 'このプロファイルのセットで、これを使っているものはありません。',
  einSet: '1つのセットが使っています。',
  vieleSets: '{anzahl}個のセットが使っています。',
  setsUnbekannt:
    'どのセットが必要としているかは、どのセットのヘッダにも記録されていません。',
  geaendert:
    '組み込み後にローカルで変更されています。置き換える前に確認が入り、変更内容は前のバージョンとして保持されます。',
  woher: '出所',
  woherMitgeliefert: 'アプリケーションに同梱されています。',
  woherBezogen: '{herkunft}から取得しました。',
  woherUebernommen: '{herkunft}から引き継ぎました。',
  woherEigen: 'ここで作成したか、パッケージフォルダに手動で置いたものです。',
  pruefsumme: 'チェックサム',
  lizenz: 'ライセンス',
  lizenzUnbekannt:
    '記載なし — MSEパッケージにはライセンスのフィールドがなく、内容は権利者のものです',
  ort: '場所',
  haengtAb: '依存先',
  gebrauchtVon: '依存元',
  nichts: 'なし',

  einbinden: 'パッケージを組み込む…',
  beziehen: 'パッケージを取得…',
  beziehenTitel: 'パッケージを取得',
  loeschen: '削除…',
  zurueck: '前のバージョンに戻す…',
  weitergeben: 'インストーラとして共有…',
  weitergegeben: 'インストーラとして保存しました：{datei}',

  titel: {
    loeschen: '{paket}を削除',
    einbinden: '{paket}を組み込む',
    ersetzen: '{paket}を置き換える',
    zurueck: '{paket}を前のバージョンに戻す',
  },
  fassung: 'バージョン{von} → {zu}',
  ohneFassung: 'バージョンなし',
  sets: '影響を受けるセット',
  keineSets: 'このプロファイルのセットで、影響を受けるものはありません。',
  konflikte: '実行後に満たされなくなるもの',
  keineKonflikte: '壊れる依存関係はありません。',
  planGeaendert: 'このパッケージは、組み込み後にローカルで変更されています。',
  ueberschreiben:
    'ローカルの変更を置き換える — 前のバージョンとして保持されます',
  nichtMoeglich: {
    loeschen:
      'このパッケージは管理対象のフォルダにありません。ここで削除できるのは、このアプリケーションが組み込んだものだけです。',
    einbinden: 'このパッケージは、ここでは組み込めません。',
    ersetzen: 'このパッケージは、ここでは置き換えられません。',
    zurueck:
      '前のバージョンはありません。このパッケージの上に新しいバージョンを組み込むと、作成されます。',
  },
  bestaetigen: {
    loeschen: '削除',
    einbinden: '組み込む',
    ersetzen: '置き換える',
    zurueck: '元に戻す',
  },
  verwerfen: '変更しない',
  bedarf: {
    ueberschrift: 'このセットに不足しているパッケージ',
    erklaerung:
      'これらがないと、プレビューは代替の描画になるか、パッケージが提供するものが省かれます。フォルダからパッケージを組み込むか、パッケージを検索して取得またはMSE2から引き継いでください。',
    suchen: 'パッケージを検索',
  },
  quelle: '{quelle}から',
  keineRechte:
    'パッケージにはテンプレートとMSE-Scriptが含まれます。スクリプトは、ファイル、ネットワーク、プログラムにアクセスできず、パッケージが必要とするものが一緒に取得されることもありません。それは上に競合として表示されます。',
}
