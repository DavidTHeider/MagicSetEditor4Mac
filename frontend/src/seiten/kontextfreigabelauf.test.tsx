// **Eine Antwort gehört zu dem Stand, aus dem sie stammt** — F03 aus dem
// Review zu J05
// (#2269),
// AGENTS.md §7, Regel 2.
//
// # Was bis dahin geschah
//
// Eine Eingabe räumte das Paket — **aber die Felder blieben änderbar, während
// der Ruf lief**, und seine Antwort trug das Paket danach bedingungslos
// wieder ein. Wer das Budget auf „8192.5" stellte (der Knopf ist dann
// angehalten), sah trotzdem ein Paket, und zwar eines für ein Budget, das
// gar nicht gilt. Dasselbe galt für Anbieter und Modell — und für den
// Fehler, der über einer Eingabe stand, die er nie betraf.
//
// # Wogegen hier geprüft wird
//
// Eine Antwort, die eintrifft, **nachdem** sich eine Eingabe geändert hat,
// wird verworfen — Erfolg wie Fehler. Die gegenprobe steht daneben: dieselbe
// Antwort **ohne** Änderung trägt sich ein, und ein neuer Ruf nach dem
// verworfenen auch. Eine Verneinung ohne diese Hälfte bewiese nur, dass
// nichts mehr ankommt.
//
// **Das Warten ist das auf die Antwort** (`antwortAbwarten`): „es wurde nichts
// eingetragen" ist erst eine Aussage, wenn die Antwort da war (§3).
import { render, screen } from '@testing-library/react'
import userEvent, { type UserEvent } from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Kontextfreigabe } from './Kontextfreigabe'
import { einPaket } from './kontextpaketprobe'

const texte = kataloge.de.seiten.recherche

type Paket = ReturnType<typeof einPaket>

interface Haltende {
  gerufen: ReturnType<typeof vi.fn>

  // antworte beantwortet den n-ten Ruf (von null an) — wann der Test es will.
  antworte: (nummer: number, paket: Paket) => void
  scheitere: (nummer: number) => void
}

// haltend belegt den Dienst mit Rufen, die **hängen**, bis der Test sie
// beantwortet. Nur so lässt sich eine Eingabe zwischen Ruf und Antwort legen.
function haltend(): Haltende {
  const offene: {
    gut: (paket: Paket) => void
    schlecht: (grund: unknown) => void
  }[] = []

  const gerufen = vi.fn(
    () =>
      new Promise<Paket>((gut, schlecht) => {
        offene.push({ gut, schlecht })
      }),
  )

  belegeApi({ KontextZusammenstellen: gerufen })

  return {
    gerufen,
    antworte: (nummer, paket) => {
      offene[nummer]?.gut(paket)
    },
    scheitere: (nummer) => {
      offene[nummer]?.schlecht(new Error('kaputt'))
    },
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Kontextfreigabe freigegeben={[]} fall="regeln" frage="trampelschaden" />
    </Sprachraum>,
  )
}

// zusammenstellen tippt eine Aufgabe und drückt den Knopf. Der Ruf ist danach
// draußen — die Fassade wird sofort gerufen und antwortet später (§3).
async function zusammenstellen(nutzer: UserEvent) {
  await nutzer.type(
    screen.getByLabelText(texte.freigabeAufgabe),
    'Erkläre Trampelschaden',
  )
  await nutzer.click(screen.getByRole('button', { name: texte.freigabeZeigen }))
}

// aenderungen sind die Eingaben, die ein anderes Paket meinen (§7, Regel 1) —
// und damit die, nach denen eine Antwort nicht mehr passt.
const aenderungen: [string, (nutzer: UserEvent) => Promise<void>][] = [
  [
    'der Anbieter wechselt',
    async (nutzer) => {
      await nutzer.selectOptions(
        screen.getByLabelText(texte.freigabeAnbieter),
        'anthropic',
      )
    },
  ],
  [
    'das Modell wechselt',
    async (nutzer) => {
      await nutzer.type(screen.getByLabelText(texte.freigabeModell), 'g')
    },
  ],
  [
    'die Ausgabereserve wechselt',
    async (nutzer) => {
      await nutzer.type(screen.getByLabelText(texte.freigabeReserve), '0')
    },
  ],

  // **Auch die Aufgabe** (Befund 5 des Nachlaufs über J05, #2269): sie geht
  // in das Paket ein wie Anbieter und Modell.
  [
    'die Aufgabe wechselt',
    async (nutzer) => {
      await nutzer.type(screen.getByLabelText(texte.freigabeAufgabe), '!')
    },
  ],

  // **Der Fall, an dem es aufgefallen ist**: das Budget wird ungültig, der
  // Knopf ist angehalten, und die Antwort trug trotzdem ein Paket ein.
  [
    'das Fenster ungültig wird',
    async (nutzer) => {
      const feld = screen.getByLabelText(texte.freigabeFenster)

      await nutzer.clear(feld)
      await nutzer.type(feld, '8192.5')
    },
  ],
]

