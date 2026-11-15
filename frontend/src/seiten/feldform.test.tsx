// **Jedes Feld bekommt das Steuerelement seiner Art**
// (#560).
//
// Vorher war alles ein einzeiliges Textfeld — der Regeltext, die Seltenheit,
// die Manakosten, die Kennung. Was ein Feld ist, steht im Spielpaket und kam
// nie im Frontend an.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

function beschreibung(
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name: 'name',
    aufschrift: 'Name',
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    ...teile,
  }
}

function zeige(karte: KarteDTO, felder: FeldbeschreibungDTO[]) {
  belegeApi({
    Vorschau: () =>
      Promise.resolve({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: '/vorschau/A.png',
          breite: 1,
          hoehe: 1,
          dpi: 150,
          befunde: [],
          flaechen: [],
          nennungen: [],
          vorgaben: [],
        },
      }),
    Probevorschau: () =>
      Promise.resolve({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: '/vorschau/B.png',
          breite: 1,
          hoehe: 1,
          dpi: 150,
          befunde: [],
          flaechen: [],
          nennungen: [],
          vorgaben: [],
        },
      }),
    FeldSetzen: () => Promise.resolve({}),
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={felder}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

const eineKarte: KarteDTO = {
  id: '5',
  nummer: 0,
  werte: [{ feld: 'name', wert: 'Feuerball' }],
}

describe('Die Form eines Feldes', () => {
  // **Einzeilig bleibt einzeilig** — der Normalfall.
  it('gibt einem einzeiligen Textfeld eine Zeile', async () => {
    zeige(eineKarte, [beschreibung()])

    const feld = await screen.findByLabelText('name')
    expect(feld.tagName).toBe('INPUT')
  })

  // **`multi line: true` gibt einen Kasten.** Über `magic.mse-game` steht es
  // 68-mal — darunter Regeltext und Flavortext.
  it('gibt einem mehrzeiligen Feld einen Kasten', async () => {
    zeige({ ...eineKarte, werte: [{ feld: 'rule text', wert: 'Fliegend.' }] }, [
      beschreibung({ name: 'rule text', mehrzeilig: true }),
    ])

    const feld = await screen.findByLabelText('rule text')
    expect(feld.tagName).toBe('TEXTAREA')
  })

  // **Eine Auswahl wird gewählt, nicht getippt.**
  it('gibt einer Auswahl ein Auswahlfeld', async () => {
    zeige({ ...eineKarte, werte: [{ feld: 'rarity', wert: 'common' }] }, [
      beschreibung({
        name: 'rarity',
        art: 'choice',
        wahlen: ['common', 'rare'],
      }),
    ])

    const feld = await screen.findByLabelText('rarity')
    expect(feld.tagName).toBe('SELECT')
    // **Im Feld gezählt** (UXR-06, #1914): der Maßstab der Vorschau ist seither
    // ebenfalls eine Auswahl, mit eigenen Optionen.
    expect(within(feld).getAllByRole('option')).toHaveLength(2)
  })

  // **Ein Wahrheitswert wird angekreuzt.**
  it('gibt einem booleschen Feld ein Ankreuzfeld', async () => {
    zeige({ ...eineKarte, werte: [{ feld: 'foil', wert: 'yes' }] }, [
      beschreibung({ name: 'foil', art: 'boolean', wahlen: ['yes', 'no'] }),
    ])

    const feld = await screen.findByLabelText('foil')
    expect(feld).toHaveAttribute('type', 'checkbox')
    expect(feld).toBeChecked()
  })

  // **Ein `info`-Feld ist eine Auskunft, kein Eingabefeld.**
  //
  //	class InfoField : public Field {
  //	  … // Only for showing a caption
  //
  // `src/data/field/information.hpp`, Fassung f9d9356 vom 2021-07-06.
  it('macht ein info-Feld unveränderlich', async () => {
    zeige({ ...eineKarte, werte: [{ feld: 'hinweis', wert: 'Nur Text' }] }, [
      beschreibung({ name: 'hinweis', art: 'info' }),
    ])

    expect(await screen.findByLabelText('hinweis')).toHaveAttribute('readonly')
  })

  // **Die Beschreibung hängt am Feld**, nicht als Absatz daneben.
  it('hängt die Beschreibung an das Feld', async () => {
    zeige(eineKarte, [beschreibung({ beschreibung: 'Der Name der Karte.' })])

    const feld = await screen.findByLabelText('name')
    expect(feld).toHaveAccessibleDescription('Der Name der Karte.')
  })

  // **Ein Feld ohne Beschreibung im Paket bleibt ein Textfeld.** Der Rückfall
  // ist der Normalfall: ein Set kann Felder führen, die sein Spiel nicht
  // erklärt.
  it('behandelt ein unbeschriebenes Feld als Text', async () => {
    zeige({ ...eineKarte, werte: [{ feld: 'fremd', wert: 'x' }] }, [])

    const feld = await screen.findByLabelText('fremd')
    expect(feld.tagName).toBe('INPUT')
  })

  // **Eine Wahl schreibt sofort.** Ein `select` hat kein „fertig getippt": wer
  // wählt, hat gewählt — und danach zeigt die gewöhnliche Vorschau, was im
  // Bestand steht.
  it('schreibt eine Wahl ohne Verlassen', async () => {
    const FeldSetzen = vi.fn(() => Promise.resolve({}))

    belegeApi({
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/A.png',
            breite: 1,
            hoehe: 1,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [],
          },
        }),
      Probevorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/B.png',
            breite: 1,
            hoehe: 1,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [],
          },
        }),
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={{ ...eineKarte, werte: [{ feld: 'rarity', wert: 'common' }] }}
          formen={[
            beschreibung({
              name: 'rarity',
              art: 'choice',
              wahlen: ['common', 'rare'],
            }),
          ]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.selectOptions(
      await screen.findByLabelText('rarity'),
      'rare',
    )

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'rarity',
        'rare',
        expect.any(Array),
      )
    })
  })

  // **Ein Ankreuzfeld schreibt `yes` und `no`** — die Wahlen, die MSE einem
  // booleschen Feld gibt (`src/data/field/boolean.cpp`, Zeilen 14–18, Fassung
  // f9d9356 vom 2021-07-06).
  //
  // **Zwei Fälle und nicht ein Hin und Her.** Wer ankreuzt und wieder abhakt,
  // steht beim Ausgangswert — und der wird nicht geschrieben (§7 und der
  // Verlauf). Aus einem Klickpaar folgt also **kein** zweiter Schreibvorgang;
  // der erste Anlauf dieses Tests hat genau darauf gewartet.
  it.each([
    ['no', 'yes'],
    ['yes', 'no'],
  ])('schreibt aus %s ein %s', async (vorher, nachher) => {
    const FeldSetzen = vi.fn(() => Promise.resolve({}))

    belegeApi({
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={{ ...eineKarte, werte: [{ feld: 'foil', wert: vorher }] }}
          formen={[
            beschreibung({
              name: 'foil',
              art: 'boolean',
              wahlen: ['yes', 'no'],
            }),
          ]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByLabelText('foil'))

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'foil',
        nachher,
        expect.any(Array),
      )
    })
  })

  // **Der Kasten schreibt beim Verlassen** — wie das einzeilige Feld, und die
  // Vorschau folgt schon dem Tippen.
  it('schreibt aus dem Kasten beim Verlassen', async () => {
    const FeldSetzen = vi.fn(() => Promise.resolve({}))
    const Probevorschau = vi.fn(() => Promise.resolve({ vorschau: null }))

    belegeApi({
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau,
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={{ ...eineKarte, werte: [{ feld: 'rule text', wert: 'Alt' }] }}
          formen={[beschreibung({ name: 'rule text', mehrzeilig: true })]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const kasten = await screen.findByLabelText('rule text')
    await userEvent.type(kasten, 'X')

    await waitFor(() => {
      expect(Probevorschau).toHaveBeenCalledWith('5', 'rule text', 'AltX', 1)
    })

    await userEvent.tab()

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'rule text',
        'AltX',
        expect.any(Array),
      )
    })
  })
})
