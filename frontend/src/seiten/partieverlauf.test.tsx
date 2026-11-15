// **Was bisher geschah** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
//   - Der Verlauf wird **auf Zuruf** geholt und nicht von selbst — sonst
//     liefe er nach jedem Zug mit.
//   - Er wird **für den gezeigten Sitz** geholt: die Sicht ist ein Argument
//     und keine Annahme, und beim Spiel gegen sich selbst hängt alles daran
//     (G8, #1465).
//   - Ein Sitzwechsel räumt ihn — über den `key`, der §7 Regel 1 und 2
//     zugleich einlöst.
//   - Die vier Zustände sind vier: läuft, ungeholt, leer, voll.
import { render, screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Partiefuehrung } from './Partiefuehrung'
import { Partieverlauf } from './Partieverlauf'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { einLauf, einePartie } from './partieprobe'

import type { Spielereignis } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

// einEreignis gibt einen Zug, wie ihn die Naht schickt.
function einEreignis(teile: Partial<Spielereignis> = {}): Spielereignis {
  return {
    folge: '1',
    art: 'objekt.ziehen',
    oeffentlich: { sitz: 'anna' },
    privat: {},
    kette: 'glied-1',
    ...teile,
  }
}

// zeige zeichnet die Rückschau und gibt die Hand.
function zeige(fuer = 'anna') {
  render(
    <Sprachraum sprache="de">
      <Partieverlauf kennung="p1" fuer={fuer} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Der Abschnitt trägt seinen Namen.**
//
// **Sonst ist ein `section` für ein Vorleseprogramm kein Bereich** — und
// dieser hier ist einer, den man anspringen will: er steht am Ende einer
// langen Seite.
it('nennt den Abschnitt bei seinem Namen', () => {
  belegeApi({ PartieVerlauf: vi.fn() })

  zeige()

  expect(
    screen.getByRole('region', { name: eigene.verlauf }),
  ).toBeInTheDocument()
})

// **Ohne Klick geht nichts hinaus.**
//
// **Und der Beleg dafür ist nicht die Abwesenheit allein** (§3): der Satz für
// den ungeholten Stand muss dastehen, sonst hieße „nicht gerufen" auch „gar
// nicht gezeichnet".
it('holt den Verlauf nicht von selbst', () => {
  const geholt = vi.fn()
  belegeApi({ PartieVerlauf: geholt })

  zeige()

  expect(screen.getByText(eigene.verlaufUngeholt)).toBeInTheDocument()
  expect(geholt).not.toHaveBeenCalled()
})

// **Auf Zuruf geht er hinaus — mit Partie und Sitz.**
it('holt den Verlauf für den gezeigten Sitz', async () => {
  const geholt = vi.fn().mockResolvedValue({ verlauf: [], fehler: null })
  belegeApi({ PartieVerlauf: geholt })

  const hand = zeige('bert')

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))

  await waitFor(() => {
    expect(screen.getByText(eigene.verlaufLeer)).toBeInTheDocument()
  })

  expect(geholt).toHaveBeenCalledWith('p1', 'bert')
})

