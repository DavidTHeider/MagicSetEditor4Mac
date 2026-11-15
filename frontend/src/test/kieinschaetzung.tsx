// **Die Vorrichtung für die Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// **Ein eigenes Modul**, weil zwei Läufe dieselbe Vorrichtung brauchen: der
// über den Weg und der über die Freigabe. Zwei Fassungen daneben liefen beim
// ersten Sonderfall auseinander.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { KIEinschaetzung } from '../bausteine/KIEinschaetzung'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type {
  KIBalanceDTO,
  KIEinschaetzungAntwort,
} from '../antworteneinschaetzung'

export const hinweistexte = kataloge.de.seiten.einstellungen

// einLokalerKatalog führt ein Modell, das **nichts** kostet.
//
// **Voreingestellt lokal**, denn der gewöhnliche Weg dieses Kastens ist der
// über ein Modell auf diesem Rechner. Wo es um die Freigabe geht, setzt der
// Lauf `cloud`.
export function einLokalerKatalog(ort = 'lokal') {
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
        anbieter: ort === 'lokal' ? 'ollama' : 'openai',
        alias: 'gemma4:8b',
        laufzeit: 'Ollama 0.12.3',
        ort,
        faehigkeiten: [],
        herkunft: [],
        bindung: (ort === 'lokal' ? 'ollama' : 'openai') + '/gemma4:8b',
        kontextfenster: 0,
        kontextbeleg: '',
      },
    ],
    fehler: null,
  }
}

// einBeleg ist ein Ausschnitt, wie der Kontextdienst ihn vergibt.
export function einBeleg(teil: Record<string, unknown> = {}) {
  return {
    kennung: 'B1',
    quellkennung: 'regelwerk:509.1:0',
    quelle: 'regelwerk',
    kennungssystem: '',
    fachkennung: '',
    semantik: '',
    titel: '509.1 Blocks erklären',
    ausschnitt: '509.1. Zuerst prüft das Spiel, ob die Blocks legal sind.',
    inhaltshash: 'abc',
    sprache: 'de',
    regelnummer: '509.1',
    urheber: '',
    bezugsadresse: '',
    quellrevision: '2026-06-07',
    quelldatum: '2026-06-07',
    abgerufen: '2026-09-14',
    veraltet: false,
    ...teil,
  }
}

// eineEinschaetzung baut eine Antwort — voreingestellt eine belegte.
export function eineEinschaetzung(
  teil: Partial<KIEinschaetzungAntwort> = {},
): KIEinschaetzungAntwort {
  return {
    art: 'regel',
    satz: 'Der Block ist legal [B1].',
    belegt: true,
    vorbehalt:
      'Das ist die Einschätzung eines Modells und keine amtliche Regelauskunft.',
    befunde: [],
    einschraenkungen: ['Nur die mitgegebenen Stellen.'],
    belege: [einBeleg()],
    korpusstaende: [
      { quelle: 'regelwerk', revision: '2026-06-07', datum: '2026-06-07' },
    ],
    anbieter: 'ollama',
    modell: 'gemma4:8b',
    balance: null,
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

// eineFreigabe ist der ausgestellte Nachweis.
export function eineFreigabe() {
  return Promise.resolve({
    kennung: 'nw-1',
    laeuftab: '2026-09-14T10:05:00Z',
    fehler: null,
  })
}

// zeigeHinweiskasten stellt ihn hin.
export function zeigeHinweiskasten() {
  render(
    <Sprachraum sprache="de">
      <KIEinschaetzung karteID="4" />
    </Sprachraum>,
  )

  return userEvent.setup()
}

// aufschlagenUndFragen klappt auf, füllt aus und drückt.
export async function aufschlagenUndFragen(
  hand: ReturnType<typeof userEvent.setup>,
  bindung = 'ollama/gemma4:8b',
) {
  await hand.click(screen.getByText(hinweistexte.kiEinschaetzungTitel))

  await hand.selectOptions(
    await screen.findByLabelText(hinweistexte.kiVorschlagAnbieter),
    bindung,
  )

  await hand.type(
    screen.getByLabelText(hinweistexte.kiEinschaetzungFrage),
    'Wie wirkt sie beim Blocken?',
  )

  await hand.click(
    screen.getByRole('button', { name: hinweistexte.kiEinschaetzungFragen }),
  )
}

// eineBalanceantwort baut eine Antwort mit Vergleichsangaben.
//
// **Mit genau einem Vergleich und einem Unterschied**: mehr prüfte die
// Schleife und nicht die Aussage, weniger wäre ein Vergleich ohne
// Unterschied — und der ist ein eigener Fall.
export function eineBalanceantwort(
  teil: Partial<KIBalanceDTO> = {},
): KIEinschaetzungAntwort {
  return eineEinschaetzung({
    art: 'balance',
    satz: 'Sie kostet eins mehr als [B1].',
    balance: {
      strategie: 'Ähnlichkeitssuche über die Quellen des Falls »aehnlich«',
      format: 'Commander',
      staende: [
        { quelle: 'oracle', revision: '2026-09-01', datum: '2026-09-01' },
      ],
      vergleiche: [
        {
          beleg: 'B1',
          name: 'Funkenwurf',
          unterschiede: ['ein Mana billiger'],
        },
      ],
      nichtBetrachtet: ['Zusammenspiel mit Manaflut'],
      ...teil,
    },
  })
}
