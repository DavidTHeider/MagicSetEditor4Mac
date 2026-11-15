// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'コレクション',
  erklaerung:
    '実際に持っているカードです。版、状態、言語、フィニッシュごとに1行になります。あなたのセットとは関係ありません。デザインしたカードは下書きで、所持カードは実物の厚紙です。',

  aufnehmenOeffnen: '追加…',
  einfuhrOeffnen: '読み込みと書き出し…',
  einfuhrTitel: '読み込みと書き出し',
  leerTitel: 'まだ何も記録されていません',
  leer: '所持カードには、その版の Scryfall ID が必要です。カードページのアドレスにある UUID です。リスト全体は一度に読み込めます。',

  druck: '版',
  menge: '数量',
  zustand: '状態',
  sprache: '言語',
  ausfuehrung: 'フィニッシュ',
  foil: 'フォイル',
  nonfoil: '非フォイル',
  etched: 'エッチド',
  erwerbsdatum: '入手日',
  erwerbspreis: '購入価格',
  erwerbswaehrung: '通貨',
  preisMehrdeutig:
    '2通りに解釈できます：{erste} または {zweite}。あいまいにならないように書いてください。',
  notiz: 'メモ',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'カード名',
  setcode: 'セット',
  nummer: '番号',
  sicht: 'コレクションの表示',
  suche: '検索',
  suchePlatzhalter: '名前、セット、ID、またはメモ',
  sortierung: '並べ替え',
  sortierBezeichnung: 'カード',
  sortierMenge: '数量',
  sortierErwerbsdatum: '入手日',
  ausfuehrungAlle: 'すべてのフィニッシュ',
  nurFoil: 'フォイルのみ',
  nurNonfoil: '非フォイルのみ',
  nurEtched: 'エッチドのみ',

  navigator: '表示とツール',
  inspector: '所持カード',
  exemplarmenue: '所持カード',
  inDecks: '使われているデッキ',

  inspectorLeer:
    'リストから所持カードを選択すると、数量、状態、購入情報、価格がここに表示され、ここで変更できます。',
  inspectorAusserhalb:
    'この所持カードは、現在の表示にはもうありません。変更は、それでもまさにこの所持カードに反映されます。',
  arbeitsplatz: 'コレクションのワークスペース',
  listeBezeichnung: '所持カード',
  spalteWahl: '選択',
  spalteBezeichnung: '名前',
  spalteMenge: '数量',
  spalteZustand: '状態',
  spalteSprache: '言語',
  spalteSet: 'セット',
  spalteLagerort: '保管場所',
  gruppeOhne: '未指定',
  gruppeSumme: '{anzahl} 件 · {menge} 枚',
  kachelWaehlen: '{bezeichnung} を選択',

  ansicht: '表示',
  ansichtTabelle: 'テーブル',
  ansichtGalerie: 'ギャラリー',
  ansichtGruppen: 'グループ',
  gruppierung: 'グループ化',
  gruppierungKeine: 'グループ化しない',
  gruppierungSet: 'セット',
  gruppierungLagerort: '保管場所',
  gruppierungZustand: '状態',
  gruppierungSprache: '言語',
  ohneTreffer: '検索とフィルタに一致する所持カードはありません。',
  ohneTrefferZuruecksetzen: '検索とフィルタをリセット',
  gezeigt: '{gesamt} 件中 {gezeigt} 件',
  gekappt: '{passend} 件中 {gezeichnet} 件を描画',
  mehrZeigen: 'さらに表示',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'リストを貼り付け',
  listeErklaerung:
    '1行に1枚のカードを入力します。たとえば「4 Lightning Bolt (M11) 149」のようにします。空行、「Deck」、「Sideboard」、および # または // で始まる行は無視されます。',
  listeFeld: 'カードリスト',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: '照合',
  listeEinlesen: '追加',
  listeOffen:
    '{gesamt} 行のうち {offen} 行は判断が必要です。持っている版を選択してください。',
  listeAlleEindeutig: '{gesamt} 行すべてが一意に決まりました。',
  listeDruckWahl: 'どの版ですか？',
  listeOhneWahl: '— 未選択 —',
  listeDrucke: '{drucke} 種類の版',
  listeBericht:
    '新規 {angelegt} 件、増加 {erhoeht} 件、追加したカード {stueck} 枚。',
  listeNichtsGewaehlt:
    '判断した行がありません。少なくとも1つの版を選択してください。',
  listeAusgelassen: '既存の {ausgelassen} 行をスキップしました。',
  listeKeinDruck: 'この名前のカードはありません。綴りを確認してください。',
  listeZuViele:
    'この名前の印刷は {drucke} 種類以上あります。「(M11)」のようにセットを括弧で指定すると、一つに絞れます。',
  listeOhneKarte: '{anzahl} 行はカードが見つからないため、スキップされます。',
  listeOhneKarteEins: '1 行はカードが見つからないため、スキップされます。',
  vorschauTitel: '取り込む前に',
  vorschauErklaerung:
    '一部の行は、コレクションにすでにある行と一致します（版、状態、言語、フィニッシュ、用途が同じ）。それらをどうするかを選択してください。それまでは何も書き込まれません。',
  vorschauZaehlung:
    '新規 {neu} 件 · 既存の行と一致 {zusammen} 件 · この取り込み内で重複 {wiederholt} 件',
  vorschauZusammen: '{bezeichnung}、{zustand}：{vorher} → {nachher}',
  vorschauModus: '既存の行',
  vorschauAddieren: '数量を加算',
  vorschauAuslassen: '既存の行をスキップ',
  vorschauEinlesen: '取り込む',
  vorschauAbbrechen: '取り込まない',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: '書き出しとバックアップ',
  ausfuhrErklaerung:
    '「CSV」はすべてのフィールドを含み、完全に読み戻せます。「リスト」はデッキビルダーに貼り付けられる形式で、数量、カード、版、フィニッシュだけを含みます。',
  ausfuhrAlsCSV: 'CSV として',
  ausfuhrAlsListe: 'リストとして',
  ausfuhrErgebnis: '出力',
  ausfuhrZurueck: 'CSV を読み戻す',
  ausfuhrLeer: 'まだ何も書き出されていません。',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: '保管場所',
  wunsch: '欲しい・未所持',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: '整理',
  ordnungLagerorte: '保管場所',
  ordnungDubletten: '重複',
  ordnungWuensche: 'ウィッシュリスト',
  ordnungBehalten: '残す',
  ordnungLeerLagerorte:
    'まだ所持品がありません。所持品がなければ保管場所もありません。',
  ordnungLeerDubletten: '{behalten}枚を超えて持っている版はありません。',
  ordnungLeerWuensche:
    'ウィッシュリストは空です。所持品を「欲しい・未所持」にするか、デッキの「足りないカードをウィッシュリストへ」で追加できます。',
  ordnungStueck: '{zeilen} 行に {stueck} 枚',
  ordnungOhneOrt: '未指定',
  ordnungUeberzaehlig: '{ueberzaehlig} 枚多い',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'このカードを {stueck} 枚持っています。',
  besitzNein: 'この版は持っていません。',
  besitzGewuenscht: '（ウィッシュリストに {gewuenscht} 枚）',

  // **Die Massenvorgänge** (#1292).
  masseTitel: '一括操作',
  masseWaehlen: '選択',
  masseGewaehlt: '{gesamt} 件中 {gewaehlt} 件を選択',
  masseLagerortSetzen: '保管場所を次に設定',
  masseZustandSetzen: '状態を次に設定',
  masseAendern: '選択した項目を変更',
  masseLoeschen: '選択した項目を削除',
  masseLoeschenFrage: '{anzahl} 件の所持カードを削除しますか？',
  masseLoeschenErklaerung:
    '削除したものは、もう一度入力し直す必要があります。数量、状態、保管場所、購入価格は他のどこにもありません。コレクションには、セットのような履歴がありません。',
  masseLoeschenJa: '削除',
  masseBericht: '{geaendert} 件を変更しました。',
  masseZuruecknehmen: '取り消す',
  masseZuruecknehmenVon: '{anzahl} 行の保管場所の変更を取り消す',
  masseZurueckgenommen: '{anzahl} 件を取り消しました。',
  massePreise: 'すべての価格を取得',
  massePreisschub: '{geholt} 件取得、残り {offen} 件',
  massePreiseFehlgeschlagen: '{anzahl} 件失敗',
  massePreiseSteht:
    'この処理では何も進みませんでした。もう一度試しても、同じ結果になります。',
  massePreiseFertig: 'すべての価格を取得しました。',

  aufnehmen: '追加',
  aenderungSichern: '保存',
  loeschen: '削除',
  loeschenFrage: 'この所持カードを削除しますか？',
  loeschenErklaerung:
    '数量、状態、保管場所、購入価格は他のどこにもありません。コレクションには、セットのような履歴がありません。',
  loeschenJa: '削除',
  abbrechen: 'キャンセル',
  bearbeiten: '編集',
  zumTausch: 'トレードへ',
  insDeck: 'デッキへ',
  auswahlTitel: '選択',
  auswahlInsDeck: '{anzahl} 件 → デッキ',

  neuesTitel: '所持カードを追加',
  weitereAngaben: 'その他の項目',

  dubletteFrage:
    '{bezeichnung} はすでにコレクションにあります。同じ状態、言語、フィニッシュで {menge} 枚です。',
  dubletteErklaerung:
    '「増やす」を選ぶと、新しい数量が加算されます。既存の行のその他の情報はそのままです。キャンセルしても何も変わらず、入力した内容は残ります。',
  dubletteErhoehen: '{summe} 枚に増やす',
  dubletteZurZeile: '既存の行へ移動',
  dubletteAbbrechen: 'キャンセル',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  preiseTitel: '価格',
  preiseHolen: '価格を取得',
  marktUndVerlauf: '出品と価格推移',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'マーケットプレイスのオファー',
  marktErklaerung:
    'この版について、個々の出品者が提示している価格、数量、状態、言語です。確認はネットワーク経由で行われます。何かを購入することはありません。',
  marktHolen: 'オファーを取得',
  marktHolenBei: '{quelle} からオファーを取得',
  marktLaeuft: '取得しています…',
  marktLeer: '現在、この版の出品はありません。',
  marktZugeordnet:
    'この版はカードのバージョン {fassung} に関連付けられました。次回の確認では、すぐに見つかります。',
  marktStueck: '{menge} 枚あり',
  marktBuendel: '{buendel} 枚単位',
  marktImUrlaub: '現在は発送していません',
  marktGradiert: '鑑定済み',
  marktSigniert: 'サイン入り',
  marktVeraendert: 'アルタード',
  marktBefunde: 'すべてのオファーを読み取れたわけではありません：',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde:
    'すべての価格を読み取れたわけではありません。次の系列がありません：',
  preiseLaufen: '取得しています…',
  keinePreise: 'この版の価格は記録されていません。',
  preisAlter: '{stunden} 時間前 · {quelle}',
  preisFrisch: 'たった今 · {quelle}',
  preisVeraltet: '古い情報（{tage} 日前） · {quelle}',
  preisQuellstand: '提供元の日付 {tag}',

  vorbehalt:
    '価格は Scryfall とその提携先による推定値です。いずれの価格も保証されません。基準となるのは、店舗での価格です。',

  verlauf: '履歴',
  keinVerlauf:
    'この系列にはスナップショットが1つしかありません。価格を取得し直すたびに、履歴が積み上がります。',

  bulkTitel: 'ソース全体から価格を取得',
  leiste: {
    bezeichnung: 'コレクションの数値',
    exemplare: '所持カード',
    eintraege: '{zahl} 件',
    wert: '価値',
    keinPreis: '価格なし',
    ohnePreis: '価格未取得',
    nichtGelesen: '未読込',
  },
  drucksuche: {
    karte: 'カード',
    hinweis: '名前（セットと番号も可）— 例: Lightning Bolt (M11) 149',
    suchen: '検索',
    drucke: '版',
    keinTreffer: 'この名前のカードは見つかりません。',
    weitere: '他にも版があります — セットと番号で絞り込めます。',
    treffer: '{name} · {set} ({code}) · No. {nummer}',
  },
  bulkErklaerung:
    'MTGJSON はカードごとではなく、ファイルで応答します。「今日」は当日分、「90日間」は履歴を取得します。後者は 1 GB を超えます。',
  bulkQuelle: 'ソース',
  bulkUmfang: '範囲',
  bulkHeute: '今日',
  bulkNeunzigTage: '90日間',
  bulkHolen: '取得',
  bulkHolenBei: '{quelle} から価格を取得',
  bulkLaeuft: '取得しています…',
  bulkBilanz: '{drucke} 種類の版について、{staende} 件の価格を保存しました。',
  bulkOhneDrucke:
    'このコレクションには MTGJSON の識別子を持つ版がありません。何も要求しませんでした。',
  uebersichtTitel: '概要',
  uebersichtLeer: '価格はまだ取得されていません。',
  ohnePreis: '価格なし：{zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} 枚',
  exemplare: '{zahl} 枚',

  keineUmrechnung:
    '通貨ごとに1つの合計です。換算は行いません。換算には為替レートが必要ですが、このプログラムにはありません。',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: '別のソースからの価格履歴',
  verlaufErklaerung:
    '時間の経過に沿った系列を取得して保存します。確認はネットワーク経由で行われ、利用枠を消費します。状態は、ソースでの呼び方のまま表示されます。',
  verlaufHolen: '履歴を取得',
  verlaufHolenBei: '{quelle} から履歴を取得',
  verlaufLaeuft: '取得しています…',
  verlaufBilanz:
    '{staende} 件の価格を保存しました。カードのバージョン：{fassung}。',
} as const
