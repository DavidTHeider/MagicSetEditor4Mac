// **Die Vorrichtung für den Bildvorschlag** — Werkstück A12 von Phase 5.3
// (#1451).
//
// **Ein eigenes Modul, weil zwei Läufe dieselbe Vorrichtung brauchen**: der
// über das Formular und der über den Vorschaukasten. Zwei Fassungen daneben
// liefen beim ersten Sonderfall auseinander.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { KIBild } from '../bausteine/KIBild'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { KIBildvorschlagAntwort } from '../antworten'

export const bildtexte = kataloge.de.seiten.einstellungen

// einVorschlag baut eine Antwort.
export function einVorschlag(
  teil: Partial<KIBildvorschlagAntwort> = {},
): KIBildvorschlagAntwort {
  return {
    kennung: 'bv-1',
    anbieter: 'openai',
    modell: 'gpt-image-2.5-flare',
    umschrift: 'ein Blitz über einer Ebene, Ölgemälde',
    breite: 1024,
    hoehe: 1536,
    bytes: 204800,
    vorschau: 'data:image/png;base64,AAAA',
    ablauf: '2026-09-12T10:00:00Z',
    angaben: [
      { name: 'groesse', wert: '1024x1536' },
      { name: 'guete', wert: 'medium' },
    ],
    fehler: null,
    ...teil,
  }
}

// einKatalog belegt die Modellliste.
//
// **Voreingestellt ein Modell, das Geld kostet**: Bilder erzeugt hier nur
// OpenAI, und ein Katalog voller kostenloser Modelle wäre eine Vorrichtung,
// die großzügiger ist als der Betrieb (§3).
export function einKatalog(ort = 'cloud') {
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
        ort,
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

// ohneNachweis belegt den Bildnachweis leer.
//
// **Jeder Lauf braucht ihn**: der Kasten fragt beim Aufklappen, und eine
// unbelegte Methode meldet sich als Fehler — zu Recht (§3). Wo es um den
// Nachweis selbst geht, belegt der Lauf ihn anders.
export function ohneNachweis() {
  return { angaben: [], fehler: null }
}

// zeige stellt den Kasten hin.
export function zeigeBildkasten(
  zuschneiden = vi.fn(() => Promise.resolve(true)),
  bildfelder = ['image'],
) {
  const gezeichnet = render(
    <Sprachraum sprache="de">
      <KIBild
        karteID="4"
        bildfelder={bildfelder}
        zuschneiden={zuschneiden}
        stand={0}
      />
    </Sprachraum>,
  )

  // neuZeichnen setzt den Stand neu — das tut sonst der Editor, nachdem ein
  // Bild wirklich eingesetzt wurde.
  const neuZeichnen = (stand: number) => {
    gezeichnet.rerender(
      <Sprachraum sprache="de">
        <KIBild
          karteID="4"
          bildfelder={bildfelder}
          zuschneiden={zuschneiden}
          stand={stand}
        />
      </Sprachraum>,
    )
  }

  return { hand: userEvent.setup(), zuschneiden, neuZeichnen }
}

// einBildbudget ist die Auskunft, mit der eine Freigabe möglich ist.
//
// **Jeder Lauf braucht sie**, seit ein Bildauftrag am Fenster freigegeben
// wird (#1539):
// das Modell in [einKatalog] rechnet bei OpenAI, und das kostet Geld.
export function einBildbudget() {
  const betrag = (wert: number) => ({ wert, waehrung: 'EUR' })

  return Promise.resolve({
    konto: 'openai',
    zeitraum: 'tag',
    grenze: betrag(50_000),
    reserviert: betrag(0),
    verbraucht: betrag(0),
    unklar: betrag(0),
    verfuegbar: betrag(50_000),
    fehler: null,
  })
}

// eineBildfreigabe ist die Antwort auf `KIFreigabeErteilen`.
export function eineBildfreigabe() {
  return Promise.resolve({
    kennung: derBildnachweis,
    laeuftab: '2026-09-12T12:05:00Z',
    fehler: null,
  })
}

// derBildnachweis ist die Kennung, die dabei herauskommt.
export const derBildnachweis = 'n-bild-1'

// mitFreigabe ergänzt, was ein Bildauftrag zum Freigeben braucht.
//
// **An einer Stelle und nicht in jedem Lauf**: die beiden Belegungen gehören
// zum Weg und nicht zur Frage, die ein einzelner Lauf stellt.
export function mitFreigabe<T extends object>(teile: T) {
  return {
    Kostenstand: einBildbudget,
    KIFreigabeErteilen: eineBildfreigabe,
    ...teile,
  }
}

// aufschlagenUndErzeugen klappt auf, füllt aus, drückt — und gibt frei.
//
// **Die Freigabe gehört in den Weg und nicht daneben** (#1539): der Auftrag
// geht an OpenAI, und ohne sie weist die Kostenwache ihn ab. Ein Helfer, der
// sie überspränge, wäre großzügiger als der Betrieb (§3).
export async function aufschlagenUndErzeugen(
  hand: ReturnType<typeof userEvent.setup>,
) {
  await hand.click(screen.getByText(bildtexte.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(bildtexte.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(bildtexte.kiBildAufgabe), 'ein Blitz')

  await hand.click(
    screen.getByRole('button', { name: bildtexte.kiBildErzeugen }),
  )

  await gibFrei(hand)
}

// gibFrei nennt den Höchstbetrag und gibt frei.
//
// **Gewartet wird auf das Feld** (§3): der Kasten zeichnet sich sofort, sein
// Betragsfeld erst, wenn das Kostenjournal geantwortet hat.
export async function gibFrei(hand: ReturnType<typeof userEvent.setup>) {
  await hand.type(
    await screen.findByLabelText(bildtexte.kiFreigabeBetrag),
    '1200',
  )
  await hand.click(
    screen.getByRole('button', { name: bildtexte.kiFreigabeErteilen }),
  )
}
