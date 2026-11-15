// **Der Weg zur aufgeschlagenen Fläche** — UX-03.AK-02
// (#1675).
import { describe, expect, it } from 'vitest'

import { brotkrumen } from './brotkrumen'
import { kataloge } from './sprache/kataloge'

const texte = kataloge.de

describe('die Brotkrumen', () => {
  // **Die Startseite bekommt keinen Weg** — so zeichnet es der Entwurf, und
  // ein Weg mit einer Station, die auf sich selbst zeigt, sagt nichts.
  it('gibt der Startseite keinen', () => {
    expect(brotkrumen(texte, 'start')).toEqual([])
  })

  // **Bereich › Seite**, und der Bereich führt auf seine erste Seite.
  it('nennt Bereich und Seite', () => {
    expect(brotkrumen(texte, 'turniere')).toEqual([
      { beschriftung: texte.allgemein.bereiche.spielen, ziel: 'decks' },
      { beschriftung: texte.seiten.turniere.titel },
    ])
  })

  // **Die letzte Station trägt kein Ziel.** Sie ist, wo man steht.
  it('lässt die letzte Station ohne Ziel', () => {
    const weg = brotkrumen(texte, 'sammlung')

    expect(weg[weg.length - 1]?.ziel).toBeUndefined()
  })

  // **Eine Lasche des Sets steht unter *Gestalten***, und dazwischen steht
  // das Set — so zeichnet es der Entwurf für *Karten*, *Eigenschaften*,
  // *Werkzeuge* und *Ausgeben*.
  it('schiebt das Bezugsobjekt zwischen Bereich und Fläche', () => {
    expect(
      brotkrumen(texte, 'stil', {
        name: 'Werkstatt der Dämmerung',
        ziel: 'karten',
      }),
    ).toEqual([
      { beschriftung: texte.allgemein.bereiche.gestalten, ziel: 'sets' },
      { beschriftung: 'Werkstatt der Dämmerung', ziel: 'karten' },
      {
        beschriftung:
          texte.allgemein.arbeitsplatz.schubladen.eigenschaften.knopf,
      },
    ])
  })

  // **Eine Seite in einer Schublade endet bei der Schublade** (UX-08,
  // #1680) — so steht es im Entwurf: *Gestalten › Werkstatt der Dämmerung ›
  // Eigenschaften*. Welche Lasche darin offen ist, sagt die Schublade selbst.
  // Eine Seite auf der Fläche endet weiter bei sich.
  it('endet bei der Schublade, nicht bei ihrer Lasche', () => {
    const set = { name: 'Werkstatt der Dämmerung', ziel: 'karten' } as const

    expect(brotkrumen(texte, 'konsole', set).at(-1)).toEqual({
      beschriftung: texte.allgemein.arbeitsplatz.schubladen.werkzeuge.knopf,
    })
    expect(brotkrumen(texte, 'bilder', set).at(-1)).toEqual({
      beschriftung: texte.seiten.bilder.titel,
    })
  })

  // **Und auf der Fläche des Objekts selbst führt es nirgendwohin.** Ein
  // Knopf auf die Stelle, an der man steht, ist ein Knopf, der nichts tut.
  it('lässt das Bezugsobjekt ohne Ziel, wo es die eigene Fläche ist', () => {
    const weg = brotkrumen(texte, 'karten', {
      name: 'Werkstatt der Dämmerung',
      ziel: 'karten',
    })

    expect(weg[1]).toEqual({ beschriftung: 'Werkstatt der Dämmerung' })
  })
})

describe('das Bezugsobjekt am falschen Ort', () => {
  // **Ein offenes Set ist nicht überall das Bezugsobjekt.** Es bleibt offen,
  // während jemand die Sammlung ansieht oder handelt — und dort spricht die
  // Seite über etwas anderes. „Sammeln › Werkstatt der Dämmerung ›
  // Sammlung" behauptete einen Zusammenhang, den es nicht gibt.
  //
  // **Der Entwurf zeichnet es so**: das Set steht im `brot` der Artboards
  // *Karten*, *Eigenschaften*, *Werkzeuge* und *Ausgeben* — also auf den
  // Laschen des Sets — und in keinem anderen.
  it.each(['sammlung', 'handeln', 'decks', 'pakete'] as const)(
    'lässt das Set auf %s weg',
    (seite) => {
      const weg = brotkrumen(texte, seite, {
        name: 'Werkstatt der Dämmerung',
        ziel: 'karten',
      })

      expect(weg.map((krume) => krume.beschriftung)).not.toContain(
        'Werkstatt der Dämmerung',
      )
    },
  )

  // **Und auf der Setliste ebenfalls.** Sie führt *alle* Sets; eines davon
  // hervorzuheben, weil es offen ist, wäre eine zweite Auskunft über eine
  // Liste.
  it('lässt das Set auf der Setliste weg', () => {
    expect(
      brotkrumen(texte, 'sets', {
        name: 'Werkstatt der Dämmerung',
        ziel: 'karten',
      }),
    ).toEqual([
      { beschriftung: texte.allgemein.bereiche.gestalten, ziel: 'sets' },
      { beschriftung: texte.seiten.sets.titel },
    ])
  })
})

describe('der Weg einer Seite', () => {
  // **Ein Weg der Seite steht an der Stelle ihres Namens** — UX-03.AK-02:
  // *Spielen › Partie gegen Bo › Am Tisch*, wie im Artboard *Tisch*. Den
  // Bereich gibt weiter der Rahmen; was dahinter kommt, weiß nur die Seite.
  it('setzt einen Weg der Seite an die Stelle ihres Namens', () => {
    expect(
      brotkrumen(texte, 'spielen', undefined, [
        { beschriftung: 'Partie gegen Bo' },
        { beschriftung: 'Am Tisch' },
      ]),
    ).toEqual([
      { beschriftung: texte.allgemein.bereiche.spielen, ziel: 'decks' },
      { beschriftung: 'Partie gegen Bo' },
      { beschriftung: 'Am Tisch' },
    ])
  })

  // **Das Set bleibt davor stehen** — der Weg ersetzt nur den Namen der
  // Seite. Im Artboard *Karten*: *Gestalten › Werkstatt der Dämmerung ›
  // Aschenwächter*, die aufgeschlagene Karte an der Stelle von „Karten".
  it('lässt das Bezugsobjekt vor einem Weg stehen', () => {
    expect(
      brotkrumen(
        texte,
        'karten',
        { name: 'Werkstatt der Dämmerung', ziel: 'karten' },
        [{ beschriftung: 'Aschenwächter' }],
      ),
    ).toEqual([
      { beschriftung: texte.allgemein.bereiche.gestalten, ziel: 'sets' },
      { beschriftung: 'Werkstatt der Dämmerung' },
      { beschriftung: 'Aschenwächter' },
    ])
  })

  // **Ein leerer Weg ist keiner** — die Seite steht dann mit ihrem Namen da.
  it('nimmt einen leeren Weg nicht an', () => {
    expect(brotkrumen(texte, 'spielen', undefined, []).at(-1)).toEqual({
      beschriftung: texte.seiten.spielen.titel,
    })
  })
})
