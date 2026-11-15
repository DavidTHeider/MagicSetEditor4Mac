// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: '外部への依頼にかけられる費用',
  hinweis:
    '外部サービスへの依頼には、費用がかかることがあります。' +
    '承認済みの金額と、そのうちすでに確保された分がここに表示されます。',

  kontoWaehlen: 'アカウント',
  zeitraumWaehlen: '期間',

  zeitraeume: {
    tag: '今日',
    monat: '今月',
  } as Record<string, string>,

  grenze: '承認済み',
  reserviert: '確保済み',
  verbraucht: '支出済み',
  unklar: '未確定',
  verfuegbar: '利用可能',

  unklarHinweis:
    '未確定とは、プロバイダがすでに依頼を受理した可能性がある状態で、' +
    'タイムアウトになったことを意味します。確認が済むまでは、支出済みとして扱われます。',

  ohneGrenze:
    'このアカウントには何も承認されていません。承認がない限り、' +
    '有料の依頼は送信されません。',

  lokalHinweis:
    'これはこのMac上の上限です。同じ認証情報を別の場所で使っている場合、' +
    'この台帳が把握していない利用枠が消費されます。',

  grenzeSetzen: '承認を変更',
  betragTippen: '金額 (最小単位。EURの場合はセント)',
  waehrungTippen: '通貨',
  setzen: '承認',
  setztGerade: '設定しています…',
  nullNimmtWeg: '0を指定すると承認が取り消されます。',

  postenTitel: '直近の依頼',
  keinePosten: 'まだ何も記帳されていません。',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: '確保済み',
    abgerechnet: '精算済み',
    freigegeben: '請求なし',
    unklar: '未確定',
  } as Record<string, string>,
}
