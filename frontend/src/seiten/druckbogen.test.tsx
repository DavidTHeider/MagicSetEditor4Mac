// Der Druckbogen eines Sets.
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { abgebrochen, laufauftrag } from '../test/laufauftrag'
import { laufzeitprobe } from '../test/laufzeitprobe'
import { ereignisAuftrag } from '../fenster'
import { Druckbogen } from './Druckbogen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Druckbogen setID="4" />
    </Sprachraum>,
  )
}

function gelegt(teile: Record<string, unknown> = {}) {
  return Promise.resolve({
    pfad: '/tmp/druckbogen.pdf',
    seiten: 2,
    karten: 10,
    befunde: [],
    abgebrochen: false,
    ...teile,
  })
}

async function druecke(name: string) {
  await userEvent.click(await screen.findByRole('button', { name }))
}

// **Der Fall „kein Set" steht hier nicht mehr** — siehe `bilder.test.tsx`
// und #288.

describe('Mit einem gewählten Set', () => {
  // **Die Vorgaben sind die, mit denen man drucken kann**: A4, ohne Anschnitt,
  // 300 dpi, mit Marken. Wer nichts einstellt, bekommt einen brauchbaren Bogen.
  it('legt den Bogen mit den Vorgaben', async () => {
    const BogenLegen = vi.fn(() => gelegt())

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenLegen,
    })

    zeige()

    await druecke(texte.seiten.druckbogen.legen)

    expect(BogenLegen).toHaveBeenCalledWith('4', 'a4', 0, 300, true, [])
    // **Der ganze Satz, nicht nur der Pfad** — hier stand „10 Karten Karten
    // auf 2 Seiten", seit das Zählwort das Wort selbst setzt (#1219), und
    // ein Griff nach dem Pfad sah es nicht. Gefunden am nativen Fenster
    // (Hüllenprobe `fmt03b`).
    expect(await screen.findByRole('status')).toHaveTextContent(
      '10 Karten auf 2 Seiten: /tmp/druckbogen.pdf',
    )
  })

  it('gibt Papier, Anschnitt, Auflösung und Marken weiter', async () => {
    const BogenLegen = vi.fn(() => gelegt())

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenLegen,
    })

    zeige()

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.druckbogen.papier),
      'letter',
    )

    const anschnitt = screen.getByLabelText(texte.seiten.druckbogen.anschnitt)
    await userEvent.clear(anschnitt)
    await userEvent.type(anschnitt, '3')

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.druckbogen.aufloesung),
      '600',
    )

    await userEvent.click(
      screen.getByLabelText(texte.seiten.druckbogen.schnittmarken),
    )

    await druecke(texte.seiten.druckbogen.legen)

    expect(BogenLegen).toHaveBeenCalledWith('4', 'letter', 3, 600, false, [])
  })

  // **Was nicht passte, steht beim Erfolg** (§8).
  it('nennt, was am Bogen nicht passte', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenLegen: () =>
        gelegt({ befunde: ['Feuerball: die Vorlage wird verzerrt'] }),
    })

    zeige()

    await druecke(texte.seiten.druckbogen.legen)

    expect(
      await screen.findByText('Feuerball: die Vorlage wird verzerrt'),
    ).toBeInTheDocument()
  })

  it('sagt nichts, wenn niemand einen Ort gewählt hat', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenLegen: () => gelegt({ abgebrochen: true, pfad: '' }),
    })

    zeige()

    await druecke(texte.seiten.druckbogen.legen)
    await antwortAbwarten()

    // **Nach dem Stand gefragt, nicht nach einem Wort** — hier stand die
    // Suche nach dem Katalogsatz, und seit er nur noch „auf" heißt, fände sie
    // jeden Hilfetext der Seite.
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('meldet einen Fehler beim Legen', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenLegen: () =>
        Promise.resolve({
          pfad: '',
          seiten: 0,
          karten: 0,
          befunde: [],
          abgebrochen: false,
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'auf das Papier passt keine Karte',
            feld: '',
          },
        }),
    })

    zeige()

    await druecke(texte.seiten.druckbogen.legen)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'auf das Papier passt keine Karte',
    )
  })
})

// **Der Lauf zeigt sich, wo er angestoßen wurde, und lässt sich anhalten** —
// FMT-03.AK-02 (#1692). Ein angehaltener Bogen ist keine Fehlermeldung wert:
// es gibt keinen halben, und der Laufstand sagt es.
it('zeigt den Lauf, hält ihn an und meldet dafür keinen Fehler', async () => {
  const probe = laufzeitprobe()
  let antworten: (wert: unknown) => void = () => undefined
  const AuftragAbbrechen = vi.fn(() =>
    Promise.resolve({ auftraege: [], laufend: 0, fehler: null }),
  )

  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    BogenLegen: () =>
      new Promise((weiter) => {
        antworten = weiter as (wert: unknown) => void
      }),
    AuftragAbbrechen,
  })

  try {
    zeige()
    await druecke(texte.seiten.druckbogen.legen)
    await effekteNachholen()

    act(() => {
      probe.melde(ereignisAuftrag, laufauftrag({ erledigt: 2, gesamt: 10 }))
    })
    expect(screen.getByRole('progressbar')).toHaveAttribute(
      'aria-valuenow',
      '2',
    )

    await druecke(texte.bausteine.laufstand.abbrechen)
    expect(AuftragAbbrechen).toHaveBeenCalledWith('4')

    act(() => {
      probe.melde(ereignisAuftrag, laufauftrag({ stand: 'abgebrochen' }))
      antworten({
        pfad: '',
        seiten: 0,
        karten: 0,
        befunde: [],
        abgebrochen: false,
        fehler: abgebrochen,
      })
    })

    expect(
      await screen.findByText((text) =>
        text.startsWith(texte.bausteine.laufstand.abgebrochen),
      ),
    ).toHaveTextContent('Bis dahin: 1 von 3 Karten.')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  } finally {
    probe.raeume()
  }
})
