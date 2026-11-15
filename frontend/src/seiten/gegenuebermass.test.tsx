// **Das Prüfmaß an *Gegenüber finden*, mit echten Mengen** — UXR-16
// (#1924).
//
// # Warum hier und nicht am Fenster
//
// Die Hüllenprobe `uirevue` geht die Seiten der Palette an; *Gegenüber
// finden* ist keine — es ist ein Kasten in *Handeln*, in *Spielen* und unter
// *Einstellungen › Verbindungen*, und der Probebestand kennt kein Gegenüber.
// Gezählt wird deshalb mit denselben Wählern unter jsdom
// (`test/bedienzaehlung.ts`), an dem Bestand aus `test/gegenueberbestand.ts`:
// acht Gegenüber in allen Lagen, sechs Funde im Netz, 150 Kontakte und eine
// vCard mit 200 Karten. **„Im Bild" zählt jsdom nicht** — die Zahlen sind
// gezeichnete Bedienelemente, für R2 also eine obere Schranke.
//
// # Vorher, gezählt am 2026-10-02 auf `29a62fd4`
//
// | Fläche | Bedienelemente | Felder | Hauptknöpfe je Region |
// | --- | ---: | ---: | ---: |
// | Kasten, acht Gegenüber | 85 | 23 | **4** |
// | Einladung, Schritt „Meinen Code zeigen" | 9 | 3 | **2** |
// | Adressbuch, 150 Kontakte | 154 | 0 | 0 |
// | Adressbuch, Vorschau mit 200 Karten | 156 | **600** | 1 |
// | Teilen, vorbereitet | 4 | 0 | **0** |
//
// Die Zahlen nachher hält dieser Lauf fest; ändert sich eine, ändert sich
// die Zeile in `docs/UIREVIEW.md` mit.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Gegenueberkasten } from './Gegenueberkasten'
import { Adressbuch } from './Adressbuch'
import { Teilen } from '../bausteine/Teilen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { zaehle } from '../test/bedienzaehlung'
import { verstoesse } from '../test/axe'
import {
  achtGegenueber,
  adressbuch,
  sechsFunde,
  umfaenge,
  vcard,
  zweiSichten,
  zwecke,
} from '../test/gegenueberbestand'

const t = kataloge.de
const technik = t.seiten.technik

// **Erst abbauen, dann die Fassade räumen** — `Teilen` verwirft beim
// Abbau seine Kopie, und ohne Fassade scheiterte dieser Ruf.
afterEach(() => {
  cleanup()
  raeumeApi()
})

function belege() {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve({ kontakte: achtGegenueber(), zwecke, fehler: null }),
    Sammlungsfreigaben: (kennung: string) =>
      Promise.resolve({
        freigaben: zweiSichten(kennung),
        umfaenge,
        fehler: null,
      }),
    Nachbarn: () =>
      Promise.resolve({ nachbarn: sechsFunde(), verworfen: 0, fehler: null }),
    KontakteinladungErstellen: () =>
      Promise.resolve({
        einladung: {
          code: 'mse4mac-kontakt:1:abc',
          ablauf: '2026-10-09T12:00:00Z',
          fingerabdruck: '1111-2222',
        },
        fehler: null,
      }),
    Kontakte: () => Promise.resolve({ kontakte: adressbuch(), fehler: null }),
    KontakteEinlesen: () =>
      Promise.resolve({
        karten: vcard(),
        befunde: [],
        abgebrochen: false,
        fehler: null,
      }),
    TeilenVorbereiten: () =>
      Promise.resolve({
        teilgut: {
          kennung: 't1',
          art: 'deck',
          titel: 'Mono-Rot',
          dateiname: 'Mono-Rot.mse4mac-deck',
          groesse: 12000,
          text: '',
          mitgehendes: ['karten', 'mengen'],
          verluste: ['notizen'],
          zuGrossFuerMail: false,
        },
        fehler: null,
      }),
    TeilenVerwerfen: () => Promise.resolve({ teilgut: null, fehler: null }),
  })
}

async function kasten() {
  belege()

  const { container } = render(
    <Sprachraum sprache="de">
      <Gegenueberkasten />
    </Sprachraum>,
  )
  const hand = userEvent.setup()

  await hand.click(
    screen.getByRole('button', { name: t.bausteine.gegenueberTitel }),
  )
  // **Gemessen wird gefaltet** (#2035): vorn stehen fünf, Emil ist der
  // fünfte.
  await screen.findByText('Emil')

  return { container, hand }
}

