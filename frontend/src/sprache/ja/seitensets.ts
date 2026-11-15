// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'すべてのセットを表示',
  erklaerung:
    '開いているセットはすべてデータベースにあります。ここからカードを開いたり、ファイルに書き戻したりできます。',
  titel: 'セット',
  leer: 'まだセットが開かれていません。.mse-set ファイルを開いて始めましょう。',
  setOeffnen: 'セットを開く…',
  neuTitel: '新規セット',
  neuKopf: '新規セット',
  einlesenTitel: '他の場所からセットを読み込む',
  neuHinweis:
    '空のセットです。カードの見た目は、ゲームとスタイルで決まります。スタイルはあとから変更できます。',
  neuName: 'セットの名前',
  neuStil: '新規セットのスタイル',
  neuAnlegen: 'セットを作成',
  neuSchritte: '新規セットの作成手順',
  neuSchrittArt: '種類',
  neuSchrittAussehen: '外観',
  neuSchrittName: '名前',
  neuArtFrage: 'カードは何のためのものですか？ パッケージ名は自動で探します。',
  neuArtZusatz: '· {stile}個利用可能',
  neuEigenes: '完全にオリジナル — ゲームパッケージを作成',
  neuWeiter: '次へ',
  neuSchrittZurueck: '戻る',
  neuUeberspringen: 'スキップ',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'ゲームパッケージがインストールされていません。ゲームパッケージがないと、カードのフィールドがありません。パッケージはパッケージのページから取得します。',
  neuOhneStile:
    '{spiel} 用のスタイルパッケージがインストールされていません。スタイルがないと、セットを作成できません。パッケージはパッケージのページから取得します。',
  neuZuDenPaketen: 'パッケージへ移動',
  neuZurueck: '⌘[ を押すとここに戻れます。入力した内容はそのまま残っています。',
  beispielsets: '試用のために同梱',
  geoeffnet: '開きました',
  schonOffen: 'このファイルはすでに開かれています。こちらです。',
  teilerfolg:
    'ファイルは最後まで読み込めませんでした。それでも{anzahl}個のセットがデータベースに入っています。下のリストにあります。',
  beimOeffnenBefunde:
    '開くときに読み込めない部分がありました。セットは開かれていますが、次の部分が欠けています:',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: '読み込み前: {name}',
    umfang: '{karten} · ゲーム {spiel} · スタイル {stil}',
    umfangOhneStil: '{karten} · ゲーム {spiel}',
    schonOffen:
      'このファイルはすでにライブラリにあります。「開く」を選ぶと、2つ目を追加せずに既存のセットが表示されます。',
    befunde: '次の部分は読み込めません。セットはそれらなしで読み込まれます:',
    bedarf:
      'パッケージが足りません。それでも読み込めます。パッケージが揃うまでは、カードは代替の表示で描画されます:',
    einlesen: '読み込む',
    aufschlagen: '開く',
    verwerfen: '読み込まない',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: '削除…',
  loeschenTitel: 'このセットを削除',
  loeschenWarnung:
    '{name} は、{karten}とともにデータベースから削除されます。ゴミ箱はなく、元に戻せません。ファイルにあるものは、そのまま残ります。削除されるのはデータベース内のコピーです。',
  loeschenSammlung:
    'コレクションは残ります。所持カードはセットではなく、版に紐付いているためです。',
  loeschenFeld: '確認のため、セットの名前を入力してください',
  loeschenPasstNicht: '名前が {name} とまだ一致しません。',
  loeschenJa: '完全に削除',
  loeschenNein: 'キャンセル',
  loeschenBefunde: 'セットは削除されました。次の残りは片付けられませんでした:',

  spalten: {
    name: '名前',
    spiel: 'ゲーム',
    karten: 'カード',
    herkunft: '出所',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'mtgjson.comのカード定義です。1つのファイルでも、フォルダでも読み込めます。どのカードフィールドに対応するかは、ゲームによって決まります。',
  mtgjsonSpiel: 'ゲーム',
  mtgjsonDatei: 'ファイルを開く…',
  mtgjsonVerzeichnis: 'フォルダを開く…',
  mtgjsonKeinSpiel:
    'MTG JSONのフィールドの対応付けを持つゲームがありません。それがないと、何も読み込めません。',
  scryfallsatzUeberschrift: 'Scryfallのセット',
  scryfallsatzErklaerung:
    'セットコードによって、どのカードが取り込まれるかが決まります。テキストは取り込まれますが、画像は取り込まれません。名前、マナ・コスト、タイプ、ルール・テキスト、フレーバー・テキスト、パワー、タフネス、レアリティ、イラストレーターが取り込まれます。画像はカードごとに取得します。スタイルは、あとで選びます。',
  scryfallsatzKeinSpiel:
    '対応付けを持つゲームがないと、セットを読み込めません。',
  scryfallsatzSpiel: '対応付けに使うゲーム',
  scryfallsatzCode: 'Scryfallでのセットコード',
  scryfallsatzPlatzhalter: '例: lea',
  scryfallsatzEinlesen: 'セットをインポート',
  vorlageAusfuhr: 'テンプレートで書き出す…',
  vorlageWaehlen: '書き出しテンプレート',
  vorlageAusfuehren: '実行',
  vorlageSchliessen: '閉じる',
  vorlageKeine:
    'このゲームには書き出しテンプレートがインストールされていません。テンプレートは、取得したパッケージに含まれています。',
  vorlageGeschrieben: '書き出し先:',
  vorlageBefunde:
    '実行中に気になる点がありましたが、ファイルは書き出されました:',
  vorlageGescheitert:
    'このテンプレートは最後まで実行できません。不足している内容は下に表示されます:',
  mtgjsonSichern: 'MTG JSONとして保存…',
  mtgjsonGesichert: '枚のカードを保存しました:',
  mtgjsonVerluste:
    '書き出し時に表現できなかった内容があります。ファイルは作成されましたが、次の詳細が欠けています:',
  mtgjsonSpalte: '書き出し',
  speichern: '保存',
  speichernUnter: '別名で保存…',
  speicherSpalte: '保存',
  gespeichert: '保存先:',
  beimSpeichernBefunde:
    '保存時に含められなかったものがあります。セットファイルは書き込まれましたが、その隣の次のファイルが欠けています:',
  oeffnen: 'このセットのカードを表示',
  handlungenAn: '操作 — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: '未解釈 ({anzahl})',
    ueberschrift: '未解釈',
    erklaerung:
      'インポートでは、これらの行を理解できませんでした。変更されずに引き継がれ、保存しても失われません。',
    keine: 'なし — このセットは完全に読み込まれました。',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'エンベロープとして保存',
  umschlagGeschnuert:
    '保存しました。エンベロープには署名されています。受け取った人は、それがここから来たものであることを確認できます。',
  umschlagTitel: 'エンベロープを取り込む',
  umschlagErklaerung:
    'エンベロープは、他の誰かが保存したセットです。最初に内容を確認し、そのあとで取り込みます。それまでは、データベースは何も変更されません。',
  umschlagWaehlen: 'エンベロープを確認…',
  umschlagUrheber: '署名者',
  umschlagAbdruckHinweis:
    'フィンガープリントが示すのは、誰が署名したかであり、その人を信頼すべきかどうかではありません。その人を知っている場合は、本人と照合してください。',
  umschlagInhalt: '内容',
  umschlagUebernehmen: '取り込む',
  umschlagVerwerfen: 'やめる',
  umschlagUebernommen:
    '取り込みました。セットは新規に作成されます。独自の識別子を持ち、何も上書きしません。',
}
