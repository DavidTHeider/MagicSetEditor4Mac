// **Späte Antworten und die Ränder der Tabelle** — Werkstück O3 von Phase 5.2
// (#1427).
//
// Sichern, Einlesen und Übernehmen teilen sich **einen** Lauf (§7, Regel 2):
// wer eines beginnt, macht die Antwort des anderen ungültig. Ein Vorgang je
// Art verhindert zwei gleiche zugleich — die Kreuzung zweier verschiedener
// verhindert er nicht, und genau die steht hier.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Einstellungsaustausch } from './Einstellungsaustausch'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import type {
  Einfuhrzeile,
  Einstellungsbericht,
  Einstellungssicht,
} from '../api'

afterEach(() => {
  raeumeApi()
})

const sichern = 'Einstellungen sichern'
const einlesen = 'Datei einlesen'
const uebernehmen = 'Gewählte übernehmen'

const sprachzeile: Einfuhrzeile = {
  ebene: 'profil',
  schluessel: 'sprache',
  vorher: 'de',
  nachher: 'en',
  herkunft: 'profil',
  wirkung: 'sofort',
}

const einzeilig: Einstellungssicht = {
  inhalt: '{"fassung":1,"bereiche":{}}',
  fassung: 1,
  revision: 'abc123',
  zeilen: [sprachzeile],
  befunde: [],
  unveraendert: 0,
}

function gehalten<T>() {
  let aufloesen: (wert: T) => void = () => undefined
  let ablehnen: (ursache: unknown) => void = () => undefined
  const promise = new Promise<T>((a, b) => {
    aufloesen = a
    ablehnen = b
  })

  return {
    promise,
    aufloesen: (wert: T) => {
      aufloesen(wert)
    },
    ablehnen: (u: unknown) => {
      ablehnen(u)
    },
  }
}

function zeige(teile: Parameters<typeof belegeApi>[0]) {
  belegeApi(teile)
  render(
    <Sprachraum sprache="de">
      <Einstellungsaustausch />
    </Sprachraum>,
  )
}

it('lässt eine späte Antwort des Sicherns fallen — gut wie schlecht', async () => {
  const erste = gehalten<{ fehler: null }>()
  const zweite = gehalten<{ fehler: null }>()
  let ruf = 0
  zeige({
    EinstellungenSichern: () => (ruf++ === 0 ? erste.promise : zweite.promise),
    EinstellungenEinlesen: () =>
      Promise.resolve({ sicht: einzeilig, fehler: null }),
  })

  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await screen.findByText('profil:sprache')

  erste.ablehnen(new Error('zu spät und falsch'))
  await antwortAbwarten()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()

  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await screen.findByText('profil:sprache')

  zweite.aufloesen({ fehler: null })
  await antwortAbwarten()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})

it('lässt eine späte Vorschau fallen — gut wie schlecht', async () => {
  const erste = gehalten<{ sicht: Einstellungssicht; fehler: null }>()
  const zweite = gehalten<{ sicht: Einstellungssicht; fehler: null }>()
  let ruf = 0
  zeige({
    EinstellungenEinlesen: () => (ruf++ === 0 ? erste.promise : zweite.promise),
    EinstellungenSichern: () => Promise.resolve({ fehler: null }),
  })

  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await antwortAbwarten()

  erste.aufloesen({ sicht: einzeilig, fehler: null })
  await antwortAbwarten()
  expect(screen.queryByText('profil:sprache')).not.toBeInTheDocument()

  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await antwortAbwarten()

  zweite.ablehnen(new Error('zu spät und falsch'))
  await antwortAbwarten()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})

it('lässt eine späte Übernahme fallen — gut wie schlecht', async () => {
  const erste = gehalten<{ bericht: Einstellungsbericht; fehler: null }>()
  const zweite = gehalten<{ bericht: Einstellungsbericht; fehler: null }>()
  let ruf = 0
  zeige({
    EinstellungenEinlesen: () =>
      Promise.resolve({ sicht: einzeilig, fehler: null }),
    EinstellungenSichern: () => Promise.resolve({ fehler: null }),
    EinstellungenUebernehmen: () =>
      ruf++ === 0 ? erste.promise : zweite.promise,
  })

  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await screen.findByText('profil:sprache')
  await userEvent.click(screen.getByRole('button', { name: uebernehmen }))
  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await antwortAbwarten()

  erste.aufloesen({
    bericht: { uebernommen: 1, revision: 'neu', zeilen: [], befunde: [] },
    fehler: null,
  })
  await antwortAbwarten()
  // **Die Vorschau bleibt stehen**: der Bericht galt einem Stand von vorhin.
  expect(screen.queryByText('1 Zeile übernommen.')).not.toBeInTheDocument()
  expect(screen.getByText('profil:sprache')).toBeInTheDocument()

  await userEvent.click(screen.getByRole('button', { name: uebernehmen }))
  await userEvent.click(screen.getByRole('button', { name: sichern }))
  await antwortAbwarten()

  zweite.ablehnen(new Error('zu spät und falsch'))
  await antwortAbwarten()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})

it('nennt einen Befund über einen ganzen Abschnitt und eine unbekannte Herkunft roh', async () => {
  zeige({
    EinstellungenEinlesen: () =>
      Promise.resolve({
        sicht: {
          ...einzeilig,
          zeilen: [{ ...sprachzeile, herkunft: 'anderswo' }],
          befunde: [
            {
              ebene: 'fremd',
              schluessel: '',
              grund: 'Diesen Abschnitt kennt dieses Programm nicht.',
            },
          ],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: einlesen }))

  expect(await screen.findByText('fremd')).toBeInTheDocument()
  expect(screen.getByText('anderswo')).toBeInTheDocument()
})

// **Der Neustartbedarf steht an der Zeile** — OPS-01.AK-02 (#1718). Heute
// reist keine Einstellung, die einen Neustart braucht (alle vier gehören dem
// Gerät); die Vorschau sagt es trotzdem, sobald eine es tut — und an einer
// Zeile, die sofort wirkt, sagt sie nichts.
it('nennt an einer Zeile, dass sie erst nach einem Neustart wirkt', async () => {
  zeige({
    EinstellungenEinlesen: () =>
      Promise.resolve({
        sicht: {
          ...einzeilig,
          zeilen: [
            sprachzeile,
            {
              ...sprachzeile,
              schluessel: 'kuenftig',
              wirkung: 'neustart',
            },
          ],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: einlesen }))
  await screen.findByText('profil:kuenftig')

  expect(screen.getAllByText('Wirkt erst nach einem Neustart.')).toHaveLength(1)
})
