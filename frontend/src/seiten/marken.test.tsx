// **Marken an Karten** — Werkstück 3 von Phase 3.5
// (#1208).
//
// Eine Marke ist ein Wort für die eigene Arbeit — „fertig", „prüfen". Sie
// steht **neben** der Karte und nicht in ihr.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'
import { lageAus, vergessen } from '../speicherstand'

afterEach(() => {
  raeumeApi()
  vergessen()
})

const texte = {
  marken: 'Marken',
  setzen: 'Marke setzen',
  filter: 'Nach Marke filtern',
}

// zwei gibt die Kartenwege über zwei Karten.
function zwei() {
  const karten = [
    { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Erste' }] },
    { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Zweite' }] },
  ]

  return {
    ...ausKarten(() => Promise.resolve({ karten })),
    Karten: () => Promise.resolve({ karten }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
  }
}

// mitMarken belegt die Fassade samt Marken an Karte 1.
function mitMarken(marken: Record<string, string[]>) {
  const proben = zwei()

  const setzen = vi.fn().mockResolvedValue({ fehler: null })
  const entfernen = vi.fn().mockResolvedValue({ fehler: null })

  belegeApi({
    ...proben,
    Kartenlistenaufbau: async (setID: string) => {
      const aufbau = await proben.Kartenlistenaufbau(setID)

      return {
        ...aufbau,
        zeilen: aufbau.zeilen.map((zeile) => ({
          ...zeile,
          marken: marken[zeile.karteID] ?? [],
        })),
      }
    },
    MarkeSetzen: setzen,
    MarkeEntfernen: entfernen,
  })

  return { setzen, entfernen }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

// **Die Marken einer Karte stehen in ihrer Zeile.**
//
// **In der Zeile gesucht und nicht auf der Seite**: „fertig" steht auch am
// Filter darüber, und ein Test, der beide trifft, sagt nicht, welchen er
// meint.
it('zeigt die Marken in der Kartenliste', async () => {
  mitMarken({ '1': ['fertig'] })

  zeige()

  const zeile = await screen.findByRole('button', { name: /Erste/ })

  expect(within(zeile).getByText('fertig')).toBeInTheDocument()

  // **Und an der Karte ohne Marke steht keine.**
  expect(
    within(screen.getByRole('button', { name: /Zweite/ })).queryByText(
      'fertig',
    ),
  ).not.toBeInTheDocument()
})

// **Und wer eine setzt, schreibt sie über die Bindung.**
it('setzt eine Marke an der aufgeschlagenen Karte', async () => {
  const { setzen } = mitMarken({})

  zeige()

  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))

  const feld = await screen.findByLabelText(texte.setzen)

  await userEvent.type(feld, 'prüfen')
  await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

  await waitFor(() => {
    expect(setzen).toHaveBeenCalledWith('1', '1', 'prüfen')
  })
})

// **Und wer eine abnimmt, nennt sie beim Namen.**
//
// „Abnehmen" allein sagte an fünf Knöpfen nebeneinander nicht, welcher welcher
// ist.
it('nimmt eine Marke wieder ab', async () => {
  const { entfernen } = mitMarken({ '1': ['fertig'] })

  zeige()

  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))

  await userEvent.click(
    await screen.findByRole('button', { name: 'Marke fertig abnehmen' }),
  )

  await waitFor(() => {
    expect(entfernen).toHaveBeenCalledWith('1', '1', 'fertig')
  })
})

// **Der Filter nimmt weg, was die Marke nicht trägt.**
it('filtert die Liste nach einer Marke', async () => {
  mitMarken({ '1': ['fertig'] })

  zeige()

  expect(
    await screen.findByRole('button', { name: /Zweite/ }),
  ).toBeInTheDocument()

  await userEvent.click(await screen.findByRole('radio', { name: 'fertig' }))

  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: /Zweite/ }),
    ).not.toBeInTheDocument()
  })

  expect(screen.getByRole('button', { name: /Erste/ })).toBeInTheDocument()

  // **Und »Alle« bringt sie zurück.**
  await userEvent.click(screen.getByRole('radio', { name: 'Alle' }))

  expect(
    await screen.findByRole('button', { name: /Zweite/ }),
  ).toBeInTheDocument()
})

// **Ohne Marken im Set zeichnet der Filter nichts.**
//
// Ein Filter über nichts ist eine Zeile, die erklärt werden müsste.
it('lässt den Filter weg, wenn es keine Marken gibt', async () => {
  mitMarken({})

  zeige()

  expect(
    await screen.findByRole('button', { name: /Erste/ }),
  ).toBeInTheDocument()
  expect(screen.queryByRole('radio', { name: 'Alle' })).not.toBeInTheDocument()
})

