// **Das Profil wechseln** — Werkstück 3 von Phase 4.12
// (#1351).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Profilraum } from '../profil'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'

import type { EinProfilAntwort, NutzerDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
})

const zwei: NutzerDTO[] = [
  {
    id: '1',
    name: 'Ich',
    bild: '',
    email: '',
    telefon: '',
    anonym: false,
    anrede: '',
    namenszusatz: '',
    vorname: '',
    weiterenamen: '',
    nachname: '',
    slogan: '',
    kurzebio: '',
    langebio: '',
    fassung: '1',
    profile: [
      probeprofil({
        id: '10',
        nutzerID: '1',
        name: 'Mein Profil',
        fassung: '1',
      }),
      probeprofil({ id: '11', nutzerID: '1', name: 'Turnier', fassung: '1' }),
    ],
  },
]

function zeige(gewechselt = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Profilraum gewechselt={gewechselt}>
        <Nutzerseite />
      </Profilraum>
    </Sprachraum>,
  )

  return gewechselt
}

// zumProfil öffnet die Lasche *Profile*.
async function zumProfil() {
  await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
}

describe('das Profil wechseln', () => {
  // **Das gezeigte Profil trägt keinen Knopf** (§6): eine Handlung, die
  // nichts tut, ist keine.
  it('nennt das gezeigte Profil und bietet es nicht zum Wechsel an', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
    })
    zeige()
    await zumProfil()

    expect(
      await screen.findByText('Hier arbeitest du gerade'),
    ).toBeInTheDocument()

    // **Genau einer**, nämlich der für das andere Profil.
    expect(
      await screen.findAllByRole('button', { name: /wechseln/ }),
    ).toHaveLength(1)
  })

  // **Und ein Wechsel meldet nach oben** — dort baut sich die Anwendung neu
  // auf (§7, Regel 1).
  it('wechselt und meldet es', async () => {
    const wechseln = vi
      .fn()
      .mockResolvedValue({ profil: zwei[0]?.profile[1] ?? null })

    belegeApi({
      ...belegeAnmeldung(),
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
      ProfilWechseln: wechseln,
    })

    const gewechselt = zeige()
    await zumProfil()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    // **Gewartet wird auf die Wirkung und nicht auf den Ruf** (§3): der Ruf
    // ist geschehen, die Antwort nicht — und was sie bewirkt, steht dann noch
    // nicht da.
    await vi.waitFor(() => {
      expect(gewechselt).toHaveBeenCalled()
    })

    expect(wechseln).toHaveBeenCalledWith('11', '')
  })

  // **Ein gescheiterter Wechsel meldet und baut nichts neu auf.**
  //
  // §7, Regel 3 von der anderen Seite: was **nicht** geschrieben wurde, darf
  // die Oberfläche nicht so behandeln, als wäre es geschehen — ein Neuaufbau
  // sähe aus wie ein geglückter Wechsel.
  it('meldet einen gescheiterten Wechsel und baut nichts neu auf', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
      ProfilWechseln: vi.fn(() =>
        Promise.resolve({
          profil: null,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'die Platte schweigt',
          },
        }),
      ),
    })

    const gewechselt = zeige()
    await zumProfil()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    // **Der Satz kommt aus dem Katalog und nicht aus Go** (§9): die Bindung
    // schickt einen **Schlüssel** neben ihrer deutschen Meldung, und das
    // Fenster sagt den Satz in seiner Sprache. Hier den Go-Text zu erwarten
    // prüfte, dass zwei Kopien gleich sind.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )

    expect(gewechselt).not.toHaveBeenCalled()
  })

  // **Ohne Profilraum geschieht nichts, und es scheitert auch nichts.**
  //
  // Ein Bauteil außerhalb — in einer Vorschau, in einem Test über nur diese
  // Seite — hat keine Anwendung, die sich neu aufbauen ließe. Der Vorgabewert in
  // `profil.tsx` ist deshalb eine leere Funktion; dass sie **gerufen** wird und
  // nicht `undefined` ist, prüft dieser Lauf.
  it('kommt ohne Profilraum aus', async () => {
    const wechseln = vi.fn().mockResolvedValue({ profil: zwei[0]?.profile[1] })

    belegeApi({
      ...belegeAnmeldung(),
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
      ProfilWechseln: wechseln,
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await zumProfil()
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    await vi.waitFor(() => {
      expect(wechseln).toHaveBeenCalledWith('11', '')
    })
  })

  // **Ein Bestand ohne gezeigtes Profil zeigt jede Wahl als wählbar.**
  //
  // Der Fall kommt im Betrieb nicht vor — es gibt immer eines (Wanderung 0034).
  // Käme er doch, wäre eine Seite mit **keinem** Knopf die schlechtere Antwort
  // als eine mit zweien: wer nirgends steht, muss irgendwohin können.
  it('bietet jedes Profil an, wenn keines gezeigt wird', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: null })),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await zumProfil()

    expect(
      await screen.findAllByRole('button', { name: /wechseln/ }),
    ).toHaveLength(2)
  })
})

// **Ein Vorgang für den ZEIGER, nicht je Ziel** — der Nachlauf zum Zweitblick
// über Phase 4.12
// (#1351).
//
// # Der Befund
//
// **`useVorgang` sperrt gleiche Schlüssel**, und der Schlüssel war die
// **Ziel**-Kennung. Zwei schnelle Klicks auf **verschiedene** Profile hatten
// damit zwei Schlüssel und liefen nebeneinander — obwohl beide denselben
// Datensatz schreiben: den Profilzeiger. §7, Regel 4 verlangt einen Vorgang je
// **Datensatz**, und der Datensatz ist hier der Zeiger.
describe('zwei Wechsel zugleich', () => {
  it('lässt nur einen laufen, auch auf verschiedene Ziele', async () => {
    let loese: (() => void) | undefined

    const gewechselt = vi.fn(
      () =>
        new Promise<EinProfilAntwort>((fertig) => {
          loese = () => {
            fertig({ profil: null, fehler: null })
          }
        }),
    )

    // **Drei Profile, damit ZWEI Knöpfe dastehen.** Das gezeigte trägt
    // keinen; mit nur einem übrigen prüfte der Lauf unten eine Liste aus
    // einem Element und wäre auch mit dem alten Schlüssel grün (§10).
    const [einer] = zwei
    if (einer === undefined) {
      throw new Error('die Vorrichtung führt keinen Nutzer')
    }

    const drei: NutzerDTO[] = [
      {
        ...einer,
        profile: [
          ...einer.profile,
          probeprofil({
            id: '12',
            nutzerID: '1',
            name: 'Drittes',
            fassung: '1',
          }),
        ],
      },
    ]

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: drei[0]?.profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: drei })),
      ProfilWechseln: gewechselt,
    })

    zeige()
    await zumProfil()

    // **Über die Beschriftung und nicht über den sichtbaren Text**: der
    // Knopf trägt „Zu … wechseln" als zugängliche Bezeichnung, damit eine
    // Liste gleichlautender Knöpfe unterscheidbar bleibt (§6).
    const knoepfe = await screen.findAllByRole('button', {
      name: /wechseln$/,
    })

    // **Ein Knopf, während der erste Wechsel noch läuft** — und der zweite
    // zeigt auf ein **anderes** Profil.
    const [erster] = knoepfe
    if (erster === undefined) {
      throw new Error('kein Knopf zum Wechseln')
    }

    await userEvent.click(erster)

    for (const knopf of knoepfe) {
      expect(knopf).toBeDisabled()
    }

    loese?.()

    expect(gewechselt).toHaveBeenCalledTimes(1)
  })
})
