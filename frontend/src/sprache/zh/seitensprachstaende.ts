// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: '各语言的进度',

  kopf: {
    sprache: '语言',
    stand: '状态',
    herkunft: '来源',
    redaktion: '审校',
  },

  stand: {
    ausgangssprache: '源语言',
    maschinell: '机器翻译',
    geprueft: '经编辑审校',
    abgenommen: '已验收',
  },

  herkunft: {
    ausgangssprache: '源语言：最先在此编写，其他语言均由此翻译',
    mitDemCode: '与程序一同产生，没有单独的翻译流程',
    sprachmodell: '由语言模型于 {datum} 翻译',
  },

  geprueft: '由编辑于 {datum} 审校',
  ungeprueft: '未经任何人审校',
}
