// **Die Ordnung einer Sammlung** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk 0057).
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sammlungsordnung } from './Sammlungsordnung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = {
  lagerorte: 'Lagerorte',
  dubletten: 'Dubletten',
  wuensche: 'Wunschliste',
  behalten: 'Behalten',
}

// zeige stellt den Kasten mit einer belegten Fassade auf.
function zeige(zusatz = {}) {
  belegeApi({
    SammlungLagerorte: () =>
      Promise.resolve({
        lagerorte: [{ lagerort: 'Kiste 4', zeilen: 2, stueck: 7 }],
        fehler: null,
      }),
    SammlungDubletten: () =>
      Promise.resolve({
        dubletten: [
          {
            exemplar: {
              id: 1,
              fassung: 1,
              printing: 'abc',
              druckID: '',
              vorlaeufig: false,
              menge: 4,
              zustand: 'nearMint',
              sprache: 'en',
              finish: 'nonfoil',
              erwerbsdatum: '',
              erwerbspreis: '',
              erwerbswaehrung: '',
              notiz: '',
              name: 'Lightning Bolt',
              setcode: 'm11',
              nummer: '149',
              lagerort: '',
              wunsch: false,
              bezeichnung: 'Lightning Bolt · M11 #149',
              herkunft: keineHerkunft,
            },
            ueberzaehlig: 3,
          },
        ],
        fehler: null,
      }),
    SammlungWuensche: () =>
      Promise.resolve({
        exemplare: [
          {
            id: 2,
            fassung: 1,
            printing: 'def',
            druckID: '',
            vorlaeufig: false,
            menge: 2,
            zustand: 'played',
            sprache: 'en',
            finish: 'nonfoil',
            erwerbsdatum: '',
            erwerbspreis: '',
            erwerbswaehrung: '',
            notiz: '',
            name: 'Counterspell',
            setcode: 'lea',
            nummer: '54',
            lagerort: '',
            wunsch: true,
            bezeichnung: 'Counterspell · LEA #54',
            herkunft: keineHerkunft,
          },
        ],
        fehler: null,
      }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsordnung />
    </Sprachraum>,
  )
}

// **Die Lagerorte stehen mit ihren Zahlen da.**
it('zeigt die Lagerorte', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.lagerorte }))

  expect(await screen.findByText('Kiste 4')).toBeInTheDocument()
  expect(screen.getByText('7 Stück in 2 Zeilen')).toBeInTheDocument()
})

// **Ein Lagerort ohne Angabe heißt so** — und nicht „leer".
it('nennt einen Lagerort ohne Angabe', async () => {
  zeige({
    SammlungLagerorte: () =>
      Promise.resolve({
        lagerorte: [{ lagerort: '', zeilen: 1, stueck: 1 }],
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.lagerorte }))

  expect(await screen.findByText('ohne Angabe')).toBeInTheDocument()
})

// **Die Dubletten nennen, wie viele zu viel sind.**
it('zeigt die Dubletten samt Überzahl', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.dubletten }))

  expect(
    await screen.findByText('Lightning Bolt · M11 #149'),
  ).toBeInTheDocument()
  expect(screen.getByText('3 zu viel')).toBeInTheDocument()
})

// **Die behaltene Zahl geht mit** — sie ist der Kern der Frage (Vermerk 0057).
it('schickt die behaltene Zahl mit', async () => {
  const dubletten = vi.fn(() =>
    Promise.resolve({ dubletten: [], fehler: null }),
  )

  zeige({ SammlungDubletten: dubletten })

  await userEvent.clear(screen.getByLabelText(texte.behalten))
  await userEvent.type(screen.getByLabelText(texte.behalten), '4')
  await userEvent.click(screen.getByRole('button', { name: texte.dubletten }))

  await waitFor(() => {
    expect(dubletten).toHaveBeenCalledWith(4)
  })
})

// **Die Wunschliste steht da.**
it('zeigt die Wunschliste', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.wuensche }))

  expect(await screen.findByText('Counterspell · LEA #54')).toBeInTheDocument()
})

// **Ein Wechsel der Frage räumt die vorige Antwort ab** (§7, Regel 1).
//
// Ohne das stünde unter »Dubletten« noch die Lagerortliste.
it('räumt die vorige Antwort beim Wechsel ab', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.lagerorte }))
  await screen.findByText('Kiste 4')

  await userEvent.click(screen.getByRole('button', { name: texte.wuensche }))
  await screen.findByText('Counterspell · LEA #54')

  expect(screen.queryByText('Kiste 4')).not.toBeInTheDocument()
})

// **Eine leere Antwort erklärt sich** (§6) — je Sicht mit eigenem Satz,
// geprüft in `sammlungsordnungleer.test.tsx` (G02).

// **Ein Fehler steht am Kasten.**
it('zeigt einen Fehler', async () => {
  zeige({
    SammlungDubletten: () =>
      Promise.resolve({
        dubletten: [],
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Das ist keine Stückzahl',
          feld: '',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.dubletten }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Das ist keine Stückzahl',
  )
})

// **Eine negative Stückzahl erreicht den Dienst nicht mehr** — Werkstück U3
// von Phase 5.1 (#1414).
//
// Die Regel steht in `service.Sammlungssichten.Dubletten`: `behalten < 0`
// wird mit `ErrValidation` abgewiesen („%d ist keine Stückzahl, die man
// behält"). Bis heute trug das Feld nur `min={0}` — und `min` ohne Formular
// meldet nichts (§10).
//
// **Die rohe Zeichenkette und nicht der geparste Wert**: getippt wird `-2`,
// und genau das muss am Feld ankommen.
it('meldet eine negative Stückzahl am Feld und schickt sie nicht', async () => {
  const dubletten = vi.fn(() =>
    Promise.resolve({ dubletten: [], fehler: null }),
  )

  zeige({ SammlungDubletten: dubletten })

  const feld = screen.getByLabelText(texte.behalten)
  await userEvent.clear(feld)
  await userEvent.type(feld, '-2')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toBeInTheDocument()

  await userEvent.click(screen.getByRole('button', { name: texte.dubletten }))

  await waitFor(() => {
    expect(dubletten).toHaveBeenCalled()
  })

  // **Die alte Zahl geht mit und nicht die abgewiesene** (§7, Regel 3): was
  // nicht gilt, erreicht den Aufrufer nicht.
  expect(dubletten).not.toHaveBeenCalledWith(-2)
})

// **Und der Grenzfall gilt**: null ist eine zulässige Stückzahl — „behalte
// keine" ist eine Frage, die der Dienst beantwortet.
it('lässt null als Stückzahl durch', async () => {
  const dubletten = vi.fn(() =>
    Promise.resolve({ dubletten: [], fehler: null }),
  )

  zeige({ SammlungDubletten: dubletten })

  const feld = screen.getByLabelText(texte.behalten)
  await userEvent.clear(feld)
  await userEvent.type(feld, '0')

  expect(feld).toHaveAttribute('aria-invalid', 'false')

  await userEvent.click(screen.getByRole('button', { name: texte.dubletten }))

  await waitFor(() => {
    expect(dubletten).toHaveBeenCalledWith(0)
  })
})