describe('Gegenüber finden mit acht Gegenübern und sechs Funden', () => {
  it('zählt den Kasten nach dem Prüfmaß', async () => {
    const { container, hand } = await kasten()

    await hand.click(
      screen.getByRole('button', { name: technik.nachbarnSuchen }),
    )
    await screen.findByText('Idas Mac')

    // **Vorher 85 Bedienelemente, 23 Felder und vier Hauptknöpfe in einer
    // Region — ohne die Funde**, die es im Kasten nicht gab. Das eine
    // Hauptgewicht ist Annas „Stimmt überein": sie wartet als Erste. **Seit
    // #2035 31 statt 33**: die Liste faltet nach fünf — drei Zeilenknöpfe
    // weniger, der Knopf „Weitere 3 zeigen" dazu.
    expect(zaehle(container)).toEqual({
      bedien: 31,
      felder: 0,
      haupt: 1,
      hauptJeRegion: 1,
      ueberschriften: 0,
    })
    expect(await verstoesse(container)).toEqual([])

    // **Eine bestätigte Zeile offen**: drei Schalter, kein Hauptknopf.
    await hand.click(
      screen.getByRole('button', {
        name: `${technik.gegenueber.verwalten}: Emil`,
      }),
    )

    expect(zaehle(container)).toMatchObject({ felder: 3, hauptJeRegion: 0 })
  })

  it('führt im Schritt „Meinen Code zeigen" mit einem Knopf', async () => {
    const { container, hand } = await kasten()

    await hand.click(
      screen.getByRole('button', { name: technik.einladung.zeigen }),
    )
    await hand.type(screen.getByLabelText(technik.einladung.name), 'Anna')

    expect(zaehle(container).hauptJeRegion).toBe(1)

    await hand.click(
      screen.getByRole('button', { name: technik.einladung.erzeugen }),
    )
    await screen.findByLabelText(technik.einladung.code)

    // **Mit dem Code führt „Code erzeugen" nicht mehr** — der nächste
    // Schritt ist das Weitergeben, und „Teilen …" bereitet ihn erst vor.
    // (Das eine Hauptgewicht, das bleibt, ist Annas „Stimmt überein" in der
    // Liste darunter — eine andere Region.)
    expect(
      screen.getByRole('button', { name: technik.einladung.erzeugen }),
    ).not.toHaveClass('bg-marke')
    expect(zaehle(container).hauptJeRegion).toBe(1)
  })
})

describe('das Adressbuch mit 150 Kontakten und einer vCard mit 200 Karten', () => {
  it('faltet Liste und Vorschau', async () => {
    belege()

    const { container } = render(
      <Sprachraum sprache="de">
        <Adressbuch />
      </Sprachraum>,
    )
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', { name: t.seiten.adressbuch.titel }),
    )
    await screen.findByText('Kontakt 020')

    // **Vorher 154 Bedienelemente** — je Kontakt ein „Entfernen".
    expect(zaehle(container)).toEqual({
      bedien: 25,
      felder: 0,
      haupt: 0,
      hauptJeRegion: 0,
      ueberschriften: 1,
    })

    await hand.click(
      screen.getByRole('button', { name: t.seiten.adressbuch.waehlen }),
    )
    await screen.findByText(t.seiten.adressbuch.vorschauTitel)

    // **Vorher 156 Bedienelemente und 600 Kästchen.**
    expect(zaehle(container)).toEqual({
      bedien: 39,
      felder: 10,
      haupt: 1,
      hauptJeRegion: 1,
      ueberschriften: 2,
    })
    expect(await verstoesse(container)).toEqual([])
  })
})

describe('Teilen', () => {
  it('führt mit dem Anbieten', async () => {
    belege()

    const { container } = render(
      <Sprachraum sprache="de">
        <Teilen art="deck" bezug="d1" />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: t.bausteine.teilen.knopf }),
    )
    await screen.findByText(t.bausteine.teilen.ueberschrift)

    expect(zaehle(container).hauptJeRegion).toBe(1)
    expect(
      screen.getByRole('button', { name: t.bausteine.teilen.anbieten }),
    ).toHaveClass('bg-marke')
  })
})
