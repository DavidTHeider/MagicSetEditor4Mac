// **Der Inspector der Sammlung** — Werkstück U8 von Phase 5.1 (#1419): das
// gewählte Exemplar mit Menge, Zustand, Erwerb, Lagerort und Preis; hier wird
// geändert und gelöscht. Ohne Wahl sagt er, was zu tun ist; steht das
// Exemplar nach einem Filterwechsel nicht mehr in der Sicht, sagt er das —
// und die Änderung trifft trotzdem dieses Exemplar.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import type { ExemplarDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlungsinspector } from './Sammlungsinspector'

const texte = kataloge.de.seiten.sammlung

function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 3,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: 'Kiste A',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
    ...teil,
  }
}

function zeige(
  gewaehlt: ExemplarDTO | undefined,
  teile: { inSicht?: boolean; laeuft?: boolean } = {},
) {
  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })
  const aufAendern = vi.fn(() => Promise.resolve(true))
  const aufLoeschen = vi.fn(() => Promise.resolve(true))

  render(
    <Sprachraum sprache="de">
      <Sammlungsinspector
        exemplar={gewaehlt}
        inSicht={teile.inSicht ?? true}
        aufAendern={aufAendern}
        aufLoeschen={aufLoeschen}
        aufGeholt={vi.fn()}
        laeuft={teile.laeuft ?? false}
      />
    </Sprachraum>,
  )

  return { aufAendern, aufLoeschen }
}

afterEach(() => {
  raeumeApi()
})

