// **Ein Handel, wie die Leitung ihn nennt** — herausgelöst aus
// `handelsvorgang.test.tsx`, als die Datei an die Zeilengrenze kam (§2).
import type { Handelsvorgangantwort } from '../apihandelsleitung'

// meineKennung ist, was der Ausweis dieses Rechners sagt.
export const meineKennung = '12D3KooWIch'

// einVorgang baut eine Antwort der Leitung.
export function einVorgang(
  anders: Partial<Handelsvorgangantwort['vorgang']> = {},
): Handelsvorgangantwort {
  return {
    vorgang: {
      vorgang: 'v-1',
      revision: 1,
      eltern: '',
      abdruck: 'abcdef0123456789',
      art: 'tausch',
      initiator: meineKennung,
      parteien: [meineKennung, 'bgegenueber'],
      ich: meineKennung,
      zustand: 'entwurf',
      positionen: [],
      bewertungGeteilt: false,
      zustimmungen: [],
      ende: false,
      beschluss: '',
      lage: 'dran',
      naechster: 'anbieten',
      moegliche: ['anbieten', 'abbrechen'],
      phase: 'angebot',
      ...anders,
    },
    fehler: null,
  }
}

// derAusweis ist die Belegung, ohne die die Fassade mit dem Namen der
// Methode scheitert (§3).
export const derAusweis = () =>
  Promise.resolve({
    stand: 'bereit',
    kennung: meineKennung,
    fingerabdruck: 'AB-CD',
    seit: '2026-09-19T08:00:00Z',
    vorher: '',
    vorherFingerabdruck: '',
    befunde: [],
    fehler: null,
  })
