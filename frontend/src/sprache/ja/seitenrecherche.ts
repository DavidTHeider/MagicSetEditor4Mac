// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: '調べる',
  rechercheErklaerung:
    'カードのテキスト、裁定、ルールを、このMac上で検索します。全文検索とベクトル検索を統合して使います。埋め込みモデルはローカルで動作し、文章を生成するモデルには問い合わせません。何もこのMacの外に出ず、何も変更されません。',

  rechercheFall: '対象',
  rechercheFallKarten: 'カード',
  rechercheFallRegeln: 'ルール',
  rechercheFallRulings: '裁定',

  rechercheFrage: '質問',
  rechercheSuchen: '調べる',
  rechercheSprache: '言語',
  rechercheAlleSprachen: 'すべて',

  rechercheModus: '検索方法',
  rechercheModusHybrid: '全文検索とベクトル検索を統合',
  rechercheModusLexikalisch: '全文検索のみ (埋め込みモデルなし)',
  rechercheModusOhneIndex: '検索していません (検索インデックスがありません)',

  rechercheLeereNichtsGefunden: '別の語を試しますか？',
  rechercheLeereOhneIndex: '「設定 › 診断」でインデックスを作成できます。',
  rechercheLeereOhneModell:
    '埋め込みモデルがないと、検索では単語は見つかりますが、類似性は見つかりません。',
  rechercheLeereOhneQuelle:
    'この質問には、インデックスにない情報源が必要です。',
  rechercheLeereGesperrt: 'この情報源へのアクセスは許可されていません。',
  rechercheLeereBudget: 'この実行の上限に達しました。',

  rechercheHerkunft:
    'リビジョン {revision}、{datum}から有効、{abgerufen}に取得',
  rechercheWeitere: '同じカードの他の項目が{anzahl}件あります',
  rechercheOeffnen: 'カードを開く',
  rechercheEinschraenkung: '実行されなかった内容',
  rechercheStaende: 'リビジョン',
  rechercheIndex: '検索インデックス',
  rechercheMehr: 'さらに読み込む',

  aehnlichTitel: '類似のカード',
  aehnlichErklaerung:
    'カード名ではなく、このカードのテキストで検索します。このカード自体は一覧に含まれません。',
  aehnlichSuchen: '類似のカードを探す',
  aehnlichLeer: 'まだ何も検索していません。',

  freigabeTitel: 'モデルに渡される内容',
  freigabeErklaerung:
    'モデルに見える内容をまとめて表示します。送信はしません。このあと生成を実行する場合、何が渡されるかを事前に確認できます。',
  freigabeAufgabe: 'タスク',
  freigabeAnbieter: 'プロバイダ',
  freigabeModell: 'モデル',
  freigabeFenster: 'コンテキストウインドウ',
  freigabeReserve: '出力の予約分',
  freigabeZeigen: 'まとめる',
  freigabeOrtLokal: 'このMac上で処理されます。',
  freigabeOrtCloud: '外部プロバイダで処理されます。',
  freigabeBudget:
    '{grenze}トークン中{verbraucht}トークン、{reserve}トークンを確保',
  freigabeZaehlweise: '{zaehlweise}による推定',
  freigabeZieladresse: '送信先アドレス',
  freigabeKategorien: 'データの種類',
  freigabePrivat: '自分のコンテンツが含まれています。',
  freigabeOeffentlich: '公開されている情報源のみ。',
  freigabeBelege: '出典',
  freigabeKeineBelege:
    '出典がありません。この状態でまとめると、モデルには何も見えません。',
  freigabeVeraltet: '現行のリビジョンではありません',
  freigabeLaeuftAb: '{zeit}まで有効',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: '自分のカードも検索に含める',
  freigabeBestandErklaerung:
    'このチェックを入れない場合、自分のカードは対象外になります。検索は公開されている情報源のみが対象となり、回答にもその旨が表示されます。含めた内容は、モデルがこのMac以外で動作する場合、このMacの外に出ます。',
}
