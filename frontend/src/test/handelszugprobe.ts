// **Ein Handel, wie ihn der Zug sieht** — P13
// (#2381).
//
// Die Vorrichtung für die Läufe über `handelszug.ts`: ein Stand, der so
// aussieht, wie ihn der Dienst schickt, und ein Dienst, der jeden Schritt so
// beantwortet, wie die Matrix ihn geht. **Er rechnet die Matrix nicht nach**
// — er kennt nur die Folge, die ein Test durchspielt, und zählt mit, was in
// welcher Reihenfolge gerufen wurde.
import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'

export const ich = 'k-anna'
export const du = 'k-bert'

// einStand baut einen Stand dieses Handels.
export function einStand(teil: Partial<Vorgangsstand>): Vorgangsstand {
  return {
    vorgang: 'v-1',
    revision: 1,
    eltern: '',
    abdruck: 'abdruck',
    art: 'tausch',
    initiator: ich,
    parteien: [ich, du],
    ich,
    zustand: 'entwurf',
    positionen: [],
    bewertungGeteilt: false,
    zustimmungen: [],
    ende: false,
    beschluss: '',
    lage: 'dran',
    naechster: '',
    moegliche: [],
    phase: 'angebot',
    ...teil,
  }
}

const ohneFehler = { fehler: null }

// einFehler ist die Antwort einer Bindung, die nicht geht.
export function einFehler(meldung: string) {
  return { code: 'konflikt', meldung, schluessel: '', feld: '' }
}

// Ruf ist, was der Dienst gefragt wurde — in der Reihenfolge der Fragen.
export type Ruf = string

// einDienst antwortet auf jeden Schritt mit dem Stand, der in `folge`
// danach kommt, und zeichnet jeden Ruf auf.
//
// **`anfang` ist der Stand, den der Mensch sah** — der Zug liest ihn vor dem
// ersten Schritt nach (Befund 1 des Zweitblicks über Phase 8). Ohne ihn
// antwortet der Dienst mit einem anderen, und der Zug geht nichts.
export function einDienst(
  folge: Record<string, Vorgangsstand>,
  scheitert: Partial<Record<string, string>> = {},
  anfang?: Vorgangsstand,
) {
  const rufe: Ruf[] = []
  // gesehen sind die Abdrücke, die jeder Schritt nannte (#2496).
  const gesehen: string[] = []
  let jetzt: Vorgangsstand | undefined = anfang
  let gelesen = 0

  const antwort = <T extends object>(
    name: string,
    inhalt: T,
    fehlername = name,
  ) => {
    rufe.push(name)
    const grund = scheitert[fehlername]

    return Promise.resolve(
      grund === undefined
        ? { ...inhalt, ...ohneFehler }
        : { ...inhalt, fehler: einFehler(grund) },
    )
  }

  const nachSchritt = (schritt: string) => {
    const neu = folge[schritt]

    if (neu !== undefined && scheitert[schritt] === undefined) {
      jetzt = neu
    }

    return neu ?? einStand({})
  }

  const api = {
    HandelSchritt: (
      _vorgang: string,
      schritt: string,
      _wer: string,
      abdruck: string,
    ) => {
      gesehen.push(abdruck)

      return antwort(schritt, { vorgang: nachSchritt(schritt) })
    },
    HandelFreigabeErteilen: (_vorgang: string, schritt: string, wer: string) =>
      antwort(`freigabe:${schritt}:${wer}`, {
        kennung: `f-${schritt}-${wer}`,
        laeuftab: '',
      }),
    Handelsverbuchung: () => {
      nachSchritt('verbuchen')

      return antwort('verbuchen', {})
    },
    Handelsabgleich: () => {
      nachSchritt('abgleichen')

      return antwort('abgleichen', {})
    },
    Handelsbestaetigung: (_vorgang: string, wer: string) => {
      nachSchritt(`bestaetigen:${wer}`)

      return antwort(`bestaetigen:${wer}`, {})
    },
    // **`stand-danach` scheitert erst nach dem ersten Lesen** — für einen
    // Lauf, in dem gegangen ist und nur das Nachlesen danach nicht geht.
    HandelStand: () => {
      gelesen += 1

      return antwort(
        'stand',
        { vorgang: jetzt ?? einStand({}) },
        gelesen > 1 ? 'stand-danach' : 'stand',
      )
    },
    HandelSchicken: () => antwort('ausgang', { nachrichtId: 'n-1' }),
    HandelZustellen: (kennung: string) =>
      antwort('zustellen', { gegenueber: kennung, getragen: 1 }),
  }

  return { api, rufe, gesehen }
}
