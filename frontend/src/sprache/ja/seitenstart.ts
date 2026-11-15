// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: '概要',
    mitProblemen: '問題あり {anzahl}',
    geprueftVon: '{gesamt}件中{geprueft}件を確認済み',
    mitBefund: '指摘あり {anzahl}',
    nichtGeprueft: '未確認あり',
    problemeGeprueft: '問題を確認 {wann} · {gesamt}枚中{geprueft}枚',
    problemeOffen: '問題: {gesamt}セット中{offen}セットが未確認',
    pruefen: '今すぐ確認',
    setsMitProblemen: '問題のある{sets}',
    kartenMitBefund: '{sets}内の指摘のある{karten}',
    aeltesterPreis: '最も古い価格 {wann}',
    ohnePreis: '価格なし {anzahl}',
    mitFehlenden: 'カード不足 {anzahl}',
    decksMitFehlenden: 'カードが不足している{decks}',
    alter: {
      gerade: 'たった今',
      minuten: '{minuten}分前',
      stunden: '{stunden}時間前',
      tage: '{tage}日前',
      tag: '1日前',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: '確認が必要な項目',
    nichts: '問題なし — {sets}、{decks}、{pakete}を確認しました。',
    durchgesehen: '{sets}、{decks}、{pakete}を確認しました。',
    arten: {
      setVorbehalte: '{anzahl}件が未解釈',
      deckFehlend: '{anzahl}件がコレクションにありません',
      paketBefund: 'スクリプトに{anzahl}件',
    },
    paketeMitBefund: 'スクリプトに問題がある{pakete}',
    gruendeZeigen: '理由を表示',
    gruendeVerbergen: '理由を隠す',
    setBefunde: '指摘のある{karten}',
    befunde: '確認できなかった項目:',
    alle: 'すべての{anzahl}件を表示',
    weniger: '表示を減らす',
    gruende: {
      vorbehalt: '{0}、{1}行目',
    },
    amOrt: '{ort}: {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'ライブラリの空き容量は残り {frei} です',
        sicherungAlt: '最後のバックアップは {tage} 日前です',
        dienstBeendet: 'ログイン時のサービスがコード {ausgang} で終了しました',
        webprozessNeu: 'Web 表示が {anzahl} 回再読み込みされました',
      },
      rat: {
        platzKnapp:
          '空きが約 2 GB または 5 % を下回ると、バックアップの余裕がなくなります。「データとバックアップ」で整理してください。',
        sicherungAlt:
          '14 日より古くなっています。「データとバックアップ」でライブラリをバックアップしてください。',
        dienstBeendet:
          'エラーの後は launchd が再起動します。終了した理由は診断で確認してください。',
        webprozessNeu:
          'WebKit がプロセスを終了しました。開いていたフォームの未保存の内容は失われている可能性があります。診断で確認してください。',
      },
    },
  },
  auftraege: {
    titel: 'バックグラウンドタスク',
    staende: {
      laeuft: '実行中',
      fertig: '完了',
      gescheitert: '失敗',
      abgebrochen: 'キャンセル',
    },
    arten: {
      preise: 'コレクションの価格を取得',
      sammeldaten: 'Scryfallの一括データを読み込む',
      fassung: '新しいバージョンを取得',
      bogen: '印刷シートを作成',
      set: 'セットを読み込む',
      mtgjson: 'MTG JSONを読み込む',
      vorlage: 'テンプレートで書き出す',
      installer: 'インストーラを読み込む',
      weitergabe: 'インストーラとして共有',
      probleme: 'セットの問題を確認',
    },
    abbrechen: '{art}をキャンセル',
    dauer: '{sekunden} 秒',
    laufend: { eine: '件のタスクを実行中', viele: '件のタスクを実行中' },
  },
  weiterarbeiten: {
    titel: '作業を再開',
    weiterAn: '{art}「{name}」の作業を続ける',
    zuletztWann: '{zeit}に最後に{vorgang}',
    angeheftet: '固定済み',
    angeheftetLeer:
      '固定したものはありません。行を固定すると、ここに表示され続けます。',
    leer: 'ここには、最近の作業が表示されます。今はまだ空です。最初のステップは、セットまたはテンプレートパッケージです。',
    handlungSets: 'セットを作成または開く',
    handlungPakete: 'パッケージを取得',
    anheften: '固定',
    loesen: '固定を解除',
    anheftenVon: '「{name}」を固定',
    loesenVon: '「{name}」の固定を解除',
    arten: { set: 'セット', deck: 'デッキ', paket: 'パッケージ' },
    vorgaenge: {
      geoeffnet: '開いた',
      eingefuehrt: 'インポートした',
      ausgegeben: '書き出した',
    },
  },
  titel: 'ホーム',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Magic とその他のゲームのカードを、Magic Set Editor 2 が使うテンプレートで、デザイン、確認、印刷できます。',
  ersteSchritte: '最初のステップ',
  einfuehrung: {
    willkommen:
      '3つのステップで、最初のカードが完成します。このウインドウは、設定からいつでも再表示できます。',
    wiederzeigen: '起動時に再び表示',
    schliessen: 'はじめる',
    bezeichnung: 'はじめに',
  },
  schritte: {
    setOeffnen: {
      titel: 'セットを開く',
      erklaerung:
        '既存の .mse-set ファイルを開くか、MTG JSONからセットをインポートします。',
      handlung: 'セットへ移動',
    },
    kartenBearbeiten: {
      titel: 'カードを編集',
      erklaerung:
        '開いているセットのカードを表示・変更します。プレビューには、変更がすぐに反映されます。',
      handlung: 'カードへ移動',
      ersteinSet: 'そのためには、先にセットを開く必要があります。',
      stattdessen: '先にセットを開く',
    },
    paketeHolen: {
      titel: 'テンプレートを取得',
      erklaerung:
        'テンプレートパッケージが8つ、すでに用意されています。さらに多くのゲームやスタイルが必要な場合は、ここで取得できます。',
      handlung: 'パッケージへ移動',
    },
  },
}
