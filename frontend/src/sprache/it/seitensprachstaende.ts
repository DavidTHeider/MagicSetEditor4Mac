// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: 'A che punto è ogni lingua',

  kopf: {
    sprache: 'Lingua',
    stand: 'Stato',
    herkunft: 'Origine',
    redaktion: 'Revisione',
  },

  stand: {
    ausgangssprache: 'lingua di partenza',
    maschinell: 'tradotta automaticamente',
    geprueft: 'revisionata da un redattore',
    abgenommen: 'accettata',
  },

  herkunft: {
    ausgangssprache:
      'lingua di partenza: scritta qui per prima, tutte le altre sono tradotte da essa',
    mitDemCode: 'nata insieme al programma, senza una traduzione a parte',
    sprachmodell: 'tradotta da un modello linguistico il {datum}',
  },

  geprueft: 'revisionata da un redattore il {datum}',
  ungeprueft: 'non revisionata da nessuna persona',
}
