// **Anzahl, Marke und Spielstein klingen verschieden** — PLAY-03.AK-04
// (#1711).
//
// # Der Befund
//
// **Drei Zahlen an einer Karte, und keine sagte, was sie zählt.** Gemessen
// am 2026-09-21 an `Tischkartenknopf.tsx`: der zugängliche Name baut sich aus
//
//	name · zone · [verdeckt] · [getappt] · "<zahl> <art>" … · [hängt an X]
//
//   - **Die Haufenzahl kam nicht vor.** Sieben Wälder auf einem Haufen sagten
//     „Wald · Schlachtfeld" — wer hört statt sieht, erfuhr nicht, dass es
//     sieben sind.
//   - **Die Marken trugen den rohen Schlüssel**: `3 +1/+1`. Kein Wort sagte,
//     dass das **Marken** sind; neben einer Haufenzahl klang beides gleich.
//   - **Ein Spielstein sagte nicht, dass er einer ist**, obwohl
//     `Spielobjekt.Art` es weiß — und die Domäne kennt **vier** Arten.
//
// Das Kriterium verlangt genau diese Trennung: „Anzahl-Badges,
// Spielmarken/Counter und Spielsteine sind sprachlich und grafisch
// unterscheidbar."
//
// # Warum das kein Bild braucht
//
// **Weil es um das geht, was ein Vorleseprogramm sagt.** Der Kommentar am
// Knopf nennt den Grund seit G3 und löste ihn nicht ein: „CR 400.5 verlangt,
// dass für **alle** klar bleibt, was woran hängt und was getappt ist, und
// ‚alle' schließt ein, wer die Karte nicht sieht, sondern hört."
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Kartenknopf } from './Tischkartenknopf'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Tischlage } from '../tisch/tischplan'

const eigene = kataloge.de.seiten.spielen

// eineLage ist eine Karte auf dem Schlachtfeld.
function eineLage(teile: Partial<Tischlage> = {}): Tischlage {
  return {
    kennung: 'k1',
    x: 0,
    y: 0,
    breite: 63,
    hoehe: 88,
    drehung: 0,
    hoehenrang: 1,
    verdeckt: false,
    zone: 'tisch',
    sitz: 'sitz-1',
    seite: 'vorne',
    karteID: '7',
    name: 'Wald',
    geheim: false,
    zaehler: {},
    haengtAn: '',
    haufen: '',
    untergelegt: false,
    art: 'karte',
    haufenzahl: 1,
    ...teile,
  }
}

function zeige(teile: Partial<Tischlage>) {
  render(
    <Sprachraum sprache="de">
      <Kartenknopf
        lage={eineLage(teile)}
        eigene={eigene}
        gezeichnet={false}
        gewaehlt={false}
        aufDruck={() => undefined}
      />
    </Sprachraum>,
  )

  return screen.getByRole('button')
}

describe('was ein Knopf über seine Karte sagt', () => {
  it('nennt die Zahl auf einem Haufen', () => {
    expect(zeige({ haufen: 'k1', haufenzahl: 7 })).toHaveTextContent(
      'auf einem Haufen',
    )
  })

  // **Und nur da.** Eine Karte, die allein liegt, liegt auf keinem Haufen —
  // „1 auf einem Haufen" wäre eine Zahl ohne Sache.
  it('schweigt, wo kein Haufen ist', () => {
    expect(zeige({})).not.toHaveTextContent('auf einem Haufen')
  })

  it('nennt Marken als Marken', () => {
    const knopf = zeige({ zaehler: { '+1/+1': 3 } })

    expect(knopf).toHaveTextContent('3 Marken +1/+1')
  })

  // **Und der Spielstein sagt, dass er einer ist** — die Domäne kennt vier
  // Arten, und drei davon sind keine Karte.
  it('nennt einen Spielstein beim Namen', () => {
    expect(zeige({ art: 'spielstein' })).toHaveTextContent(
      eigene.tischArt.spielstein,
    )
  })

  it('nennt ein Emblem beim Namen', () => {
    expect(zeige({ art: 'emblem' })).toHaveTextContent(eigene.tischArt.emblem)
  })

  // **Eine Karte sagt nicht, dass sie eine ist.** Das ist der Normalfall,
  // und wer ihn ansagte, machte aus jeder Zeile eine Wiederholung.
  it('schweigt über die Art einer Karte', () => {
    const knopf = zeige({})

    for (const wort of Object.values(eigene.tischArt)) {
      expect(knopf).not.toHaveTextContent(wort)
    }
  })

  // **Die drei zugleich bleiben auseinanderzuhalten.** Das ist das
  // Kriterium: nicht dass jedes für sich einen Namen hat, sondern dass man
  // sie nebeneinander noch trennen kann.
  it('hält die drei nebeneinander auseinander', () => {
    const knopf = zeige({
      haufen: 'k1',
      haufenzahl: 7,
      zaehler: { '+1/+1': 3 },
      art: 'spielstein',
    })

    expect(knopf).toHaveTextContent('7 auf einem Haufen')
    expect(knopf).toHaveTextContent('3 Marken +1/+1')
    expect(knopf).toHaveTextContent(eigene.tischArt.spielstein)
  })
})
