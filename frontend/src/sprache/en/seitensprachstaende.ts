// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: 'How far each language is',

  kopf: {
    sprache: 'Language',
    stand: 'Status',
    herkunft: 'Origin',
    redaktion: 'Review',
  },

  stand: {
    ausgangssprache: 'source language',
    maschinell: 'machine-translated',
    geprueft: 'reviewed by an editor',
    abgenommen: 'accepted',
  },

  herkunft: {
    ausgangssprache:
      'source language: written here first, every other language is translated from it',
    mitDemCode:
      'created together with the program, without a separate translation run',
    sprachmodell: 'translated by a language model on {datum}',
  },

  geprueft: 'reviewed by an editor on {datum}',
  ungeprueft: 'not reviewed by any human editor',
}
