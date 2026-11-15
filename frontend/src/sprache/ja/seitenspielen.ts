// The texts of the **Play** area — work item G1 of phase 5.5
// (#1458).
import { abzugworte } from './spielenabzug'
import { zugworte } from './spielenzug'
import { menueworte } from './spielenmenue'
import { vorhangworte } from './spielenvorhang'
import { partienwoerter } from './spielenpartien'
import { regelwoerter } from './spielenregeln'
import { anfangworte } from './spielenanfang'
import { beginnworte } from './spielenbeginn'
import { gastworte } from './spielengast'
import { netzworte } from './spielennetz'
import { leihewoerter } from './spielenleihe'

export const spielen = {
  ...abzugworte,
  ...zugworte,
  ...menueworte,
  ...vorhangworte,
  ...partienwoerter,
  ...regelwoerter,
  ...anfangworte,
  ...beginnworte,
  ...gastworte,
  ...netzworte,
  ...leihewoerter,

  titel: 'プレイ',
  erklaerung:
    'ゲームは、デッキが固定されたテーブルです。ここで起きることはイベントの流れで、それぞれが直前のイベントにつながり、上書きされるものはありません。カードを動かすと、他の人にそのカードがどう見えているかがすぐにわかります。',
  anfangTitel: 'ゲームを開始',
  anfangErklaerung:
    'どの席もデッキを持ち寄り、そのデッキは固定されます。あとでデッキを変更しても、このゲームは変わりません。ゲームは、最初に選択したデッキから決まります。',
  leerDecksTitel: 'デッキがまだありません',
  leerDecks: 'プレイにはデッキが必要です。「デッキ」で作成してください。',
  sitz: '席',
  sitzName: 'テーブルでの名前',
  sitzDeck: 'デッキ',
  sitzLeihe: '…{von} から借用 · {karten} 枚',
  deckWaehlen: 'デッキを選択',
  anschlussWartet:
    '「{name}」が一緒に来て、最初の席に座ります。2番目の席は空いています。',
  sitzHinzu: '席を追加',
  sitzWeg: 'この席を削除',
  sitzGruppe: '席 {nummer}',
  spiel: 'ゲーム',
  anfangen: 'ゲームを開始',
  tischTitel: 'テーブルで',
  partieGegen: '{namen} との対戦',
  partieZwischen: '{namen} の対戦',
  partieVon: '一人でのゲーム：{namen}',
  fassung: 'リビジョン',
  wirt: 'ホスト',
  wirtHinweis:
    'ホストはこのゲームの完全な状態を保持しているため、自分の視点で見えるより多くのことを見られます。参加した後ではなく、参加する前に知っておくべきことなので、ここに記載しています。',
  tischAufHaufen: '山に {zahl} 枚',
  tischMarkeEine: '{art}カウンター {zahl} 個',
  tischMarkeViele: '{art}カウンター {zahl} 個',
  tischArt: {
    spielstein: 'トークン',
    emblem: '紋章',
    faehigkeit: 'スタック上の能力',
  },

  sicht: '視点',
  sichtZuschauer: '観戦者（席なし）',
  sichtWechseln: '席を切り替える',
  sichtLaedt: '選択した視点を読み込んでいます…',
  verlassen: 'ゲームを閉じる',
  zonen: '何がどこにあるか',
  zoneBibliothek: 'ライブラリー',
  zoneHand: '手札',
  zoneTisch: '戦場',
  zoneFriedhof: '墓地',
  zoneExil: '追放領域',
  zoneStapel: 'スタック',
  zoneKommando: '統率領域',
  zonekurzBibliothek: 'ライブラリ',
  zonekurzFriedhof: '墓地',
  zonekurzExil: '追放',
  zonekurzKommando: '統率',
  zoneLeer: '空',

  unbekannterTyp: '不明なタイプ：{woerter}',
  geheim: '裏向き',
  getappt: 'タップ状態',
  befehle: '操作',
  regelnTitel: '適用されるもの',
  regelwerk: 'ルールセット',
  wirksamAb: '有効',
  hoechsteStrenge: '最高の厳格さ',
  strenge: '厳格さ',
  andereSeite: '反対の面',
  grossAnsehen: '拡大して表示',
  ersteSeite: '最初の面',
  stelle: '置く位置',
  stelleOben: '一番上',
  stelleUnten: '一番下',
  stelleZufaellig: '切り直して混ぜる',
  mischpflichtTitel: '切り直しが必要なもの',

  legendeTitel: '未決の選択：伝説のルール',
  legendeSatz:
    '{wer} は、「{name}」という名前の伝説のパーマネントのうち、残すものを選ぶ必要があります（CR 704.5j）。',
  legendeBehalten: 'これを残す',

  reihungTitel: '未決の選択：能力の順番',
  reihungSatz:
    '{wer} は、同時に誘発した複数の能力をコントロールしており、それらをスタックに置く順番を決めます（CR 603.3b）。最初に置いたものが一番下になり、最後に解決されます。',
  reihungLegen: '次に置く：',
  stand: '状態',
  standLaeuft: '進行中',
  standPausiert: '一時停止中',
  standBeendet: '終了',
  pausieren: '一時停止',
  fortsetzen: '再開',
  beenden: '終了',
  beendenFrage:
    'このゲームを終了しますか？終了は取り消せません。終了したゲームは再開できません。',
  beendenJa: '終了',
  beendenNein: 'キャンセル',
  ereignisse: '今起きたこと',
  ereignisseLeer: 'まだ何もありません。',
  ereignisseLuecke:
    'ここには一部の手が表示されていない可能性があります。ページが非表示の間に、取り戻せる以上の通知が届きました。',
  ereignisPrivat: 'あなただけに表示',
  ereignisSatz: '{sitz}：{handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}：{handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'これまでの経過',
  verlaufAnsehen: '履歴を表示',
  verlaufNachladen: '再読み込み',
  verlaufLaeuft: '取得しています…',
  verlaufUngeholt:
    '履歴はまだ取得されていません。この席から見える範囲で、このゲームのすべての動きが表示されます。',
  verlaufLeer: 'このゲームでは、まだ何も起きていません。',
  verlaufGrenze:
    '他の席がその席だけに見えた内容は、ここにはありません。後からでも同じです。',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'このゲームの記録',
  protokollErklaerung:
    'このゲームのすべての動きは、このコンピュータに保存され、各席は自分の分だけを見られます。自分の分は、後で見られるように公開できます。相手1人だけにも、全員にも公開できます。',
  protokollAnsehen: '公開を表示',
  protokollNiemand: '誰も何も公開していません。',
  protokollFuer: '公開先',
  protokollAlle: '全員',
  protokollSitz: '席',
  protokollEmpfaenger: '公開先（カンマ区切り。全員の場合は *）',
  protokollFreigeben: '公開',
  protokollZuruecknehmen: '取り消す',
  protokollNurSpaeter:
    '公開が有効になるのは、ゲームの後だけです。テーブルでは、引き続き各自が自分の分しか見えません。',
  protokollKeineRueckholung:
    '取り消すと以後は見られなくなりますが、すでに誰かが読んだ内容を取り戻すことはできません。',
  protokollAufbewahrung:
    '終了した記録は、90日後にデータの整理で確認の対象として表示されます。あなたが指示したときにだけ削除されます。',
  protokollLoeschen: '記録を削除',
  protokollLoeschenErklaerung:
    'このゲームの状態、履歴、公開設定がこのコンピュータから削除されます。ほかの場所にあるコピーについては、何も影響しません。',
  protokollGeloescht: '記録を削除しました。',
  protokollLoeschenFrage: 'このゲームの記録を削除しますか？',
  protokollLoeschenJa: '記録を削除',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: '接続の状態',
  diagnoseErklaerung:
    'このテーブルが何を通じて通信し、どれだけ待機しているかを表示します。あなたのカードについても、他の人のカードについても、何も表示されません。',
  diagnoseAnsehen: '接続を確認',
  diagnoseStill: '動作中のテーブルはありません。',
  diagnoseAlsWirt: 'このコンピュータがテーブルをホストしています。',
  diagnoseAlsGast: 'このコンピュータは、他の人のテーブルに座っています。',
  diagnosePausiert: '接続が切れました。席はそのまま残ります。',
  diagnoseWeg: '経路',
  diagnoseGegenueber: 'ピア',
  diagnoseSeit: '接続時間',
  diagnoseWartend: '待機中',
  diagnoseVerworfen: '破棄',
  diagnoseStand: 'シーケンス',
  diagnoseRunde: '直近の往復時間',
  diagnoseBildzeit: 'このウインドウのフレーム間隔',
  diagnoseUnbekannt: '不明',
  diagnoseOhneLeitung: '接続はありません。',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: '他の人がこのテーブルでプレイできるようにする',
  freigabeErklaerung:
    '接続できる状態にすることと、他の人にプレイさせることは、別の判断です。テーブルを開くと、ピアがこのゲームで操作できるようになります。席を割り当てるのは相手ではなく、あなたです。',
  freigabeOffen: 'このテーブルは開いています。',
  freigabeZu: 'このテーブルは開いていません。',
  freigabeStarten: 'テーブルを開く',
  freigabeBeenden: 'テーブルを閉じる',
  freigabeNachsehen: '再確認',
  freigabeSitze: '空席',
  freigabeNiemand: 'まだ誰も座っていません。',
  freigabePartner: 'この人：',
  freigabePartnerFremd: 'このピアは、あなたの連絡先にいません。',
  freigabeBesetzt: '着席中：',
  freigabeOhneHorchposten:
    'ピアが接続してくるには、このコンピュータにも接続できる状態が必要です。これは「設定」で行います。',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'テーブル',
  tischListe: 'リスト表示',
  tischGrafisch: 'テーブル表示',
  tischDarstellung: '表示',
  tischEinpassen: 'すべて表示',
  tischNaeher: '拡大',
  tischWeiter: '縮小',
  tischDrehen: 'テーブルを回転',
  tischAuswahl: '選択へ移動',
  tischEigenerSitz: '自分の席へ移動',
  tischBlick: 'テーブルの表示',
  tischZiel: '移動先',
  tischOhneWebgl:
    'このコンピュータでは、ハードウェアアクセラレーションされたグラフィックスを利用できません。テーブルは下にリストとして表示され、描画されたテーブルでできることはすべて、そこでもできます。',
  tischOhneRenderer:
    'テーブルを描画できませんでした。下のリストにはすべてのカードがあり、それ以外の操作もそこでできます。',
  tischVerloren:
    'グラフィックスとの接続が失われました。ウインドウを開き直すと、テーブルが戻ります。それまでは、下のリストにすべてのカードがあります。',
  tischHeimatlos:
    'これらのオブジェクトは、このプログラムが知らない領域にあるため、テーブルには表示されません：',
  tischVerdeckt: '裏向き',
  tischGetappt: 'タップ状態',
  tischHaengtAn: 'つけられている先：',
  tischbild: '背景',
  tischbildHolz: '木目',
  tischbildDunklesHolz: '濃い木目',
  tischbildPicknick: 'ピクニック',
  tischbildDungeon: 'ダンジョン',
  tischbildTaverne: '酒場',
  tischbildTurniertisch: 'トーナメントテーブル',
  tischEigenes: '自分の画像…',
  tischEigenesWeg: '自分の画像を削除',
  gespraech: 'テーブルでの会話',
  gespraechVerlauf: 'テーブルで話された内容',
  gespraechEingabe: 'メッセージを入力',
  gespraechSenden: '発言',
  gespraechLeer: 'まだ誰も何も話していません。',
  gespraechNiemandZeigt: '今、何かを指している人はいません。',
  gespraechZeigt: '{sitz}：{tut}',
  gespraechZeigtAuf: '{sitz}：{objekt}を{tut}',
  zeigerZeiger: '指している',
  zeigerGreift: 'ドラッグしている',
  zeigerSchwebt: '見ている',
  zeigerPing: '呼びかけている',
  spielart: 'モード',
  spielartGoldfish: '一人で練習',
  spielartSelfplay: '複数の席',
  format: 'フォーマット',
  fehltDeck:
    'すべての席にデッキを選択してください。最初のデッキがゲームを決めます。',
  fehltEigenesDeck:
    '少なくとも1席に自分のデッキが必要です。貸し出しにはゲームが含まれず、ゲームは最初の自分のデッキから決まります。',
  fehltName:
    'すべての席に名前が必要です。テーブルでプレイヤーを区別するために使われます。',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'デッキの貸し借り',
  leiheOeffnen: 'デッキの貸し借り …',
  leiheVerleihenTitel: 'デッキを貸す',
  leiheAnnehmenTitel: '貸し出しの受け入れ',
  leiheListeTitel: 'あなたの貸し借り',
  leiheErklaerung:
    '貸し出すと、相手はあなたのデッキを固定したスナップショットでプレイできます。デッキはここに残り、コレクションのカードは1枚も変わりません。',
  leiheKeinEigentum:
    'これは使用許可であり、譲渡ではありません。デジタルでも実物でも、持ち主は何も変わりません。',
  leiheKeinKopierschutz:
    'また、コピー防止でもありません。一度カードを見た人は、そのカードを見たことになります。貸し出しで決まるのは、このプログラムがどのデッキでゲームを開始するかです。',
  leiheAnsehen: '貸し出しを表示',
  leiheKeine: '貸し出しはありません。',
  leiheEntleiher: '誰に貸しますか？',
  leiheEntleiherLeer:
    '確認済みのピアはまだいません。下の「相手の追加と確認」のボックスで追加してください。',
  leiheDeck: 'どのデッキか',
  leiheDeckWaehlen: 'デッキを選択…',
  leiheKeinDeck: 'デッキがまだありません。「デッキ」で作成します。',
  leiheFachleute: '特定の1ゲームだけに貸し出す',
  leiheFachleuteErklaerung:
    '指定しない場合、貸し出しは相手の次のゲームに適用されます。ちょうど1ゲームに限定するには、そのゲームの ID を入力します。相手は自分のゲーム一覧で確認できます。',
  leiheWoran: 'ゲーム ID（空欄の場合は次のゲーム）',
  leiheStunden: '有効時間（時間）',
  leiheZiel: 'エンベロープの保存先',
  leiheAnbieten: '貸し出して保存',
  leiheAbgelegt: 'エンベロープはここにあります：',
  leiheUmschlag: '貸し出しのエンベロープ',
  leiheAnnehmen: '貸し出しを受け入れる',
  leiheWiderrufen: '取り消す',
  leiheZurueckgeben: '返す',
  leiheVerliehen: '貸し出し先：',
  leiheGeliehen: '借りている相手：',
  leiheBis: '有効期限：',
  leiheKarten: '枚',
  leiheVerbraucht: '使用済み：',
  leiheOffen: '未使用',
  leiheZurueckgenommen: '取り消し済み',
  leiheModusZuende: '進行中のゲームは最後まで続けられます。',
  leiheModusAbbruch: '進行中のゲームは中止されます。',
  leiheRuecknahmeGrenze:
    '取り消すと、新たに開始することができなくなります。進行中のゲームからは何も取り除かれず、相手側のコピーも削除されません。',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: '結果',
  ausgangSieger: '勝者：',
  ausgangUnentschieden:
    '引き分けです。ゲームに残っていた全員が同時に敗北しました。',
  grundAufgabe: '投了しました',
  standLeiste: 'プレイヤーの状況',
  standDu: 'あなた',
  standLeben: 'ライフ',
  standGift: '毒',
  standMana: 'マナ',
  standStrahlung: '放射線',
  standSteuer: '統率者税',
  standCommanderschaden: '{name}：{schaden}',
  standCommanderTod: '21点でゲーム終了',
  standLeergezogen: '空のライブラリーからカードを引いた',
  grundLeben: 'ライフが0以下になりました',
  grundBibliothek: '空のライブラリーからカードを引くことになりました',
  grundGift: '毒カウンターが10個になりました',
  aufgeben: '投了',
  aufgebenFrage:
    '投了しますか？この席のゲームはすぐに終わり、そのカードも一緒に退きます。',
  aufgebenJa: '投了する',
  aufgebenNein: '続ける',
  aufgebenHinweis:
    'いつでも投了でき、すぐに有効になります。ゲームから退出し、あなたのカードも一緒にゲームから取り除かれます。',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'プロファイル',
  sitzprofilKeines: 'プロファイルなし',
  sitzprofilErklaerung:
    'プロファイルを2つ持っている場合は、席ごとに1つずつ割り当てられます。ゲームは、現在のプロファイルに対応する席の視点で開始します。ほかの視点には、テーブルの視点選択から移れます。もう一方の視点で見えていた内容は失われます。どの席にもプロファイルを割り当てない場合は、現在のプロファイルが最初の席に着きます。',
  sitzprofilBefunde: 'これについて、補足があります：',
}
