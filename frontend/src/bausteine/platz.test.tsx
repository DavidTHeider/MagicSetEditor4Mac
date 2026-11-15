// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Platz } from './Platz'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Platz />
    </Sprachraum>,
  )
}

// vier gibt eine Antwort mit allen Posten.
function vier(vorsicherungen = 3) {
  return {
    posten: [
      { posten: 'bestand', dateien: 1, bytes: 127_000_000, raeumbar: false },
      { posten: 'bilder', dateien: 12, bytes: 4_500_000, raeumbar: false },
      { posten: 'pakete', dateien: 900, bytes: 1_000_000_000, raeumbar: false },
      {
        posten: 'vorsicherungen',
        dateien: vorsicherungen,
        bytes: vorsicherungen * 116_000_000,
        raeumbar: true,
      },

      // **Und die beiden aus Werkstück 6** — Befund 7 des Zweitblicks über
      // Phase 4.13: die Vorrichtung führte noch den Vier-Posten-Vertrag,
      // während die Domäne längst sechs hat. Ein Lauf gegen eine zu kurze
      // Aufzählung ist grün, und niemand sieht es.
      { posten: 'protokoll', dateien: 2, bytes: 964_000, raeumbar: true },
      { posten: 'sonstiges', dateien: 3, bytes: 2_600_000, raeumbar: false },
    ],
    fehler: null,
  }
}

describe('Der Platz', () => {
  // **Die Zahl, wegen der es diese Seite gibt.**
  it('nennt jeden Posten mit seinem Gewicht', async () => {
    belegeApi({ Platz: vi.fn().mockResolvedValue(vier()) })

    zeige()

    expect(await screen.findByText(/Bestand: 127\.0 MB/)).toBeInTheDocument()

    expect(
      screen.getByText(/Sicherungen vor Migrationen: 348\.0 MB/),
    ).toBeInTheDocument()
  })

  // **Nur der räumbare Posten trägt einen Knopf** (§6): eine Handlung, die
  // nichts tut, ist keine.
  it('bietet nur die räumbaren Posten zum Wegräumen an', async () => {
    belegeApi({ Platz: vi.fn().mockResolvedValue(vier()) })

    zeige()

    await screen.findByText(/Bestand:/)

    // **Zwei von sechs** — die Vorsicherungen und das Protokoll. *Sonstiges*
    // bekommt keinen: was die Anwendung nicht kennt, löscht sie nicht.
    //
    // **Gezählt werden die Räumknöpfe und nicht jeder Knopf der Fläche** —
    // seit UX-15
    // (#1771)
    // steht neben der Erläuterung ein Hilfeknopf. Eine Zählung über alle
    // sagte etwas anderes, als ihr Name behauptet (§10).
    const knoepfe = screen.getAllByRole('button', {
      name: /die älteren wegräumen$/,
    })
    expect(knoepfe).toHaveLength(2)

    for (const posten of ['Sicherungen vor Migrationen', 'Protokoll']) {
      expect(
        screen.getByRole('button', { name: new RegExp(posten) }),
      ).toBeInTheDocument()
    }
  })

  // **Und eine einzelne Sicherung bekommt keinen** — sie ist der Rückweg und
  // bliebe ohnehin stehen. Ein Knopf, der nichts täte, wäre eine Lüge.
  it('bietet die letzte Sicherung nicht zum Wegräumen an', async () => {
    belegeApi({ Platz: vi.fn().mockResolvedValue(vier(1)) })

    zeige()

    await screen.findByText(/Bestand:/)

    expect(
      screen.queryByRole('button', { name: /Sicherungen vor Migrationen/ }),
    ).not.toBeInTheDocument()
  })

  // **Das Wegräumen zeigt die neue Zählung** — und nicht die alte.
  it('zeigt nach dem Wegräumen, was übrig ist', async () => {
    const geraeumt = vi.fn().mockResolvedValue(vier(1))

    belegeApi({
      Platz: vi.fn().mockResolvedValue(vier()),
      PlatzFreigeben: geraeumt,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Sicherungen vor Migrationen/,
      }),
    )
    await bestaetigen()

    await waitFor(() => {
      expect(geraeumt).toHaveBeenCalledWith('vorsicherungen')
    })

    expect(
      await screen.findByText(/Sicherungen vor Migrationen: 116\.0 MB/),
    ).toBeInTheDocument()
  })

  // **Gelöscht wird erst nach einem Ja** (OPS-06.AK-04) — und ein Nein
  // löscht nichts. Der Beleg, dass jemand zuhört, ist der Test darüber.
  it('fragt, bevor es wegräumt', async () => {
    const PlatzFreigeben = vi.fn().mockResolvedValue(vier(1))

    belegeApi({ Platz: vi.fn().mockResolvedValue(vier()), PlatzFreigeben })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Sicherungen vor Migrationen/,
      }),
    )

    expect(
      screen.getByText(
        kataloge.de.seiten.bestandspflege.platzFreigebenErklaerung,
      ),
    ).toBeInTheDocument()
    expect(PlatzFreigeben).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )

    expect(
      screen.queryByText(
        kataloge.de.seiten.bestandspflege.platzFreigebenErklaerung,
      ),
    ).toBeNull()
    expect(PlatzFreigeben).not.toHaveBeenCalled()
  })

  // **Ein Fehler beim Wegräumen steht da** (§8) und nicht im Nichts.
  it('meldet einen Fehler beim Wegräumen', async () => {
    belegeApi({
      Platz: vi.fn().mockResolvedValue(vier()),
      PlatzFreigeben: vi.fn().mockResolvedValue({
        posten: [],
        fehler: {
          code: 'validierung',
          schluessel: 'validierung',
          meldung: 'geht nicht',
          feld: '',
          werte: {},
        },
      }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Sicherungen vor Migrationen/,
      }),
    )
    await bestaetigen()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })

  // **Und ein Lesefehler ebenso.**
  it('meldet einen Lesefehler', async () => {
    belegeApi({
      Platz: vi.fn().mockResolvedValue({
        posten: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'nichts geht',
          feld: '',
          werte: {},
        },
      }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Ein Posten, den der Katalog nicht kennt, steht mit seinem Schlüssel
  // da** — und nicht ohne Namen.
  it('nennt einen unbekannten Posten bei seinem Schlüssel', async () => {
    belegeApi({
      Platz: vi.fn().mockResolvedValue({
        posten: [
          {
            // **Räumbar und mit zwei Dateien**, damit auch die Beschriftung
            // des Knopfes über den unbekannten Schlüssel geht.
            posten: 'zwischenspeicher',
            dateien: 2,
            bytes: 500,
            raeumbar: true,
          },
          // **Und einer ohne Dateien** — der ehrliche Fall „noch kein Bild".
          { posten: 'bilder', dateien: 0, bytes: 0, raeumbar: false },
        ],
        fehler: null,
      }),
    })

    zeige()

    expect(
      await screen.findByText('zwischenspeicher: 500 B · 2 Datei(en)'),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: /zwischenspeicher/ }),
    ).toBeInTheDocument()

    // **Ohne Dateien steht nur das Gewicht da** — „0 Datei(en)" wäre eine
    // Zahl, die niemand braucht.
    expect(screen.getByText('Eingesetzte Bilder: 0 B')).toBeInTheDocument()
  })
})

