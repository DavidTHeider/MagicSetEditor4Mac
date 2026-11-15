import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert:
      'この欄はその間に別の場所で「{wert}」に変更されました。下書きはまだ保存されていません。',
    behalten: '下書きを保存',
    nehmen: '新しい値を使う',
  },
  listenansicht: {
    suche: '検索',
    suchePlatzhalter: '名前またはゲーム',
    sortierung: '並べ替え',
    ansicht: '表示',
    liste: 'リスト',
    kacheln: 'タイル',
    zahl: '{gesamt}件中{gezeigt}件',
    nichtsPasst: '「{suche}」に一致するものはありません。',
    sucheLeeren: '検索をクリア',
    filterZuruecksetzen: '検索とフィルターをリセット',
    weitere: 'さらに{anzahl}件を表示',
    wieImBestand: '保存順',
    zuletztBenutzt: '最近使った順',
    nachName: '名前順',
    nachKarten: 'カード数順',
  },
  speicherstand: {
    schreibt: '保存中…',
    geschrieben: 'すべての変更が保存されています。',
    fehler:
      '「{feld}」は保存されていません。理由は該当のフィールドに表示されています。',
  },
  gegenueberTitel: '相手の追加または確認',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: '相手を検索 …',
  feldhilfe: '{feld}の説明',
  mehrZu: '「{satz}」の詳細',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'プロファイル{quelle}から引き継ぎました。',
  herkunftAusEinfuhr: '{quelle}から読み込みました。',

  pflichtfeldLeer: '{feld}に値を入力してください。',

  waehrung: {
    beispiel: '例：EUR',
    keineGestalt:
      '「{wert}」は通貨コードではありません。通貨コードはEURのように3文字です。',
    grossSchreiben: '「{wert}」— 通貨コードは大文字で書きます：{gross}。',
  },

  drucksprachen: {
    beispiel: '例：en',
    namen: {
      en: '英語',
      es: 'スペイン語',
      fr: 'フランス語',
      de: 'ドイツ語',
      it: 'イタリア語',
      pt: 'ポルトガル語',
      ja: '日本語',
      ko: '韓国語',
      ru: 'ロシア語',
      zhs: '簡体字中国語',
      zht: '繁体字中国語',
      he: 'ヘブライ語',
      la: 'ラテン語',
      grc: '古代ギリシャ語',
      ar: 'アラビア語',
      sa: 'サンスクリット語',
      ph: 'ファイレクシア語',
      qya: 'クウェンヤ語',
      dw: 'ドワーフ語',
      tlh: 'クリンゴン語',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'パスワードをそのまま表示',
    zeigenKurz: '表示',
    verbergen: 'パスワードを再び隠す',
    verbergenKurz: '隠す',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'コード',
    abschicken: '確定',
  },

  aktionen: {
    abbrechen: 'キャンセル',
    oeffnen: '開く',
    schliessen: '閉じる',
    speichern: '保存',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— スタイルを選択 —',
  },
  druckwahl: {
    drucke: '版',
    hinweis:
      '相手には、ここで選んだ版だけが、所持している枚数とともに見えます。',
    suchen: '版を検索',
    gewaehlt: '選択中: {anzahl}',
    nurSuche:
      'コレクションには{anzahl}あります。名前で検索して選んでください。',
    keiner: '「{suche}」に一致する版はありません。',
    genauer:
      '表示しきれないほど多くの版が一致しています。検索を絞り込んでください。',
    leer: 'まだ共有できる版がコレクションにありません。先にカードを追加してください。',
    fehlt: '版を1つ以上選んでください。版を選ばないと何も共有されません。',
    behalten: '版ごとに手元に残す枚数',
    behaltenHinweis:
      '相手に見えるのは、その枚数を超えた分です。1枚残して3枚持っていれば2枚です。',
    detailBehalten: '{behalten}枚を残す',
  },
  suchauswahl: {
    keiner: '「{suche}」に一致するものはありません。',
    treffer: '一致: {anzahl} 件',
    begrenzt: '{anzahl} 件中 {gezeigt} 件を表示 — 検索を絞り込んでください …',
  },
  zeitzonenwahl: {
    system: 'システムに合わせる',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: '作成', heisst: '新しいものを作成します。' },
      bearbeiten: { kurz: '編集', heisst: 'その場で変更します。' },
      sichern: { kurz: '保存', heisst: '現在の状態を保持します。' },
      zuruecknehmen: { kurz: '取り消す', heisst: '直前の操作を取り消します。' },
      duplizieren: {
        kurz: '複製',
        heisst: '隣にもう1つ作ります。',
      },
      kopieren: { kurz: 'コピー', heisst: 'クリップボードに入れます。' },
      entfernen: {
        kurz: '取り除く',
        heisst: 'この表示から外します。ほかの場所には残ります。',
      },
      loeschen: { kurz: '削除', heisst: 'すべての場所から完全に削除します。' },
    },
  },
  nurlesen: {
    hinweis:
      'このプロファイルは閲覧者として表示しています。すべて見ることはできますが、変更はできません。作業するには自分のプロファイルに切り替えるか、このプロファイルの管理者に「共同作業者」の役割を依頼してください。',
    knopf: '閲覧のみ — このプロファイルでは何も変更できません。',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: '作業コンテキスト',
    profilVon: '{eigner}のプロファイル',
    rollen: {
      leser: '閲覧者',
      mitarbeiter: '共同作業者',
      verwalter: '管理者',
    },
    nurLesen: '読み取り専用',
    zuDenNutzern: 'プロファイルを切り替える…',
    zeile: '{name}・{profil}',
  },

  zustaende: {
    laedt: '読み込み中…',
    leer: '何もありません',
    fehler: '読み込めませんでした。',
    fehlerMitGrund: '読み込めませんでした。{grund}',
    wiederholen: '再読み込み',
    neuLaden: '再読み込み',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'テキスト内を検索',
    schliessen: '検索を閉じる',
    suchen: '検索',
    ersetzen: '置換後の文字列',
    treffer: '{ganz}件中{bei}件目',
    keinTreffer: '一致なし',
    weiter: '次の一致',
    zurueck: '前の一致',
    einsetzen: '置換',
    alleEinsetzen: 'すべて置換',
    grossKlein:
      '大文字と小文字を区別します。MSE-Scriptでは「card」と「Card」は別のものです。',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: '対象のカード',
    ganzesSet: 'セット全体（{zahl}枚）',
    ganzesSetOhneZahl: 'セット全体',
    teilmenge: '選択したカードのみ（{zahl}枚）',
    ohneNamen: '名称未設定',
    eingrenzen: 'カードを絞り込む',
  },

  vorschlaege: {
    bezeichnung: '候補',
    uebernehmen: '↹で確定',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: '「{region}」の幅',
    breiteWert: '{punkte} ポイント',
    navigatorAuf: 'ナビゲータを表示',
    navigatorZu: 'ナビゲータを隠す',
    inspectorAuf: 'インスペクタを表示',
    inspectorZu: 'インスペクタを隠す',
    navigatorEng:
      '3列すべてを表示するには幅が足りないため、ナビゲータは中央の上に重ねて開きます',
    inspectorEng:
      '3列すべてを表示するには幅が足りないため、インスペクタは中央の上に重ねて開きます',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: '暫定',
    vorlaeufigErklaerung:
      'この版にはまだScryfall識別子がありません。確定するまで、価格と画像は表示されません。',
  },
  kommandozentrale: {
    titel: 'コマンドセンター',
    erklaerung:
      'コマンドはMCPサーバのツールで、ここではウインドウ内から呼び出します。MCPやコマンドラインから呼び出した場合と同じ結果になります。シェルは一切介さず、入力した内容は常にデータとして扱われます。',
    lesend: '読み取り',
    schreibend: '変更あり',
    gefaehrlich: '簡単には元に戻せません',
    faehigkeit: '機能',
    cli: 'コマンドライン',
    rest: 'サービス',
    parameter: 'パラメータ',
    keineParameter: 'このコマンドにパラメータはありません。',
    pflicht: '必須',
    jsonHinweis:
      'JSON形式で入力してください。リストは角括弧、オブジェクトは波括弧で囲みます。',
    befundPflicht: 'このフィールドに値を入力してください。',
    befundZahl: '数値ではありません。',
    befundJson: 'JSONではありません。',
    vorschau: '送信される内容',
    bestaetigen: 'プレビューを確認しました。このまま実行します。',
    ausfuehren: '実行',
    abbrechen: 'キャンセル',
    laeuft: 'コマンドを実行中…',
    ergebnis: '結果',
    ausgabe: 'コマンドの出力',
    alsText: 'テキスト',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      'まだ何も実行していません。トークンやパスワードなどの機密情報は、ここに平文で表示されることはありません。',
    verlauf: 'このセッションの履歴',
    wiederholen: '{name}を再読み込み',
    gelungen: '成功',
    gescheitert: '失敗',
    abgebrochen: 'キャンセル済み',
    schliessen: '閉じる',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: '診断を開く',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'AI を設定',
  },
  kikasten: {
    titel: 'このカードの AI',
  },
} as const
