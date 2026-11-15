// **Die Symbolbibliothek** — DOM-02
// (#1697):
// Vorschau, Suche, Marken, Aufnehmen und Einlesen, getrennt von den
// Symbolschriften (AK-01), und Fehler am verursachenden Feld (§6).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import {
  krone,
  mitSymbolen,
  ohneMarken,
  symbolantwort,
  zacke,
  zeigeBibliothek,
} from '../test/symbolbibliothek'

const texte = kataloge.de.seiten.karten.symbolbibliothek

afterEach(() => {
  raeumeApi()
})

describe('Die Symbolbibliothek', () => {
  it('erklärt den leeren Zustand und nimmt das Setsymbol auf', async () => {
    const liste = vi
      .fn()
      .mockResolvedValueOnce(mitSymbolen())
      .mockResolvedValue(mitSymbolen(krone))
    const aufnehmen = vi
      .fn()
      .mockResolvedValue(symbolantwort({ id: '1', name: 'Krone' }))
    belegeApi({
      Symbolbibliothek: liste,
      BibliothekssymbolAufnehmen: aufnehmen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.name }),
      'Krone',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(
      await screen.findByText(einsetzen(texte.aufgenommen, { name: 'Krone' })),
    ).toBeInTheDocument()
    expect(aufnehmen).toHaveBeenCalledWith('7', 'Krone')
    expect(screen.getByRole('textbox', { name: texte.name })).toHaveValue('')

    const eintrag = await screen.findByRole('listitem', { name: 'Krone' })
    expect(
      within(eintrag).getByRole('img', {
        name: einsetzen(texte.vorschau, { name: 'Krone' }),
      }),
    ).toHaveAttribute('src', krone.vorschau)
    expect(
      within(eintrag).getByText(
        einsetzen(texte.verwendet, { sets: 'Kronenset, Leerset' }),
      ),
    ).toBeInTheDocument()
  })

  it('zeigt einen doppelten Namen am Namensfeld und anderes darunter', async () => {
    const aufnehmen = vi
      .fn()
      .mockResolvedValueOnce({
        ...symbolantwort({}),
        fehler: {
          code: 'konflikt',
          schluessel: 'doppelterName',
          meldung: 'Diesen Namen gibt es hier schon.',
          feld: 'name',
          werte: {},
        },
      })
      .mockResolvedValueOnce({
        ...symbolantwort({}),
        fehler: {
          code: 'validierung',
          schluessel: 'setOhneSymbol',
          meldung: 'Dieses Set hat noch kein Symbol.',
          feld: 'symbol',
          werte: {},
        },
      })
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      BibliothekssymbolAufnehmen: aufnehmen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    await screen.findByRole('listitem', { name: 'Krone' })

    const feld = screen.getByRole('textbox', { name: texte.name })
    await userEvent.type(feld, 'krone')
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    await waitFor(() => {
      expect(feld).toHaveAttribute('aria-invalid', 'true')
    })
    expect(feld).toHaveAccessibleDescription(kataloge.de.fehler.doppelterName)

    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.setOhneSymbol,
    )
    expect(feld).not.toHaveAttribute('aria-invalid', 'true')
  })

  it('sucht über Go und sagt, wenn nichts passt', async () => {
    const liste = vi
      .fn()
      .mockImplementation((suche: string) =>
        Promise.resolve(
          suche === '' ? mitSymbolen(krone, zacke) : mitSymbolen(),
        ),
      )
    belegeApi({ Symbolbibliothek: liste, MarkenJeObjekt: ohneMarken })

    zeigeBibliothek()
    await screen.findByRole('listitem', { name: 'Zacke' })

    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.suche }),
      'q',
    )

    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
    expect(liste).toHaveBeenLastCalledWith('q')

    // **Und führt zurück** (G02, #2249) — mit einem Knopf statt nur dem Rat,
    // eines von beiden zu ändern.
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.listenansicht.filterZuruecksetzen,
      }),
    )

    expect(
      await screen.findByRole('listitem', { name: 'Zacke' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('searchbox', { name: texte.suche })).toHaveValue('')
  })

  it('liest eine Datei ein — und ein weggeklickter Dialog sagt nichts', async () => {
    const einlesen = vi
      .fn()
      .mockResolvedValueOnce(symbolantwort({ abgebrochen: true }))
      .mockResolvedValueOnce(symbolantwort({ id: '3', name: 'reif' }))
    const liste = vi.fn().mockResolvedValue(mitSymbolen(krone))
    belegeApi({
      Symbolbibliothek: liste,
      BibliothekssymbolEinlesen: einlesen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    await screen.findByRole('listitem', { name: 'Krone' })
    const knopf = screen.getByRole('button', { name: texte.einlesen })

    await userEvent.click(knopf)
    await antwortAbwarten()
    expect(einlesen).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()

    await userEvent.click(knopf)
    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.eingelesen, { name: 'reif' }),
    )
  })

  it('trennt die Symbolschriften und zeigt sie nur', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen()),
      MarkenJeObjekt: ohneMarken,
      Pakete: vi.fn().mockResolvedValue({
        pakete: [
          {
            art: 'mse-symbol-font',
            name: 'magic-mana-small',
            anzeigename: 'Magic Mana Small',
            version: '2020-01-01',
          },
          {
            art: 'mse-symbol-font',
            name: 'eigene-zeichen',
            anzeigename: 'Eigene Zeichen',
            version: '',
          },
          { art: 'mse-game', name: 'magic', anzeigename: 'Magic', version: '' },
        ],
        befunde: [],
        fehler: null,
      }),
    })

    zeigeBibliothek()

    // **Die Liste steht eingeklappt** — UXR-07 (#1915): das Full-Magic-Pack
    // bringt 76 Symbolschriften mit, und offen machten sie die Bibliothek
    // anderthalb Fenster hoch. Überschrift und Erklärung bleiben stehen.
    expect(
      await screen.findByRole('heading', { name: texte.schriften }),
    ).toBeInTheDocument()
    const knopf = await screen.findByRole('button', {
      name: einsetzen(texte.schriftenZeigen, { zahl: '2' }),
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('list', { name: texte.schriftenListe }),
    ).not.toBeInTheDocument()

    await userEvent.click(knopf)

    const schriften = await screen.findByRole('list', {
      name: texte.schriftenListe,
    })
    expect(
      within(schriften).getByText(
        einsetzen(texte.schriftMitFassung, {
          name: 'Magic Mana Small',
          fassung: '2020-01-01',
        }),
      ),
    ).toBeInTheDocument()
    expect(within(schriften).getByText('Eigene Zeichen')).toBeInTheDocument()
    expect(within(schriften).queryByText('Magic')).not.toBeInTheDocument()
    expect(within(schriften).queryByRole('button')).not.toBeInTheDocument()
  })

  it('sagt, wenn keine Symbolschrift eingebunden ist', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen()),
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()

    expect(await screen.findByText(texte.keineSchriften)).toBeInTheDocument()
  })

  it('zeigt einen Ladefehler mit dem Weg zurück', async () => {
    const liste = vi
      .fn()
      .mockRejectedValueOnce(new Error('klemmt'))
      .mockResolvedValue(mitSymbolen(krone))
    belegeApi({ Symbolbibliothek: liste, MarkenJeObjekt: ohneMarken })

    zeigeBibliothek()

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    )

    expect(
      await screen.findByRole('listitem', { name: 'Krone' }),
    ).toBeInTheDocument()
  })
})
