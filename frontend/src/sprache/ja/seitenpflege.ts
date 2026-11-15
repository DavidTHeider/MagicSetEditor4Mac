// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: '整理',
  hinweis:
    '使い続けるうちに、ライブラリには何も指していない行がたまっていきます。ここでは、それが何なのか、そのうち安全に削除できるものはどれかを確認します。',

  nachsehen: '確認する',
  nachsehenLaeuft: '確認しています…',
  sauber: '何も見つかりませんでした。ライブラリは整っています。',
  gueltigbis: 'この概要は {zeit} まで有効です。',

  arten: {
    kartenwerteOhneKarte: 'カードのない値',
    setangabenOhneSet: 'セットのないセット情報',
    suchzeilenOhneKarte: 'カードのない検索行',
    deckplaetzeOhneDeck: 'デッキのないデッキ枠',
    zuordnungenOhneExemplar: '所持カードのないデッキへの割り当て',
    verlaufOhneSet: '存在しなくなったセットの履歴',
    abgelaufeneNachweise: '期限切れの承認記録',
    abgelaufeneToken: '期限切れのアクセストークン',
    moeglicheDubletten: '重複の可能性があるもの',
    offeneRueckkehren: '未完了の復元',
  } as Record<string, string>,

  klassen: {
    sicher: '安全に削除できる',
    entscheidung: 'あなたが判断する',
    nurInformieren: '手を加えない',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'これらの行は何も指しておらず、何も保持していません。あらかじめ選択されています。',
    entscheidung:
      '残りかすのように見えますが、記録である可能性もあります。選択する前に、例を確認してください。',
    nurInformieren:
      'これは削除されません。カード番号が同じでも重複とは限りません。状態、言語、フィニッシュ、入手、保管場所で区別されます。また、未完了の復元は記録です。',
  } as Record<string, string>,

  anzahl: 'ライブラリ内の {anzahl} か所',
  beispieleTitel: '例',

  aufraeumen: '選択した項目を削除',
  aufraeumenLaeuft: '削除しています…',
  aufraeumenHinweis:
    '先にライブラリのバックアップが作成されます。削除後、もう一度確認されます。',

  bilanzTitel: '実行結果',
  gesichert: '{ort} にバックアップしました',

  ausgaenge: {
    erledigt: '削除しました',
    konflikt: '概要の作成後に変更されたため、手を加えていません',
    schonWeg: 'すでにありませんでした',
    uebersprungen: 'スキップしました',
    gescheitert: '失敗しました',
    bereit: '準備完了',
  } as Record<string, string>,

  bilanzZeile: '{art}：{ausgang}（{anzahl}）',
  bilanzStaende: '削除前は {vorher} 件、削除後は {nachher} 件でした。',
}
