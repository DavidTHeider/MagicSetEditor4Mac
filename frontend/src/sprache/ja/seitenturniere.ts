// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'すべてのトーナメント（{anzahl}）',
  titel: 'トーナメント',
  erklaerung:
    'トーナメントは、時間の経過に沿った進行です。ラウンド、ペアリング、結果があります。順位表はそこから計算され、どこにも保存されません。結果を訂正すると、新しい順位表がすぐに表示されます。',

  leer: 'まだトーナメントがありません。上の「トーナメントを作成…」で始めるか、「トーナメントをインポート…」でファイルから取り込めます。',

  neu: 'トーナメントを作成',
  neuOeffnen: 'トーナメントを作成…',
  name: '名前',
  spiel: 'ゲーム',
  notiz: 'メモ',
  einzelspiel: 'マッチは1ゲームのみ',
  einzelspielHinweis:
    'この場合、ゲーム勝率と対戦相手のゲーム勝率は意味を持たず、マッチの数値をそのまま写します。それでも計算はされます。',
  anlegen: '作成',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: '新しいトーナメントを4ステップで作成',
  schrittTurnier: 'トーナメント',
  schrittRegeln: 'ルール',
  schrittTeilnehmende: '参加者',
  schrittErsteRunde: '第1ラウンド',
  weiter: '次へ',
  zurueck: '戻る',
  verlassen: 'あとで続ける',
  vorlage: 'テンプレート',
  ohneVorlage: 'テンプレートなし',
  vorlageHinweis:
    'テンプレートには、ゲーム、フォーマット、ルール、メモが含まれます。希望すれば名前も含まれますが、デッキは含まれません。提出されたリストは、1つのトーナメントに属するためです。',
  mitTeilnehmenden: 'プレイヤーの名前を引き継ぐ',
  nameWarum:
    '必須です。一覧で見つけられるようにするためです。同じゲームの2つのトーナメントは、異なる名前にします。',
  spielWarum: '必須です。どのフォーマットとデッキが適合するかが決まります。',
  format: 'フォーマット',
  ohneFormat: 'フォーマットなし',
  formatWarum: '任意です。このゲームのフォーマットから選べます。',
  regelnAusVorlage: 'ルールはテンプレートから取り込まれます。',
  schweizerSystem:
    'ペアリングはスイスドロー方式に従います。同じポイントのプレイヤー同士が対戦し、同じ相手と2回対戦することはありません。アプリは、プレイヤー数に応じたラウンド数を提案します。',
  zuWenige:
    'ラウンドには対戦する2人が必要です。少なくとも2人を登録してください。',
  ersteRundeHinweis:
    '第1ラウンドは、このようになります。ボタンを押すまでは、ペアリングは行われません。',
  ersteRundeAuslosen: '第1ラウンドのペアリングを行う',

  waehlen: '開く',
  loeschen: 'トーナメントを削除',
  handlungenAn: '操作 — {name}',
  handlungenKurz: '…',
  turniermenue: 'トーナメント',
  loeschenFrage:
    'このトーナメントを削除しますか？ ラウンド、ペアリング、結果も一緒に削除されます。プレイヤーのデッキは残ります。',
  loeschenJa: '削除',
  loeschenNein: 'キャンセル',

  teilnehmer: '参加者',
  teilnehmerLeer: 'まだ誰も登録していません。',
  anmelden: '登録',
  gebiete: 'このトーナメントの項目',
  gebietTeilnehmer: '参加者',
  gebietRunden: 'ラウンド',
  gebietRangliste: '順位表',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: '終了',
  abgeschlossenAm:
    '{datum}に終了しました。プレイヤー、ラウンド、結果は確定しています。',
  wiedereroeffnen: '再開',
  nichtWiedereroeffnen: '再開しない',
  wiedereroeffnenFrage: 'このトーナメントを再開しますか？',
  wiedereroeffnenErklaerung:
    'そのあと、プレイヤー、ラウンド、結果を再び変更できます。再開したことは記録に残ります。',
  gebietProtokoll: '記録',
  protokollLeer: 'まだ何も起きていません。',
  protokollAngelegt: '「{angaben}」として作成',
  protokollAngemeldet: '{angaben}が登録',
  protokollAusgestiegen: '{angaben}がドロップ',
  protokollEingestiegen: '{angaben}が再登録',
  protokollRunde: 'ラウンド{angaben}のペアリング',
  protokollErgebnis: '結果 {paarung}: {angaben}',
  protokollZurueckgenommen: '結果 {paarung} を取り消し',
  protokollAbgeschlossen: '終了',
  protokollWiedereroeffnet: '再開',
  protokollEingelesen: '「{angaben}」としてインポート',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: '書き出す…',
  berichtSichern: 'レポートを保存…',
  einlesen: 'トーナメントをインポート…',
  gesichertUnter: '{pfad} に保存しました。',
  eingelesen: 'インポートして開きました。',
  einlesebefunde: 'インポート中に気づいた点:',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'スイスドロー · 第{runde}ラウンド / 全{von}ラウンド · {anzahl}人',
  kopfzeileOhneRunde: 'スイスドロー · {anzahl}人',
  kopfzeileOhneEmpfehlung: 'スイスドロー · 第{runde}ラウンド · {anzahl}人',
  gebietDecklisten: 'デッキリスト',
  decklistenLeer:
    'まだ誰もリストを提出していません。登録時にデッキを選ぶと、そのリストが提出されます。',
  deckAufschlagen: 'デッキを開く',
  partieOeffnen: 'ゲームを開く',
  deck: 'デッキ',
  ohneDeck: 'デッキなし',
  deckHinweis:
    'デッキを指定すると、そのリストは参照ではなくコピーされます。トーナメントのデッキリストは、提出された書類だからです。',
  deckabzug: '提出されたリスト',
  deckabzugLeer: 'デッキは提出されていません。',
  deckabzugUnbenannt: 'カード',
  aussteigen: 'ドロップ',
  einsteigen: '復帰',
  ausgestiegen: 'ドロップ済み',
  ausgestiegenHinweis:
    'これ以降ペアリングされませんが、順位表には残ります。その勝率は、対戦相手の勝率にも引き続き算入されます。',

  runden: 'ラウンド',
  rundenLeer: 'まだペアリングされたラウンドがありません。',
  runde: '第{nummer}ラウンド',
  empfehlung:
    '{gelegt}/{empfohlen}ラウンド — トーナメントルールでは、{spieler}人のプレイヤーに対して{empfohlen}ラウンドを推奨しています。あくまで提案であり、規則ではありません。',
  vorschlag: '次のラウンドをプレビュー',
  vorschlagTitel: '次のラウンドは、このようになります',
  vorschlagHinweis:
    'プレビューであり、作成はされていません。下のボタンを押した場合にのみ作成されます。',
  rundeLegen: 'ラウンドを作成',
  bye: 'BYE (2–0)',
  gegen: 'vs',
  offen: 'まだ対戦していません',
  spiele: 'ゲーム',
  spieleA: '1人目のゲーム数',
  spieleB: '2人目のゲーム数',
  unentschiedenSpiele: '引き分け',
  eintragen: '記録',
  zuruecknehmen: '取り消す',
  ergebnis: '結果',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'この訂正により、順位表が変わります。',
  folgenPlatz:
    '{name}: {vorher}位 → {nachher}位、{punkteVorher} → {punkteNachher}ポイント',
  folgenSpaeter:
    '第{runden}ラウンドは古い順位表でペアリングされており、そのまま変更されません。',
  folgenEintragen: 'それでも記録',
  folgenNicht: '記録しない',
  ergebnisZahlen: 'ゲーム数は、0から始まる整数で数えます。',
  ergebnisOhneSpiel:
    'ゲームが1つもない結果は、結果ではありません。まだ対戦していない人には、まだ結果がありません。',
  ergebnisZuViele: 'ここでは、1マッチは最大{hoechstens}ゲームまでです。',

  rangliste: '順位表',
  ranglisteLeer: 'まだ対戦が行われていません。',
  platz: '順位',
  bilanz: '戦績',
  matchpunkte: 'ポイント',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Magic トーナメントルールの順序による、4つのタイブレーカーです。ポイント、対戦相手のマッチ勝率、ゲーム勝率、対戦相手のゲーム勝率の順です。各勝率には、0.33の下限があります。',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'ここで対戦',
  matchErklaerung:
    'このペアリングをこのプログラムで対戦する場合は、ゲームに紐付けてください。結果は両側から報告され、両側が同じ内容を報告した場合にのみ記録されます。',
  matchNichtGebunden: 'このペアリングには、ゲームが紐付けられていません。',
  matchAnsehen: '確認',
  matchPartie: 'ゲームID',
  matchSitzA: '1人目のプレイヤーの席',
  matchSitzB: '2人目のプレイヤーの席',
  matchBinden: 'このゲームに紐付ける',
  matchGebunden: '紐付け先のゲーム',
  matchSitz: '席',
  matchDeckabdruck: 'デッキのフィンガープリント',
  matchKeineMeldung: 'まだ誰も報告していません。',
  matchMeldungVon: '報告者:',
  matchMelden: 'この席として報告',
  matchAusDerPartie: 'ゲームから結果を報告',
  matchAusDerPartieHilfe:
    'スコアと証拠は、終了したゲームから取得されます。誰が報告するかを指定するだけです。',
  matchNachweis: '証拠 (最後のチェーンリンク)',
  matchUebernehmen: '結果を記録',
  matchUebernommen: '結果がトーナメントに記録されました。',
  matchGrenze:
    '証拠が示すのは、両側が同じ履歴を見たということです。ゲームが正しく行われたことを示すものではなく、非公式のラウンドは公式なトーナメント認定にはなりません。',
}