// **Und die Anzeige rechnet in Tausendern** — wie der Finder.
describe('inBytes', () => {
  it('rechnet wie der Finder', async () => {
    const { inBytes } = await import('./Platz')

    expect(inBytes(999)).toBe('999 B')
    expect(inBytes(1000)).toBe('1.0 kB')
    expect(inBytes(127_000_000)).toBe('127.0 MB')
    expect(inBytes(1_000_000_000)).toBe('1.0 GB')
    expect(inBytes(2_500_000_000_000)).toBe('2.5 TB')
    expect(inBytes(3_000_000_000_000_000)).toBe('3.0 PB')
  })
})

// **Zwei Posten zugleich gibt es nicht.**
//
// # Der Befund
//
// **Nachlauf zum Zweitblick über Phase 4.13** (2026-09-07): `useVorgang`
// sperrte **je Posten**, während jede Antwort die **ganze** Zählung trägt.
// Zwei Posten durften parallel laufen, und dann entschied die Reihenfolge der
// Antworten über die angezeigte Zahl.
//
// **`useLauf` reichte nicht**: es bevorzugt die später **gestartete** Antwort,
// und der neuere **Stand** kann von der früher gestarteten kommen — wer zuerst
// klickt und zuletzt löscht, hat die jüngere Zählung.
//
// **Deshalb läuft nur eines zugleich.** Der Datensatz ist die Anzeige, nicht
// der Posten (§7, Regel 4).
describe('Zwei Läufe zugleich', () => {
  it('lässt den zweiten gar nicht erst starten', async () => {
    let ersterLoesen: ((wert: unknown) => void) | undefined

    const PlatzFreigeben = vi.fn().mockImplementationOnce(
      () =>
        new Promise((loesen) => {
          ersterLoesen = loesen
        }),
    )

    belegeApi({ Platz: vi.fn().mockResolvedValue(vier()), PlatzFreigeben })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Sicherungen vor Migrationen/,
      }),
    )
    await bestaetigen()

    // **Beide Knöpfe sind angehalten**, solange einer läuft — auch der des
    // anderen Postens.
    for (const name of [/Sicherungen vor Migrationen/, /Protokoll/]) {
      expect(screen.getByRole('button', { name })).toBeDisabled()
    }

    // **Und ein zweiter Klick löst nichts aus.**
    await userEvent.click(screen.getByRole('button', { name: /Protokoll/ }))

    expect(PlatzFreigeben).toHaveBeenCalledTimes(1)

    ersterLoesen?.({ posten: zwei(9), fehler: null })

    expect(await screen.findByText(/9 Datei/)).toBeInTheDocument()
  })
})

// zwei gibt eine Zählung mit einer erkennbaren Dateizahl.
function zwei(dateien: number) {
  return [
    {
      posten: 'vorsicherungen',
      dateien,
      bytes: 2048,
      raeumbar: true,
    },
    { posten: 'protokoll', dateien: 7, bytes: 1024, raeumbar: true },
  ]
}

// bestaetigen sagt Ja zur Rückfrage vor dem Wegräumen (OPS-06.AK-04).
async function bestaetigen() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.bestandspflege.platzLoeschenJa,
    }),
  )
}
