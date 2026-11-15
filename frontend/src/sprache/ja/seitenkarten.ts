// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'カードを追加',
  nachOben: '上へ移動',
  nachUnten: '下へ移動',
  duplizieren: 'カードを複製',
  loeschen: 'カードを削除',
  loeschenFrage: 'このカードを削除しますか？ ⌘Zで取り消せます。',
  loeschenJa: '削除',
  loeschenNein: 'キャンセル',
  loeschenTrotzdem: 'それでも削除',
  symbole: 'シンボルを挿入',
  hergeleitet: '導出',
  hergeleitetWert: '導出: {wert}',
  bildWaehlen: '画像を選択',
  stilfelder: 'スタイルの項目',
  stilfelderHilfe:
    'これらの項目はカードのスタイルに含まれています。値はこのスタイルにのみ適用されます。',
  symbolZeichnen: 'シンボルを描く',
  kartenskripteTitel: 'スクリプトからのカード',
  kartenskripteErklaerung:
    'ゲームパッケージに用意された処理です。1回のクリックで複数のカードをまとめて追加します。' +
    '取り消しも1回でまとめて元に戻ります。',
  kartenskriptEine: 'カードを1枚追加しました。',
  kartenskriptMehrere: 'カードを{anzahl}枚追加しました。',
  bildEinsetzen: '画像を挿入',
  pack: {
    ueberschrift: 'ランダムパック',
    erklaerung:
      'このセットのカードから、ゲームパッケージの規則に従ってブースターを作ります。同じシードなら同じ結果になります。結果を再現できるのはそのためです。',
    laedtGrund:
      'パックの種類ごとに、すべてのカードが条件に合うかを確認します。初回は少し時間がかかりますが、その後はすぐに使えます。',
    art: 'パックの種類',
    same: 'シード',
    sameZahl: 'シードは 0 から 18446744073709551615 までの整数です。',
    ziehen: '引く',
    neuerSame: '新しいシード',
    gezogen: '引いた結果',
    keineArten:
      'このセットのゲームには、パックの種類がありません。またはそのパッケージがインストールされていません。',
    keineKarten: 'このパックの種類からはカードが出ませんでした。',
    artMitZahl: '{name} ({zahl}枚)',
    nurBaustein: '構成要素 — 他のパックの種類から引かれます',
    vorbehalte: 'フィルタを評価できなかったパックの種類',
  },
  symbol: {
    ueberschrift: 'シンボル',
    erklaerung:
      'このセットのシンボルです。読み込むのではなく、描画します。プレビューにはすぐ反映され、セットを保存するとファイルに書き込まれます。どの図形もキーボードで配置・移動できます。',

    grundform: '基本図形を追加',

    verlauf: '描画の履歴',
    nachzeichnen: '画像をトレース',
    groesser: '大きく',
    nachzeichnenHilfe:
      'シルエットがはっきりしたPNG、JPEG、GIFは、ここで編集を続けられるベジェ点を持つ図形になります。',
    zuruecknehmen: '操作を取り消す',
    wiederherstellen: '操作をやり直す',
    kreis: '円',
    rechteck: '長方形',
    dreieck: '三角形',
    stern: '星形',

    teile: 'パーツ',
    keineTeile:
      'まだ何も描かれていません。まず基本図形を追加してください。あとから、どの点も変更できます。',
    teilLoeschen: 'パーツを削除',
    teilMitPunkten: '{name} — {zahl}点',
    teilMitKopien: '{name} — 対称、コピー{zahl}個',
    gruppieren: 'グループ化',
    gruppierenHilfe:
      '2つ以上のパーツを選択すると、グループになります。⌘を押しながらクリックすると複数選択できます。',
    aufloesen: 'グループ解除',
    nachOben: '上へ移動',
    nachUnten: '下へ移動',
    nachObenFuer: '{name}を上へ移動',
    nachUntenFuer: '{name}を下へ移動',
    teilLoeschenFuer: '{name}を削除',
    teilMitKindern: '{name} — パーツ{zahl}個',
    teilKind: '{name}、グループ内',
    wahlHilfe:
      '⌘を押しながらクリックすると複数のパーツを選択できます。順序は描画順です。一覧の上にあるものが前面に表示されます。',

    zeichnung: '描画エリア',
    punkt: '{name}のポイント{nummer}',
    griffDavor: 'ポイント{nummer}の手前のハンドル',
    griffDanach: 'ポイント{nummer}の後ろのハンドル',
    punktHinzufuegen: 'ポイントを挿入',
    punktLoeschen: 'ポイントを削除',

    tastatur: 'キーボードで操作',
    tastaturhilfe:
      'Tabキーでポイント間を移動します。矢印キーで100分の1ずつ、Shiftキーを押しながらだと10分の1ずつ移動します。Returnキーで選択中のポイントの後ろに新しいポイントを挿入し、deleteキーで削除します。',

    kombination: '合成',
    ueberdecken: '重ねる',
    vereinen: '結合',
    abziehen: '減算',
    schneiden: '交差',
    unterschied: '差分',
    umranden: '縁取り',

    symmetrie: '対称',
    symmetrieAnlegen: '対称を追加',
    symmetrieAufloesen: '対称を解除',
    symmetrieErklaerung:
      '対称は、その中にあるものを複製します。コピーは描画エリアに薄く表示され、編集されるのは常に元の図形だけです。',
    symmetrieHinweis: 'パーツを選択すると、対称にできます。',
    symmetrieart: '種類',
    drehung: '回転',
    achse: '反転',
    kopien: 'コピー数',
    kopienHinweis:
      '反転にはペアが必要です。奇数だと1つのコピーが単独になり、MSEは切り捨てます。',
    mitteX: '中心 (水平)',
    mitteY: '中心 (垂直)',
    griffX: 'ハンドル (水平)',
    griffY: 'ハンドル (垂直)',
    mitte: '対称の中心',
    griff: '対称のハンドル',

    sichern: 'セットに保存',
    gesichert:
      '{name}として保存しました。セットを保存するとファイルに書き込まれます。',
  },
  konsole,

  wortlisten: {
    ueberschrift: '選択肢',
    liste: '{liste}から選択',
    keine: 'いずれでもない',
    gruppe: '{name} (グループ)',
  },
  regeln: {
    ueberschrift: 'Scryfallの裁定',
    holen: '裁定を取得',
    keine: 'このカードには裁定がありません。',
    offiziell: '公式',
    anmerkung: 'Scryfallの注記',
    anmerkungErklaerung:
      '注記はScryfallチームによるもので、非公式なフォーマットを説明していることがよくあります。ルールではありません。',
    quelle: 'このカードをScryfallで見る',
    alter: '{stunden}時間前の情報',
    manaPruefen: 'マナ・コストを確認',
    manaGut: 'マナ・コストに問題はありません。',
    manaSchlecht: '不明: {symbole}',
    manaLeer: 'このカードにはマナ・コストがありません。',
  },
  abgleich: {
    ueberschrift: 'Scryfallのカードデータ',
    erklaerung:
      '変更される内容を表示します。チェックを入れた項目だけが書き込まれます。',
    holen: '提案を取得',
    uebernehmen: 'チェックした項目を適用',
    nichtsZuTun: '変更はありません。カードはScryfallと一致しています。',
    alt: '現在',
    neu: '新規',
    leer: '(空)',
    alter: '{stunden}時間前の情報',
    frisch: 'たった今取得した情報',
    quelle: 'このカードをScryfallで見る',
  },
  scryfall: {
    ueberschrift: 'Scryfallの画像',
    erklaerung:
      'カード名で画像が決まります。アーティストもカードに書き込まれます。Scryfallでは、イラストの隣にアーティスト名を表示することが求められています。',
    name: 'Scryfallでのカード名',
    art: '画像の種類',
    artKlein: '小 (プレビュー)',
    artNormal: 'カード全体',
    artAusschnitt: 'イラストのみ',
    holen: '画像を取得',
    ausschnittHinweis:
      'イラストのみの切り抜きは、イラストだけを表示します。Scryfallがアーティストを示していない場合は拒否されます。',
    quelle: 'このカードをScryfallで見る',
  },
  scryfallkarte: {
    name: 'Scryfallでのカード名',
    platzhalter: '例: Lightning Bolt',
    mitBild: 'イラスト付き',
    anlegen: 'Scryfallから作成',
  },
  zuschnitt: {
    ueberschrift: '切り抜きを選択',
    erklaerung:
      '枠は、画像のどの部分がカードに入るかを示します。ドラッグで移動し、角をドラッグしてサイズを変更します。',
    quelle: '選択した切り抜き付きの元画像',
    vorschau: 'フィールドサイズでのプレビュー',
    links: '左',
    oben: '上',
    breite: '幅',
    hoehe: '高さ',
    seitenverhaeltnis: 'フィールドの縦横比を保持',
    uebernehmen: '挿入',
    abbrechen: 'キャンセル',
    masse: '{breite} × {hoehe} ピクセル',
  },
  feldWaehlen: 'このフィールドを編集',
  leerTitel: 'カードがまだありません',
  listeBezeichnung: 'セットのカード',
  mehr: 'その他',
  gruppeAnlage: 'カード',
  gruppeAnsicht: '表示',
  ausScryfallAufklapp: 'Scryfallから作成…',
  ausScryfallErklaerung:
    '空のカードの代わりに、名前を入力するかScryfallで検索します。フィールドはそこから取り込まれます。',
  dieseKarte: 'このカード',
  kartenmenue: 'カード',
  auswahlLoesen: 'すべてのカードを表示',
  bezugAufklapp: '印刷版への紐付け…',
  dieseKarteLeer: 'カードを開くと編集できます。',
  keineKarteTitel: 'カードが選択されていません',
  titel: 'カード',
  leer: 'カードを作成するか、コピーしたカードをペーストしてください。',
  ohneName: '名前なし',
  ohneFelder: 'このカードにはフィールドの値がまったくありません。',
  nummer: '番号',
  vorschau: 'カードのプレビュー',
  keineKarte:
    'リストから選択すると、そのフィールドとプレビューがここに表示されます。',
  bearbeiten: '編集',

  notizen: {
    ueberschrift: 'カードのメモ',
    platzhalter: 'あなた専用です。カードには表示されません。',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'アプリケーションが把握している情報',
  tiefe: '表示の詳しさ',
  tiefeNormal: '標準',
  tiefeErweitert: '詳細',
  tiefeExperte: 'エキスパート',
  tiefeErweitertTitel: 'スクリプトが入力する内容',
  tiefeErweitertHinweis:
    'これらはゲームパッケージが計算します。カードに何かが表示される理由を示しており、編集はできません。',
  befundVerdeckt: '「{tiefe}」に検出事項があります。',
  befundSpringen: 'そこへ移動',
  buchfuehrungHinweis:
    'これらの項目はカード自体ではなく、その管理情報です。どのテンプレートを使い、何が設定されているかを示します。理由がわかる場合にのみ変更してください。',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'このカードの作業状況',
  arbeitsstandHinweis:
    'メモとマークは、カードではなく、カードに対するあなたの作業に属します。共有するファイルには含まれません。',

  marken: {
    ueberschrift: 'マーク',
    erklaerung:
      '自分の作業用の言葉です。カードの中ではなく、カードの横に付きます。共有するファイルには含まれません。',
    platzhalter: '例: 完了',
    setzen: 'マークを追加',
    entfernen: 'マーク {marke} を外す',
    keine: 'このカードにはまだマークがありません。',
    schonGesetzt:
      'このカードには別の場所で「{marke}」がすでに付けられています。',
    filter: 'マークで絞り込む',
    alle: 'すべて',
  },

  scryfallsuche: {
    frage: 'Scryfallで検索',
    platzhalter: '例: Lightning Bolt、または t:creature c:r cmc<=3',
    erklaerung:
      'カード名だけで検索できます。Scryfallの検索構文を知っている場合は、ここで使えます。',
    suchen: '検索',
    nichtsGefunden: '見つかりませんでした。別の語を試しますか？',
    mehr: 'Scryfallにはさらに検索結果があります。検索条件を絞り込んでください。自動では取得されません。',
    alter: '{stunden}時間前の情報',
    frisch: 'たった今取得した情報',
    befund: '検出事項: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: '紐付け先',
    keiner:
      'このカードは特定の版に紐付けられていません。この状態でScryfallと同期すると、任意の版が選ばれます。上で検索して紐付けてください。',
    druck: '{setcode} {nummer} より',
    gesetzt: '— {wann}に紐付け',
    setzen: '{name} ({setcode} {nummer}) に紐付ける',
    loesen: '紐付けを解除',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'コピー ({zahl})',
    ausschneiden: 'カット ({zahl})',
    einfuegen: 'ペースト',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'リストの列',
    jeSpiel: 'この選択は、同じゲームのすべてのセットに適用されます。',
    filtern: '列を検索',
    gezeigt: '{alle}列中{gezeigt}列',
    ausgeblendet: 'ほかに {anzahl} 列が非表示です — 上の欄で検索してください。',
    genauer: 'さらに一致する列があります — 絞り込むと表示されます。',
    nachOben: '左へ移動',
    nachUnten: '右へ移動',
    sortieren: '並べ替え',
    garNicht: 'セットの順序',
    aufsteigend: '昇順',
    absteigend: '降順',
    ohneVorlage:
      'ゲームパッケージを読み込めません。列はカード自体から作られ、見出しやパッケージの順序はありません。',
    spalteAusblenden: '列を非表示',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— フィールドを選択 —',
    keinWert: '— 値を選択 —',
    ueberschrift: '一括変更',
    erklaerung:
      'チェックを入れたすべてのカードの1つのフィールドを、1回の操作で設定します。途中で問題が起きた場合、1枚も変更されません。',
    auswaehlen: '一括変更のためにチェック',
    mehrereWaehlen: '複数選択',
    ausgewaehlt: 'チェック済み',
    keineAuswahl:
      'チェックを入れたカードがありません。リストで複数のカードにチェックを入れると、まとめて変更できます。',
    feld: 'フィールド',
    wert: '新しい値',
    anwenden: '適用',
    betroffen: 'これらのカードで変更に失敗しました。1枚も変更されていません:',
    gescheitert: '失敗',
    ohneWert: 'まだ値なし',
  },
  stilTitel: 'スタイル',
  stilWaehlen: 'このセットのスタイル',
  stilLeer:
    'スタイルがインストールされていません。スタイルがないと、プレビューには代替の表示が使われます。',
  stilFrage: 'このスタイルに切り替えますか？',
  stilSpielwechsel: '新しいスタイルは別のゲームのものです。',
  stilUngedeutet:
    '新しいゲームはこれらのフィールドを認識しません。値はデータベースに残りますが、描画はされなくなります:',
  stilVerloreneOptionen:
    '新しいスタイルにはこれらのスタイルオプションがありません。値は失われます:',
  stilUebernehmen: 'スタイルを切り替える',
  stilAbbrechen: '現在のスタイルを維持',
  stilKeiner:
    'このセットが指定しているスタイルはインストールされていません。インストール済みのものから選択してください。',
  stiloptionen: 'スタイルオプション',
  stiloptionenSichtbar: '{gesamt} 件中 {zahl} 件',
  stiloptionenWeitere: 'その他のオプション（{zahl}）',
  marktTitel: 'このカードの入手先',
  marktErklaerung:
    'これらのリンクは、カード名または提供元の識別子から、このMac上で作られます。クリックするまで何も送信されず、お好みのブラウザで開きます。キーは不要です。そこで販売されているかどうかは、提供元次第です。',
  marktGenau: 'このカードそのものへ移動',
  marktSuche: '検索へ移動',
  marktNurAnbieter: '提供元のサイトへのみ移動',
  umschlagAnnehmen: 'エンベロープからのカード',
  umschlagAngenommen: 'カードがこのセットに追加されました。',
}
