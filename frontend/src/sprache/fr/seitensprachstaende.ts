// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: 'Où en est chaque langue',

  kopf: {
    sprache: 'Langue',
    stand: 'État',
    herkunft: 'Origine',
    redaktion: 'Relecture',
  },

  stand: {
    ausgangssprache: 'langue source',
    maschinell: 'traduite automatiquement',
    geprueft: 'relue par un rédacteur',
    abgenommen: 'validée',
  },

  herkunft: {
    ausgangssprache:
      'langue source : écrite ici en premier, toutes les autres sont traduites à partir d’elle',
    mitDemCode: 'née avec le programme, sans traduction distincte',
    sprachmodell: 'traduite par un modèle de langage le {datum}',
  },

  geprueft: 'relue par un rédacteur le {datum}',
  ungeprueft: 'relue par aucun humain',
}
