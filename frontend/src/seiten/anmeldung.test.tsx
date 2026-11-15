// **Die Anmeldefläche** — Werkstück 4 von Phase 4.12
// (#1351).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'

import type { NutzerDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
})

const einer: NutzerDTO = {
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
    probeprofil({ id: '10', nutzerID: '1', name: 'Mein Profil', fassung: '1' }),
  ],
}

const gezeigtesProfil = () =>
  Promise.resolve({ profil: einer.profile[0] ?? null })

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// zumKonto öffnet die Lasche *Konto* — und seit UXR-20 (#1928) den Kasten
// der Anbieter: er steht dort eingeklappt hinter Sitzung und zweitem Faktor.
async function zumKonto() {
  await userEvent.click(await screen.findByRole('tab', { name: 'Konto' }))
  await userEvent.click(
    await screen.findByRole('button', {
      name: 'Anmeldung über einen Anbieter',
    }),
  )
}

describe('die Anmeldung', () => {
  // **Vier Anbieter, vier Knöpfe** — und keiner davon hat etwas angestoßen.
  //
  // **Das ist die Zusage aus §16, Regel 5**: wer die Seite öffnet, hat nichts
  // getan. Erst ein Klick macht einen Browser auf.
  it('bietet jeden Anbieter an, ohne etwas anzustoßen', async () => {
    const anmelden = vi.fn()

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      Anmelden: anmelden,
    })
    zeige()
    await zumKonto()

    // **Im Abschnitt der Anbieter und nicht auf der ganzen Lasche.** Seit
    // Werkstück 6 (#1351) steht daneben ein „Anmelden" für die eigene
    // Kennung; eine Suche über die Seite zählte es mit und sagte damit etwas
    // anderes, als ihr Name behauptet.
    const abschnitt = within(
      await screen.findByRole('region', {
        name: 'Anmeldung über einen Anbieter',
      }),
    )

    // **„Bei X anmelden" und nicht bloß „anmelden"** — seit UX-15
    // (#1771)
    // steht neben der Erläuterung ein Hilfeknopf, und der heißt nach dem
    // Satz, den er fortsetzt — in dem das Wort „anmelden" vorkommt. Eine
    // Suche nach einem Wort zählte ihn mit und sagte damit etwas anderes,
    // als ihr Name behauptet (§10).
    expect(
      await abschnitt.findAllByRole('button', { name: /^Bei .+ anmelden$/ }),
    ).toHaveLength(4)

    expect(anmelden).not.toHaveBeenCalled()
  })

  // **Und sie sagt, was sie braucht** (§6): dieses Programm liefert keine
  // Zugangsdaten mit.
  it('sagt, dass Zugangsdaten selbst zu registrieren sind', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()
    await zumKonto()

    // **Und sagt auch, WO** — Nacharbeit zu diesem Werkstück (#1351): der
    // Satz verwies auf „die Einstellungen", und dort stand nichts. Er nennt
    // jetzt das Gebiet, und die Einstellung dahinter gibt es auch.
    //
    // **Und zwar im ersten Satz** — UX-15.AK-02
    // (#1771):
    // was man zum Entscheiden braucht, darf nicht hinter dem Fragezeichen
    // stehen. Der Satz fing bis zum 2026-09-22 mit „liefert keine
    // Zugangsdaten mit" an — eine Auskunft, kein Weg —, und das Wo wäre
    // damit verschwunden.
    expect(
      await screen.findByText(/Wer sich über einen Anbieter anmelden will/),
    ).toBeInTheDocument()

    // **Und das Gebiet steht hervorgehoben da, nicht mit Sternchen.**
    //
    // Hier stand `/Anmeldung über einen Anbieter\* ein/` — mit dem wörtlichen
    // Sternchen, denn der Katalog schrieb `*…*`, und `Satz` setzt nur
    // `**fett**` (#699). Bei der Sichtprüfung am 2026-09-07 fielen siebzehn
    // solche Stellen auf; seither steht dort `**…**`, und der Satz ist in
    // zwei Knoten geteilt. **Dieser Test war der Beleg, dass die Änderung
    // ankommt** — er wurde rot.
    //
    // **Und seit OPS-07.NAT-01 (#1724) mit dem ganzen Weg**: hier stand
    // „Einstellungen → Anmeldung über einen Anbieter", und das ist keine
    // Gruppe der Einstellungen, sondern ein Kasten darin.
    expect(
      screen.getByText(
        'Einstellungen › Nutzer und Profil › Anmeldung über einen Anbieter',
      ),
    ).toBeInTheDocument()
  })

  // **Ein Klick meldet an und lädt neu.**
  it('meldet an und lädt danach neu', async () => {
    const anmelden = vi.fn().mockResolvedValue({ nutzer: einer })
    const nutzer = vi.fn(() => Promise.resolve({ nutzer: [einer] }))

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: nutzer,
      Anmelden: anmelden,
    })
    zeige()
    await zumKonto()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Bei GitHub anmelden' }),
    )

    await vi.waitFor(() => {
      expect(anmelden).toHaveBeenCalledWith('github')
    })

    // **Und die Seite liest neu** — das Konto steht danach da.
    await vi.waitFor(() => {
      expect(nutzer.mock.calls.length).toBeGreaterThan(1)
    })
  })

  // **Ein verbundenes Konto zeigt seinen Namen und einen Weg zurück.**
  //
  // **Der Rückweg gehört dazu**: wer sein Konto beim Anbieter schließt, darf
  // nicht ausgesperrt sein.
  it('zeigt ein verbundenes Konto und löst es', async () => {
    const loesen = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      ...belegeAnmeldung(),
      Anbieterkonten: vi.fn(() =>
        Promise.resolve({
          konten: [{ anbieter: 'github', anzeigename: 'Wer' }],
        }),
      ),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      KontoLoesen: loesen,
    })
    zeige()
    await zumKonto()

    expect(await screen.findByText('Verbunden als Wer')).toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Die Verbindung zu GitHub lösen',
      }),
    )

    await vi.waitFor(() => {
      expect(loesen).toHaveBeenCalledWith('1', 'github')
    })
  })

  // **Ein verbundenes Konto ohne Anzeigenamen sagt trotzdem, dass es steht.**
  it('zeigt ein verbundenes Konto ohne Namen', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      Anbieterkonten: vi.fn(() =>
        Promise.resolve({ konten: [{ anbieter: 'x', anzeigename: '' }] }),
      ),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()
    await zumKonto()

    expect(await screen.findByText('Verbunden')).toBeInTheDocument()
  })

  // **Ein gescheiterter Ruf meldet sich am Ort und lädt nicht neu.**
  //
  // §7, Regel 3 von der anderen Seite: nach einem gescheiterten Schreiben wird
  // gar nicht nachgeladen — sonst überschriebe ein Lesefehler die Meldung, die
  // sagt, was zu tun ist.
  it('meldet einen gescheiterten Ruf', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      Anmelden: vi.fn(() =>
        Promise.resolve({
          nutzer: null,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'validierung',
            meldung: 'für github ist keine Kennung eingetragen',
          },
        }),
      ),
    })
    zeige()
    await zumKonto()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Bei GitHub anmelden' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })

  // **Und ein gescheitertes Lösen ebenso.**
  it('meldet ein gescheitertes Lösen', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      Anbieterkonten: vi.fn(() =>
        Promise.resolve({
          konten: [{ anbieter: 'github', anzeigename: 'Wer' }],
        }),
      ),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      KontoLoesen: vi.fn(() =>
        Promise.resolve({
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'die Platte schweigt',
          },
        }),
      ),
    })
    zeige()
    await zumKonto()

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Die Verbindung zu GitHub lösen',
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Ein Anbieter, den der Katalog nicht kennt, steht mit seinem rohen
  // Namen da.**
  //
  // Der Fall ist keiner von heute — aber Vermerk 0066 sagt, OpenAI wäre „eine
  // Zeile, sobald es einen Weg gibt". Bis der Katalog nachzieht, ist ein roher
  // Name besser als eine leere Zeile (§6).
  it('zeigt einen unbekannten Anbieter mit seinem rohen Namen', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      Anbieter: vi.fn(() => Promise.resolve({ anbieter: ['openai'] })),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()
    await zumKonto()

    expect(await screen.findByText('openai')).toBeInTheDocument()
  })

  // **Je Anbieter: eingerichtet, und wogegen geprüft** — ID-05.AK-01
  // (#1704).
  it('sagt je Anbieter, ob er eingerichtet und wogegen er geprüft ist', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      Anbieter: vi.fn(() =>
        Promise.resolve({
          anbieter: ['github', 'google'],
          staende: [
            {
              anbieter: 'github',
              eingerichtet: true,
              geprueft: 'simuliert' as const,
            },
            {
              anbieter: 'google',
              eingerichtet: false,
              geprueft: 'realGeprueft' as const,
            },
          ],
          fehler: null,
        }),
      ),
    })
    zeige()
    await zumKonto()

    const texte = kataloge.de.seiten.nutzer
    expect(
      await screen.findByText(
        `${texte.eingerichtet} · ${texte.geprueft.simuliert}`,
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText(
        `${texte.nichtEingerichtet} · ${texte.geprueft.realGeprueft}`,
      ),
    ).toBeInTheDocument()
  })
})
