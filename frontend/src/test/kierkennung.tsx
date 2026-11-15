// **Die Vorrichtung für die Kartenerkennung** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// **Ein eigenes Modul**, weil drei Läufe dieselbe Vorrichtung brauchen: der
// über den Weg, der über die Freigabe und der über die Anzeige.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { KIKartenerkennung } from '../bausteine/KIKartenerkennung'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type {
  KIBeobachtungDTO,
  KIBildkandidatDTO,
  KIErkennungAntwort,
} from '../antwortenerkennung'

export const erkennungstexte = kataloge.de.seiten.einstellungen

// einKatalog führt ein Modell — voreingestellt eines auf diesem Rechner.
export function einKatalog(ort = 'lokal') {
  const anbieter = ort === 'lokal' ? 'ollama' : 'openai'

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
        anbieter,
        alias: 'gemma4:8b',
        laufzeit: 'Ollama 0.12.3',
        ort,
        faehigkeiten: [],
        herkunft: [],
        bindung: `${anbieter}/gemma4:8b`,
        kontextfenster: 0,
        kontextbeleg: '',
      },
    ],
    fehler: null,
  }
}

// eineBeobachtung ist eine gelesene Stelle — voreingestellt der Name.
export function eineBeobachtung(
  teil: Partial<KIBeobachtungDTO> = {},
): KIBeobachtungDTO {
  return { art: 'name', wert: 'Blitzschlag', lesbar: true, ...teil }
}

// einKandidat ist ein passender Druck.
export function einKandidat(
  teil: Partial<KIBildkandidatDTO> = {},
): KIBildkandidatDTO {
  return {
    kennung: 'p1',
    name: 'Blitzschlag',
    setcode: 'lea',
    setname: 'Limited Edition Alpha',
    nummer: '161',
    seltenheit: 'common',
    bild: '',
    gestuetzt: ['name'],
    ...teil,
  }
}

// eineErkennung baut eine Antwort — voreingestellt eine eindeutige.
export function eineErkennung(
  teil: Partial<KIErkennungAntwort> = {},
): KIErkennungAntwort {
  return {
    beobachtungen: [eineBeobachtung()],
    nichtGelesen: ['nummer'],
    befunde: [],
    kandidaten: [einKandidat()],
    drucke: 1,
    weitere: false,
    eindeutig: true,
    alter: '3m0s',
    anbieter: 'ollama',
    modell: 'gemma4:8b',
    abgebrochen: false,
    fehler: null,
    ...teil,
  }
}

// einBudget ist die Auskunft, mit der eine Freigabe möglich ist.
export function einBudget() {
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

// zeigeErkennungskasten stellt ihn hin.
export function zeigeErkennungskasten() {
  render(
    <Sprachraum sprache="de">
      <KIKartenerkennung />
    </Sprachraum>,
  )

  return userEvent.setup()
}

// aufschlagenUndLesen klappt auf, wählt ein Modell und drückt.
export async function aufschlagenUndLesen(
  hand: ReturnType<typeof userEvent.setup>,
  bindung = 'ollama/gemma4:8b',
) {
  await hand.click(screen.getByText(erkennungstexte.kiErkennungTitel))

  await hand.selectOptions(
    await screen.findByLabelText(erkennungstexte.kiVorschlagAnbieter),
    bindung,
  )

  await hand.click(
    screen.getByRole('button', { name: erkennungstexte.kiErkennungLesen }),
  )
}
