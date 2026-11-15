// **Die fünf Zahlen des Sitzes an der Spielerleiste** — PLAY-03
// (#1711).
//
// # Der Anlass
//
// `domain.Sitz` führt zwölf Felder; bis zum 2026-09-21 erreichten **acht**
// das Fenster. Strahlung (CR 728.1), Commander (CR 903.3), Commandersteuer
// (CR 903.8), Commanderschaden (CR 903.10a) und der Griff in eine leere
// Bibliothek (CR 704.5b) lagen in der Domäne — gerechnet, gespeichert und
// nie gezeigt.
//
// **Eine eigene Datei**, damit `spielerleiste.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, was
// die Leiste **ist**, hier, was sie über eine Commanderpartie sagt.
//
// # Dieselbe Regel wie beim Gift daneben
//
// **Eine Null ist keine Auskunft.** Die meisten Partien haben nie eine
// Strahlungsmarke und keinen Commander; „Steuer: 0" an jedem Sitz wäre
// Lärm, in dem die eine Zahl untergeht, auf die es ankommt.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Spielerleiste } from './Spielerleiste'
import { Sprachraum } from '../sprache/kontext'
import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Sitz } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

function zeige(sitze: Sitz[]) {
  render(
    <Sprachraum sprache="de">
      <Spielerleiste partie={einePartie({ sitze, fuer: 'sitz-1' })} />
    </Sprachraum>,
  )
}