describe('Der Sammlungsinspector', () => {
  it('sagt ohne Wahl, was zu tun ist', () => {
    zeige(undefined)

    expect(screen.getByText(texte.inspectorLeer)).toBeInTheDocument()
  })

  it('zeigt das gewählte Exemplar mit Menge, Zustand und Lagerort', async () => {
    zeige(exemplar())

    expect(screen.getByText('probe-druck')).toBeInTheDocument()
    expect(
      screen.getByText((_, knoten) => knoten?.textContent === '3 × Near Mint'),
    ).toBeInTheDocument()
    expect(screen.getByText('Kiste A')).toBeInTheDocument()
    expect(await screen.findByText(texte.keinePreise)).toBeInTheDocument()
  })

  // **Ein benannter Druck zeigt seine Kennung, ein unbekannter Zustand liest
  // sich als Near Mint, und ohne Lagerort steht keine leere Zeile.**
  it('zeigt Kennung und Rückfallzustand, lässt einen leeren Lagerort weg', () => {
    zeige(exemplar({ name: 'Karte', zustand: 'kaputt', lagerort: '' }))

    expect(screen.getByText('probe-druck', { selector: 'span' })).toHaveClass(
      'font-mono',
    )
    expect(
      screen.getByText((_, knoten) => knoten?.textContent === '3 × Near Mint'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Kiste A')).not.toBeInTheDocument()
  })

  it('ändert Menge und Lagerort und gibt den Entwurf ab', async () => {
    const { aufAendern } = zeige(exemplar())

    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )
    const menge = screen.getByLabelText(texte.menge)
    await userEvent.clear(menge)
    await userEvent.type(menge, '9')
    const lagerort = screen.getByLabelText(texte.lagerort)
    await userEvent.clear(lagerort)
    await userEvent.type(lagerort, 'Kiste B')
    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungSichern }),
    )

    await waitFor(() => {
      expect(aufAendern).toHaveBeenCalledWith(
        expect.objectContaining({ id: 1, menge: 9, lagerort: 'Kiste B' }),
      )
    })
    // Nach dem Gelingen ist das Formular wieder zu.
    await waitFor(() => {
      expect(screen.queryByLabelText(texte.menge)).not.toBeInTheDocument()
    })
  })

  // **Jedes Feld schreibt in den Entwurf** — eines nach dem anderen, damit
  // ein vergessener Setter auffällt, statt still das alte Feld zu lassen.
  it('nimmt jedes Feld in den Entwurf', async () => {
    const { aufAendern } = zeige(exemplar())

    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )
    // **Bezug, Erwerb und Notiz stehen seit UXR-10 (#1918) unter „Weitere
    // Angaben"** — eingeklappt, bis jemand sie braucht.
    await userEvent.click(
      screen.getByRole('button', { name: texte.weitereAngaben }),
    )
    for (const [feld, wert] of [
      [texte.druck, 'neuer-druck'],
      [texte.sprache, 'de'],
      [texte.name, 'Name'],
      [texte.setcode, 'khm'],
      [texte.nummer, '12'],
      [texte.erwerbspreis, '1.50'],
      [texte.erwerbswaehrung, 'EUR'],
      [texte.notiz, 'Notiz'],
    ] as const) {
      const eingabe = screen.getByLabelText(feld)
      await userEvent.clear(eingabe)
      await userEvent.type(eingabe, wert)
    }
    await userEvent.selectOptions(screen.getByLabelText(texte.zustand), 'good')
    await userEvent.selectOptions(
      screen.getByLabelText(texte.ausfuehrung),
      'etched',
    )
    // Ein Datumsfeld nimmt keine Tastatur, aber einen Wert.
    const datum = screen.getByLabelText(texte.erwerbsdatum)
    await userEvent.clear(datum)
    await userEvent.type(datum, '2026-09-09')
    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungSichern }),
    )

    await waitFor(() => {
      expect(aufAendern).toHaveBeenCalledWith(
        expect.objectContaining({
          printing: 'neuer-druck',
          sprache: 'de',
          name: 'Name',
          setcode: 'khm',
          nummer: '12',
          erwerbspreis: '1.50',
          erwerbswaehrung: 'EUR',
          notiz: 'Notiz',
          zustand: 'good',
          finish: 'etched',
          erwerbsdatum: '2026-09-09',
        }),
      )
    })
  })

  it('verwirft den Entwurf beim Abbrechen', async () => {
    zeige(exemplar())

    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )
    const menge = screen.getByLabelText(texte.menge)
    await userEvent.clear(menge)
    await userEvent.type(menge, '9')
    await userEvent.click(screen.getByRole('button', { name: texte.abbrechen }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )

    expect(screen.getByLabelText(texte.menge)).toHaveValue(3)
  })

  it('lässt das Formular offen, wenn das Ändern scheitert', async () => {
    const { aufAendern } = zeige(exemplar())
    aufAendern.mockResolvedValue(false)

    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungSichern }),
    )

    // **Die Antwort abwarten, nicht den Ruf** (§3): das Formular soll danach
    // noch stehen — eine Verneinung, die erst nach der Antwort etwas sagt.
    await expect(aufAendern.mock.results[0]?.value).resolves.toBe(false)
    expect(screen.getByLabelText(texte.menge)).toBeInTheDocument()
  })

  // **Ein Exemplar zu löschen fragt vorher** — UX-10.AK-04
  // (#1682).
  //
  // **Weil es keine Rücknahme gibt**: Menge, Zustand, Lagerort und
  // Erwerbspreis sind von Hand eingetragen, und der Sammlungsbestand hat
  // keinen Verlauf wie ein Set.
  it('fragt, bevor es das Exemplar löscht', async () => {
    const { aufLoeschen } = zeige(exemplar())

    await userEvent.click(screen.getByRole('button', { name: texte.loeschen }))

    expect(await screen.findByText(texte.loeschenFrage)).toBeVisible()
    expect(aufLoeschen).not.toHaveBeenCalled()
  })

  it('löscht das Exemplar nach dem Bestätigen', async () => {
    const { aufLoeschen } = zeige(exemplar())

    await userEvent.click(screen.getByRole('button', { name: texte.loeschen }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.loeschenJa }),
    )

    expect(aufLoeschen).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }))
  })

  // **Abbrechen lässt das Exemplar stehen** und räumt die Frage weg (§7,
  // Regel 1).
  it('lässt das Exemplar stehen, wenn jemand abbricht', async () => {
    const { aufLoeschen } = zeige(exemplar())

    await userEvent.click(screen.getByRole('button', { name: texte.loeschen }))
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )

    await waitFor(() => {
      expect(screen.queryByText(texte.loeschenFrage)).toBeNull()
    })

    expect(aufLoeschen).not.toHaveBeenCalled()
  })

  // **Escape gibt den Fokus dem Löschknopf zurück** — UX-10.AK-04, gefunden
  // in der WebKit-Abnahme (QA01, #2382): der Knopf wurde während der Frage
  // abgebaut, und der gemerkte Auslöser war danach ein losgelöster Knoten.
  it('gibt den Fokus nach Escape dem Löschknopf zurück', async () => {
    zeige(exemplar())

    await userEvent.click(screen.getByRole('button', { name: texte.loeschen }))
    await screen.findByText(texte.loeschenFrage)
    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByText(texte.loeschenFrage)).toBeNull()
    })
    expect(screen.getByRole('button', { name: texte.loeschen })).toHaveFocus()
  })

  it('sagt es, wenn das Exemplar nicht mehr in der Sicht steht', () => {
    zeige(exemplar(), { inSicht: false })

    expect(screen.getByText(texte.inspectorAusserhalb)).toBeInTheDocument()
    // Ändern geht trotzdem — an diesem Exemplar.
    expect(screen.getByRole('button', { name: texte.bearbeiten })).toBeEnabled()
  })

  // **Ein Wechsel des Exemplars räumt den Entwurf** (§7, Regel 1).
  it('räumt den Entwurf, wenn ein anderes Exemplar gewählt wird', async () => {
    const { rerender } = render(
      <Sprachraum sprache="de">
        <Sammlungsinspector
          exemplar={exemplar()}
          inSicht
          aufAendern={vi.fn(() => Promise.resolve(true))}
          aufLoeschen={vi.fn(() => Promise.resolve(true))}
          aufGeholt={vi.fn()}
          laeuft={false}
        />
      </Sprachraum>,
    )
    belegeApi({
      Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.bearbeiten }),
    )
    expect(screen.getByLabelText(texte.menge)).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Sammlungsinspector
          exemplar={exemplar({
            id: 2,
            bezeichnung: 'zweiter',
            printing: 'zweiter',
          })}
          inSicht
          aufAendern={vi.fn(() => Promise.resolve(true))}
          aufLoeschen={vi.fn(() => Promise.resolve(true))}
          aufGeholt={vi.fn()}
          laeuft={false}
        />
      </Sprachraum>,
    )

    expect(screen.queryByLabelText(texte.menge)).not.toBeInTheDocument()
    expect(screen.getByText('zweiter')).toBeInTheDocument()
  })
})

const marke = kataloge.de.bausteine.druck.vorlaeufig

// **Ein Druck ohne Alias ist vorläufig** — Werkstück A2 von Phase 5.3 (#1441).
describe('der Inspector und die vorläufige Identität', () => {
  it('zeigt die Marke, wenn der Druck keinen Alias hat', () => {
    zeige(exemplar({ printing: '', druckID: '9', vorlaeufig: true }))

    expect(screen.getByText(marke)).toBeInTheDocument()
  })

  it('zeigt sie nicht bei einem Druck mit Alias', () => {
    zeige(exemplar())

    expect(screen.queryByText(marke)).not.toBeInTheDocument()
  })
})
