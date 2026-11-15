// **Die Nennung bleibt stehen, wenn der Zuschnitt zu ist** —
// #1025.
//
// Scryfall verlangt, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`: „list the artist name and copyright
// elsewhere in the same interface presenting the art crop […] Users should be
// able to identify the artist and source of the image somehow."
//
// Erfüllt war das bis hierher **nur, solange das Zuschnittfenster offen war**.
// Danach stand das Bild auf der Karte, und der Kasten mit der Nennung war
// fort — die Bedingung überlebte das Schließen eines Kastens nicht.
//
// # Was dieser Lauf nicht prüft
//
// **Ob die Nennung stimmt.** Sie kommt fertig aus Go (`service.Nennungen`),
// und dort steht auch, woran eine Kartenseite bei Scryfall erkannt wird.
// Hier wird geprüft, dass sie **ankommt und stehen bleibt**.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

function vorschau(nennungen: VorschauAntwort['vorschau']): VorschauAntwort {
  return { vorschau: nennungen }
}

const mitNennung = {
  stil: '',
  gezeichnet: '',
  bild: '/vorschau/AAAA.png',
  breite: 10,
  hoehe: 14,
  dpi: 150,
  befunde: [],
  nennungen: [
    {
      adresse: 'https://scryfall.com/card/lea/232/black-lotus',
      felder: ['kunst'],
      kuenstler: 'Christopher Rush',
      text: 'Illustration: Christopher Rush. ™ & © Wizards of the Coast',
    },
  ],
  flaechen: [{ feld: 'kunst', links: 0, oben: 0, breite: 10, hoehe: 5 }],
  vorgaben: [],
}

function karte(id = '5'): KarteDTO {
  return { id, nummer: 0, werte: [] }
}

const bildform = {
  name: 'kunst',
  aufschrift: 'Kunst',
  art: 'image',
  mehrzeilig: false,
  wahlen: [],
  beschreibung: '',
  vorgabe: '',
  teilfelder: [],
  reihe: '',
  stelle: 0,
  editierbar: true,
  gespeichert: true,
}

it('zeigt Künstler, Copyright und Quelle ohne offenen Zuschnitt', async () => {
  belegeApi({
    Vorschau: () => Promise.resolve(vorschau(mitNennung)),
    Probevorschau: () => Promise.resolve(vorschau(mitNennung)),
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        formen={[bildform]}
        geaendert={() => undefined}
        karte={karte()}
        setID="1"
      />
    </Sprachraum>,
  )

  expect(
    await screen.findByText(/Illustration: Christopher Rush/),
  ).toBeInTheDocument()

  expect(screen.getByText(/Wizards of the Coast/)).toBeInTheDocument()

  const quelle = screen.getByRole('link', {
    name: /scryfall/i,
  })

  expect(quelle).toHaveAttribute(
    'href',
    'https://scryfall.com/card/lea/232/black-lotus',
  )
})

it('zeigt nichts, wenn nichts übernommen wurde', async () => {
  const ohne = { ...mitNennung, nennungen: [] }

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau(ohne)),
    Probevorschau: () => Promise.resolve(vorschau(ohne)),
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        formen={[bildform]}
        geaendert={() => undefined}
        karte={karte()}
        setID="1"
      />
    </Sprachraum>,
  )

  await screen.findByRole('img')

  expect(screen.queryByText(/Wizards of the Coast/)).not.toBeInTheDocument()
})
