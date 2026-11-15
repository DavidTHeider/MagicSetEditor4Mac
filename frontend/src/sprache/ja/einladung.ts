// **Invite, compare, allow** — NET-01
// (#1705).
// Maschinell übersetzt.
export const einladung = {
  titel: '相手を招待',
  erklaerung:
    '同じネットワークで互いを見つけられないときは、あなたの招待コードを送ってください。相手に届くなら、どちらか一方のコードだけで足ります。そうでなければ互いに交換し、そのあと30桁の数字を照合します。',
  zeigen: '自分のコードを表示',
  eingeben: 'コードを入力',

  name: '他の人にどう表示しますか？',
  nameHinweis:
    '名前だけでは何も証明できません。そのため、このあと数字を照合します。',
  zweck: '何のために招待しますか？',
  erzeugen: 'コードを作成',
  code: 'あなたの招待コード',
  codeHinweis:
    '相手にメッセージなどで送ってください。{zeitpunkt}まで有効で、1回だけ使えます。',
  kopiert: 'コピーしました。',

  eingabe: '相手の招待コード',
  eingabeBeispiel: 'mse4mac-kontakt:1:…',
  pruefen: 'チェック',

  bestaetigenTitel: '{name}を確認',
  vergleichen:
    'この数字を互いに照合してください。声に出して、同じ部屋で、または電話で。一致すれば、確かにお互いに話している相手です。',
  vergleichenHinweis:
    '相手があなたを認識すると、同じ数字が表示されます。あなたのコードによる場合も、あなたから名乗り出た場合も同じです。',
  vorgaenger:
    '新しい鍵が{name}を名乗っています。これまで{name}に許可した内容は引き継がれません。あらためて照合してください。',
  bekannt: 'このコンピュータはすでに確認済みです。',
  gesperrt:
    'このコンピュータはブロック中です。もう一度追加するには、先にブロックを解除してください。',
  angenommen:
    'この招待はすでに承諾済みです。新しく接続するには、新しいコードを依頼してください。',
  eigenerName: '相手を何と呼びますか？',
  eigenerNameHinweis:
    'この名前はあなたの側だけで使われ、外部には送信されません。',
  freigeben: '確認後に許可: {zweck}',
  nichtsGeteilt:
    'あなたが同意するまで、何も共有されません。確認後も、相手に見えるのは、あなたが明示的に共有したものだけです。',

  technik: '技術情報',
  technikFingerabdruck: 'フィンガープリント',
  technikAdresse: 'アドレス',
  technikOhneAdresse: 'なし。同じネットワーク内で見つけ合います',
  technikAblauf: '有効期限',

  abbrechen: 'キャンセル',
  spaeter: 'あとで照合',
  stimmt: '一致しています',
  aufgenommen: '{name}を追加しました。まだ未確認です。',
  bestaetigt: '{name}を確認しました。',

  lageSchluesselGeaendert: '鍵が変更されました。もう一度照合してください',
  zeileVergleichen: 'この数字を照合してください',
  zeileStimmt: '一致しています',
  zeileVonHand: 'フィンガープリントを手動で照合',
  vonHand: '識別子を使って手動で追加',

  // **One invitation is enough** (NET-01b, #1954).
  meinName: '相手にどう表示しますか？',
  meinNameHinweis:
    'この名前で{name}に名乗り出ます。名前だけでは何も証明できません。そのため、数字を照合します。',
  gemeldet: '{name}もあなたを認識しました。数字を照合するまでは未確認です。',
  nichtGemeldet:
    '{name}にもあなたを認識してもらうには、あなた自身のコードを送り返してください。',
  meldungGescheitert:
    '{name}に追加されませんでした。理由：{grund}あなた自身のコードを送り返してください。',
  meldegruende: {
    unbeantwortet: '応答がなかったため。',
    verbraucht: '招待がすでに相手側で使用済みだったため。',
    abgelaufen: '招待の有効期限が切れていたため。',
    fremd: '招待が相手側のものではなかったため。',
    gesperrt: '相手側でブロックされているため。',
    gedrosselt:
      '試行回数が多すぎるため。しばらくしてからもう一度お試しください。',
    fassung: '相手側で別のバージョンが動作しているため。',
    gescheitert: '相手側で問題が発生したため。',
  },
}
