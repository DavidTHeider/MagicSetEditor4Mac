// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'ファイル情報',
    kopfHinweis:
      'Magic Set Editor がこのセットの保存時に記録した情報です。そのまま引き継がれ、変更されません。',
    mseFassung: '作成に使用したMSE',
    spielfassung: 'ゲームパッケージのバージョン',
    stilfassung: 'スタイルシートのバージョン',
    versionskontrolle: 'バージョン管理',
    apprenticecode: 'Apprenticeコード',
    titel: 'セット情報',
    leerTitel: 'セットフィールドなし',
    leer: 'このセットのゲームパッケージには `set field:` ブロックが宣言されていません。宣言しているゲームを選択してください。',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: '画像',
    erklaerung:
      'すべてのカードが1枚ずつ描画され、書き出されます。処理はこのウインドウ内で実行され、進行状況が表示され、いつでも停止できます。',
    ordnerWaehlen: 'フォルダを選択…',
    keinOrdner: 'フォルダがまだ選択されていません。',
    gemerkt: '前回選択 — すべてのセットで同じフォルダです。',
    nichtGemerkt:
      'このフォルダは今回の書き出しに使われますが、記憶されませんでした：',
    vorlage: 'ファイル名テンプレート',
    konflikt: 'ファイルがすでにある場合',
    konfliktHilfe:
      '同じ名前のカードが2枚あると、同じファイル名になります。「番号を振り直す」は Magic Set Editor 2 の初期設定です。1回の実行内で番号を増やしながら付け、以前の実行で残ったファイルは上書きします。',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: '既存のファイルを残す',
      ueberschreiben: '上書き',
      nummerieren: '番号を付ける',
      nummerierenneu: '番号を振り直す',
    },
    vorlageHilfe:
      '{nummer} は3桁のカード番号に、{name} はカード名になります。同じ名前には番号が追加されます。',
    format: 'フォーマット',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: '品質',
    gueteFast: '高速 — 手早く確認する場合',
    gueteMedium: 'Magic Set Editor 2 の規則に準拠',
    gueteHigh: '高 — 細い線がちらつきません',
    gueteUltra: '最高 — フル解像度での印刷向け',
    breite: '幅 (ピクセル)',
    breiteHilfe:
      'dpi指定はありません。カードの幅が何ミリメートルなのかは、どのテンプレートにも書かれていないためです。幅63 mmのカードで750ピクセルは300 dpiです。',
    ausgeben: '画像を書き出す',
    abbrechen: '停止',
    laeuft: '枚書き出し済み / 全',
    fertig: '枚の画像を書き出しました:',
    abgebrochen: '停止しました。',
    uebersprungen:
      '{anzahl}枚のカードをスキップしました。ファイルがすでにあり、「既存のファイルを残す」が選択されています:',
    befunde:
      '描画が完全ではありませんでした。画像はありますが、次の部分が欠けています:',
  },
  druckbogen: {
    titel: '印刷シート',
    erklaerung:
      'カードは実寸のままです。63 × 88 mm で、裁断するサイズそのものです。トンボは用紙の端に付き、カード上には付きません。',
    papier: '用紙',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: '裁ち落とし (mm)',
    anschnittHilfe:
      '裁ち落としとは、画像が裁断線の外側にはみ出す部分のことです。裁断したカードに白い縁が出るのを防ぎます。カードは、裁ち落としの2倍の間隔をあけて配置されます。',
    aufloesung: '解像度',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'トンボ',
    legen: '印刷シートを作成…',
    fertig: '→',
    seiten: 'ページ:',
    befunde:
      'すべてがうまく収まったわけではありません。シートは作成済みです。次の箇所を確認してください:',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Scryfallの一括データ',
      erklaerung:
        'すべてのカード情報を含むファイルです。これがここにあれば、Scryfall関連の機能はネットワークなしでも動作します。データは最大1日前のものです。',
      fehlt: 'まだ取得されていません。',
      vorhanden: '取得済み、{mb} MB。',
      angeboten: 'Scryfall の提供サイズは {mb} MB、{stand} 時点です。',
      geholt: '取得日時 {zeit}。',
      holen: '一括データを取得',
      laedt: '取得中…',
    },

    titel: 'パッケージ',
    bezugUeberschrift: 'テンプレートを取得',
    bezugErklaerung:
      'MSEのテンプレートはリポジトリで公開されています。取得するのは特定のリビジョン (リリースまたはコミット) であり、ブランチではありません。ブランチは変化するため、同じ指定でも翌日には別のパッケージになりかねないからです。',
    bezugAdresse: 'リポジトリのアドレス',
    bezugStand: 'リビジョン (リリースまたはコミット)',
    bezugUnterverzeichnis: 'サブフォルダ (パッケージのルート)',
    bezugName: 'パッケージフォルダ内の名前',
    bezugPruefen: '取得元を表示…',
    bezugStarten: '取得',
    bezugAbbrechen: '取得しない',
    bezugGeholt: '取得先:',
    bezugPruefsumme: 'チェックサム',
    bezugUebergangen: '展開しませんでした。実行ファイルは含まれません:',
    bezugUnvollstaendig:
      'アドレスとリビジョンが未入力です。両方がないと取得するものがありません。',
    bezugHerkunft: '{quelle} をリビジョン {stand} で取得します。',
    bezugLizenz:
      'これらの内容は MagicSetEditor4Mac のものではなく、当プロジェクトのライセンスの対象外です。カードフレーム、シンボル、フォントは保護されている場合があり、使用は自己責任となります。実行ファイルは展開されません。',
    mse2Ueberschrift: 'MSE2のインストールから取り込む',
    mse2Erklaerung:
      'MSE2をお使いの場合、テンプレートはすでにこのMacにあります。通常の場所を下に示します。選択するまでは、どこも読み込まれません。',
    mse2Uebernehmen: 'ここから取り込む',
    mse2VonHier: 'ここから取り込む — {ort}',
    mse2Waehlen: 'フォルダを選択…',
    mse2KeinVorschlag:
      '通常の場所にMSE2のインストールが見つかりませんでした。テンプレートが入っているフォルダを自分で選択してください。MSE2では「data」という名前です。',
    installerErklaerung:
      'または .mse-installer ファイルから。MSE2はこの方法でテンプレートを配布します。すでに組み込まれているものは、そのまま残ります。',
    installerWaehlen: 'インストーラを読み込む…',
    mse2Ergebnis: '{anzahl}個のテンプレートを取り込みました。',
    mse2Uebergangen:
      'すでに存在するためスキップしました。何も上書きされていません:',
    abhaengigUeberschrift: '依存関係',
    abhaengigAllesGut: 'すべての依存関係が満たされています。',
    abhaengigFehlt:
      '{paket} には {braucht} が必要ですが、インストールされていません。取得するか、MSE2のインストールから取り込んでください。',
    abhaengigZuAlt:
      '{paket} にはバージョン {verlangt} の {braucht} が必要ですが、インストールされているのは {da} です。新しいものがないと、テンプレートは何も知らせずに別のものを描画します。',
    abhaengigZyklus: '{paket} と {braucht} が循環して互いを必要としています。',
    abhaengigUnlesbar:
      '{paket} には {braucht} が必要ですが、バージョンを比較できません。要求: {verlangt}、インストール済み: {da}。',
    quellenErklaerung:
      '4つのコレクションをワンクリックで選べます。アドレスとリビジョンは自動で入力され、同意するまで何も取得されません。テンプレートのフォントも一緒に取得されます。フォントがないと、どのカードも正しく描画されません。4つのいずれにもライセンスはなく、内容は権利者のものであり、当プロジェクトのものではありません。',
    quellenGroesse: '約 {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15: '一般的なM15フレーム。カードを描画するのに最も手早い方法です。',
      fullNonMagic: 'Magic以外のゲーム用のテンプレート。',
      m15: '公開されているM15テンプレート (324パッケージ)。ほとんどの方に適した出発点です。',
      fullMagic:
        '公開されているほぼすべてのMagicテンプレート (526パッケージ)。このコレクションは、このプロジェクトの基準です。',
    },
    katalogUeberschrift: 'インストール済みパッケージ',
    katalogLeer:
      'パッケージがインストールされていません。ゲームとスタイルがないと、プレビューには代替の表示が使われます。',
    katalogName: '名前',
    katalogArt: '種類',
    katalogVersion: 'バージョン',
    katalogHandlung: '操作',
    katalogBefunde: 'パッケージフォルダの読み込み中:',
    anlegen: 'パッケージを作成',
    anlegenArt: '新しいパッケージの種類',
    anlegenArtSpiel: 'ゲーム (.mse-game)',
    anlegenArtStil: 'スタイル (.mse-style)',
    anlegenArtInclude: 'インクルード (.mse-include)',
    anlegenArtExport: '書き出しテンプレート (.mse-export-template)',
    anlegenName: '短い名前',
    anlegenNameHinweis:
      '英字、数字、ハイフン、アンダースコアが使えます。フォルダ名になります',
    anlegenVollname: '正式名称',
    anlegenVollnameHinweis: '空欄の場合は短い名前と同じになります',
    anlegenSpiel: '新しいパッケージのゲーム',
    anlegenSpielHinweis:
      'スタイルはそのゲームのフィールドを配置し、書き出しテンプレートはそれを出力します',
    anlegenSpielLeer: '選択してください',
    anlegenKnopf: '作成',
    anlegenGelungen: '作成しました: {name}',
    loeschen: '削除',
    loeschenAn: '削除 — {name}',
    handlungenAn: '操作 — {name}',
    handlungenKurz: '…',
    katalogSuche: '検索',
    auswahlLoesen: 'すべてのパッケージを表示',
    leiste: {
      bezeichnung: 'パッケージの数値',
      installiert: 'インストール済み',
      befund: '指摘あり',
      nichtGelesen: '未読込',
      auswahl: '指摘のある{pakete}',
    },
    katalogSucheHinweis: '名前、種類、取得元、またはグループ',
    katalogArtAlle: 'すべての種類',
    katalogOrdnung: '並べ替え',
    katalogOrdnungGruppen: 'グループ',
    katalogOrdnungName: '名前',
    katalogOrdnungVersion: 'バージョン',
    katalogOrdnungHerkunft: '取得元',
    katalogGefunden: 'パッケージ: {gesamt}件中{gefunden}件',
    katalogNichtsGefunden:
      'この検索に一致するパッケージはありません。別の表記を試すか、種類のフィルタを解除してください。',
    katalogLeerTitel: 'まだパッケージがありません',
    katalogNichtsTitel: '該当なし',
    katalogOhneGruppe: 'グループなし',
    katalogGruppeAn: 'グループ {name}',
    katalogGruppe: 'グループ',
    katalogOhneVersion: 'バージョンなし',
    katalogZustandFehlt: '依存先がありません',
    katalogZustandZuAlt: '依存先が古すぎます',
    katalogZustandUnlesbar: 'バージョンを比較できません',
    katalogZustandZyklus: '依存関係が循環しています',
    katalogEinVorbehalt: '検出事項 1件',
    katalogVorbehalte: '検出事項 {anzahl}件',
    umschlagTitel: 'エンベロープからのテンプレート',
    umschlagErklaerung:
      'エンベロープにはテンプレートの**ファイル**が入っています。拡張子付きの名前を、たとえば `mygame.mse-game` や `mystyle.mse-style` のように指定します。テンプレートはその名前でここに保存されます。同じパッケージがすでにある場合は何も上書きされず、拒否されるので、別の名前を指定してください。',
    umschlagName: 'パッケージの名前',
    umschlagOhneNamen: 'まず、パッケージの名前を入力してください。',
    umschlagAnnehmen: 'テンプレートを受け入れる',
    umschlagAngenommen: 'テンプレートがパッケージフォルダに保存されました。',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: '置換',
    erklaerung:
      'このセットのテキストフィールド内を置換します。最初にプレビューを確認してから適用します。操作全体は1回の取り消しで元に戻せます。',
    suchtext: '検索する文字列',
    ersatz: '置換後の文字列',
    grossKlein: '大文字と小文字を区別',
    ganzesWort: '単語全体のみ',
    vorschauen: '変更内容を表示',
    ersetzen: '今すぐ置換 ({treffer})',
    nichtsGefunden: '見つかりませんでした。別の語を試しますか？',
    zusammenfassung: '一致: {treffer} · 影響を受けるカード: {karten}',
    getan: '置換しました — 一致: {treffer} · カード: {karten}',
    ruecknahme: '⌘Zで操作全体を取り消せます。',
    stellen: '影響を受ける箇所',
    kartenfeld: 'カード {nummer} · {feld}',
    weitere: 'これ以外の箇所は一覧に表示されませんが、同様に置換されます。',
  },

  statistik: {
    worueber: '集計の基準',
    belegungWahl: 'フィールドごとの入力率',
    gruppeSpielpaket: 'ゲームパッケージによる',
    gruppeFelder: 'フィールドの値による',
    verteilungVon: '「{feld}」の分布 — 値を持つカードに対する割合',
    nachDimension: '「{dimension}」による統計',
    titel: '統計',
    karten: 'カード',
    felder: '異なるフィールド',
    leere: '値がまったくないカード',
    belegung: 'フィールドごとの入力率 — セット内の全カードに対する割合',
    belegungLeer: 'このセットには、まだフィールドの値がありません。',
    verteilungLeer: 'このセットでは、このフィールドに値がありません。',
    anteil: '割合',
    dimensionenErklaerung:
      'ゲームパッケージ自身が、何を基準に集計できるかを示します。色区分、マナ総量、レアリティなどです。各ディメンションはカードごとに値を計算するため、少し時間がかかります。',
    dimensionLeer: 'このディメンションでは、値を持つカードがありません。',
    gruppeHinweis: 'バーを選ぶと、そのカードがカード一覧に表示されます。',
    gruppeGewaehlt: '選択中:「{gruppe}」 · カード:{karten}',
    gruppeZeigen: 'カードを表示',
    gruppeAufheben: '選択を解除',
    gruppeLeer: 'この値を持つカードはありません。',
    filterBezeichnung: '統計からのフィルター',
    filterKopf: '統計から:「{quelle}」=「{gruppe}」',
    filterZurStatistik: '統計に戻る',
    filterAufheben: 'フィルターを解除',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'スタイル',
    vorschauTitel: 'プレビュー',
    vorschauLeer:
      'このセットにはまだカードがありません。カードを作成すると、スタイルを確認できます。',
    eigeneFarbe: 'カスタムカラー',
    vorgabe: '初期設定',
  },
  suche: {
    titel: '検索',
    eingabe: '何を探しますか？',
    recherche: 'リサーチ…',
    feld: 'このフィールドのみ',
    alleFelder: 'すべてのフィールド',
    set: 'このセットのみ',
    alleSets: 'すべてのセット',
    leer: 'まだ何も検索していません。入力してカードを探しましょう。',
    nichts: '見つかりませんでした。別の語を試しますか？',
    gekuerzt:
      '表示しきれない件数の一致があります。検索条件を絞り込むと表示されます。',
    treffer: '検索結果',
  },
  ausgeben,
  hilfe: {
    titel: 'ヘルプ',
    erklaerung:
      'このアプリケーションで使う用語の意味と、初回起動時にすることを説明します。',
    glossar: '用語',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'このページ内',
    seitenuebersicht: '各ページの機能',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
