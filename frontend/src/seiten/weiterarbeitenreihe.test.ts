// **Eine Reihe statt vier Listen** — UXR-03
// (#1911).
//
// Geprüft wird, was die Reihe zusagt: Angeheftetes oben, darunter jedes
// Objekt **einmal**, an der Stelle seines jüngsten Schritts — und dass dabei
// keines verloren geht, das die Ablage genannt hat.
import { describe, expect, it } from 'vitest'

import type {
  ArbeitsschrittDTO,
  WeiterarbeitenAntwort,
} from '../apiweiterarbeiten'
import { reihe } from './weiterarbeitenreihe'

function schritt(
  kennung: string,
  vorgang: ArbeitsschrittDTO['vorgang'],
  zeitpunkt: string,
): ArbeitsschrittDTO {
  return {
    objekt: { art: 'set', kennung, name: kennung, angeheftet: false },
    vorgang,
    zeitpunkt,
  }
}

const leer: WeiterarbeitenAntwort = {
  zuletzt: null,
  geoeffnet: [],
  eingefuehrt: [],
  ausgegeben: [],
  angeheftet: [],
  fehler: null,
}

// namen ist die Reihe als Kennungen mit ihrem Vorgang — „angeheftet", wo
// keiner ist.
function namen(daten: WeiterarbeitenAntwort): string[] {
  return reihe(daten).map(
    ({ objekt, schritt: jung }) =>
      `${objekt.kennung}:${jung?.vorgang ?? 'angeheftet'}`,
  )
}

describe('die Reihe von Weiterarbeiten', () => {
  it('ist leer, wenn die Ablage nichts nennt', () => {
    expect(reihe(leer)).toEqual([])
  })

  it('stellt Angeheftetes oben und nennt es nicht noch einmal', () => {
    expect(
      namen({
        ...leer,
        geoeffnet: [schritt('a', 'geoeffnet', '2026-09-09T12:00:00Z')],
        angeheftet: [
          { art: 'set', kennung: 'a', name: 'a', angeheftet: true },
          { art: 'deck', kennung: 'b', name: 'b', angeheftet: true },
        ],
      }),
    ).toEqual(['a:angeheftet', 'b:angeheftet'])
  })

  // **Der jüngste Schritt gilt** — über alle drei Vorgänge hinweg.
  it('nennt jedes Objekt einmal, mit seinem jüngsten Schritt', () => {
    expect(
      namen({
        ...leer,
        geoeffnet: [
          schritt('a', 'geoeffnet', '2026-09-09T11:00:00Z'),
          schritt('b', 'geoeffnet', '2026-09-09T08:00:00Z'),
        ],
        eingefuehrt: [schritt('b', 'eingefuehrt', '2026-09-09T07:00:00Z')],
        ausgegeben: [schritt('a', 'ausgegeben', '2026-09-09T12:00:00Z')],
      }),
    ).toEqual(['a:ausgegeben', 'b:geoeffnet'])
  })

  // **Gerechnet und nicht verglichen**: als Zeichenkette stünde
  // `…10:00:00.5Z` vor `…10:00:00Z` — hier ist es eine halbe Sekunde später
  // —, und `+02:00` liegt zwei Stunden früher, als es aussieht.
  it('ordnet nach der Zeit, auch mit Bruchteilen und Versatz', () => {
    expect(
      namen({
        ...leer,
        geoeffnet: [
          schritt('versatz', 'geoeffnet', '2026-09-09T11:00:00+02:00'),
          schritt('ganz', 'geoeffnet', '2026-09-09T10:00:00Z'),
        ],
        ausgegeben: [
          schritt('bruchteil', 'ausgegeben', '2026-09-09T10:00:00.5Z'),
        ],
      }),
    ).toEqual(['bruchteil:ausgegeben', 'ganz:geoeffnet', 'versatz:geoeffnet'])
  })

  it('lässt gleich alte Schritte in der Reihenfolge der Antwort', () => {
    const jetzt = '2026-09-09T10:00:00Z'

    expect(
      namen({
        ...leer,
        geoeffnet: [schritt('a', 'geoeffnet', jetzt)],
        eingefuehrt: [schritt('b', 'eingefuehrt', jetzt)],
        ausgegeben: [schritt('c', 'ausgegeben', jetzt)],
      }),
    ).toEqual(['a:geoeffnet', 'b:eingefuehrt', 'c:ausgegeben'])
  })

  // **Keines geht verloren**: was die Ablage in irgendeinem Teil nennt, steht
  // in der Reihe — bei allem, was sie hergibt.
  it('verliert kein Objekt, das die Ablage nennt', () => {
    const sechs = (vorgang: ArbeitsschrittDTO['vorgang'], ab: number) =>
      Array.from({ length: 6 }, (_, stelle) =>
        schritt(
          `${vorgang}-${String(stelle)}`,
          vorgang,
          `2026-09-09T${String(ab - stelle).padStart(2, '0')}:00:00Z`,
        ),
      )

    expect(
      reihe({
        ...leer,
        geoeffnet: sechs('geoeffnet', 23),
        eingefuehrt: sechs('eingefuehrt', 17),
        ausgegeben: sechs('ausgegeben', 11),
        angeheftet: [
          { art: 'paket', kennung: 'p', name: 'p', angeheftet: true },
        ],
      }),
    ).toHaveLength(19)
  })
})
