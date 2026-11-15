// **Massenvorgänge über eine Auswahl** — Werkstück 6 von Phase 4.5
// (#1292).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Massenvorgaenge } from './Massenvorgaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
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
it('nennt die Zahl der gewählten Zeilen', () => {
  zeige(zwei)

  expect(screen.getByText('2 von 5 gewählt')).toBeInTheDocument()
})

// **Die Fassungen gehen mit** (§7, Regel 2).
//
// Ohne sie überschriebe ein Massenvorgang jede Änderung seit dem Laden der
// Liste — fünfzigfach auf einmal.
it('schickt jede Zeile samt ihrer Fassung', async () => {
  const aendern = vi.fn(() =>
    Promise.resolve({
      bericht: { geaendert: 2, befunde: [], vorher: [] },
      fehler: null,
    }),
  )

  zeige(zwei, { MasseAendern: aendern })

  await userEvent.type(
    screen.getByLabelText(texte.masseLagerortSetzen),
    'Kiste 9',
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  await waitFor(() => {
    expect(aendern).toHaveBeenCalledWith({
      zeilen: zwei,
      menge: null,
      zustand: null,
      lagerort: 'Kiste 9',
      wunsch: null,
    })
  })
})

// **Ein leerer Lagerort ist eine Angabe**, keine Auslassung.
it('schickt auch einen leeren Lagerort', async () => {
  const aendern = vi.fn(() =>
    Promise.resolve({
      bericht: { geaendert: 2, befunde: [], vorher: [] },
      fehler: null,
    }),
  )

  zeige(zwei, { MasseAendern: aendern })

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  await waitFor(() => {
    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({ lagerort: '' }),
    )
  })
})

// **Der Bericht steht danach da.**
it('sagt, wie viele geändert wurden', async () => {
  zeige(zwei)

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  expect(await screen.findByText('2 geändert.')).toBeInTheDocument()
})

// **Ohne Auswahl steht nichts da, was eine bräuchte** — UXR-10 (#1918).
// Bis hierher standen Feld und Knöpfe da und sagten auf Klick nur
// „Nichts gewählt."; die Preise gelten allen und bleiben.
it('zeigt ohne Auswahl weder Lagerort noch Ändern und Löschen', () => {
  zeige([])

  expect(screen.getByText('0 von 5 gewählt')).toBeInTheDocument()
  expect(screen.queryByLabelText(texte.masseLagerortSetzen)).toBeNull()
  expect(screen.queryByRole('button', { name: texte.masseAendern })).toBeNull()
  expect(screen.queryByRole('button', { name: texte.masseLoeschen })).toBeNull()
  expect(
    screen.getByRole('button', { name: texte.massePreise }),
  ).toBeInTheDocument()
})

// **Löschen ruft den anderen Weg.**
it('löscht die gewählten Zeilen', async () => {
  const loeschen = vi.fn(() =>
    Promise.resolve({
      bericht: { geaendert: 2, befunde: [], vorher: [] },
      fehler: null,
    }),
  )

  zeige(zwei, { MasseLoeschen: loeschen })

  // **Seit UX-10.AK-04 liegt eine Rückfrage dazwischen.** Sie gehört hierher
  // und nicht weg: dieser Lauf prüft den Weg bis zum Wirt, und der geht jetzt
  // über eine Bestätigung.
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

// **Der Preisschub sagt, was bleibt** — daran hängt der Fortschritt.
it('nennt den Rest nach einem Preisschub', async () => {
  zeige(zwei)

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  expect(await screen.findByText('3 geholt, 7 offen')).toBeInTheDocument()
})

// **Und wenn nichts mehr bleibt, sagt er das.**
it('sagt, wenn alle Preise da sind', async () => {
  zeige(zwei, {
    MassePreise: () =>
      Promise.resolve({
        schub: {
          geholt: 0,
          offen: 0,
          fehlgeschlagen: 0,
          fertig: true,
          steht: false,
          befunde: [],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  expect(await screen.findByText('Alle Preise sind da.')).toBeInTheDocument()
})

// **Der Preisschub braucht keine Auswahl** — er gilt der ganzen Sammlung.
it('holt Preise auch ohne Auswahl', async () => {
  const preise = vi.fn(() =>
    Promise.resolve({
      schub: {
        geholt: 1,
        offen: 0,
        fehlgeschlagen: 0,
        fertig: true,
        steht: false,
        befunde: [],
      },
      fehler: null,
    }),
  )

  zeige([], { MassePreise: preise })

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  await waitFor(() => {
    expect(preise).toHaveBeenCalledWith(24)
  })
})

// **Nach jedem geglückten Vorgang wird nachgeladen** (§7, Regel 3).
it('lädt nach einem Vorgang nach', async () => {
  const aufGeaendert = vi.fn()

  zeige(zwei, {}, aufGeaendert)

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  await waitFor(() => {
    expect(aufGeaendert).toHaveBeenCalled()
  })
})

// **Ein Fehler steht am Kasten** (§6).
it('zeigt einen Fehler', async () => {
  zeige(zwei, {
    MasseAendern: () =>
      Promise.resolve({
        bericht: null,
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Das ist kein Zustand',
          feld: '',
        },
      }),
  })

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Das ist kein Zustand',
  )
})

// **Und einer beim Preisschub ebenso.**
it('zeigt einen Fehler beim Preisschub', async () => {
  zeige(zwei, {
    MassePreise: () =>
      Promise.resolve({
        schub: null,
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Kein Netz',
          feld: '',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Kein Netz')
})

// **Ein Bericht, der `null` ist, gibt keine Zeile** — und keinen Absturz.
it('kommt mit einem leeren Bericht zurecht', async () => {
  zeige(zwei, {
    MasseAendern: () => Promise.resolve({ bericht: null, fehler: null }),
    MassePreise: () => Promise.resolve({ schub: null, fehler: null }),
  })

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )
  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  await waitFor(() => {
    expect(screen.queryByText(/geändert\./)).not.toBeInTheDocument()
  })
})

// **Was nicht durchkam, steht da** — Befund 4 des Zweitblicks über Phase 4.5.
//
// # Warum dieser Test fehlte
//
// Weil **jede** Vorrichtung `befunde: []` gab. Die DTOs trugen das Feld, die
// Oberfläche las es nicht, und kein Lauf hat es bemerkt: bei einem
// Fassungskonflikt stand nur »N geändert«, und die übergangene Zeile war
// unsichtbar. Danach räumte die Seite die Auswahl — auch die fehlgeschlagene.
it('zeigt die Zeilen, die nicht durchkamen', async () => {
  zeige(zwei, {
    MasseAendern: () =>
      Promise.resolve({
        bericht: {
          geaendert: 1,
          befunde: ['Exemplar 2: Konflikt mit dem gegenwärtigen Stand'],
          vorher: [],
        },
        fehler: null,
      }),
  })

  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  expect(
    await screen.findByText('Exemplar 2: Konflikt mit dem gegenwärtigen Stand'),
  ).toBeInTheDocument()
})
