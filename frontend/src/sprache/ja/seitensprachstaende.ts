// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: '各言語の状況',

  kopf: {
    sprache: '言語',
    stand: '状態',
    herkunft: '由来',
    redaktion: '校閲',
  },

  stand: {
    ausgangssprache: '原文の言語',
    maschinell: '機械翻訳',
    geprueft: '編集者による校閲済み',
    abgenommen: '受け入れ済み',
  },

  herkunft: {
    ausgangssprache:
      '原文の言語：最初にここで書かれ、他の言語はすべてここから翻訳されます',
    mitDemCode: 'プログラムと一緒に作成され、個別の翻訳工程はありません',
    sprachmodell: '{datum} に言語モデルが翻訳',
  },

  geprueft: '{datum} に編集者が校閲',
  ungeprueft: '人による校閲なし',
}
