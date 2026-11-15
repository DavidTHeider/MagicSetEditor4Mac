// **Der Assistent in einer Schublade** — UXR-12
// (#1920).
//
// # Der Befund
//
// Gemessen am 2026-10-02 am nativen Fenster (`docs/UIREVIEW.md`): die Seite
// zeigte **fünf Felder und keinen Hauptknopf** — der Assistent stand immer
// offen über der Liste, auch über einem laufenden Turnier, und schob dessen
// Werkbank unter die Falz. Angelegt wird ein Turnier einmal je Abend;
// gearbeitet wird den ganzen Abend an einem.
//
// # Was jetzt gilt
//
// *Turnier anlegen …* steht im Kopf der Seite und öffnet den Assistenten in
// einer Schublade — wie *+ Aufnehmen …* in der Sammlung (UXR-10). Es führt,
// solange kein Turnier offen ist; steht eines da, führt dessen Runde.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { buehne, melde, schritt } from '../test/turnierassistent'
import { raeumeApi } from '../test/api'
import { zaehle } from '../test/bedienzaehlung'
import { verstoesse } from '../test/axe'
import { kataloge } from '../sprache/kataloge'
import { turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// anlegenKnopf ist *Turnier anlegen …* im Kopf der Seite.
function anlegenKnopf() {
  return screen.findByRole('button', { name: texte.neuOeffnen })
}

describe('der Assistent in einer Schublade', () => {
  it('zeigt ohne Turnier kein Feld und führt mit »Turnier anlegen …«', async () => {
    buehne()

    const knopf = await anlegenKnopf()

    expect(knopf).toHaveClass('bg-marke')
    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
    expect(zaehle(document.body).felder).toBe(0)
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('öffnet den Assistenten und schließt ihn nach der ersten Runde', async () => {
    const doppel = buehne()

    await userEvent.click(await anlegenKnopf())
    const schublade = await screen.findByRole('dialog', { name: texte.neu })

    expect(schritt()).toContain(texte.schrittTurnier)
    const name = await within(schublade).findByLabelText(texte.name)

    // **Eine führende Handlung in der Schublade** — gezählt über die ganze
    // Schublade: sie ist selbst die Region, und `zaehle` sieht nur, was
    // darunter liegt.
    expect(zaehle(schublade).haupt).toBe(1)
    expect(await verstoesse(schublade)).toEqual([])

    await userEvent.type(name, 'Freitag')
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.anlegen }),
    )
    await waitFor(() => {
      expect(schritt()).toContain(texte.schrittTeilnehmende)
    })
    await melde('Anna')
    await melde('Bert')
    await waitFor(() => {
      expect(
        within(schublade).getByRole('button', { name: texte.weiter }),
      ).toBeEnabled()
    })
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      await within(schublade).findByRole('button', {
        name: texte.ersteRundeAuslosen,
      }),
    )

    await waitFor(() => {
      expect(doppel.legen).toHaveBeenCalledWith('8')
    })

    // **Danach ist die Schublade zu, und das Turnier steht in der Werkbank.**
    expect(
      await screen.findByRole('region', { name: 'Freitag' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('behält eine halbe Eingabe über das Schließen', async () => {
    buehne()

    await userEvent.click(await anlegenKnopf())
    const schublade = await screen.findByRole('dialog', { name: texte.neu })
    await userEvent.type(
      await within(schublade).findByLabelText(texte.name),
      'Freitag',
    )

    await userEvent.click(
      within(schublade).getByRole('button', {
        name: kataloge.de.allgemein.arbeitsplatz.schubladen.schliessen,
      }),
    )
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    await userEvent.click(await anlegenKnopf())
    const wieder = await screen.findByRole('dialog', { name: texte.neu })

    expect(within(wieder).getByLabelText(texte.name)).toHaveValue('Freitag')
  })

  // **Ab dem dritten Schritt gibt es das Turnier schon.** Wer die Schublade
  // dann schließt, macht später weiter — sonst bliebe die Werkbank für ein
  // Turnier fort, dessen Assistent niemand mehr sieht.
  it('macht beim Schließen ab den Teilnehmenden später weiter', async () => {
    buehne()

    await userEvent.click(await anlegenKnopf())
    const schublade = await screen.findByRole('dialog', { name: texte.neu })
    await userEvent.type(
      await within(schublade).findByLabelText(texte.name),
      'Freitag',
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.anlegen }),
    )
    await waitFor(() => {
      expect(schritt()).toContain(texte.schrittTeilnehmende)
    })

    await userEvent.keyboard('{Escape}')

    expect(
      await screen.findByRole('region', { name: 'Freitag' }),
    ).toBeInTheDocument()

    await userEvent.click(await anlegenKnopf())
    await screen.findByRole('dialog', { name: texte.neu })

    expect(schritt()).toContain(texte.schrittTurnier)
  })

  it('gibt den Vortritt ab, solange ein Turnier offen ist', async () => {
    buehne([turnierprobe({ name: 'Freitag' })])

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Freitag` }),
    )
    await screen.findByRole('region', { name: 'Freitag' })

    expect(await anlegenKnopf()).not.toHaveClass('bg-marke')
  })

  // **Ein Fehler steht, wo gearbeitet wird** — unter der Schublade sähe ihn
  // niemand.
  it('zeigt ein Nein in der Schublade, solange sie steht', async () => {
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

    await userEvent.click(await anlegenKnopf())
    const schublade = await screen.findByRole('dialog', { name: texte.neu })
    await userEvent.type(
      await within(schublade).findByLabelText(texte.name),
      'Freitag',
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.anlegen }),
    )

    expect(await within(schublade).findByRole('alert')).toHaveTextContent(
      'Der Bestand schweigt.',
    )
  })
})
