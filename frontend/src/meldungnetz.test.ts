// **Ein Netzfehler nennt Dienst, Rat und Vorgang** — D02
// (#2233).
//
// Die Seiten machen aus einer Ursache sofort Text (`meldungAus`), und der
// Knopf „Diagnose öffnen" steht nur an drei Stellen. Der Vorgang, unter dem
// das Protokoll sagt, was die Verbindung meldete, muss deshalb **im Satz**
// stehen — in jeder Sprache.
import { describe, expect, it } from 'vitest'

import { Bindungsfehler } from './api'
import { meldungAus } from './meldung'
import { kataloge, sprachen } from './sprache/kataloge'

const faelle = [
  'netzName',
  'netzNichtErreichbar',
  'netzKeineAntwort',
  'netzAntwortStockt',
  'netzAbgerissen',
  'netzProxyAnmeldung',
  'gedrosselt',
] as const

describe('Ein Netzfehler', () => {
  for (const sprache of sprachen) {
    for (const schluessel of faelle) {
      it(`nennt in ${sprache} bei ${schluessel} den Vorgang`, () => {
        const ursache = new Bindungsfehler({
          code: 'netz',
          schluessel,
          meldung: 'roh',
          werte: {
            wirt: 'api.scryfall.com',
            sekunden: '30',
            vorgang: 'a1b2c3d4',
          },
        })

        const satz = meldungAus(ursache, kataloge[sprache])

        expect(satz).toContain('a1b2c3d4')
        expect(satz).not.toMatch(/\{\w+\}/)
      })
    }
  }
})
