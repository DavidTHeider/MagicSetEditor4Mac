// **Ein Massenlöschen fragt vorher** — UX-10.AK-04
// (#1682).
//
// Aus `Massenvorgaenge.test.tsx` herausgelöst, als die Datei über 400 Zeilen
// wuchs (§2). Der Schnitt folgt der Frage: dort steht, **was ein Massenvorgang
// tut** — ändern, bepreisen, berichten; hier, **was vor dem Löschen steht**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Massenvorgaenge } from './Massenvorgaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { MassenzeileDTO } from '../antwortensammlung'

afterEach(() => {
  raeumeApi()
})

// **Die Aufschriften kommen aus dem Katalog und nicht aus der Erinnerung** —
// #1551.
//
// Hier standen vier abgeschriebene Zeichenketten, und zwei davon — „Ändern"
// und „Löschen" — waren die des **Inspektors** und nicht die dieses Kastens.
// Getroffen haben sie trotzdem, weil beide Paare dasselbe deutsche Wort
// trugen; im englischen Katalog hätte derselbe Griff nichts gefunden.
const texte = kataloge.de.seiten.sammlung

// zwei sind zwei gewählte Zeilen samt ihrer gelesenen Fassung.
const zwei: MassenzeileDTO[] = [
  { id: 1, fassung: 3 },
  { id: 2, fassung: 1 },
]

// zeige stellt den Kasten auf.
function zeige(
  gewaehlt: MassenzeileDTO[],
  zusatz = {},
  aufGeaendert = vi.fn(),
) {
  belegeApi({
    MasseAendern: () =>
      Promise.resolve({
        bericht: { geaendert: gewaehlt.length, befunde: [], vorher: [] },
        fehler: null,
      }),
    MasseLoeschen: () =>
      Promise.resolve({
        bericht: { geaendert: gewaehlt.length, befunde: [], vorher: [] },
        fehler: null,
      }),
    MassePreise: () =>
      Promise.resolve({
        schub: {
          geholt: 3,
          offen: 7,
          fehlgeschlagen: 0,
          fertig: false,
          steht: false,
          befunde: [],
        },
        fehler: null,
      }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Massenvorgaenge
        gewaehlt={gewaehlt}
        gesamt={5}
        aufGeaendert={aufGeaendert}
      />
    </Sprachraum>,
  )

  return aufGeaendert
}

// **Wie viele gewählt sind, steht da.**
describe('Ausgewählte löschen', () => {
  it('fragt, bevor es löscht', async () => {
    const loeschen = vi.fn(() =>
      Promise.resolve({
        bericht: { geaendert: 2, befunde: [], vorher: [] },
        fehler: null,
      }),
    )

    zeige(zwei, { MasseLoeschen: loeschen })

    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschen }),
    )

    // **Der Beleg, dass die Frage dasteht** — und erst danach die Verneinung
    // (§3): ohne ihn wäre ein „nicht gerufen" auch dann grün, wenn der Knopf
    // gar nichts täte.
    expect(
      await screen.findByText(
        einsetzen(texte.masseLoeschenFrage, { anzahl: '2' }),
      ),
    ).toBeVisible()
    expect(loeschen).not.toHaveBeenCalled()
  })

  it('löscht erst nach dem Bestätigen', async () => {
    const loeschen = vi.fn(() =>
      Promise.resolve({
        bericht: { geaendert: 2, befunde: [], vorher: [] },
        fehler: null,
      }),
    )

    zeige(zwei, { MasseLoeschen: loeschen })

    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.masseLoeschenJa }),
    )

    await waitFor(() => {
      expect(loeschen).toHaveBeenCalledWith(zwei)
    })
  })

  // **Abbrechen lässt die Auswahl stehen** — und räumt die Frage weg, statt
  // sie stehen zu lassen (§7, Regel 1).
  it('lässt die Auswahl stehen, wenn jemand abbricht', async () => {
    const loeschen = vi.fn(() =>
      Promise.resolve({
        bericht: { geaendert: 2, befunde: [], vorher: [] },
        fehler: null,
      }),
    )

    zeige(zwei, { MasseLoeschen: loeschen })

    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschen }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )

    await waitFor(() => {
      expect(
        screen.queryByText(
          einsetzen(texte.masseLoeschenFrage, { anzahl: '2' }),
        ),
      ).toBeNull()
    })

    expect(loeschen).not.toHaveBeenCalled()
  })

  // **Ein Wechsel der Auswahl schließt die Rückfrage** (§7, Regel 1) — sonst
  // bestätigte, wer „2 löschen?" las, das Löschen einer anderen Auswahl
  // (Befund 3 des Nachlaufs über UX-12).
  it('schließt die Rückfrage, wenn sich die Auswahl ändert', async () => {
    const loeschen = vi.fn(() =>
      Promise.resolve({
        bericht: { geaendert: 1, befunde: [], vorher: [] },
        fehler: null,
      }),
    )
    const drei: MassenzeileDTO[] = [...zwei, { id: 3, fassung: 2 }]
    belegeApi({ MasseLoeschen: loeschen })

    const kasten = (gewaehlt: MassenzeileDTO[]) => (
      <Sprachraum sprache="de">
        <Massenvorgaenge
          gewaehlt={gewaehlt}
          gesamt={5}
          aufGeaendert={vi.fn()}
        />
      </Sprachraum>
    )
    const { rerender } = render(kasten(zwei))

    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschen }),
    )
    expect(
      await screen.findByRole('button', { name: texte.masseLoeschenJa }),
    ).toBeVisible()

    rerender(kasten(drei))

    expect(
      screen.queryByRole('button', { name: texte.masseLoeschenJa }),
    ).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschen }),
    )
    expect(
      await screen.findByText(
        einsetzen(texte.masseLoeschenFrage, { anzahl: '3' }),
      ),
    ).toBeVisible()
    await userEvent.click(
      screen.getByRole('button', { name: texte.masseLoeschenJa }),
    )
    await waitFor(() => {
      expect(loeschen).toHaveBeenCalledWith(drei)
    })
  })
})
