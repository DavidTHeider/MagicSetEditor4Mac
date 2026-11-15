// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: '⌘0～⌘9の割り当て',
  belegungHinweis:
    '⌘とともに数字キーを押したときに開くページです。初期設定はサイドバーの各エリアです。「0.6.0以前」を選ぶと、0.5.0までの割り当てに戻せます。その一覧は変更履歴に記載されています。',
  belegungFuer: '{kuerzel}で開くページ',
  belegungFrei: '—',
  belegungVor06: '0.6.0以前',
  belegungVorgabe: '初期設定に戻す',
  kuerzel: {
    palette: 'コマンドパレットを開く',
    suche: 'セット内を検索',
    zurueck: '変更を取り消す',
    vor: '変更をやり直す',
    sichern: 'セットを保存',
    sichernUnter: 'セットに名前を付けて保存…',
    oeffnen: 'セットを開く…',
    neu: '新規セットを作成',
    schliessen: 'セットを閉じる',
    drucken: '印刷シートへ移動',
    navigator: 'ナビゲータの表示/非表示',
    inspector: 'インスペクタの表示/非表示',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: '前のページ',
    seiteVor: '次のページ',
  },
} as const
