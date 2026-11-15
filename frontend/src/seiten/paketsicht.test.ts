// **Was die Paketliste zeigt** —
// #244.
//
// Die Sicht beantwortet eine Frage: *welche Zeilen stehen jetzt da?* Sie
// besteht aus Suche, Artfilter und Ordnung, und sie ist von der Anzeige
// getrennt, weil sie die Stelle ist, an der man sich vertut — nicht die
// Klassennamen.
import { describe, expect, it } from 'vitest'

import type { PaketDTO } from '../api'
import { arten, sicht } from './paketsicht'

function paket(teile: Partial<PaketDTO> & { name: string }): PaketDTO {
  return {
    art: 'mse-style',
    anzeigename: teile.name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 1000000,
    artrang: 30,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    ...teile,
  }
}

const bestand = [
  paket({ name: 'magic', art: 'mse-game', artrang: 20, gruppe: 'magic' }),
  paket({
    name: 'magic-m15',
    anzeigename: 'magic-m15',
    gruppe: 'magic/m15 style',
    version: '2026-01-01',
    herkunft: 'https://example.invalid/pack',
  }),
  paket({
    name: 'vtes-standard',
    anzeigename: 'vtes-standard',
    gruppe: 'vtes',
    version: '2020-01-01',
  }),
]

