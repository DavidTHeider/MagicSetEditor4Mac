// **Die Werkbank für die Turnierseite** — sie stellt die Seite hin und
// schlägt ein Turnier auf.
//
// **Eine eigene Datei**, weil zwei Testdateien sie brauchen und beide unter
// der Zeilengrenze aus §2 bleiben sollen: `turnierlauf` fragt, was die Seite
// über die Zeit **tut**, `turnierstand`, was sie über einen Stand **sagt**.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { zuGebiet } from '../test/gebiete'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { rangprobe } from './turnierprobe'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

export async function aufgeschlagen(
  turnier: TurnierDTO,
  weitere: Record<string, unknown> = {},

  // gebiet ist die Lasche, die danach offen sein soll.
  //
  // **Seit Werkstück I von Phase 4.10 hat die Werkbank drei**
  // (#1325):
  // Teilnehmer, Runden, Rangliste. Wer eine Paarung sucht, geht denselben Weg
  // wie ein Anwender.
  gebiet:
    | 'teilnehmer'
    | 'runden'
    | 'rangliste'
    | 'decklisten'
    | 'protokoll' = 'teilnehmer',
) {
  belegeApi({
    Turniere: () => Promise.resolve({ turniere: [turnier], fehler: null }),
    Turnier: () => Promise.resolve({ turnier, fehler: null }),
    TurnierRangliste: () =>
      Promise.resolve({ rangliste: [rangprobe()], fehler: null }),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', {
      name: `${texte.waehlen}: ${turnier.name}`,
    }),
  )

  // **Immer wechseln, auch auf »Teilnehmer«.** Ein Turnier mit Runden schlägt
  // dort auf, wo es steht — bei den Runden; wer die Teilnehmer will, muss also
  // hin.
  await zumTurniergebiet(gebiet)

  return screen.findByRole('region', { name: turnier.name })
}

// zumTurniergebiet wählt eine Lasche des offenen Turniers.
//
// **Seit Werkstück I von Phase 4.10 hat die Werkbank drei Gebiete**
// (#1325):
// Teilnehmer, Runden, Rangliste.
export async function zumTurniergebiet(
  gebiet: 'teilnehmer' | 'runden' | 'rangliste' | 'protokoll' | 'decklisten',
): Promise<void> {
  const namen = kataloge.de.seiten.turniere

  await zuGebiet(
    {
      teilnehmer: namen.gebietTeilnehmer,
      runden: namen.gebietRunden,
      rangliste: namen.gebietRangliste,
      protokoll: namen.gebietProtokoll,
      decklisten: namen.gebietDecklisten,
    }[gebiet],
  )
}

// ergebnisOeffnen klappt das Eintragen einer Paarung auf.
//
// **Seit UXR-12 steht es hinter *Ergebnis* an ihrer Zeile**
// (#1920),
// und nur die laufende Runde ist offen — `runde` nennt eine frühere, die
// dafür erst aufgeklappt wird. Derselbe Weg wie ein Anwender.
export async function ergebnisOeffnen(
  werkbank: HTMLElement,
  wer: string,
  runde?: number,
): Promise<void> {
  if (runde !== undefined) {
    await userEvent.click(
      within(werkbank).getByRole('button', {
        name: einsetzen(texte.runde, { nummer: String(runde) }),
      }),
    )
  }

  await userEvent.click(
    within(werkbank).getByRole('button', {
      name: `${texte.ergebnis}: ${wer}`,
    }),
  )
}

// alleTurniere klappt die Liste unter einem offenen Turnier auf — seit der
// Messung am Fenster (`uxr12`) steht sie dort hinter „Alle Turniere (N)" (R4).
// **Ohne offenes Turnier tut es nichts**: dann steht die Liste offen.
export async function alleTurniere() {
  const anfang = texte.alleTurniere.slice(0, texte.alleTurniere.indexOf(' ('))
  const falte = screen.queryByRole('button', {
    name: (name) => name.startsWith(anfang),
  })

  if (falte?.getAttribute('aria-expanded') === 'false') {
    await userEvent.click(falte)
  }
}
