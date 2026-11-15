// **Die Bühne des Turnierassistenten** — DOM-04.AK-01
// (#1699).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen und beide unter der
// Zeilengrenze aus §2 bleiben sollen: `turnierassistent` fragt nach dem Weg,
// `turnierassistentfehler` nach den Fehlerwegen.
//
// Das Doppel hält **einen** Stand, den jeder schreibende Ruf fortschreibt —
// sonst prüfte ein Lauf Schritte, die nie aufeinander folgen.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { Turniere } from '../seiten/Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'
import { kataloge } from '../sprache/kataloge'
import {
  paarungsprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from '../seiten/turnierprobe'
import type { Api } from '../api'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

export const pauper = {
  id: '4',
  fassung: '1',
  name: 'Pauper',
  spiel: 'magic',
  mindestgroesse: 60,
  hoechstgroesse: 0,
  sideboardgroesse: 15,
  hoechstzahlGleich: 4,
  sideboardOffen: false,
  scryfallname: 'pauper',
  bauart: 'konstruiert',
}

// buehne stellt die Seite mit einem Doppel hin, das mitschreibt.
export function buehne(
  vorhanden: TurnierDTO[] = [],
  weitere: Partial<Api> = {},
) {
  let turniere = [...vorhanden]
  const finde = (kennung: string) =>
    turniere.find((turnier) => turnier.id === kennung) ?? null
  const ersetze = (neu: TurnierDTO) => {
    turniere = [...turniere.filter((t) => t.id !== neu.id), neu]
  }

  const anlegen = vi.fn<Api['TurnierAnlegen']>(
    (name, spiel, notiz, einzelspiel, formatID) => {
      const neu = turnierprobe({
        id: '8',
        name,
        spiel,
        notiz,
        einzelspiel,
        formatID,
      })
      ersetze(neu)

      return Promise.resolve({ turnier: neu, fehler: null })
    },
  )
  const kopieren = vi.fn<Api['TurnierKopieren']>((quelle, name, mit) => {
    const vorlage = finde(quelle) ?? turnierprobe()
    const neu = turnierprobe({
      ...vorlage,
      id: '9',
      name,
      runden: [],
      teilnehmer: mit
        ? vorlage.teilnehmer.map((teilnehmer) => ({
            ...teilnehmer,
            deckID: '',
            deckabzug: [],
          }))
        : [],
    })
    ersetze(neu)

    return Promise.resolve({ turnier: neu, fehler: null })
  })
  const anmelden = vi.fn<Api['TurnierAnmelden']>((kennung, name) => {
    const turnier = finde(kennung) ?? turnierprobe()
    const teilnehmer = teilnehmerprobe({
      id: String(turnier.teilnehmer.length + 20),
      name,
      stelle: turnier.teilnehmer.length + 1,
    })
    ersetze({ ...turnier, teilnehmer: [...turnier.teilnehmer, teilnehmer] })

    return Promise.resolve({ teilnehmer, fehler: null })
  })
  const legen = vi.fn<Api['TurnierRundeLegen']>((kennung) => {
    const turnier = finde(kennung) ?? turnierprobe()
    const runde = rundenprobe({
      paarungen: [paarungsprobe({ a: '20', b: '21' })],
    })
    ersetze({ ...turnier, runden: [runde] })

    return Promise.resolve({ runde, fehler: null })
  })

  belegeApi({
    Turniere: () => Promise.resolve({ turniere, fehler: null }),
    Turnier: (kennung) =>
      Promise.resolve({ turnier: finde(kennung), fehler: null }),
    TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    Formate: () =>
      Promise.resolve({
        formate: [pauper, { ...pauper, id: '5', spiel: 'pokemon' }],
        fehler: null,
      }),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    TurnierPaarungsvorschlag: () =>
      Promise.resolve({
        paarungen: [paarungsprobe({ a: '20', b: '21' })],
        fehler: null,
      }),
    TurnierAnlegen: anlegen,
    TurnierKopieren: kopieren,
    TurnierAnmelden: anmelden,
    TurnierRundeLegen: legen,
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )

  return { anlegen, kopieren, anmelden, legen }
}

// schritt sagt, welcher Schritt gerade gilt.
export function schritt() {
  const folge = screen.getByRole('list', { name: texte.assistent })

  return within(folge)
    .getAllByRole('listitem')
    .find((eintrag) => eintrag.getAttribute('aria-current') === 'step')
    ?.textContent
}

// assistent öffnet den Assistenten über den Kopf der Seite und gibt seine
// Schublade — seit UXR-12 (#1920) steht er nicht mehr offen über der Liste.
export async function assistent(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.neuOeffnen }),
  )

  return screen.findByRole('dialog', { name: texte.neu })
}

export async function melde(name: string) {
  const bereich = screen.getByRole('dialog', { name: texte.neu })

  // **Auf das Feld warten** — der Schritt *Teilnehmende* steht schon in der
  // Schrittfolge, bevor das angelegte Turnier geladen und sein Feld
  // gezeichnet ist (unter Last am 2026-10-02 gesehen).
  await userEvent.type(await within(bereich).findByLabelText(texte.name), name)
  await userEvent.click(
    within(bereich).getByRole('button', { name: texte.anmelden }),
  )
}