// **Und die Überschrift des Kastens steht da, sobald eine Karte offen ist.**
it('zeigt den Markenkasten an der aufgeschlagenen Karte', async () => {
  mitMarken({})

  zeige()

  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))

  expect(
    await screen.findByRole('region', { name: texte.marken }),
  ).toBeInTheDocument()
})

// **Ein gescheitertes Setzen sagt, warum** (§8).
//
// Ohne die Meldung stünde der Entwurf weiter im Feld, die Marke fehlte, und
// niemand wüsste, ob der Knopf gedrückt wurde.
it('meldet einen gescheiterten Schreibvorgang', async () => {
  const proben = zwei()

  belegeApi({
    ...proben,
    MarkeSetzen: () =>
      Promise.resolve({
        fehler: {
          code: 'validierung',
          meldung: 'Zu lang.',
          feld: 'marke',
          schluessel: '',
        },
      }),
  })

  zeige()

  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))

  await userEvent.type(await screen.findByLabelText(texte.setzen), 'x')
  await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

  // **Im Markenkasten gesucht**: die Seite trägt mehr als eine Meldung, und
  // ein Test, der irgendeine nimmt, prüft irgendetwas.
  const kasten = screen.getByRole('region', { name: texte.marken })

  await waitFor(() => {
    expect(within(kasten).getByRole('alert')).toHaveTextContent('Zu lang.')
  })

  // **Und der Speicherstand weiß es** — die Marken schrieben bis zum
  // Zweitblick über Phase 6.3 an ihm vorbei (M8).
  expect(lageAus('1')).toEqual({ art: 'fehler', feld: texte.marken })
})

// **Eine eben gesetzte Marke steht sofort in der Liste** (#1208).
//
// # Warum dieser Test sein muss
//
// Der erste Anlauf lud über `notizGeaendert` nach — und das holt nur die
// **aufgeschlagene Karte**. Die Marken reisen mit den **Zeilen** der Liste;
// die eben gesetzte blieb unsichtbar, bis jemand das Set wechselte. **Gesehen
// hat es das laufende Fenster und keiner der Tests** (§3), und deshalb steht
// hier einer.
it('zeigt eine eben gesetzte Marke ohne Umweg', async () => {
  const proben = zwei()

  // **Die Vorrichtung antwortet wie der Bestand**: nach dem Schreiben trägt
  // die Zeile die Marke.
  let gesetzt = false

  belegeApi({
    ...proben,
    Kartenlistenaufbau: async (setID: string) => {
      const aufbau = await proben.Kartenlistenaufbau(setID)

      return {
        ...aufbau,
        zeilen: aufbau.zeilen.map((zeile) => ({
          ...zeile,
          marken: gesetzt && zeile.karteID === '1' ? ['fertig'] : [],
        })),
      }
    },
    MarkeSetzen: () => {
      gesetzt = true

      return Promise.resolve({ fehler: null })
    },
  })

  zeige()

  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))

  await userEvent.type(await screen.findByLabelText(texte.setzen), 'fertig')
  await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

  await waitFor(() => {
    expect(
      within(screen.getByRole('button', { name: /Erste/ })).getByText('fertig'),
    ).toBeInTheDocument()
  })
})

// **Verschwindet die gefilterte Marke, sieht man wieder alle** (#1208).
//
// # Warum dieser Test sein muss
//
// Wer die letzte Karte entmarkiert, nimmt die Marke aus dem Set — die
// Kachelreihe zeichnet sie nicht mehr. Blieb die Wahl im Zustand stehen, war
// die Liste **leer und ohne Weg zurück**: der Knopf für »Alle« war mit ihr
// verschwunden. **Gesehen hat es das laufende Fenster** (§3).
it('gibt den Filter frei, wenn seine Marke verschwindet', async () => {
  const proben = zwei()

  let markiert = true

  belegeApi({
    ...proben,
    Kartenlistenaufbau: async (setID: string) => {
      const aufbau = await proben.Kartenlistenaufbau(setID)

      return {
        ...aufbau,
        zeilen: aufbau.zeilen.map((zeile) => ({
          ...zeile,
          marken: markiert && zeile.karteID === '1' ? ['fertig'] : [],
        })),
      }
    },
    MarkeEntfernen: () => {
      markiert = false

      return Promise.resolve({ fehler: null })
    },
  })

  zeige()

  await userEvent.click(await screen.findByRole('radio', { name: 'fertig' }))

  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: /Zweite/ }),
    ).not.toBeInTheDocument()
  })

  // **Die Marke abnehmen** — danach gibt es sie im Set nicht mehr.
  await userEvent.click(await screen.findByRole('button', { name: /Erste/ }))
  await userEvent.click(
    await screen.findByRole('button', { name: 'Marke fertig abnehmen' }),
  )

  // **Und beide Karten stehen wieder da.**
  expect(
    await screen.findByRole('button', { name: /Zweite/ }),
  ).toBeInTheDocument()
})
