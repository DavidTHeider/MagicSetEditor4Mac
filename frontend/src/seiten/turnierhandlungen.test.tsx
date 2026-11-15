// **Was man an einem Turnier tut, steht am offenen** — UXR-12
// (#1920).
//
// Jede Zeile der Turnierliste trug *Öffnen*, *Turnier löschen* und die
// Marken: bei acht Turnieren 24 Bedienelemente in der Liste (R2 in
// `docs/UIREVIEW.md`), und das Löschen — das seltenste, was man an einem
// Turnier tut — stand an jedem. Dasselbe Muster wie an der Deckliste
// (UXR-11): die Zeile öffnet, das offene Turnier trägt seine Handlungen.
import { fireEvent, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { alleTurniere, aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { abend, liste, rangliste } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import { turnierhandlung, turniermenue } from '../test/turnierhandlung'
import { effekteNachholen } from '../test/effekte'
import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// amAbend schlägt den Dämmerpokal auf, mit sieben weiteren in der Liste.
function amAbend(weitere: Partial<Api> = {}) {
  return aufgeschlagen(abend, {
    Turniere: () => Promise.resolve({ turniere: liste, fehler: null }),
    TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
    ...weitere,
  })
}

// menueNennt prüft die Einträge des offenen Menüs an ihren Namen — das
// Zeichen davor ist Schmuck.
function menueNennt(namen: readonly string[]) {
  expect(screen.getAllByRole('menuitem')).toHaveLength(namen.length)
  for (const name of namen) {
    expect(screen.getByRole('menuitem', { name })).toBeInTheDocument()
  }
}

// zeilen gibt die Zeilen der Turnierliste.
function zeilen() {
  return within(screen.getByRole('region', { name: texte.titel })).getAllByRole(
    'listitem',
  )
}

describe('die Handlungen am offenen Turnier', () => {
  it('lässt in der Liste je Turnier nur Öffnen und die Marken', async () => {
    await amAbend()
    await alleTurniere()

    expect(zeilen()).toHaveLength(8)
    for (const zeile of zeilen()) {
      expect(within(zeile).getAllByRole('button')).toHaveLength(2)
    }
    expect(
      screen.queryByRole('button', {
        name: `${texte.loeschen}: Aprilturnier`,
      }),
    ).toBeNull()
  })

  // **„Abschließen“ bleibt, das übrige steht hinter „…“** — AP-6 (#2363),
  // B-04: Ausgeben, Bericht sichern, Abschließen und Löschen standen
  // gleichrangig nebeneinander über den Laschen. **Und der Rechtsklick auf
  // den Kopf zeigt dieselbe Liste** wie der Knopf (Entwurf `Kontext09`).
  it('legt Ausgeben, Bericht und Löschen hinter „…“ und lässt Abschließen stehen', async () => {
    const werkbank = await amAbend()

    expect(
      within(werkbank).getByRole('button', { name: texte.abschliessen }),
    ).toBeInTheDocument()
    for (const name of [
      texte.ausgeben,
      texte.berichtSichern,
      `${texte.loeschen}: ${abend.name}`,
    ]) {
      expect(within(werkbank).queryByRole('button', { name })).toBeNull()
    }

    await userEvent.click(await turniermenue(abend.name))
    menueNennt([texte.ausgeben, texte.berichtSichern, texte.loeschen])
    await userEvent.keyboard('{Escape}')

    fireEvent.contextMenu(
      within(werkbank).getByRole('heading', { name: abend.name }),
    )
    await screen.findAllByRole('menuitem')
    menueNennt([texte.ausgeben, texte.berichtSichern, texte.loeschen])
  })

  it('löscht das offene Turnier erst nach Rückfrage', async () => {
    const TurnierLoeschen = vi.fn<Api['TurnierLoeschen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const werkbank = await amAbend({ TurnierLoeschen })

    await turnierhandlung(abend.name, texte.loeschen)
    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.loeschenNein }),
    )
    expect(TurnierLoeschen).not.toHaveBeenCalled()

    await turnierhandlung(abend.name, texte.loeschen)
    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(TurnierLoeschen).toHaveBeenCalledWith('7')
    })
    await waitFor(() => {
      expect(screen.queryByRole('region', { name: abend.name })).toBeNull()
    })
  })

  // **Wer A löscht und vor der Antwort B öffnet, behält B** (§7, Regel 2).
  // Seit UXR-12 steht Löschen am offenen Turnier; den Schutz in `vergiss`
  // betrat vorher das Löschen aus der Liste, jetzt nur noch dieser Lauf.
  it('schließt nach einem späten Löschen nicht das inzwischen geöffnete', async () => {
    let loesche: (() => void) | undefined
    const TurnierLoeschen = vi.fn<Api['TurnierLoeschen']>(
      () =>
        new Promise((fertig) => {
          loesche = () => {
            fertig({ fehler: null })
          }
        }),
    )
    const Turniere = vi.fn<Api['Turniere']>(() =>
      Promise.resolve({ turniere: liste, fehler: null }),
    )
    const werkbank = await amAbend({
      TurnierLoeschen,
      Turniere,
      Turnier: (kennung: string) =>
        Promise.resolve({
          turnier: liste.find((turnier) => turnier.id === kennung) ?? null,
          fehler: null,
        }),
    })

    await turnierhandlung(abend.name, texte.loeschen)
    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.loeschenJa }),
    )

    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Märzturnier` }),
    )
    await screen.findByRole('region', { name: 'Märzturnier' })

    // **Das Löschen ist unterwegs** — gerufen und noch nicht beantwortet.
    expect(TurnierLoeschen).toHaveBeenCalledWith('7')
    const vorher = Turniere.mock.calls.length

    loesche?.()

    // **Erst der Beleg, dass der Löschweg zu Ende lief** — er lädt die Liste
    // nach; erst danach sagt das Stehenbleiben etwas.
    await waitFor(() => {
      expect(Turniere.mock.calls.length).toBeGreaterThan(vorher)
    })
    await effekteNachholen()

    expect(screen.getByRole('region', { name: 'Märzturnier' })).toBeVisible()
  })
})
