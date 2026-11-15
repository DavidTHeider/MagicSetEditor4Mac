// Wie weit jede Sprache ist — H04 (#2258); Gestalt wie `../de/seitensprachstaende.ts`.

export const sprachstaende = {
  titel: 'Насколько готов каждый язык',

  kopf: {
    sprache: 'Язык',
    stand: 'Состояние',
    herkunft: 'Происхождение',
    redaktion: 'Редактура',
  },

  stand: {
    ausgangssprache: 'исходный язык',
    maschinell: 'машинный перевод',
    geprueft: 'проверен редактором',
    abgenommen: 'принят',
  },

  herkunft: {
    ausgangssprache:
      'исходный язык: пишется здесь первым, все остальные переводятся с него',
    mitDemCode: 'создан вместе с программой, без отдельного перевода',
    sprachmodell: 'переведён языковой моделью {datum}',
  },

  geprueft: 'проверен редактором {datum}',
  ungeprueft: 'не проверялся ни одним человеком',
}
