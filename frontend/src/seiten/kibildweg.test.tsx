// **Der Weg vom erzeugten Bild bis in die Karte** — Werkstück A12 von
// Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass es derselbe Weg ist.** Das Issue verlangt „Übernahme nutzt
// vorhandene Crop-/Ablage-/Renderfunktion" — und ein Lauf, der nur das
// Bauteil prüfte, sähe davon nichts: dort endet er beim Aufruf. Hier geht er
// durch **die Seite**, durch dasselbe Zuschnittfenster wie eine Datei von der
// Platte, und kommt bei `KIBildEinsetzen` heraus.
//
// **Und dass nur der dritte Weg genommen wird.** Drei Bindungen setzen ein —
// Platte, Scryfall, Modell —, und welche es ist, entscheidet der offene
// Zuschnitt. Wer sich hier vertut, schreibt ein Bild ohne seine Herkunft.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Karteneditor } from './Karteneditor'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { gibFrei, mitFreigabe, ohneNachweis } from '../test/kibild'
import { zuschnittantwort } from '../test/zuschnitt'

import type { KarteDTO, VorschauAntwort } from '../api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.seiten.einstellungen

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [{ feld: 'kunst', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
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

function karte(): KarteDTO {
  return { id: '5', nummer: 0, werte: [{ feld: 'kunst', wert: 'alt.png' }] }
}

function einKatalog() {
  return {
    nurLokal: false,
    vorrang: [],
    vorrangbefunde: [],
    einrichtung: 'bereit',
    anbieter: [],
    cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
    befunde: [],
    modelle: [
      {
        anbieter: 'openai',
        alias: 'gpt-image-2.5-flare',
        laufzeit: 'Responses-API',
        ort: 'cloud',
        faehigkeiten: [],
        herkunft: [],
        bindung: 'openai/gpt-image-2.5-flare',
        kontextfenster: 0,
        kontextbeleg: '',
      },
    ],
    fehler: null,
  }
}

function einVorschlag() {
  return {
    kennung: 'bv-1',
    anbieter: 'openai',
    modell: 'gpt-image-2.5-flare',
    umschrift: '',
    breite: 1024,
    hoehe: 1536,
    bytes: 2048,
    vorschau: 'data:image/png;base64,AAAA',
    ablauf: '2026-09-12T10:00:00Z',
    angaben: [],
    fehler: null,
  }
}

function zeige(teile = {}) {
  belegeApi(
    mitFreigabe({
      Vorschau: () => Promise.resolve(vorschau()),
      Probevorschau: () => Promise.resolve(vorschau()),
      Zuschnittgrenze: () =>
        Promise.resolve({ x: 0, y: 0, breite: 125, hoehe: 100 }),
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () => Promise.resolve(einVorschlag()),
      KIBildZuschnitt: () => Promise.resolve(zuschnittantwort()),
      KIFeldherkunft: () => Promise.resolve({ herkunft: [], fehler: null }),
      ...teile,
    }),
  )

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte()}
        formen={[bildform]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

// erzeuge klappt den KI-Kasten und darin den Bildkasten auf und lässt ein
// Bild erzeugen — der Weg eines Menschen seit UXR-21 (#1929).
async function erzeuge(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(await screen.findByText(texte.bausteine.kikasten.titel))
  await hand.click(await screen.findByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Blitz')
  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  // **Der Auftrag geht an OpenAI und kostet Geld** (#1539) — ohne Freigabe
  // weist die Kostenwache ihn ab, und der Weg endete vor dem Bild.
  await gibFrei(hand)

  await screen.findByRole('figure')
}

// **Der Vorschlag geht durch dasselbe Zuschnittfenster wie eine Datei.**
it('ein erzeugtes Bild geht durch dasselbe Zuschnittfenster', async () => {
  const einsetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ KIBildEinsetzen: einsetzen })

  const hand = userEvent.setup()

  await erzeuge(hand)

  await hand.click(
    screen.getByRole('button', { name: eigene.kiBildUebernehmen }),
  )

  // **Das Fenster ist dasselbe** — derselbe Knopf, derselbe Text.
  await hand.click(
    await screen.findByRole('button', {
      name: texte.seiten.karten.zuschnitt.uebernehmen,
    }),
  )

  await waitFor(() => {
    expect(einsetzen).toHaveBeenCalledWith('1', '5', 'kunst', 'bv-1', {
      x: 0,
      y: 0,
      breite: 125,
      hoehe: 100,
    })
  })
})

// **Und nicht über den Weg der Platte.**
//
// **Der Lauf darüber allein belegte das nicht**: er sieht, dass die eine
// Bindung gerufen wurde, nicht dass die andere es nicht wurde. Drei Wege
// enden in einem Fenster, und die Wahl zwischen ihnen ist die Stelle, an der
// man sich vertut.
it('ein erzeugtes Bild geht nicht über den Weg der Platte', async () => {
  const ueberDiePlatte = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({
    KIBildEinsetzen: () => Promise.resolve({ fehler: null }),
    BildEinsetzen: ueberDiePlatte,
  })

  const hand = userEvent.setup()

  await erzeuge(hand)

  await hand.click(
    screen.getByRole('button', { name: eigene.kiBildUebernehmen }),
  )

  await hand.click(
    await screen.findByRole('button', {
      name: texte.seiten.karten.zuschnitt.uebernehmen,
    }),
  )

  await waitFor(() => {
    expect(
      screen.queryByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    ).not.toBeInTheDocument()
  })

  expect(ueberDiePlatte).not.toHaveBeenCalled()
})