// **Ein voller Verlauf steht als geordnete Liste da** — mit Folge und Art.
it('zeigt jeden Zug mit seiner Folge', async () => {
  belegeApi({
    PartieVerlauf: vi.fn().mockResolvedValue({
      verlauf: [
        einEreignis(),
        einEreignis({
          folge: '2',
          art: 'zone.blick',
          privat: { karten: ['Wald'] },
        }),
      ],
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))

  const zeilen = await screen.findAllByRole('listitem')

  expect(zeilen).toHaveLength(2)
  expect(zeilen[0]).toHaveTextContent('1. objekt.ziehen')
  expect(zeilen[1]).toHaveTextContent('2. zone.blick')

  // **Die eigene Ergänzung steht dabei** — sie ist der Grund, warum ein
  // Verlauf überhaupt empfängergebunden ist.
  expect(zeilen[1]).toHaveTextContent(eigene.ereignisPrivat)
  expect(zeilen[0]).not.toHaveTextContent(eigene.ereignisPrivat)

  // **Und die Grenze steht darunter**, damit man ihr ansieht, dass es sie
  // gibt.
  expect(screen.getByText(eigene.verlaufGrenze)).toBeInTheDocument()
})

// **Nach dem ersten Holen heißt der Knopf anders.**
it('nennt den Knopf danach »neu laden«', async () => {
  belegeApi({
    PartieVerlauf: vi.fn().mockResolvedValue({ verlauf: [], fehler: null }),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))

  expect(
    await screen.findByRole('button', { name: eigene.verlaufNachladen }),
  ).toBeInTheDocument()
})

// **Solange er unterwegs ist, sagt die Seite es.**
it('sagt, dass der Verlauf geholt wird', async () => {
  // **Eine Antwort, die auf Zuruf kommt** — nur so lässt sich der Zustand
  // „ist unterwegs" überhaupt sehen. Der Anfangswert wird nie gerufen; er
  // steht da, weil TypeScript sonst ein `undefined` durchreichen ließe.
  let antworte: (antwort: unknown) => void = () => undefined

  belegeApi({
    PartieVerlauf: vi.fn().mockImplementation(
      async () =>
        new Promise((fertig) => {
          antworte = fertig
        }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))

  expect(await screen.findByText(eigene.verlaufLaeuft)).toBeInTheDocument()

  antworte({ verlauf: [], fehler: null })

  await waitFor(() => {
    expect(screen.getByText(eigene.verlaufLeer)).toBeInTheDocument()
  })
})

// **Ein Fehler steht am Bauteil und nimmt nichts zurück** (§7, Regel 3).
it('meldet einen Fehler und lässt das Gelesene stehen', async () => {
  const geholt = vi
    .fn()
    .mockResolvedValueOnce({ verlauf: [einEreignis()], fehler: null })
    .mockResolvedValueOnce({
      verlauf: [],
      fehler: { code: 'intern', meldung: 'Der Bestand ist zu.', feld: '' },
    })

  belegeApi({ PartieVerlauf: geholt })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))
  await screen.findAllByRole('listitem')

  await hand.click(
    screen.getByRole('button', { name: eigene.verlaufNachladen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Bestand ist zu.',
  )

  // **Und der erste Zug steht noch da.**
  expect(screen.getAllByRole('listitem')).toHaveLength(1)
})

// **Ein Sitzwechsel räumt die Rückschau** — §7, Regel 1 und 2.
//
// # Warum der Lauf über `Partiefuehrung` geht und nicht über dieses Bauteil
//
// **Weil der `key` dort steht.** Ein Lauf, der `Partieverlauf` selbst mit
// einem neuen `fuer` neu zeichnete, prüfte nichts: React behielte den Zustand,
// und der Lauf wäre rot — oder man setzte den `key` im Lauf selbst, und dann
// prüfte er die eigene Vorrichtung. Geprüft wird die Stelle, an der die Zusage
// eingelöst wird (§10).
it('räumt die Rückschau beim Sitzwechsel', async () => {
  const geholt = vi
    .fn()
    .mockResolvedValue({ verlauf: [einEreignis()], fehler: null })

  belegeApi({ PartieVerlauf: geholt })

  const { rerender } = render(
    <Sprachraum sprache="de">
      <Partiefuehrung lauf={einLauf()} partie={einePartie({ fuer: 'sitz1' })} />
    </Sprachraum>,
  )

  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))
  await screen.findAllByRole('listitem')

  rerender(
    <Sprachraum sprache="de">
      <Partiefuehrung lauf={einLauf()} partie={einePartie({ fuer: 'sitz2' })} />
    </Sprachraum>,
  )

  // **Annas Rückschau steht nicht mehr da** — sie war empfängergebunden.
  expect(screen.getByText(eigene.verlaufUngeholt)).toBeInTheDocument()
  expect(screen.queryAllByRole('listitem')).toHaveLength(0)

  // **Die Gegenprobe** (§10): der nächste Ruf fragt für den neuen Sitz.
  await hand.click(screen.getByRole('button', { name: eigene.verlaufAnsehen }))

  await waitFor(() => {
    expect(geholt).toHaveBeenCalledWith('p1', 'sitz2')
  })
})
