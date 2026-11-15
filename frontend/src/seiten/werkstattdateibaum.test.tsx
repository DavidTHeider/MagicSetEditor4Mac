// **Der Dateibaum der Werkstatt** — FMT-02.AK-01
// (#1691).
//
// Anlegen, Hinzufügen, Umbenennen und Entfernen stehen in **einer** Zeile über
// dem Editor und wirken auf die gewählte Datei und ihr Paket; Eingabe und
// Rückfrage stehen in der Zeile selbst (§6: kein Dialog über dem Dialog). Das
// Beiwerk steht im Baum und zeigt statt des Editors eine Tafel.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

import type { Api } from '../api'

const texte = kataloge.de.seiten.vorlagen
const pfad = '/pakete/probe.mse-style'

function zeige(eigene: Partial<Api> = {}) {
  const liste = vi.fn(() =>
    Promise.resolve({
      dateien: ['style', 'card_fields', 'skripte/farben'],
      beiwerk: ['rahmen.png'],
      fehler: null,
    }),
  )
  const lesen = vi.fn((_paket: string, datei: string) =>
    Promise.resolve({ inhalt: `# ${datei}\n`, stand: 'alt', fehler: null }),
  )
  const anlegen = vi.fn(() => Promise.resolve({ stand: 'neu', fehler: null }))
  const umbenennen = vi.fn(() => Promise.resolve({ fehler: null }))
  const entfernen = vi.fn(() => Promise.resolve({ fehler: null }))
  const einfuegen = vi.fn(() =>
    Promise.resolve({ datei: 'maske.png', abgebrochen: false, fehler: null }),
  )

  belegeWerkstatt({
    Paketdateien: liste,
    PaketdateiLesen: lesen,
    PaketdateiAnlegen: anlegen,
    PaketdateiUmbenennen: umbenennen,
    PaketdateiEntfernen: entfernen,
    PaketdateiEinfuegen: einfuegen,
    ...eigene,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  return { liste, lesen, anlegen, umbenennen, entfernen, einfuegen }
}

async function oeffne(datei: string) {
  await userEvent.click(await screen.findByRole('button', { name: datei }))
  await screen.findByLabelText(texte.inhalt)
}

function aktionen() {
  return within(screen.getByRole('group', { name: texte.dateiaktionen }))
}

// handgriff wählt einen Handgriff aus dem Menü »Datei«. **Seit UXR-18
// (#1926) stehen die vier hinter einem Knopf** und nicht mehr als Reihe über
// dem Editor — sie sind selten, und der Editor ist es nicht.
async function handgriff(name: string) {
  await userEvent.click(aktionen().getByRole('button', { name: texte.datei }))
  await userEvent.click(aktionen().getByRole('menuitem', { name }))
}

describe('Der Dateibaum der Werkstatt', () => {
  afterEach(raeumeApi)

  it('nennt das Beiwerk im Baum und zeigt es ohne Editor', async () => {
    const { lesen } = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `rahmen.png — ${texte.beiwerk}`,
      }),
    )

    expect(
      await screen.findByText(kurzform(texte.beiwerkHinweis)),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(texte.inhalt)).not.toBeInTheDocument()
    expect(lesen).not.toHaveBeenCalledWith(pfad, 'rahmen.png')
  })

  it('legt eine neue Vorlage an und schlägt sie auf', async () => {
    const { anlegen, lesen, liste } = zeige()
    await oeffne('card_fields')
    const vorher = liste.mock.calls.length

    await handgriff(texte.dateiNeu)
    expect(screen.getByLabelText(texte.dateiName)).toHaveAttribute(
      'spellcheck',
      'false',
    )
    await userEvent.type(
      screen.getByLabelText(texte.dateiName),
      'skripte/neu{Enter}',
    )

    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(pfad, 'skripte/neu')
    })
    expect(anlegen).toHaveBeenCalledWith(pfad, 'skripte/neu', '')
    expect(liste.mock.calls.length).toBeGreaterThan(vorher)
  })

  it('benennt die offene Datei mit ihrem Stand um', async () => {
    const { umbenennen, lesen } = zeige()
    await oeffne('card_fields')

    await handgriff(texte.dateiUmbenennen)
    const feld = screen.getByLabelText(texte.dateiNeuerName)
    expect(feld).toHaveValue('card_fields')
    await userEvent.clear(feld)
    await userEvent.type(feld, 'karten{Enter}')

    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(pfad, 'karten')
    })
    expect(umbenennen).toHaveBeenCalledWith(
      pfad,
      'card_fields',
      'karten',
      'alt',
    )
  })

  it('entfernt die offene Datei nach Rückfrage und kehrt zur Hauptdatei zurück', async () => {
    const { entfernen, lesen } = zeige()
    await oeffne('card_fields')

    await handgriff(texte.dateiLoeschen)
    expect(screen.getByText(texte.dateiLoeschenFrage)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: texte.dateiLoeschenJa }),
    )

    await waitFor(() => {
      expect(lesen).toHaveBeenLastCalledWith(pfad, 'style')
    })
    expect(entfernen).toHaveBeenCalledWith(pfad, 'card_fields', 'alt')
  })

  // **Ein Knopf statt vier** (UXR-18, #1926): über dem Editor stand eine
  // Reihe gleichrangiger Knöpfe für Handgriffe, die selten gebraucht werden.
  it('trägt die Handgriffe hinter einem Menüknopf', async () => {
    zeige()
    await oeffne('card_fields')

    const knopf = aktionen().getByRole('button', { name: texte.datei })
    expect(aktionen().getAllByRole('button')).toEqual([knopf])
    expect(knopf).toHaveAttribute('aria-haspopup', 'menu')

    knopf.focus()
    await userEvent.keyboard('{ArrowDown}')

    const eintraege = aktionen().getAllByRole('menuitem')
    const namen = [
      texte.dateiNeu,
      texte.dateiHinzufuegen,
      texte.dateiUmbenennen,
      texte.dateiLoeschen,
    ]
    expect(eintraege).toHaveLength(namen.length)
    namen.forEach((name, stelle) => {
      expect(eintraege[stelle]).toHaveAccessibleName(name)
    })
    expect(eintraege[0]).toHaveFocus()
  })

  it('lässt die Hauptdatei weder umbenennen noch entfernen', async () => {
    zeige()
    await oeffne('style')

    await userEvent.click(aktionen().getByRole('button', { name: texte.datei }))

    expect(
      aktionen().queryByRole('menuitem', { name: texte.dateiUmbenennen }),
    ).not.toBeInTheDocument()
    expect(
      aktionen().queryByRole('menuitem', { name: texte.dateiLoeschen }),
    ).not.toBeInTheDocument()
  })

  it('sperrt Umbenennen und Entfernen, solange etwas ungesichert ist', async () => {
    zeige()
    await oeffne('card_fields')

    await userEvent.type(screen.getByLabelText(texte.inhalt), 'x')
    await userEvent.click(aktionen().getByRole('button', { name: texte.datei }))

    for (const name of [texte.dateiUmbenennen, texte.dateiLoeschen]) {
      expect(aktionen().getByRole('menuitem', { name })).toHaveAttribute(
        'aria-disabled',
        'true',
      )
    }
    expect(aktionen().getByText(texte.dateiErstSichern)).toBeInTheDocument()
  })

  it('fügt eine Datei aus dem Dialog in den Ordner der offenen Datei ein', async () => {
    const { einfuegen } = zeige()
    await oeffne('skripte/farben')

    await handgriff(texte.dateiHinzufuegen)

    expect(
      await screen.findByText(kurzform(texte.beiwerkHinweis)),
    ).toBeInTheDocument()
    expect(einfuegen).toHaveBeenCalledWith(pfad, 'skripte')
  })

  it('ändert nichts, wenn der Dialog abgebrochen wird', async () => {
    zeige({
      PaketdateiEinfuegen: () =>
        Promise.resolve({ datei: '', abgebrochen: true, fehler: null }),
    })
    await oeffne('card_fields')

    await handgriff(texte.dateiHinzufuegen)

    expect(screen.getByLabelText(texte.inhalt)).toBeInTheDocument()
    expect(
      screen.queryByText(kurzform(texte.beiwerkHinweis)),
    ).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler in der Zeile und lässt die Eingabe stehen', async () => {
    zeige({
      PaketdateiAnlegen: () =>
        Promise.resolve({
          stand: '',
          fehler: {
            code: 'konflikt' as const,
            schluessel: '',
            meldung: 'skripte/neu steht schon da',
            werte: {},
          },
        }),
    })
    await oeffne('card_fields')

    await handgriff(texte.dateiNeu)
    await userEvent.type(
      screen.getByLabelText(texte.dateiName),
      'skripte/neu{Enter}',
    )

    expect(await aktionen().findByRole('alert')).toBeInTheDocument()
    expect(screen.getByLabelText(texte.dateiName)).toHaveValue('skripte/neu')
  })
})
