// Wie weit jede Sprache ist — H04
// (#2258).
//
// Aus `seiten.ts` herausgelöst, weil die Einstellungskataloge an der
// Zeilengrenze stehen (§2); eingesetzt wird es in `seiten.ts`.

export const sprachstaende = {
  titel: 'Wie weit jede Sprache ist',

  kopf: {
    sprache: 'Sprache',
    stand: 'Stand',
    herkunft: 'Herkunft',
    redaktion: 'Redaktion',
  },

  stand: {
    ausgangssprache: 'Ausgangssprache',
    maschinell: 'maschinell übersetzt',
    geprueft: 'redaktionell geprüft',
    abgenommen: 'abgenommen',
  },

  herkunft: {
    ausgangssprache:
      'Ausgangssprache: hier zuerst geschrieben, jede andere wird aus ihr übersetzt',
    mitDemCode:
      'zusammen mit dem Programm entstanden, ohne eigenen Übersetzungslauf',
    sprachmodell: 'von einem Sprachmodell übersetzt am {datum}',
  },

  geprueft: 'redaktionell geprüft am {datum}',
  ungeprueft: 'von keinem Menschen redaktionell geprüft',
}