describe('was die Spielerleiste über eine Commanderpartie sagt', () => {
  it('zeigt Strahlungsmarken nur, wo welche liegen', () => {
    zeige([
      einSitz({ kennung: 'sitz-1', name: 'Anna', strahlung: 3 }),
      einSitz({ kennung: 'sitz-2', name: 'Bert', strahlung: 0 }),
    ])

    const marken = screen.getAllByText(eigene.standStrahlung, { exact: false })

    expect(marken).toHaveLength(1)
    expect(marken[0]).toHaveTextContent('3')
  })

  // **Und die Mulligans** (#1839): wer neu gezogen hat, erklärt es am
  // Tisch (CR 103.5), und die Zahl sagt, mit wie vielen Karten er spielt.
  it('zeigt die Mulligans nur, wo einer genommen wurde', () => {
    zeige([
      einSitz({ kennung: 'sitz-1', name: 'Anna', mulligans: 2 }),
      einSitz({ kennung: 'sitz-2', name: 'Bert', mulligans: 0 }),
    ])

    const mulligans = screen.getAllByText(eigene.standMulligans, {
      exact: false,
    })

    expect(mulligans).toHaveLength(1)
    expect(mulligans[0]).toHaveTextContent('2')
  })

  it('zeigt die Commandersteuer nur, wo schon gewirkt wurde', () => {
    zeige([
      einSitz({ kennung: 'sitz-1', name: 'Anna', commandersteuer: 2 }),
      einSitz({ kennung: 'sitz-2', name: 'Bert', commandersteuer: 0 }),
    ])

    const steuer = screen.getAllByText(eigene.standSteuer, { exact: false })

    expect(steuer).toHaveLength(1)
    expect(steuer[0]).toHaveTextContent('2')
  })

  // **Je Commander und nicht als Summe** — CR 903.10a spricht von
  // „desselben Commanders". Zwei mit je zwanzig sind vierzig und trotzdem
  // keine einundzwanzig.
  it('zählt den Commanderschaden je Commander', () => {
    zeige([
      einSitz({
        kennung: 'sitz-1',
        name: 'Anna',
        commanderschaden: { 'sitz-2/drache/0': 7, 'sitz-2/baer/0': 20 },
      }),
    ])

    // **Über den ganzen Satz und nicht über die Zahl**: „20" steht auch am
    // Lebensstand, und ein Test, der sie dort fände, prüfte den falschen
    // Knoten.
    const faelle: [string, string][] = [
      ['sitz-2/drache/0', '7'],
      ['sitz-2/baer/0', '20'],
    ]

    for (const [wessen, wieviel] of faelle) {
      expect(
        screen.getByText(
          einsetzen(eigene.standCommanderschaden, {
            name: wessen,
            schaden: wieviel,
          }),
        ),
      ).toBeInTheDocument()
    }
  })

  // **Ab einundzwanzig ist die Partie vorbei** — CR 903.10a, und der
  // Entwurf sagt es wörtlich im Artboard `TischCommander`. Die Zahl steht
  // dann hervorgehoben da; ein Stand, der einen Spieler aus der Partie
  // nimmt, ist keine Nebenauskunft.
  it('hebt einundzwanzig Commanderschaden hervor', () => {
    zeige([
      einSitz({
        kennung: 'sitz-1',
        name: 'Anna',
        commanderschaden: { 'sitz-2/drache/0': 21 },
      }),
    ])

    expect(
      screen.getByText(eigene.standCommanderTod, { exact: false }),
    ).toBeInTheDocument()
  })

  it('hebt zwanzig noch nicht hervor', () => {
    zeige([
      einSitz({
        kennung: 'sitz-1',
        name: 'Anna',
        commanderschaden: { 'sitz-2/drache/0': 20 },
      }),
    ])

    expect(
      screen.queryByText(eigene.standCommanderTod, { exact: false }),
    ).not.toBeInTheDocument()
  })

  // **Der Griff in eine leere Bibliothek steht da** — CR 704.5b. Er
  // entscheidet die Partie bei der nächsten Prüfung der zustandsbasierten
  // Aktionen, und bis dahin sieht ihn niemand.
  it('sagt, wer aus einer leeren Bibliothek ziehen wollte', () => {
    zeige([
      einSitz({ kennung: 'sitz-1', name: 'Anna', leergezogen: true }),
      einSitz({ kennung: 'sitz-2', name: 'Bert', leergezogen: false }),
    ])

    const vermerk = screen.getAllByText(eigene.standLeergezogen, {
      exact: false,
    })

    expect(vermerk).toHaveLength(1)
  })

  // **Der Commander selbst wird benannt, nicht seine Kennung** — soweit er
  // am Tisch liegt. Wo er es nicht tut, bleibt die Kennung (§8): sie ist
  // schlechter als ein Name und besser als eine leere Stelle.
  it('nennt den Commander beim Namen, wo er am Tisch liegt', () => {
    const partie = einePartie({
      sitze: [
        einSitz({
          kennung: 'sitz-1',
          name: 'Anna',
          commanderschaden: { held: 4 },
        }),
      ],
      fuer: 'sitz-1',
      objekte: [
        einObjekt({
          kennung: 'held',
          beherrscher: 'sitz-2',
          eigner: 'sitz-2',
          zone: 'tisch',
          name: 'Ada, Erste',
        }),
      ],
    })

    render(
      <Sprachraum sprache="de">
        <Spielerleiste partie={partie} />
      </Sprachraum>,
    )

    expect(
      screen.getByText(
        einsetzen(eigene.standCommanderschaden, {
          name: 'Ada, Erste',
          schaden: '4',
        }),
        { exact: false },
      ),
    ).toBeInTheDocument()
  })
})

// **Ein verdeckter Commander zeigt keinen Namen** — dann bleibt die
// Kennung (§8). Der Fall ist echt: die Sicht gibt den Namen eines Objekts
// nur her, wo sie ihn zeigen darf.
it('behält die Kennung, wo das Objekt keinen Namen zeigt', () => {
  const partie = einePartie({
    sitze: [
      einSitz({
        kennung: 'sitz-1',
        name: 'Anna',
        commanderschaden: { held: 4 },
      }),
    ],
    fuer: 'sitz-1',
    objekte: [einObjekt({ kennung: 'held', name: '' })],
  })

  render(
    <Sprachraum sprache="de">
      <Spielerleiste partie={partie} />
    </Sprachraum>,
  )

  expect(
    screen.getByText(
      einsetzen(eigene.standCommanderschaden, { name: 'held', schaden: '4' }),
    ),
  ).toBeInTheDocument()
})
