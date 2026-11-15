// **Meinen Code zeigen, einen Code prüfen** — NET-01
// (#1705).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Antwort zu dem Code gehört, der geprüft wurde** (§7, Regel 2)
// — und dass jeder Fehler an seinem Schritt steht und nicht irgendwo (§6).
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import {
  alsAntwort,
  eigene,
  einFehler,
  eineVorschau,
  pruefe,
  stelleAuf,
  technik,
} from '../test/einladungsprobe'

import type { Einladungsvorschauantwort } from '../apikontakteinladung'

afterEach(() => {
  raeumeApi()
})

it('stellt einen Code aus und kopiert ihn', async () => {
  const erstellen = vi.fn(() =>
    Promise.resolve({
      einladung: {
        code: 'mse4mac-kontakt:1:abc.def',
        ablauf: '2026-10-01T12:00:00Z',
        fingerabdruck: '9f4b',
      },
      fehler: null,
    }),
  )
  belegeApi({ KontakteinladungErstellen: erstellen })

  // **Nach `userEvent.setup()`**: es legt eine eigene Zwischenablage an,
  // und eine davor gesetzte wäre überschrieben.
  const { hand } = stelleAuf()
  const schreiben = vi.fn(() => Promise.resolve())
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: schreiben },
  })
  await hand.click(screen.getByRole('button', { name: eigene.zeigen }))

  // **Ohne Namen geht nichts hinaus** — er ist das, was der andere liest.
  expect(screen.getByRole('button', { name: eigene.erzeugen })).toBeDisabled()

  await hand.type(screen.getByLabelText(eigene.name), 'Anna')
  await hand.selectOptions(
    screen.getByLabelText(eigene.zweck),
    technik.zweckPartie,
  )
  await hand.click(screen.getByRole('button', { name: eigene.erzeugen }))

  expect(await screen.findByLabelText(eigene.code)).toHaveValue(
    'mse4mac-kontakt:1:abc.def',
  )
  expect(erstellen).toHaveBeenCalledWith('Anna', 'partie')

  await hand.click(
    screen.getByRole('button', {
      name: new RegExp(
        kataloge.de.bausteine.aktionskatalog.aktionen.kopieren.kurz,
      ),
    }),
  )
  expect(await screen.findByText(eigene.kopiert)).toBeInTheDocument()
  expect(schreiben).toHaveBeenCalledWith('mse4mac-kontakt:1:abc.def')

  // **Ein zweiter Klick auf denselben Schritt schließt ihn.**
  await hand.click(screen.getByRole('button', { name: eigene.zeigen }))
  expect(screen.queryByLabelText(eigene.name)).toBeNull()
})

it('zeigt einen Fehler des Ausstellens am Schritt', async () => {
  belegeApi({
    KontakteinladungErstellen: () =>
      Promise.resolve({
        ...einFehler('ohne Identität keine Einladung'),
        einladung: { code: '', ablauf: '', fingerabdruck: '' },
      }),
  })

  const { hand } = stelleAuf()
  await hand.click(screen.getByRole('button', { name: eigene.zeigen }))
  await hand.type(screen.getByLabelText(eigene.name), 'Anna')
  await hand.click(screen.getByRole('button', { name: eigene.erzeugen }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'ohne Identität keine Einladung',
  )
  expect(screen.getByLabelText(eigene.name)).toHaveAttribute(
    'aria-invalid',
    'true',
  )
})

it('zeigt einen Fehler der Prüfung am Feld', async () => {
  belegeApi({
    KontakteinladungPruefen: () =>
      Promise.resolve({
        ...einFehler('die Unterschrift passt nicht'),
        vorschau: eineVorschau(),
      }),
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'die Unterschrift passt nicht',
  )
  expect(screen.queryByRole('button', { name: eigene.stimmt })).toBeNull()

  // **Wer den Code ändert, räumt den Fehler mit ab** (§7, Regel 1).
  await hand.type(screen.getByLabelText(eigene.eingabe), 'x')
  expect(screen.queryByRole('alert')).toBeNull()
})

it('trägt eine Antwort zu einem früheren Code nicht ein', async () => {
  const offen: {
    erfuellen: (wert: Einladungsvorschauantwort) => void
    ablehnen: (grund: unknown) => void
  }[] = []

  belegeApi({
    KontakteinladungPruefen: () =>
      new Promise((erfuellen, ablehnen) => {
        offen.push({ erfuellen, ablehnen })
      }),
  })

  const { hand } = stelleAuf()
  const pruefknopf = () => screen.getByRole('button', { name: eigene.pruefen })

  await pruefe(hand)
  await waitFor(() => {
    expect(offen).toHaveLength(1)
  })

  // **Der Code ändert sich, bevor die Antwort kommt** — sie kommt, der
  // Vorgang endet, und eingetragen wird nichts.
  await hand.type(screen.getByLabelText(eigene.eingabe), 'x')
  offen[0]?.erfuellen({ vorschau: eineVorschau(), fehler: null })
  await waitFor(() => {
    expect(pruefknopf()).toBeEnabled()
  })
  expect(screen.queryByRole('heading', { name: 'Anna bestätigen' })).toBeNull()

  // **Dasselbe mit einem Fehler**, der zu spät kommt.
  await hand.click(pruefknopf())
  await waitFor(() => {
    expect(offen).toHaveLength(2)
  })
  await hand.type(screen.getByLabelText(eigene.eingabe), 'y')
  offen[1]?.ablehnen(new Error('zu spät'))
  await waitFor(() => {
    expect(pruefknopf()).toBeEnabled()
  })
  expect(screen.queryByRole('alert')).toBeNull()

  // **Beleg, dass jemand zuhört**: eine dritte Prüfung trägt ihr Ergebnis
  // ein — die ersten beiden wurden also nicht still verschluckt.
  await hand.click(pruefknopf())
  await waitFor(() => {
    expect(offen).toHaveLength(3)
  })
  offen[2]?.erfuellen({ vorschau: eineVorschau(), fehler: null })

  expect(
    await screen.findByRole('heading', { name: 'Anna bestätigen' }),
  ).toBeInTheDocument()
})

it('warnt vor einem neuen Schlüssel und zeigt die technischen Angaben', async () => {
  belegeApi({
    KontakteinladungPruefen: () =>
      alsAntwort(
        eineVorschau({
          vorgaenger: 'Anna vom Stammtisch',
          lage: 'bestaetigt',
          zweck: 'paarung',
          adresse: '',
        }),
      ),
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  expect(
    await screen.findByText(
      /Das will Anna vom Stammtisch mit einem neuen Schlüssel sein/,
    ),
  ).toBeInTheDocument()
  expect(screen.getByText(eigene.bekannt)).toBeInTheDocument()

  // **Kennenlernen gibt nichts frei** — also gibt es keinen Haken.
  expect(screen.queryByRole('checkbox')).toBeNull()

  await hand.click(screen.getByRole('button', { name: eigene.technik }))
  expect(screen.getByText(eigene.technikOhneAdresse)).toBeInTheDocument()
  expect(screen.getByText('9f4b-8c8e-9025')).toBeInTheDocument()
})
