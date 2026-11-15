// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': 'トレース',
  'messung.abtastung': '記録する呼び出しの割合（パーセント）',
  'messung.spurgrenze': 'エントリの最大数',
  'messung.aufbewahrung': 'エントリを保持する時間（分）',
  'messung.profile': '保持するパフォーマンスプロファイルの数',
  'zwischenspeicher.grenze': 'ディスクキャッシュごとの上限（MB）',
} as const

export const messwerte = {
  'messung.spur.aus': 'オフ',
  'messung.spur.an': 'オン',
} as const

export const messhinweise = {
  'messung.spur':
    'サービスとツールサーバへの呼び出しごとに1行を記録します。操作ID、経路、所要時間、結果が含まれ、内容は含まれません。プロセスのメモリ上に保持され、次回の起動から有効になります。',
  'messung.abtastung':
    '10にすると、トレースは10回に1回の呼び出しを記録します。無作為ではなく正確に数えます。空欄は100です。',
  'messung.spurgrenze':
    'トレースがいっぱいになると、最も古いエントリから削除されます。10～10000、空欄は1000です。',
  'messung.aufbewahrung':
    '古いエントリは除外されます。1～1440分、空欄は60です。',
  'messung.profile':
    '最新のものはデータの隣に残り、古いものは次のプロファイルの際に削除されます。1～20、空欄は3です。',
  'zwischenspeicher.grenze':
    'ScryfallからのレスポンスとScryfallの画像に、それぞれ個別に適用されます。起動時に、上限に収まるまで古いものから削除されます。1～8192、空欄は512です。',
} as const
