// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'カード', viele: 'カード' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'ページ', viele: 'ページ' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'セット', viele: 'セット' },
    decks: { eine: 'デッキ', viele: 'デッキ' },
    pakete: { eine: 'パッケージ', viele: 'パッケージ' },
    eigenePakete: { eine: '自作パッケージ', viele: '自作パッケージ' },
    stile: { eine: 'スタイル', viele: 'スタイル' },
    positionen: { eine: '項目', viele: '項目' },
    exemplare: { eine: '枚', viele: '枚' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: '版', viele: '版' },
  },
  bestandGesichert: 'データのバックアップ先：',
  zahlenbereich: '{kleinstens}から{groesstens}までの数値を入力してください。',
  zahlenAb: '{kleinstens}以上の数値を入力してください。',
  ganzeZahlBereich:
    '{kleinstens}から{groesstens}までの整数を入力してください。',
  ganzeZahlAb: '{kleinstens}以上の整数を入力してください。',
  nurZiffern:
    '数字のみ（ピリオドやカンマは使えません）。金額は最小単位で入力してください。',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'ファイル{pfad}を開けませんでした。{grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: '言語',
    de: 'Deutsch',
    en: 'English',
    es: 'Español',
    fr: 'Français',
    it: 'Italiano',
    ja: '日本語',
    ru: 'Русский',
    zh: '简体中文',
  },
  // **Die fünf Arbeitsbereiche, der Einstieg und die Dienstzone** — UX-02
  // (#1674).
  //
  // **`bereiche` und `dienste` sind die zwei Überschriften der Leiste**, nicht
  // Bereiche selbst — sie sagen, was darunter steht.
  bereiche: {
    // **„Home" und nicht „Start"** — so heißt die Seite in dieser Sprache
    // (`en/seitenstart.ts`), und der Bereich trägt ihren Namen.
    start: 'ホーム',
    gestalten: 'デザイン',
    sammeln: 'コレクション',
    spielen: 'プレイ',
    handeln: 'トレード',
    ressourcen: 'リソース',
    programm: 'アプリ',

    ueberschrift: '作業エリア',
  },

  navigation: {
    bezeichnung: 'メインナビゲーション',
    sprungZumInhalt: 'コンテンツへスキップ',
    seitenleisteEinblenden: 'サイドバーを表示',
    seitenleisteAusblenden: 'サイドバーを隠す',
    seitenleisteKuerzel: '⌥⌘S',
    weg: '現在の場所',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'この領域を描画できませんでした',
    bereichSatz:
      '参照番号 {kennung}。ほかのデータベースには影響ありません。このエラーはこの参照番号でログに記録されています。',
    neuLaden: '領域を再読み込み',
    diagnoseOeffnen: '診断を開く',
    wurzelTitel: 'ウインドウを描画できませんでした',
    wurzelSatz:
      '参照番号 {kennung}。データベースには影響ありません。このエラーはこの参照番号でログに記録されています。',
    fensterNeuLaden: 'ウインドウを再読み込み',
  },
  meldungen: {
    methodeFehlt: 'メソッド{name}は割り当てられていません。',
    vorschauLeer: 'プレビューは画像もエラーもなく返されました。',
    kontoLeer: 'アカウントはユーザーもエラーもなく返されました。',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'ウインドウのメッセージ',
    setNichtGesichert: 'セット「{name}」は保存されませんでした。{grund}',
    setNichtGesichertOhneName: 'セットは保存されませんでした。{grund}',
    bestandNichtGesichert: 'データはバックアップされませんでした。{grund}',
    verlaufGescheitert: '取り消しまたはやり直しができませんでした。{grund}',
    darstellungGescheitert: '外観をリセットできませんでした。{grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'このバージョンでお使いのデータを開けます。',
    unvertraeglich: 'このバージョンではお使いのデータを開けません。',
    ungeklaert:
      'このバージョンでお使いのデータを開けるかどうかは不明です。この確認より前のバージョンのためです。切り替える前にデータのバックアップを取り、問題があれば復元してください。',
    bereit: '切り替えの準備ができました。バックアップを取り、読み戻しました。',
    vorbereiten: '切り替えを準備',
    hindernisse: {
      keinManifest: 'このリリースには自身に関する情報がありません。',
      plattform: '別のコンピュータ向けにビルドされています。',
      wanderungFehlt:
        'お使いのデータには、このバージョンが把握していない変更が含まれています。',
      summeAnders: 'ある変更の内容が、このバージョンでは異なっています。',
      vorDerGrundlinie:
        'このデータはこのバージョンの基準より古いため、先に中間のバージョンで開いてください。',
      platz: '空き容量が足りません。',
      vorgaengeLaufen: 'まだ実行中の処理があります。',
      keineSicherung: 'バックアップが作成されていません。',
      sicherungUnlesbar: 'バックアップを読み戻せませんでした。',
      halter:
        '別のプログラムがデータを開いたままです。切り替える前に終了してください。',
    },
  },
  fassungsband: {
    bereich: '新しいバージョン',
    neu: 'バージョン{fassung}が公開されました。',
    seite: '変更内容',
    laden: 'ダウンロード',
    laeuft: 'ダウンロード中…',
    fortschritt: '{ganz}中{gelesen}をダウンロード済み',
    fortschrittOhneLaenge: '{gelesen}をダウンロード済み',
    geladen:
      'バージョン{fassung}は{pfad}にあります。インストールするには、Finderで開いて、アプリケーションを「アプリケーション」フォルダへドラッグしてください。',
    weglegen: 'このメッセージを閉じる',
  },
  wiederanlauf: {
    bereich: '前回のセッション',
    satz: '{zeit}に開始した前回のセッションは予期せず終了しました。',
    satzOhneZeit: '前回のセッションは予期せず終了しました。',
    mitSpur: 'クラッシュの記録があります。診断バンドルを作成すると含まれます。',
    ohneSpur:
      'クラッシュの記録はありません。アプリが外部から終了された可能性があります。',
    ausblenden: '非表示',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'セットの数値',
      befund: '指摘あり',
      nichtGeprueft: '未チェック',
      nichtGelesen: '未読込',
      geprueft: 'チェック済み {wann}',
    },
    laschen: 'セットの各エリア',

    gruppen: {
      bauen: 'カードを作る',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'プロパティ',
        titel: 'セットのプロパティ',
        laschen: 'プロパティの各エリア',
        hinweis: 'このセットのすべてのカードに適用される設定です。',
      },
      werkzeuge: {
        knopf: 'ツール',
        titel: 'ツール',
        laschen: 'セットのツール',
        hinweis: 'このセットの統計、ランダムパック、コンソール。',
      },
      schliessen: '閉じる',
    },

    keinSet: '開いているセットはありません',
    keinSetErklaerung:
      '「セット」からセットを選ぶか、新規作成または読み込みをしてください。カード、統計、印刷シートがここに表示されます。',
    zuDenSets: 'セットへ移動',
    zuDenPaketen: 'パッケージへ移動',
    zuDenKarten: 'カードへ移動',
    ohneKartenTitel: 'このセットにはまだカードがありません',
    ohneKarten:
      'このタブはセットのカードを使います。カードは「カード」で作成・貼り付け・読み込みできます。',
    ohneNamen: '開いているセット',
  },
  palette: {
    bezeichnung: 'コマンドパレット',
    eingabe: '移動先は？',
    keineTreffer: '見つかりませんでした。別の言葉をお試しください。',
    oeffnen: 'コマンドパレットを開く',
    seiten: 'ページ',
    befehle: 'コマンド',
    befehlEingabe: '移動先またはコマンドを入力',
    tun: '操作',
    finden: '検索',
    karteErstellen: 'カードを作成',
    setWaehlen: 'セットを作成または開く',
    deckSpielen: '{name}ですぐに対戦',
    tauschenMit: '{name}とカードをトレード',
    sicherungErstellen: 'バックアップを作成',
    schnittstellenEinrichten:
      'インターフェースを設定（コマンドライン、MCP、REST）',
    woKarte: 'デザイン · {set}',
    woSets: 'デザイン · セット',
    woPartie: 'プレイ · 対戦を開始',
    woHandel: 'トレード · 相手',
    woEinstellungen: '設定 · {gruppe}',
    alsKarte: 'カード · {set}',
    alsSet: 'セット · {karten}',
    alsDeck: 'デッキ · {spiel}',
    alsGegenueber: '相手 · 確認済み',
    fuss: '↑↓ 選択 · ↵ 実行 · esc 閉じる',
  },
} as const
