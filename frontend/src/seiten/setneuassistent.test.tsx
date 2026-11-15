// **Ein neues Set in drei Schritten** — UX-05, S07.NAT-02 und S07.NAT-03
// (#1677),
// nach dem Artboard *Sets*.
//
// *„Wizard ist überspringbar und behält Eingaben beim Zurückgehen."* und
// *„Suche findet Stil nach sichtbarem Namen und technischer ID."*
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sets } from './Sets'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { waehleZeile } from '../test/suchwahl'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

function paket(art: string, name: string, anzeigename: string) {
  return {
    art,
    name,
    anzeigename,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 0,
    herkunftsart: 'mitgeliefert' as const,
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}

// zwei Spiele mit je Stilen — damit alle drei Schritte dastehen.
function katalog() {
  return {
    pakete: [
      paket('mse-game', 'probe', 'Probe'),
      paket('mse-style', 'probe-schlicht', 'Schlicht'),
      paket('mse-game', 'magic', 'Magic'),
      paket('mse-style', 'magic-m15', 'M15 Normal'),
      paket('mse-style', 'magic-m15-altered', 'Altered Art'),
    ],
    befunde: [],
    fehler: null,
  }
}

function zeige(
  SetAnlegen = vi.fn().mockResolvedValue({ id: '9', fehler: null }),
  pakete = katalog(),
) {
  belegeApi({
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    Pakete: vi.fn().mockResolvedValue(pakete),
    SetAnlegen,
  })

  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return SetAnlegen
}

function schritt(name: string) {
  return screen.findByRole('heading', { name, level: 3 })
}

describe('Ein neues Set in drei Schritten', () => {
  it('zeigt, wo man steht', async () => {
    zeige()

    const folge = await screen.findByRole('list', { name: worte.neuSchritte })
    const schritte = within(folge).getAllByRole('listitem')

    expect(schritte.map((einer) => einer.textContent)).toEqual([
      `1${worte.neuSchrittArt}`,
      `2${worte.neuSchrittAussehen}`,
      `3${worte.neuSchrittName}`,
    ])
    expect(schritte[0]).toHaveAttribute('aria-current', 'step')
  })

  // **Überspringbar** — Spiel und Stil stehen vorgewählt, wie MSE2 sie
  // vorwählt, und der Name ist die einzige Frage. **Vorgewählt ist der erste
  // in MSE2s Ordnung**: bei gleichem Rang entscheidet der Name.
  it('lässt sich überspringen und legt mit den Vorgaben an', async () => {
    const SetAnlegen = zeige()

    await schritt(worte.neuSchrittArt)
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuUeberspringen }),
    )

    await userEvent.type(await screen.findByLabelText(worte.neuName), 'Schnell')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalledWith('Schnell', 'magic', 'magic-m15')
    })
  })

  // **Zurück behält alles** — Name, Spiel, Stil und die Suche.
  it('behält beim Zurückgehen, was eingegeben war', async () => {
    const SetAnlegen = zeige()

    await schritt(worte.neuSchrittArt)
    await userEvent.click(screen.getByRole('radio', { name: /Magic/ }))
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))

    // **Der Stil steht in einer Auswahl** seit UXR-04 (#1912).
    await schritt(worte.neuSchrittAussehen)
    await waehleZeile(
      screen.getByRole('combobox', { name: worte.neuStil }),
      'Altered Art magic-m15-altered',
    )
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))

    await userEvent.type(await screen.findByLabelText(worte.neuName), 'Zurück')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuSchrittZurueck }),
    )

    await schritt(worte.neuSchrittAussehen)
    expect(screen.getByRole('combobox', { name: worte.neuStil })).toHaveValue(
      'Altered Art',
    )

    await userEvent.click(
      screen.getByRole('button', { name: worte.neuSchrittZurueck }),
    )
    await schritt(worte.neuSchrittArt)
    expect(screen.getByRole('radio', { name: /Magic/ })).toBeChecked()

    await userEvent.click(
      screen.getByRole('button', { name: worte.neuUeberspringen }),
    )
    expect(await screen.findByLabelText(worte.neuName)).toHaveValue('Zurück')

    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )
    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalledWith(
        'Zurück',
        'magic',
        'magic-m15-altered',
      )
    })
  })

  // **Was bei MSE2 zuerst steht, ist vorgewählt** — Spiele und Stile in der
  // Ordnung von `position hint:` und Name (`domain.VergleicheNachRang`). Nach
  // dem Namen allein stand für Magic „magic-DayNight" vorn.
  it('ordnet Spiele und Stile wie MSE2 und wählt den ersten vor', async () => {
    const mitRang = (eines: ReturnType<typeof paket>, rang: number) => ({
      ...eines,
      rang,
    })

    zeige(undefined, {
      pakete: [
        mitRang(paket('mse-game', 'archenemy', 'Archenemy'), 2),
        mitRang(paket('mse-style', 'archenemy-standard', 'Standard'), 1),
        mitRang(paket('mse-game', 'magic', 'Magic'), 1),
        mitRang(paket('mse-style', 'magic-DayNight', 'M15 Leveler D/N'), 34),
        mitRang(paket('mse-style', 'magic-m15', 'M15 Style'), 10),
        mitRang(paket('mse-style', 'magic-m15-altered', 'M15 Mainframe'), 1),
      ],
      befunde: [],
      fehler: null,
    })

    await schritt(worte.neuSchrittArt)
    const spiele = within(
      screen.getByRole('group', { name: worte.neuArtFrage }),
    ).getAllByRole('radio')
    expect(spiele.map((eines) => eines.getAttribute('value'))).toEqual([
      'magic',
      'archenemy',
    ])
    expect(spiele[0]).toBeChecked()

    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))
    await schritt(worte.neuSchrittAussehen)
    const auswahl = screen.getByRole('combobox', { name: worte.neuStil })
    expect(auswahl).toHaveValue('M15 Mainframe')

    // **Die Zeilen stehen in derselben Ordnung** — und erst, wenn das Feld
    // offen ist (UXR-F1, #2007).
    await userEvent.click(auswahl)
    expect(
      screen.getAllByRole('option').map((einer) => einer.textContent),
    ).toEqual([
      'M15 Mainframe magic-m15-altered',
      'M15 Style magic-m15',
      'M15 Leveler D/N magic-DayNight',
    ])
  })

  // **Ein Fehlschlag ist keine leere Liste** (UX-08.NAT-01): statt einer
  // halben Schrittfolge steht, was schiefging, und ein Weg, es zu wiederholen.
  it('bietet Wiederholen, wenn die Pakete nicht kommen', async () => {
    const Pakete = vi
      .fn()
      .mockRejectedValueOnce(new Error('Katalog gesperrt'))
      .mockResolvedValue(katalog())

    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete,
    })
    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    )

    expect(await schritt(worte.neuSchrittArt)).toBeInTheDocument()
    expect(Pakete).toHaveBeenCalledTimes(2)
  })

  // **Gefunden über den Namen und über die Kennung** — S07.NAT-02.
  it('findet einen Stil über seinen Namen und seine Kennung', async () => {
    zeige()

    await schritt(worte.neuSchrittArt)
    await userEvent.click(screen.getByRole('radio', { name: /Magic/ }))
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))
    await schritt(worte.neuSchrittAussehen)

    // **Eingegrenzt wird die Auswahl** seit UXR-04 (#1912), nicht mehr eine
    // Liste von Optionsfeldern — und seit UXR-F1 (#2007) im Feld selbst.
    const auswahl = screen.getByRole('combobox', { name: worte.neuStil })
    const angeboten = () =>
      screen.queryAllByRole('option').map((zeile) => zeile.textContent)

    await userEvent.click(auswahl)
    await userEvent.keyboard('altered')
    expect(angeboten()).toEqual(['Altered Art magic-m15-altered'])

    await userEvent.clear(auswahl)
    await userEvent.keyboard('m15-alt')
    expect(angeboten()).toEqual(['Altered Art magic-m15-altered'])

    // **Was gewählt ist, bleibt stehen** — sonst sähe es aus wie keine Wahl.
    await userEvent.clear(auswahl)
    await userEvent.keyboard('nichts dergleichen')
    expect(angeboten()).toEqual([])
    await userEvent.keyboard('{Escape}')
    expect(auswahl).toHaveValue('M15 Normal')
  })
})
