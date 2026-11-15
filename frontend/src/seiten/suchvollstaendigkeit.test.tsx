// **Die Suche findet jede Einstellung** — UX-04.AK-03
// (#1676),
// zugleich NET-02.NAT-03 und S13.NAT-01
// (#1706).
//
// # Der Anlass
//
// **Die Suche nach „proxy" sagte „Keine Einstellung passt."** — gemessen am
// laufenden Fenster am 2026-09-21. Der Grund stand in
// `seiten/Anwendungsinhalt.tsx`: die Seite bekam `imBereich('anwendung')`,
// die Dienste, die Fragen, die KI-Zeilen, die Vorgaben und
// `imBereich('anmeldung')` — **den Bereich `netz` bekam sie nie**, denn
// `Netzweg` lädt seine Felder selbst.
//
// **Zwei Kriterien verlangen ausdrücklich das Gegenteil:**
//
//   - NET-02.NAT-03: „Proxy ist über Einstellungssuche auffindbar und
//     bearbeitbar oder mit klarer externer Herkunft schreibgeschützt."
//   - S13.NAT-01: „Suche nach Proxy, HTTP und Netz führt zu einem
//     editierbaren passenden Feld oder klarer schreibgeschützter Herkunft."
//
// # Zeichnen und Suchen sind zwei Fragen
//
// **Und bis heute beantwortete sie eine Liste.** Fürs Zeichnen ist sie zu
// Recht eng — jede Einstellung steht **genau einmal** da, und das war der
// Befund vom selben Tag
// (#1768).
// Fürs Suchen ist dieselbe Enge falsch.
//
// # Was er nicht prüft
//
// **Ob der Sprung ankommt.** Er sieht, dass ein Treffer dasteht, nicht dass
// der Knopf daneben zum Feld führt — das steht im Lauf daneben.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import {
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

// **Eine Einstellung, die ein eigenes Bauteil zeichnet.** `netz.proxy.http`
// steht im Bereich `netz`, und die Seite bekommt ihn in keiner ihrer sechs
// Listen.
const proxy: EinstellungDTO = {
  schluessel: 'netz.proxy.http',
  wert: '',
  wirksam: '',
  werte: [],
  bereich: 'netz',
  ebene: 'geraet',
  herkunft: 'vorgabe',
  vorgabe: '',
  wirkung: 'sofort',
}

async function suche(frage: string) {
  await userEvent.type(screen.getByLabelText(eigene.suche), frage)
}

describe('die Einstellungssuche', () => {
  it('findet, was ein eigenes Bauteil zeichnet', async () => {
    zeige({ alleEinstellungen: [...eintraege, proxy] })

    await suche('proxy')

    // **Der Treffer trägt seine Marke** — und zwar dort, wo das Feld steht,
    // nicht in einer zweiten Liste daneben (UX-04.AK-03).
    expect(
      screen
        .getByRole('tabpanel')
        .querySelector('[data-einstellung="netz.proxy.http"]'),
    ).toBeInTheDocument()

    expect(
      within(screen.getByRole('tabpanel')).getByText(eigene.gesuchtMarke),
    ).toBeInTheDocument()
  })

  // **Und sie sagt, in welcher Gruppe es steht.** Ein Treffer ohne Ort ist
  // eine halbe Antwort: wer ihn zumacht, findet ihn nicht wieder.
  it('nennt die Gruppe eines Treffers', async () => {
    zeige({ alleEinstellungen: [...eintraege, proxy] })

    await suche('proxy')

    expect(
      within(screen.getByRole('tabpanel')).getByRole('heading', {
        name: new RegExp(eigene.gruppen.verbindungen),
      }),
    ).toBeInTheDocument()
  })

  // **Der Steuerlauf** (§10): ohne ihn sagte der Lauf darüber nicht, ob
  // überhaupt noch etwas gefunden wird.
  it('findet weiterhin, was die Seite selbst zeichnet', async () => {
    zeige({ alleEinstellungen: eintraege })

    await suche('sprache')

    expect(
      within(screen.getByRole('tabpanel')).getByText(eigene.gesuchtMarke),
    ).toBeInTheDocument()
  })

  // **Ohne die volle Liste sucht sie in dem, was sie zeigt.** Ein Fenster,
  // dem niemand die Liste gibt, hat keine leere Suche — es hat die alte.
  it('kommt ohne die volle Liste aus', async () => {
    zeige({})

    await suche('sprache')

    expect(
      within(screen.getByRole('tabpanel')).getByText(eigene.gesuchtMarke),
    ).toBeInTheDocument()
  })
})
