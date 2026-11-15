// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'リセット',
  hinweis:
    '軽いものから重いものまで、6種類あります。' +
    '影響を受ける内容は、事前に件数と名前つきで表示されます。',

  artWaehlen: '何をリセットしますか？',

  arten: {
    layout: 'ウインドウの外観のみ',
    profileinstellungen: 'このプロファイルの設定',
    nutzereinstellungen: '自分の設定',
    profil: 'このプロファイルと、そのすべての内容',
    nutzer: '自分と、すべての自分のプロファイル',
    anwendung: 'アプリケーション全体',
  } as Record<string, string>,

  nachsehen: '影響を受ける内容を確認',
  nachsehenLaeuft: '確認しています…',

  nichtsBetroffen: 'ここには、リセットできるものがありません。',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: '設定',
    profile: 'プロファイル',
    sets: 'セット',
    karten: 'カード',
    decks: 'デッキ',
    exemplare: 'コレクションの所持カード',
    token: 'アクセストークン',
    freigaben: '承認の記録',
    zugaenge: 'プロファイルへのアクセス権',
    bilder: '画像',
  } as Record<string, string>,

  sicherung:
    'ここではデータが失われます。先に、上のエリアでデータベースのバックアップを作成してください。',

  hindernisse: {
    offeneRueckkehr:
      'ロールバックのあと、このプロファイルの数量が未確定です。' +
      '削除する前に、数量を確認し、ロールバックを承認してください。',
    fremdesZiel: 'これはあなたのものではありません。',
    letzterNutzer:
      'このデータベースに残っている最後のユーザです。削除すると、' +
      '誰も残らなくなります。',
    laufenderAuftrag:
      'まだ実行中の処理があります。完了するまでお待ちください。',
    nachweisFehlt:
      'アプリケーション全体をリセットするには、このMacでの' +
      '明示的な承認が必要です。',
  } as Record<string, string>,

  nameTippen: '確認のため、名前を入力してください',
  freigabeSatz:
    'アプリケーション全体のリセットには、このコンピュータでのあなたの承認が必要です。承認は5分間、1回の新規開始に有効です。',
  freigeben: '承認する',
  freigebenLaeuft: '承認しています…',
  freigegeben: '{zeit}まで承認済みです。',
  kennungFuerKommandozeile:
    '同じ承認はコマンドラインでも使えます（mse zuruecksetzen anwendung）：',

  tun: 'リセット',
  tunLaeuft: 'リセットしています…',

  bilanz: '{anzahl}件の設定を元に戻しました。',
  beiseite: '以前のデータベースは {ort} にあります。',
}
