// **Ein Turnier in vier Schritten** — DOM-04.AK-01
// (#1699):
// *„Ein Assistent führt durch Format, Teilnehmer, Regeln und erste Runde;
// Pflichtangaben sind begründet."*
//
// Das Doppel hält **einen** Stand, den jeder schreibende Ruf fortschreibt —
// sonst prüfte der Lauf Schritte, die nie aufeinander folgen.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { assistent, buehne, melde, schritt } from '../test/turnierassistent'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { rundenprobe, teilnehmerprobe, turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

describe('der Turnierassistent', () => {
  it('führt vom Namen bis zur ersten Runde', async () => {
    const doppel = buehne()
    const bereich = await assistent()

    expect(schritt()).toContain(texte.schrittTurnier)

    // **Nur die Formate dieses Spiels.**
    const format = within(bereich).getByLabelText(texte.format)
    await waitFor(() => {
      expect(within(format).getByRole('option', { name: 'Pauper' }))
    })
    expect(within(format).getAllByRole('option')).toHaveLength(2)

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.selectOptions(format, '4')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )

    expect(schritt()).toContain(texte.schrittRegeln)
    await userEvent.click(within(bereich).getByLabelText(texte.einzelspiel))
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(doppel.anlegen).toHaveBeenCalledWith(
        'Freitag',
        'magic',
        '',
        true,
        '4',
      )
    })

    // **Die Teilnehmenden, und „Weiter" sagt, warum es noch ruht.**
    await waitFor(() => {
      expect(schritt()).toContain(texte.schrittTeilnehmende)
    })
    expect(
      screen.getByRole('button', { name: texte.weiter }),
    ).toHaveAccessibleDescription(texte.zuWenige)
    expect(screen.getByRole('button', { name: texte.weiter })).toBeDisabled()

    await melde('Anna')
    await melde('Bert')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.weiter })).toBeEnabled()
    })
    await userEvent.click(screen.getByRole('button', { name: texte.weiter }))

    // **Die erste Runde: erst sehen, dann auslosen.**
    expect(schritt()).toContain(texte.schrittErsteRunde)
    expect(await screen.findByText(/Anna/)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: texte.ersteRundeAuslosen }),
    )

    await waitFor(() => {
      expect(doppel.legen).toHaveBeenCalledWith('8')
    })

    // **Danach steht das Turnier in der Werkbank**, und der Assistent fängt
    // von vorn an — seit UXR-12 (#1920) in einer Schublade, die dabei
    // schließt; wer ihn wieder öffnet, steht beim ersten Schritt.
    expect(
      await screen.findByRole('region', { name: 'Freitag' }),
    ).toBeInTheDocument()
    await assistent()
    expect(schritt()).toContain(texte.schrittTurnier)
  })

  it('begründet die Pflichtangaben', async () => {
    buehne()
    const bereich = await assistent()

    expect(
      within(bereich).getByLabelText(texte.name),
    ).toHaveAccessibleDescription(texte.nameWarum)
    expect(
      within(bereich).getByLabelText(texte.spiel),
    ).toHaveAccessibleDescription(texte.spielWarum)
  })

  it('behält die Eingabe beim Zurückgehen', async () => {
    buehne()
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.zurueck }),
    )

    expect(within(bereich).getByLabelText(texte.name)).toHaveValue('Freitag')
  })

  it('legt nach einer Vorlage an — mit den Namen, ohne Decks', async () => {
    const vorlage = turnierprobe({
      id: '7',
      name: 'Freitag',
      formatID: '4',
      einzelspiel: true,
      teilnehmer: [
        teilnehmerprobe({ id: '20', name: 'Anna', deckID: '3' }),
        teilnehmerprobe({ id: '21', name: 'Bert', stelle: 2 }),
      ],
      runden: [rundenprobe()],
    })
    const doppel = buehne([vorlage])
    const bereich = await assistent()

    await waitFor(() => {
      expect(
        within(within(bereich).getByLabelText(texte.vorlage)).getByRole(
          'option',
          { name: 'Freitag' },
        ),
      )
    })
    await userEvent.selectOptions(
      within(bereich).getByLabelText(texte.vorlage),
      '7',
    )

    // **Spiel und Format kommen aus der Vorlage** — und ruhen.
    expect(within(bereich).getByLabelText(texte.spiel)).toBeDisabled()
    expect(within(bereich).getByLabelText(texte.format)).toHaveValue('4')

    await userEvent.type(
      within(bereich).getByLabelText(texte.name),
      'Freitag II',
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )

    expect(
      within(bereich).getByText(texte.regelnAusVorlage),
    ).toBeInTheDocument()
    expect(within(bereich).getByLabelText(texte.einzelspiel)).toBeChecked()
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(doppel.kopieren).toHaveBeenCalledWith('7', 'Freitag II', true)
    })
    expect(doppel.anlegen).not.toHaveBeenCalled()

    // **Die Namen sind schon da** — „Weiter" geht sofort.
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.weiter })).toBeEnabled()
    })
  })

  // **Ohne Namen nur die Einstellungen** — der Haken steht, und wer ihn
  // nimmt, bekommt ein leeres Turnier.
  it('kopiert ohne Namen, wenn der Haken fehlt', async () => {
    const vorlage = turnierprobe({
      id: '7',
      name: 'Freitag',
      teilnehmer: [teilnehmerprobe({ id: '20', name: 'Anna' })],
    })
    const doppel = buehne([vorlage])
    const bereich = await assistent()

    await waitFor(() => {
      expect(
        within(within(bereich).getByLabelText(texte.vorlage)).getByRole(
          'option',
          { name: 'Freitag' },
        ),
      )
    })
    await userEvent.selectOptions(
      within(bereich).getByLabelText(texte.vorlage),
      '7',
    )
    await userEvent.click(
      within(bereich).getByLabelText(texte.mitTeilnehmenden),
    )
    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Leer')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(doppel.kopieren).toHaveBeenCalledWith('7', 'Leer', false)
    })
  })

  // **Ein Fehler ohne Feld steht über der Seite** — und der Assistent
  // bleibt, wo er ist.
  it('meldet einen Fehler ohne Feld über der Seite', async () => {
    buehne([], {
      TurnierAnlegen: () =>
        Promise.resolve({
          turnier: null,
          fehler: {
            code: 'unbekannt',
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Bestand schweigt.',
          },
        }),
    })
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    expect(await screen.findByText('Der Bestand schweigt.')).toBeInTheDocument()
    expect(schritt()).toContain(texte.schrittRegeln)
  })

  it('lässt sich nach dem Anlegen verlassen', async () => {
    buehne()
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.verlassen }),
    )

    expect(
      await screen.findByRole('region', { name: 'Freitag' }),
    ).toBeInTheDocument()
    await assistent()
    expect(schritt()).toContain(texte.schrittTurnier)
  })

  it('führt einen Befund am Namen zurück zum ersten Schritt', async () => {
    buehne([], {
      TurnierAnlegen: () =>
        Promise.resolve({
          turnier: null,
          fehler: {
            code: 'konflikt',
            schluessel: 'fehler.konflikt',
            meldung: 'Den Namen gibt es schon.',
            feld: 'name',
          },
        }),
    })
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(schritt()).toContain(texte.schrittTurnier)
    })
    expect(within(bereich).getByLabelText(texte.name)).toHaveAttribute(
      'aria-invalid',
      'true',
    )
  })
})