describe('eine Antwort, die eine Eingabe überholt hat', () => {
  afterEach(() => {
    raeumeApi()
  })

  it.each(aenderungen)(
    'trägt kein Paket ein, wenn %s',
    async (_name, aendere) => {
      const dienst = haltend()

      zeige()

      const nutzer = userEvent.setup()

      await zusammenstellen(nutzer)

      expect(dienst.gerufen).toHaveBeenCalledTimes(1)

      await aendere(nutzer)

      dienst.antworte(0, einPaket())
      await antwortAbwarten()

      expect(screen.queryByText(texte.freigabeOrtLokal)).not.toBeInTheDocument()
    },
  )

  it.each(aenderungen)(
    'zeigt keinen Fehler, wenn %s',
    async (_name, aendere) => {
      const dienst = haltend()

      zeige()

      const nutzer = userEvent.setup()

      await zusammenstellen(nutzer)

      expect(dienst.gerufen).toHaveBeenCalledTimes(1)

      await aendere(nutzer)

      dienst.scheitere(0)
      await antwortAbwarten()

      // **Die Meldung des Dienstes und nicht „irgendein Alarm"**: ein
      // ungültiges Feld meldet sich selbst als `alert`, und das gehört nicht
      // zum verworfenen Ruf.
      expect(screen.queryByText('kaputt')).not.toBeInTheDocument()
    },
  )

  // **Die Gegenprobe, Hälfte eins**: dieselbe verspätete Antwort, ohne dass
  // sich etwas geändert hat, gilt.
  it('trägt das Paket ein, wenn sich nichts geändert hat', async () => {
    const dienst = haltend()

    zeige()
    await zusammenstellen(userEvent.setup())

    dienst.antworte(0, einPaket())

    expect(await screen.findByText(texte.freigabeOrtLokal)).toBeInTheDocument()
  })

  it('zeigt den Fehler, wenn sich nichts geändert hat', async () => {
    const dienst = haltend()

    zeige()
    await zusammenstellen(userEvent.setup())

    dienst.scheitere(0)

    expect(await screen.findByText('kaputt')).toBeInTheDocument()
  })

  // **Hälfte zwei**: das Verwerfen sperrt den Weg nicht. Nach der verworfenen
  // Antwort ist der Knopf wieder frei, und der nächste Ruf gilt für den
  // Stand, der jetzt dasteht.
  it('nimmt nach einer verworfenen Antwort den nächsten Ruf an', async () => {
    const dienst = haltend()

    zeige()

    const nutzer = userEvent.setup()

    await zusammenstellen(nutzer)
    await nutzer.selectOptions(
      screen.getByLabelText(texte.freigabeAnbieter),
      'anthropic',
    )

    dienst.antworte(0, einPaket())
    await antwortAbwarten()

    const knopf = screen.getByRole('button', { name: texte.freigabeZeigen })

    expect(knopf).toBeEnabled()

    await nutzer.click(knopf)

    expect(dienst.gerufen).toHaveBeenCalledTimes(2)

    dienst.antworte(1, einPaket())

    expect(await screen.findByText(texte.freigabeOrtLokal)).toBeInTheDocument()
  })
})

// **Ein anderer Rechercheauftrag ist ein anderer Stand** (Befund 5 des
// Nachlaufs über J05, #2269): Fall, Frage und freigegebene Korpora kommen von
// der Seite darüber, und wechseln sie, während der Ruf läuft, gehört seine
// Antwort zu einem Auftrag, den niemand mehr hat.
describe('ein Auftrag, der während des Rufs wechselt', () => {
  afterEach(() => {
    raeumeApi()
  })

  it.each<[string, { fall?: string; frage?: string; freigegeben?: string[] }]>([
    ['die Frage', { frage: 'andere frage' }],
    ['der Fall', { fall: 'karten' }],
    ['die Freigabe', { freigegeben: ['eigene'] }],
  ])('trägt kein Paket ein, wenn %s wechselt', async (_name, neu) => {
    const dienst = haltend()

    const auftrag = {
      fall: 'regeln',
      frage: 'trampelschaden',
      freigegeben: [] as string[],
    }
    const { rerender } = render(
      <Sprachraum sprache="de">
        <Kontextfreigabe {...auftrag} />
      </Sprachraum>,
    )

    const nutzer = userEvent.setup()
    await zusammenstellen(nutzer)
    expect(dienst.gerufen).toHaveBeenCalledTimes(1)

    rerender(
      <Sprachraum sprache="de">
        <Kontextfreigabe {...auftrag} {...neu} />
      </Sprachraum>,
    )

    dienst.antworte(0, einPaket())
    await antwortAbwarten()

    expect(screen.queryByText(texte.freigabeOrtLokal)).not.toBeInTheDocument()
  })

  // **Die Gegenprobe**: ein neues Zeichnen mit **demselben** Auftrag — eine
  // neue Liste mit denselben Einträgen — lässt die Antwort gelten.
  it('trägt das Paket ein, wenn derselbe Auftrag neu gezeichnet wird', async () => {
    const dienst = haltend()

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Kontextfreigabe
          fall="regeln"
          frage="trampelschaden"
          freigegeben={['eigene']}
        />
      </Sprachraum>,
    )

    const nutzer = userEvent.setup()
    await zusammenstellen(nutzer)

    rerender(
      <Sprachraum sprache="de">
        <Kontextfreigabe
          fall="regeln"
          frage="trampelschaden"
          freigegeben={['eigene']}
        />
      </Sprachraum>,
    )

    dienst.antworte(0, einPaket())
    await antwortAbwarten()

    expect(await screen.findByText(texte.freigabeOrtLokal)).toBeInTheDocument()
  })
})