describe('sicht', () => {
  // **Nach der Herkunft** — DOM-01.AK-01 (#1696): das Artboard `Pakete`
  // filtert nach mitgeliefert, bezogen und eigene.
  it('filtert nach der Herkunft', () => {
    const gemischt = [
      paket({ name: 'saat' }),
      paket({ name: 'geholt', herkunftsart: 'bezogen' }),
    ]

    const zeilen = sicht(
      gemischt,
      { suche: '', art: '', herkunft: 'bezogen', ordnung: 'name' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => zeile.paket?.name)).toEqual(['geholt'])
  })

  it('zeigt ohne Suche alles', () => {
    const zeilen = sicht(
      bestand,
      { suche: '', art: '', herkunft: '', ordnung: 'gruppen' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.filter((zeile) => zeile.paket !== undefined)).toHaveLength(3)
  })

  // **Gesucht wird über vier Angaben** — Name, Art, Herkunft und Gruppe.
  // Wer „m15" tippt, meint den Stil; wer „example.invalid" tippt, meint alles
  // von dort; wer „symbol-font" tippt, meint eine Art.
  it.each([
    ['nach dem Namen', 'm15', ['magic-m15']],
    ['nach der Art', 'mse-game', ['magic']],
    ['nach der Herkunft', 'example.invalid', ['magic-m15']],
    ['nach der Gruppe', 'vtes', ['vtes-standard']],
  ])('sucht %s', (_name, frage, erwartet) => {
    const zeilen = sicht(
      bestand,
      { suche: frage, art: '', herkunft: '', ordnung: 'gruppen' },
      'Ohne Gruppe',
      'de',
    )

    expect(
      zeilen
        .filter((zeile) => zeile.paket !== undefined)
        .map((z) => z.aufschrift),
    ).toEqual(erwartet)
  })

  // **Groß und klein ist dasselbe.** Paketnamen sind kleingeschrieben, und wer
  // „M15" tippt, hat recht.
  it('sucht ohne Rücksicht auf Groß- und Kleinschreibung', () => {
    const zeilen = sicht(
      bestand,
      { suche: 'M15', art: '', herkunft: '', ordnung: 'gruppen' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.filter((zeile) => zeile.paket !== undefined)).toHaveLength(1)
  })

  // **Die Gruppe über einem Treffer bleibt stehen.** Sie fällt von selbst an:
  // gefiltert werden Pakete, der Baum entsteht danach. Ohne diesen Test wäre
  // nicht festgehalten, dass die Reihenfolge so und nicht umgekehrt ist —
  // wer erst den Baum baut und dann Zeilen wegwirft, lässt Gruppen ohne
  // Inhalt stehen.
  //
  // **Zwei Stufen und nicht drei**: `magic/m15 style` trägt nach dem Filtern
  // genau ein Paket, und eine Gruppe mit genau einem Paket wird durch dieses
  // ersetzt (siehe [paketbaum.ausgeben]). Das ist MSEs Tiefe.
  it('behält den Weg zu einem Treffer', () => {
    const zeilen = sicht(
      bestand,
      { suche: 'm15', art: '', herkunft: '', ordnung: 'gruppen' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => [zeile.stufe, zeile.aufschrift])).toEqual([
      [0, 'magic'],
      [1, 'magic-m15'],
    ])
  })

  it('filtert nach der Art', () => {
    const zeilen = sicht(
      bestand,
      {
        suche: '',
        art: 'mse-game',
        herkunft: '',
        ordnung: 'gruppen',
      },
      'Ohne Gruppe',
      'de',
    )

    expect(
      zeilen
        .filter((zeile) => zeile.paket !== undefined)
        .map((z) => z.aufschrift),
    ).toEqual(['magic'])
  })

  // **Flach heißt: eine Stufe, keine Gruppen.** Die Baumordnung ist MSEs und
  // beantwortet „was gehört zusammen"; wer nach Version ordnet, fragt etwas
  // anderes — „was ist alt" —, und darauf gibt ein Baum keine Antwort.
  it.each([
    ['name', ['magic', 'magic-m15', 'vtes-standard']],
    ['version', ['magic-m15', 'vtes-standard', 'magic']],
    ['herkunft', ['magic-m15', 'magic', 'vtes-standard']],
  ] as const)('ordnet flach nach %s', (ordnung, erwartet) => {
    const zeilen = sicht(
      bestand,
      { suche: '', art: '', herkunft: '', ordnung },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => zeile.aufschrift)).toEqual(erwartet)
    expect(zeilen.every((zeile) => zeile.stufe === 0)).toBe(true)
    expect(zeilen.every((zeile) => zeile.paket !== undefined)).toBe(true)
  })

  // **Die neueste Version steht oben, aber „keine Version" steht unten.**
  //
  // Ein Paket ohne `version:` ist nicht das älteste — man weiß es nicht. Wer
  // die leere Zeichenkette absteigend einsortiert, stellt jedes Paket ohne
  // Angabe ans Ende der Liste; das ist dasselbe Ergebnis und der ehrlichere
  // Grund.
  it('stellt ein Paket ohne Version hinten an', () => {
    const zeilen = sicht(
      [paket({ name: 'ohne' }), paket({ name: 'mit', version: '2020-01-01' })],
      { suche: '', art: '', herkunft: '', ordnung: 'version' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => zeile.aufschrift)).toEqual(['mit', 'ohne'])
  })

  // Dieselbe Frage für die Herkunft: „von hier" ist keine Adresse.
  it('stellt ein Paket ohne Herkunft hinten an', () => {
    const zeilen = sicht(
      [paket({ name: 'ohne' }), paket({ name: 'mit', herkunft: 'https://a' })],
      { suche: '', art: '', herkunft: '', ordnung: 'herkunft' },
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => zeile.aufschrift)).toEqual(['mit', 'ohne'])
  })

  // **Bei gleicher Angabe entscheidet der Name.** Ohne diesen Fall stünde die
  // Reihenfolge zweier Pakete derselben Herkunft — der Normalfall, wenn alles
  // aus einem Pack kommt — bei jedem Aufruf woanders, und niemand sähe es.
  it.each([
    ['version', '2020-01-01', ''],
    ['herkunft', '', 'https://a'],
  ] as const)(
    'ordnet bei gleicher %s nach dem Namen',
    (ordnung, version, herkunft) => {
      const zeilen = sicht(
        [
          paket({ name: 'zweitens', version, herkunft }),
          paket({ name: 'erstens', version, herkunft }),
        ],
        { suche: '', art: '', herkunft: '', ordnung },
        'Ohne Gruppe',
        'de',
      )

      expect(zeilen.map((zeile) => zeile.aufschrift)).toEqual([
        'erstens',
        'zweitens',
      ])
    },
  )

  it('gibt für nichts nichts', () => {
    expect(
      sicht(
        [],
        { suche: '', art: '', herkunft: '', ordnung: 'gruppen' },
        'Ohne Gruppe',
        'de',
      ),
    ).toEqual([])
  })
})

describe('arten', () => {
  // **Die Auswahlliste kommt aus dem Bestand, nicht aus einer Aufzählung.**
  //
  // Sieben Arten sind möglich; installiert sind meist drei. Eine feste Liste
  // böte vier Filter an, die nichts finden — und ein achter Pakettyp, den
  // jemand ins Verzeichnis legt, fehlte darin.
  it('nennt die vorkommenden Arten, geordnet', () => {
    expect(arten(bestand, 'de')).toEqual(['mse-game', 'mse-style'])
  })

  it('nennt für nichts nichts', () => {
    expect(arten([], 'de')).toEqual([])
  })
})
