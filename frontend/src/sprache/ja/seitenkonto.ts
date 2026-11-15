// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: '自分のアカウント',
  eigenesErklaerung:
    'ここではアドレスや電話番号は識別子で、だれがサインインするかを示します。メールボックスや番号が実在するかどうかは確認されません。このプログラムは何も送信せず、そのために外部サービスも必要としません。',

  kennung: 'アドレスまたは電話番号',
  kennungHinweis: '例：who@example.org または +4915112345678',

  kennwort: 'パスワード',
  kennwortHinweis: '12文字以上',
  kennwortAnmelden: 'パスワード',

  anzeigename: '表示名（任意）',
  anzeigenameHinweis: '空欄の場合は、識別子から作られます',

  eintragen: 'アカウントを作成',
  eingetragen: '作成しました：{name}',

  anmelden: 'サインイン',
  angemeldet: '{name}としてサインインしました',

  ohneAngaben: '情報なし',
  ohneAngabenErklaerung:
    '匿名アカウントには、名前としてランダムな識別子が付けられ、パスワードもアドレスもありません。アカウントが不要な場合は、作らなくても構いません。最初の起動時のユーザーがすでにあり、アプリケーションは動作します。',
  anonym: '匿名アカウントを作成',
  anonymAngelegt: '作成しました：{name}',

  kennwortNieZurueck:
    'パスワードはハッシュ化して保存され、再表示されることはありません。忘れた場合は、古いものを読み取るのではなく、新しいものを設定します。',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: '第2要素',
  erklaerung:
    'サインイン時とプロファイルの切り替え時に、第2要素が求められます。認証アプリのコードはどこでも使え、Touch IDはこのMacでのみ使えます。',

  totp: 'アプリのコード',
  biometrie: '生体認証',

  einrichten: '設定',
  entfernen: '削除',

  totpEinrichten: 'アプリのコードを設定',
  totpEntfernen: 'アプリのコードを削除',
  biometrieEinrichten: 'Touch IDまたはFace IDを設定',
  biometrieEntfernen: 'Touch IDまたはFace IDを削除',

  totpEntfernenFrage: 'アプリのコードを削除しますか？',
  biometrieEntfernenFrage: 'Touch IDまたはFace IDを削除しますか？',
  entfernenErklaerung:
    '削除すると、サインインとプロファイルの切り替えでは求められなくなります。再び使うには、新しいシークレットを持つアプリとして設定し直します。',
  entfernenJa: '削除',

  einmal: 'これは一度しか表示されません',
  einmalErklaerung:
    'このシークレットを認証アプリに入力し、リカバリーコードを保管してください。どちらも後から取り出せません。紛失した場合は、設定し直します。',
  codesUeberschrift: 'リカバリーコード{zahl}個（それぞれ1回のみ有効）：',
  bestaetigen:
    'アプリのコードを入力してください。その後で初めて第2要素が有効になります。',
} as const
