// The **Decks** page — Werkstück 7 of phase 4.6
// (#1293).
export const decks = {
  titel: 'デッキ',
  erklaerung:
    '自作のカードと実在の版を組み合わせたものです。デッキはカードをコピーせずに参照するため、元のセットでカードを修正すると、すぐに反映されます。',

  leer: 'デッキはまだありません。上の「デッキを作成」から始められます。デッキはゲームに属します。ゲームがないと、チェックするフォーマットがありません。',

  neu: 'デッキを作成',
  name: '名前',
  spiel: 'ゲーム',
  notiz: 'メモ',
  anlegen: '作成',
  ausSetTitel: 'またはセットから',
  ausSetErklaerung:
    'セットの各カードがメインデッキの枠になります — コピーではなく参照です。セットで訂正すれば、デッキでも訂正されます。',
  ausSetSet: 'セット',
  ausSetMenge: 'カード1枚あたりの枚数',
  ausSetAnlegen: 'セットから作成',
  ausSetKeinSet:
    'カードのあるセットがまだありません — セットからデッキを作るには、セットが必要です。',

  hauptdeck: 'メインデッキ',
  sideboard: 'サイドボード',
  zaehlung: 'メインデッキに{haupt}、サイドボードに{side}',
  waehlen: '開く',
  offenMarke: '表示中',
  umbenennen: '名前を変更',
  neuerName: '新しい名前',
  umbenennenTitel: '{name}の名前を変更',
  nameSichern: '保存',
  umbenennenAbbrechen: '名前を変更しない',
  duplizieren: '複製',
  kopieTitel: '{name}を複製',
  kopieName: 'コピーの名前',
  kopieVorschlag: '{name}（コピー）',
  kopieAnlegen: 'コピーを作成',
  kopieAbbrechen: '複製しない',
  aufWunschliste: '足りないカードをウィッシュリストへ',
  wunschlisteErgebnis:
    'ウィッシュリストに新規追加：{neu} · すでにあるもの：{schon}',
  loeschen: 'デッキを削除',
  handlungenAn: '操作 — {name}',
  handlungenKurz: '…',
  deckmenue: 'デッキ',
  loeschenFrage: 'このデッキを削除しますか？カード自体は残ります。',
  loeschenJa: '削除',
  loeschenNein: 'キャンセル',

  plaetze: '枠',
  plaetzeLeer: 'デッキにカードはまだありません。',
  bezug: 'カードまたは版',
  menge: '数量',
  bereich: '区分',
  setWaehlen: 'セット',
  karteWaehlen: 'カード',
  karteSuchen: 'カードを検索',
  karteKeine: '「{suche}」に一致するカードはありません。',
  setLeer:
    '開いているセットがありません。自作のセットのカードは、セットができてから選べます。',
  ohneSet: 'セットを選択…',
  ohneKarte: 'カードを選択…',
  setOhneKarten: 'このセットにはまだカードがありません。',
  kurve: 'マナ・カーブ',
  kurveHinweis:
    'マナ総量ごとのカード枚数です。値がわかるものを数えます。Scryfallの版はそこに記録されている値、それ以外はコストから計算した値を使います。土地にはコストがないため含まれず、コスト欄が空の自作カードも含まれません。',
  farben: '色',
  farbeWeiss: '白',
  farbeBlau: '青',
  farbeSchwarz: '黒',
  farbeRot: '赤',
  farbeGruen: '緑',
  farbenHinweis:
    '2色のカードは両方の色で数えるため、合計はデッキの枚数より大きくなります。',
  karten: 'カード',
  gebietFormate: 'フォーマット',
  formateKnopf: 'フォーマット…',

  // **Taking a deck from someone else's profile** — Werkstück 12 of Phase
  // 4.12 (#1351).
  uebernahme: '別のプロファイルからデッキを取得',
  auswahlLoesen: 'すべてのデッキを表示',
  leiste: {
    bezeichnung: 'デッキの数値',
    sideboard: 'サイドボード {zahl}',
    pruefung: 'チェック',
    ohneFormat: 'フォーマットなし',
    nichtGeprueft: '未チェック',
    fehlt: 'コレクションに不足',
    unbekannt: '不明 {zahl}',
    nichtAbgeglichen: '未照合',
  },
  uebernahmeErklaerung:
    '切り替え可能なプロファイルであれば、どれからでも取得できます。所有者が第2要素を使っている場合は、そのプロファイルでの共有が必要です。「閲覧者」で十分です。一緒に取得されないもの：そのプロファイル自身のカードを指している枠と、フォーマット。どちらも向こう側のデータです。取得されなかったものは、あとでここに表示されます。',
  uebernahmeProfil: '別のプロファイルのID',
  uebernahmeDeck: 'デッキのID',
  uebernahmeProfilwahl: 'どのプロファイルから？',
  uebernahmeProfilWaehlen: 'プロファイルを選択…',
  uebernahmeProfilEintrag: '{nutzer} · {profil}',
  uebernahmeKeinProfil:
    'このコンピュータには、ほかのプロファイルがありません。',
  uebernahmeDeckwahl: 'どのデッキ？',
  uebernahmeDeckWaehlen: 'デッキを選択…',
  uebernahmeKeinDeck: 'このプロファイルにはデッキがありません。',
  uebernahmeFachleute: 'IDを直接指定',
  uebernahmeFachleuteErklaerung:
    'プロファイルとデッキは、コマンドラインやサービスが扱うIDでも指定できます。許可の内容は同じです。',
  uebernahmeHolen: '取得',
  uebernahmeGeschrieben: '「{name}」がデッキに追加されました（{karten}）。',
  uebernahmeNichtMit: '取得されなかったもの：',

  gebietKarten: 'カード',
  gebietAuswertung: '分析',
  gebietPruefung: 'チェック',
  gebieteDeck: 'このデッキのエリア',
  legen: 'デッキに入れる',
  nehmen: '取り除く',
  druck: '版（Scryfall ID）',
  anschlussOhneDeck:
    '「{name}」が待機中です。デッキを開くか作成すると、そのカード選択に表示されます。',
  holenAus: 'カードの取得元',
  holenSammlung: '自分のコレクション',
  holenSet: 'セット',
  wegeTitel: '取り込みと受け渡し',
  mitspielen: 'このデッキでプレイ',
  anschlussWartet:
    '「{name}」がコレクションから移動してきました。数量と区分を選ぶと、デッキに入ります。',

  format: 'フォーマット',
  formatWaehlen: 'フォーマットを選択',
  ohneFormat: 'フォーマットなし — 未チェック',
  formatHinweis:
    'フォーマットのないデッキは、誤りではなく未チェックです。フォーマットはデッキのゲームに属している必要があります。',

  pruefen: 'フォーマットでチェック',
  zulaessig: '使用可能です。',
  unzulaessig: '使用できません。',
  ungeprueftZahl: '{zahl}枚のカードが未チェックです。',
  ungepruefterklaerung:
    '未チェックは、使用不可という意味ではありません。自分でデザインしたカードは、どのフォーマットにも載っていません。「restricted」もデッキの妨げにはなりません。制限カードを何枚入れられるかはトーナメントルールに書かれていて、このプログラムはそれを読み込んでいません。',

  legal: '使用可能',
  not_legal: '対象外',
  restricted: '制限',
  banned: '禁止',
  ungeprüft: '未チェック',

  bestand: '手元にあるもの',
  vorhanden: '所有',
  gebraucht: '必要',
  fehlend: '不足',
  unbekannt: '不明',
  bestandVollstaendig: 'すべてそろっています。',
  bestandFehlt: '{zahl}枚のカードが不足しています。',

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
  gebietZuordnung: '割り当て',
  zuordnungen: '必要数と割り当て',
  zuordnungenLaden: '読み込み中…',
  zuordnungenLeer:
    'デッキにカードはまだありません。ここに表示される内容は、その枠から作られます。',
  zuordnungHinweis:
    'デッキリストは何も確保しません。同じ3枚のカードを持つ2つのデッキは、どちらも完成しています。割り当て済みは、このデッキ用に明示的に取り置いた分で、空きはその後に残る分です。',
  zuordnungszahlen:
    '{gebraucht}枚中{vorhanden}枚所有 · 空き{frei} · 割り当て済み{zugeordnet} · 確保済み{reserviert}',
  zuordnungPlatz: '区分',
  zuordnungExemplar: '現物',
  zuordnungExemplarWaehlen: '現物を選択',
  zuordnungMenge: '数量',
  zuordnen: '割り当てる',
  zuordnungLoesen: '解除',
  abweichend: '別の版',
  abweichendErklaerung:
    'この現物は、枠とは別の版です。意図的に選ばれています。同じ名前でも、同じカードとは限りません。',
  fehlkarten: '不足カードをリストで表示',
  fehlkartenliste: '不足しているカード',
  fehlkartenKeine: '不足しているものはありません。',

  typen: 'タイプ',
  kennzahlen: '主な数値',
  laender: '土地',
  durchschnitt: '平均マナ総量',
  keineZahl: '—',
  quellen: '色の供給源',
  quellenHinweis:
    '5種類の基本土地タイプのいずれかを持つ土地を数えます。これらはルールテキストなしでマナを生み出します（CR 305.6）。テキストで自分の色を示す土地はここには含まれません。そのテキストの解釈は推測になってしまうからです。',

  weniger: '−',
  wenigerVon: '1枚減らす',
  mehr: '+',
  mehrVon: '1枚増やす',
  insSideboard: 'サイドボードへ',
  insHauptdeck: 'メインデッキへ',
  suche: 'カードを検索',
  deckbereich: 'デッキ',
  analyse: '分析',
  analyseLeer:
    '開いているデッキの分析、チェック、割り当てがここに表示されます。',
  stapelTitel: 'コレクションから',
  stapelWartet:
    'コレクションの{name}のうち、まだどのデッキにもないもの：{liste}。',
  stapelLegen: 'すべてデッキに入れる',
  stapelTeilweise:
    '{gesamt}枚のうち{gelegt}枚がデッキに入りましたが、残りは入りませんでした：',
  stapelOhneDruck:
    '「{name}」にはScryfall IDがありません。デッキに入れられるのは、IDのある版だけです。',
  stapelGrund: '「{name}」：{grund}',
  eintauschen: '{anzahl}を交換で入手',
  deckplatz: 'デッキ構築スペース',
  pruefungOhneFormat:
    'フォーマットがないとチェックできません。デッキにフォーマットを選んでください。それまでは、誤りではなく未チェックです。',
  pruefungLaeuft: 'チェック中…',
  postenUnzulaessig: '使用不可',
  postenUnzulaessigErklaerung:
    'このカードは、選択したフォーマットでは使用できません。詳細は「チェック」タブに表示されます。',
  keinDeckOffen:
    '開いているデッキはありません。左から選ぶか、新しく作成してください。',

  formate: 'フォーマット',
  formateLeer:
    'フォーマットはまだありません。フォーマットはここでは組み込みのリストではなくデータで、同梱されるのではなく作成されます。',
  bauart: '構築タイプ',
  bauarten: '確認済みの構築タイプ',
  bauartHinweis:
    '数値はComprehensive Rulesに基づく提案です。あとでそれぞれ変更できます。',
  scryfallname: 'Scryfall名',
  scryfallnameHinweis:
    'Scryfallのlegalitiesにあるキーで、たとえば「standard」です。空の場合はScryfallがこのフォーマットを認識しておらず、すべてのカードが未チェックのままになります。',
  mindestgroesse: '最小',
  hoechstgroesse: '最大',
  sideboardgroesse: 'サイドボード',
  sideboardOffen: 'サイドボードの上限なし',
  hoechstzahlGleich: '同名カード',
  grenzenHinweis:
    '0は上限なしを意味します。サイドボードでは、0はサイドボードなしを意味します。',
  formatNeu: 'フォーマットを作成',
  formatSichern: 'フォーマットを保存',
  formatLoeschen: 'フォーマットを削除',
  formatLoeschenFrage: '{name}を削除しますか？',
  formatLoeschenJa: '削除',
  formatLoeschenHinweis:
    'このフォーマットを使っていたデッキは残りますが、その後は未チェックになります。',
  umschlagTitel: 'エンベロープとして渡す',
  umschlagErklaerung:
    'エンベロープは、デッキを別のコンピュータに運びます — 署名とチェックサム付きで、アカウントは不要です。届いたものは新しいデッキになり、上書きされるものはありません。名前、ゲーム、印刷版のある枠は**一緒に運ばれ**、メモ、由来、自作カードのある枠は**ここに残ります** — その識別子はこのデータの中でしか通用しないためです。',
  umschlagSchnueren: '開いているデッキを保存',
  umschlagAnnehmen: 'エンベロープからデッキを取り込む',
  umschlagOhneDeck: '先にデッキを開いてください。',
  umschlagGeschnuert: '保存しました。',
  umschlagAngenommen: '取り込みました',
  einfuhrTitel: 'デッキリストを読み込む',
  einfuhrMenue: 'デッキリストを読み込む…',
  einfuhrErklaerung:
    'デッキリストを貼り付けてください。1行に1枚、「4 Lightning Bolt (M11) 149」の形式です。見出しで区切ります。「Sideboard」の下がサイドボードです。プレビューでは何も書き込みません。取り込みを実行して初めて、その行が開いているデッキに入り、何も取り除かれません。',
  einfuhrFeld: 'デッキリスト',
  einfuhrVorschau: 'プレビュー',
  einfuhrUebernehmen: 'デッキリストを取り込む',
  einfuhrLeer: '先にリストを貼り付けてください。',
  einfuhrOffen: '未解決',
  einfuhrBilanz: 'メインデッキ / サイドボード',
  einfuhrGelegt: '取り込み済み',
  einfuhrStueck: '枚',
  einfuhrNichtsOffen: '曖昧でない行がないため、入れるものがありません。',
} as const
