// **Ein Set mit Schlüsselwörtern für die Tests der Schlüsselwortseite** —
// UXR-F3 (#2009).
//
// # Warum ein lebender Stand
//
// **Weil die Seite nach jedem Schreiben nachlädt.** Eine Vorrichtung, die
// immer dieselbe Liste antwortet, ist großzügiger als der Betrieb (§3): ein
// angelegtes Wort käme nie zurück, ein gelöschtes bliebe stehen — und genau
// dort entscheidet sich, welches Wort danach gewählt ist. Deshalb ändern
// `SchluesselwortAnlegen`, `-Setzen` und `-Loeschen` hier einen Stand, den
// `Schluesselwoerter` beim nächsten Ruf zurückgibt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'

import type { Api, SchluesselwortDTO } from '../api'
import { Schluesselwoerter } from '../seiten/Schluesselwoerter'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'

export const texte = kataloge.de.seiten.karten.schluesselwoerter

// wort baut ein eigenes Schlüsselwort — leer, wo der Fall nichts sagt.
export function wort(teile: Partial<SchluesselwortDTO>): SchluesselwortDTO {
  return {
    id: '',
    wort: '',
    muster: '',
    erinnerung: '',
    regeln: '',
    modus: '',
    fest: false,
    gueltig: true,
    ...teile,
  }
}

// fuenf sind fünf eigene Wörter, wie sie ein Set mitbringen kann.
export function fuenf(): SchluesselwortDTO[] {
  return ['Beben', 'Kicker', 'Rausch', 'Sog', 'Wacht'].map((name, stelle) =>
    wort({
      id: String(stelle + 1),
      wort: name,
      muster: `${name} `,
      erinnerung: `Erinnerung an ${name}`,
      modus: 'custom',
    }),
  )
}

export interface Probe {
  // bestand ist, was „das Backend" gerade hält — die eigenen Wörter.
  bestand: SchluesselwortDTO[]
  SchluesselwortSetzen: Mock<Api['SchluesselwortSetzen']>
  SchluesselwortLoeschen: Mock<Api['SchluesselwortLoeschen']>
  SchluesselwortAnlegen: Mock<Api['SchluesselwortAnlegen']>

  // vorLesen hält die nächste Antwort von `Schluesselwoerter` zurück, solange
  // das Versprechen offen ist — für den Augenblick zwischen dem Schreiben und
  // dem Nachladen.
  vorLesen: (() => Promise<void>) | undefined

  // zeichneNeu zeichnet die Seite mit einem anderen Verlaufsstand — erst
  // nach [zeige], das sie zeichnet.
  zeichneNeu: (verlaufsstand: number) => void
}

interface Wahl {
  modi?: string[]
  feste?: SchluesselwortDTO[]

  // belegen ersetzt einzelne Bindungen — für das Scheitern und das Warten.
  belegen?: Partial<Api>
}

// belege legt den lebenden Stand in die Fassade, ohne etwas zu zeichnen — für
// den, der die Seite in einem anderen Rahmen zeigen will.
export function belege(woerter: SchluesselwortDTO[], wahl: Wahl = {}): Probe {
  const probe: Probe = {
    bestand: [...woerter],
    SchluesselwortSetzen: vi.fn<Api['SchluesselwortSetzen']>(),
    SchluesselwortLoeschen: vi.fn<Api['SchluesselwortLoeschen']>(),
    SchluesselwortAnlegen: vi.fn<Api['SchluesselwortAnlegen']>(),
    vorLesen: undefined,
    zeichneNeu: () => undefined,
  }
  let naechste = 9

  probe.SchluesselwortSetzen.mockImplementation((_set, kennung, feld, wert) => {
    const treffer = probe.bestand.find((eines) => eines.id === kennung)
    if (treffer !== undefined) {
      Object.assign(treffer, { [feld]: wert })
    }

    return Promise.resolve({ fehler: null })
  })
  probe.SchluesselwortLoeschen.mockImplementation((_set, kennung) => {
    probe.bestand = probe.bestand.filter((eines) => eines.id !== kennung)

    return Promise.resolve({ fehler: null })
  })
  probe.SchluesselwortAnlegen.mockImplementation(() => {
    const kennung = String(naechste++)
    probe.bestand.push(wort({ id: kennung, modus: 'custom' }))

    return Promise.resolve({ wortID: kennung, fehler: null })
  })

  belegeApi({
    Schluesselwoerter: async () => {
      await probe.vorLesen?.()

      return {
        // **Copy, damit eine Änderung danach die Antwort von vorhin nicht
        // verändert** — der Betrieb schickt Bytes, keine Referenz.
        woerter: [
          ...probe.bestand.map((eines) => ({ ...eines })),
          ...(wahl.feste ?? []),
        ],
        modi: wahl.modi ?? ['core', 'custom'],
        fehler: null,
      }
    },
    SchluesselwortSetzen: probe.SchluesselwortSetzen,
    SchluesselwortLoeschen: probe.SchluesselwortLoeschen,
    SchluesselwortAnlegen: probe.SchluesselwortAnlegen,

    // **Die Parameterhilfen sind Beiwerk** jedes gewählten Wortes — die
    // häufigste Antwort: das Spiel kennt keine.
    Parameterhilfen: () =>
      Promise.resolve({ arten: [], verweise: [], befunde: [], fehler: null }),
    ...wahl.belegen,
  })

  return probe
}

// zeige zeichnet die Seite über einem lebenden Stand.
export function zeige(woerter: SchluesselwortDTO[], wahl: Wahl = {}): Probe {
  const probe = belege(woerter, wahl)

  const seite = (verlaufsstand: number) => (
    <Sprachraum sprache="de">
      <Schluesselwoerter setID="3" verlaufsstand={verlaufsstand} />
    </Sprachraum>
  )
  const { rerender } = render(seite(0))
  probe.zeichneNeu = (verlaufsstand) => {
    rerender(seite(verlaufsstand))
  }

  return probe
}

// liste ist das Gitter der eigenen Wörter.
export async function liste(): Promise<HTMLElement> {
  return screen.findByRole('grid', { name: texte.eigene })
}

// zeile gibt die Zeile eines Wortes nach seinem Namen.
export async function zeile(name: string): Promise<HTMLElement> {
  return within(await liste()).findByRole('row', { name })
}

// waehle wählt ein Wort mit der Maus.
export async function waehle(name: string): Promise<void> {
  const reihe = await zeile(name)
  await userEvent.click(within(reihe).getByRole('button'))
}

// editor gibt den Editor des gewählten Wortes.
export async function editor(name: string): Promise<HTMLElement> {
  return screen.findByRole('group', { name: `${texte.bearbeiten} — ${name}` })
}
