// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'オフにすると、その領域はサイドバーとパレットから消え、どの面からの操作も拒否されます。データはそのまま残り、オンに戻せばすべて元どおりです。',
    hinweis:
      'スイッチは隠すだけで、保護はしません。保護したい場合はロールと共有設定を使ってください。',
    reife: {
      stabil: '安定',
      vorschau: 'プレビュー',
      auslaufend: '廃止予定',
    },
    saetze: {
      sammeln: 'コレクション、所有物、ウィッシュリスト',
      spielen: 'デッキ、トーナメント、対戦、テーブル',
      handeln: 'トレード、共有、封筒',
      ressourcen: 'テンプレートパッケージとテンプレート',
      ki: '提案、画像、検索インデックス、アカウント、費用',
      netz: 'ID、連絡先、招待、到達性',
      markt: 'マーケットプレイスの出品、価格、アカウント',
      turniere: 'ラウンド、組み合わせ、結果のあるトーナメント',
      dienst: 'スクリプト用 REST。動作中のサービスはオフにすると停止します',
      werkzeugserver:
        'HTTP 経由の MCP。動作中のサーバーはオフにすると停止します',
    },
    abgeschaltet: {
      titel: '{seite} はオフになっています',
      erklaerung:
        '設定 › 領域と機能 で領域を再びオンにできます。データは変わらずに残っています。',
      oeffnen: '領域と機能を開く',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'コレクション',
  'schalter.spielen': 'プレイ',
  'schalter.handeln': 'トレード',
  'schalter.ressourcen': 'リソース',
  'schalter.ki': 'AI',
  'schalter.netz': 'ネットワークと相手',
  'schalter.markt': 'マーケットプレイスと価格',
  'schalter.turniere': 'トーナメント',
  'schalter.dienst': 'スクリプト用サービス',
  'schalter.werkzeugserver': 'HTTP 経由のツールサーバー',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'オン',
  'schalter.sammeln.aus': 'オフ',
  'schalter.spielen.an': 'オン',
  'schalter.spielen.aus': 'オフ',
  'schalter.handeln.an': 'オン',
  'schalter.handeln.aus': 'オフ',
  'schalter.ressourcen.an': 'オン',
  'schalter.ressourcen.aus': 'オフ',
  'schalter.ki.an': 'オン',
  'schalter.ki.aus': 'オフ',
  'schalter.netz.an': 'オン',
  'schalter.netz.aus': 'オフ',
  'schalter.markt.an': 'オン',
  'schalter.markt.aus': 'オフ',
  'schalter.turniere.an': 'オン',
  'schalter.turniere.aus': 'オフ',
  'schalter.dienst.an': 'オン',
  'schalter.dienst.aus': 'オフ',
  'schalter.werkzeugserver.an': 'オン',
  'schalter.werkzeugserver.aus': 'オフ',
}
