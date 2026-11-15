// **Die Auswahl auf der Sammlungsseite** — Werkstück 6 von Phase 4.5
// (#1292).
//
// # Warum die Auswahl an der Liste hängt und nicht am Kasten
//
// Weil eine gelöschte Zeile aus ihr verschwinden muss. Ein Kasten, der sie
// hielte, hätte einen zweiten Halter derselben Sache (§7, Regel 2) — und nach
// einem Massenvorgang zeigte er auf Kennungen, die es nicht mehr gibt.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { ExemplarDTO } from '../api'
import { Sammlung } from './Sammlung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

// **Die Aufschriften kommen aus dem Katalog** (#1551) — hier standen „Ändern"
// und „Löschen" als Zeichenketten, und beide Wörter trug bis zum 2026-09-13
// auch der Inspektor daneben.
const texte = kataloge.de.seiten.sammlung

// zwei gibt zwei Zeilen mit verschiedenen Fassungen.
const zwei: ExemplarDTO[] = [
  {
    id: 1,
    fassung: 3,
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
    lagerort: 'Kiste 4',
    wunsch: false,
    bezeichnung: 'Lightning Bolt · M11 #149',
    herkunft: keineHerkunft,
  },
  {
    id: 2,
    fassung: 1,
    printing: 'def',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
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
    wunsch: false,
    bezeichnung: 'Counterspell · LEA #54',
    herkunft: keineHerkunft,
  },
]

// zeige stellt die Seite auf.
function zeige(zusatz = {}) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: zwei, fehler: null }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// **Was angekreuzt ist, geht samt seiner Fassung mit.**
it('schickt die gewählte Zeile samt ihrer Fassung', async () => {
  const aendern = vi.fn(() =>
    Promise.resolve({
      bericht: { geaendert: 1, befunde: [], vorher: [] },
      fehler: null,
    }),
  )

  zeige({ MasseAendern: aendern })

  await userEvent.click(
    await screen.findByLabelText('Lightning Bolt · M11 #149'),
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )

  await waitFor(() => {
    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({ zeilen: [{ id: 1, fassung: 3 }] }),
    )
  })
})

// **Ein zweites Ankreuzen nimmt die Zeile wieder heraus.**
it('nimmt eine Zeile beim zweiten Klick wieder heraus', async () => {
  zeige()

  const feld = await screen.findByLabelText('Lightning Bolt · M11 #149')

  await userEvent.click(feld)
  expect(await screen.findByText('1 von 2 gewählt')).toBeInTheDocument()

  await userEvent.click(feld)
  expect(await screen.findByText('0 von 2 gewählt')).toBeInTheDocument()
})

// **Nach einem Massenvorgang ist die Auswahl leer** (§7, Regel 1).
//
// Die Kennungen zeigten sonst auf einen Stand, den es so nicht mehr gibt —
// gelöschte Zeilen zuerst.
it('räumt die Auswahl nach einem Vorgang', async () => {
  zeige({
    MasseLoeschen: () =>
      Promise.resolve({
        bericht: { geaendert: 1, befunde: [], vorher: [] },
        fehler: null,
      }),
  })

  await userEvent.click(
    await screen.findByLabelText('Lightning Bolt · M11 #149'),
  )
  await screen.findByText('1 von 2 gewählt')

  // **Seit UX-10.AK-04 liegt eine Rückfrage dazwischen** — der Vorgang
  // beginnt erst nach dem Bestätigen, und erst danach räumt er die Auswahl.
  await userEvent.click(
    screen.getByRole('button', { name: texte.masseLoeschen }),
  )
  await userEvent.click(
    await screen.findByRole('button', { name: texte.masseLoeschenJa }),
  )

  expect(await screen.findByText('0 von 2 gewählt')).toBeInTheDocument()
})

// **Ein Sichtwechsel räumt die Auswahl** — Nachlauf über Phase 4.5, Befund 5.
//
// # Warum dieser Lauf fehlte
//
// Behoben war es: der gemeinsame `setze`-Rückruf räumt die Auswahl bei jedem
// Wechsel. **Nur prüfte es niemand.** Die Läufe daneben prüfen Übermittlung,
// Abwählen und das Räumen nach einem Massenvorgang — keiner einen
// Filterwechsel. Die Korrektur war damit unbewacht: wer sie entfernt, sieht
// nichts.
//
// Der Schaden, gegen den sie steht: die Auswahl wird aus **allen** Exemplaren
// rekonstruiert, nicht aus der sichtbaren Liste. Eine gewählte, dann
// ausgefilterte Karte blieb im Massenkasten und ließ sich ändern oder löschen,
// ohne dass sie irgendwo stand (§7, Regel 1).
it('räumt die Auswahl, wenn die Sicht wechselt', async () => {
  const aendern = vi.fn(() =>
    Promise.resolve({
      bericht: { geaendert: 1, befunde: [], vorher: [] },
      fehler: null,
    }),
  )

  zeige({ MasseAendern: aendern })

  await userEvent.click(
    await screen.findByLabelText('Lightning Bolt · M11 #149'),
  )

  // **Der Massenkasten zählt sie.**
  expect(screen.getByText('1 von 2 gewählt')).toBeInTheDocument()

  // **Und die Suche filtert die gewählte Zeile weg.**
  await userEvent.type(screen.getByLabelText('Suchen'), 'Counterspell')

  // **Danach ist nichts mehr gewählt.** Ohne das Räumen stünde hier weiter
  // »1 von 2« — und »Ändern« träfe eine Zeile, die niemand sieht.
  expect(screen.getByText('0 von 2 gewählt')).toBeInTheDocument()
})
