// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: '一般',
    bereiche: '領域と機能',
    erscheinungsbild: '外観',
    nutzer: 'ユーザとプロファイル',
    fach: '専門機能',
    ki: 'AI',
    verbindungen: '接続',
    dienste: 'サービス',
    daten: 'データとバックアップ',
    diagnose: '診断',
  },

  untertitel: {
    allgemein: '言語、起動、ショートカット',
    bereiche: '作業領域、機能、成熟度',
    erscheinungsbild: 'テーマ、書体、ウインドウ',
    nutzer: 'サインイン、ロック解除',
    fach: 'カード、コレクション、プレイ',
    ki: 'プロバイダ、モデル、コスト',
    verbindungen: 'ネットワーク、接続先、CLI、MCP、REST',
    dienste: 'カードサービス、価格、マーケット',
    daten: 'データベース、パス、バックアップ',
    diagnose: 'ログ、計測、技術情報',
  },

  abschnitte: {
    fenster: 'ウインドウとサイドバー',
    spalten: '列と表示',
    handel: 'トレードの制限',
    protokoll: 'ログ',
    messung: '計測',
    speicher: 'キャッシュと検索インデックス',
    adressen: 'アドレス',
    ordner: 'フォルダとファイル名',
    eingriffe: '元に戻す・整理・リセット',
    mitnehmen: '設定を持ち出す',
  },
} as const
