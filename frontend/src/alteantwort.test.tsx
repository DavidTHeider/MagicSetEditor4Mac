// **Eine alte Antwort übermalt keinen neuen Stand — über der ganzen
// Anwendung** — C05 (#2230).
//
// `haken/ladenverzahnung.test.tsx` belegt den Schutz **im Haken**: mit
// erzeugten Folgen über einen gleichbleibenden Bezug. Der Zweitblick über
// Phase C (Befund 6) fragte, ob der Schutz auch **über die echten Übergänge**
// trägt: hier hält die Deckliste des alten Stands ihre Antwort zurück, die
// Anwendung geht wirklich über — Profilwechsel, Zurückspielen —, und erst
// danach kommt die alte Antwort an.
//
// # Was hier schützt, und was die Gegenprobe zeigte
//
// **Der Abbau des Baums** (§7, die erste Gestalt): `key={profilstand}` an
// `Anwendung` und den Räumen darüber, beim Zurückspielen das Räumen der
// aufgehobenen Seiten (`rueckspielseiten.test.tsx`). Eine alte Antwort
// findet danach kein Bauteil mehr, in das sie fallen könnte. Der Standschutz
// in `useLaden` allein abgeschaltet: grün — er ist hier nicht die Wache.
// Dazu die Schlüssel entfernt: rot, aber **an der Rückkehr zur Startseite**,
// nicht an der alten Antwort — ohne Abbau gibt es den Übergang nicht, den
// dieser Test braucht.
//
// **Was er also hält**: dass keine Antwort, die vor dem Übergang gefragt
// wurde, danach irgendwo erscheint — auch nicht über einen Zwischenspeicher,
// der den Abbau überlebt. Für die Deckliste gibt es heute keinen; die
// Befehlspalette lädt beim Öffnen und ist abgebaut, solange sie zu ist.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import type { FehlerDTO, NutzerDTO, Sicherungsauswahl } from './api'
import type { DeckDTO } from './apidecks'
import { deckprobe } from './seiten/deckprobe'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { belegeAnmeldung } from './test/anmeldebelegung'
import { effekteNachholen } from './test/effekte'
import { zuAbschnitt, zuGebiet } from './test/gebiete'
import { leisteneintrag } from './test/leistenweg'
import { probeprofil } from './test/profil'

afterEach(raeumeApi)

const alt = deckprobe({ id: '1', name: 'Aus dem alten Stand' })
const altfehler: FehlerDTO = {
  code: 'unbekannt',
  schluessel: '',
  meldung: 'Der alte Stand scheiterte',
}
const neu = deckprobe({ id: '2', name: 'Aus dem neuen Stand' })

const ich: NutzerDTO = {
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
    probeprofil({ id: '11', nutzerID: '1', name: 'Turnier', fassung: '1' }),
  ],
}

const sicherung: Sicherungsauswahl = {
  pfad: '/Users/probe/mse-sicherung.msesicherung',
  angelegtAm: '2026-08-01T09:12:33.000Z',
  schemastand: '0012_paketkurzname.sql',
  sets: 3,
  karten: 107,
  bilder: 'ersetzt',
  abgebrochen: false,
  verschluesselt: false,
  paketeEnthalten: [],
  paketeFehlen: [],
}

// belege setzt die Fassade: **jede Antwort der Deckliste, die vor dem
// Übergang gefragt wurde, wird zurückgehalten** — auch die des Überblicks auf
// der Startseite —, jede danach antwortet sofort mit dem neuen Stand.
// Deckantwort ist, was die Deckliste bekommt — Decks oder einen Fehler.
interface Deckantwort {
  decks: DeckDTO[]
  fehler: FehlerDTO | null
}

function belege(abbrechen = false) {
  const gehalten: ((antwort: Deckantwort) => void)[] = []
  let uebergegangen = false
  const Decks = vi.fn(() =>
    uebergegangen
      ? Promise.resolve<Deckantwort>({ decks: [neu], fehler: null })
      : new Promise<Deckantwort>((fertig) => {
          gehalten.push(fertig)
        }),
  )
  let arbeitsprofil = '10'
  belegeApi({
    ...belegeAnmeldung(),
    ...belegeFenster('de'),
    Decks,
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
    Nutzer: () => Promise.resolve({ nutzer: [ich] }),
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: ich.profile.find((p) => p.id === arbeitsprofil) ?? null,
      }),
    ProfilWechseln: () => {
      arbeitsprofil = '11'
      uebergegangen = true

      return Promise.resolve({ profil: ich.profile[1] ?? null })
    },
    Arbeitskontext: () =>
      Promise.resolve({
        nutzerID: '1',
        profilID: arbeitsprofil,
        eignerID: '1',
        rolle: 'verwalter',
        nurLesend: false,
        modus: 'fenster',
        revision: 1,
        fehler: null,
      }),
    SicherungWaehlen: () =>
      Promise.resolve(
        abbrechen ? { ...sicherung, abgebrochen: true } : sicherung,
      ),
    // **Der Neuanfang** (C05-Rest, #2284) — die schwerste Rücksetzung: die
    // ganze Datei geht beiseite, und danach antwortet der neue Bestand.
    Ruecksetzungsarten: () => Promise.resolve(['anwendung']),
    Ruecksetzungsvorschau: () =>
      Promise.resolve({
        art: 'anwendung',
        betroffen: [],
        hindernisse: [],
        bereit: true,
        nameNoetig: false,
        nachweisNoetig: true,
        sicherungEmpfohlen: true,
        revision: 'marke',
        fehler: null,
      }),
    RuecksetzungFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'freigabe',
        laeuftab: '2026-10-10T20:00:00Z',
        fehler: null,
      }),
    Zuruecksetzen: () => {
      uebergegangen = true

      return Promise.resolve({
        art: 'anwendung',
        zurueckgesetzt: 0,
        beiseite: '/Users/probe/vor-dem-neuanfang.db',
        fehler: null,
      })
    },
    SicherungZurueckspielen: () => {
      uebergegangen = true

      return Promise.resolve({
        ersetzt: true,
        beiseite: '/Users/probe/vorher.db',
        unveraendert: false,
      })
    },
  })

  return { gehalten }
}

// zeigtNurDenNeuenStand geht zur Deckseite und gibt **danach** die alten
// Antworten frei — die Wirkung wird abgewartet, dann geprüft.
async function zeigtNurDenNeuenStand(
  gehalten: ((antwort: Deckantwort) => void)[],
  alte: Deckantwort = { decks: [alt], fehler: null },
) {
  await userEvent.click(await leisteneintrag('decks'))

  // **Beleg, dass jemand zuhört**: der neue Stand steht da, die alten Rufe
  // hängen noch.
  expect(await screen.findByText(neu.name)).toBeInTheDocument()
  expect(gehalten.length).toBeGreaterThan(0)

  for (const freigeben of gehalten) freigeben(alte)
  await effekteNachholen()

  expect(screen.queryByText(alt.name)).toBeNull()
  expect(screen.getByText(neu.name)).toBeInTheDocument()
  // **Auch ein alter Fehler meldet sich nicht** (C05-Rest, #2284).
  expect(screen.queryByText(altfehler.meldung)).toBeNull()
}

describe('eine alte Antwort über einen echten Übergang', () => {
  it('übermalt nach einem Profilwechsel nicht die Decks des neuen Profils', async () => {
    const { gehalten } = belege()
    render(<App />)

    // Die Deckseite des alten Stands fragt — und bekommt noch nichts.
    await userEvent.click(await leisteneintrag('decks'))
    await vi.waitFor(() => {
      expect(gehalten.length).toBeGreaterThan(1)
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )
    await screen.findByRole('heading', { level: 1, name: /MagicSetEditor/ })

    await zeigtNurDenNeuenStand(gehalten)
  })

  it('übermalt nach einem Zurückspielen nicht die Decks des neuen Bestands', async () => {
    const { gehalten } = belege()
    const worte = kataloge.de.seiten.bestandspflege
    render(<App />)

    await userEvent.click(await leisteneintrag('decks'))
    await vi.waitFor(() => {
      expect(gehalten.length).toBeGreaterThan(1)
    })

    await userEvent.click(await leisteneintrag('einstellungen'))
    await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.waehlen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: worte.bestaetigen }),
    )
    expect(await screen.findByText(/vorher\.db/)).toBeInTheDocument()

    await zeigtNurDenNeuenStand(gehalten)
  })

  it('übermalt nach einem Neuanfang nicht die Decks des neuen Bestands', async () => {
    const { gehalten } = belege()
    const worte = kataloge.de.seiten.ruecksetzung
    render(<App />)

    await userEvent.click(await leisteneintrag('decks'))
    await vi.waitFor(() => {
      expect(gehalten.length).toBeGreaterThan(1)
    })

    await userEvent.click(await leisteneintrag('einstellungen'))
    await zuAbschnitt(
      kataloge.de.seiten.einstellungen.gruppen.daten,
      kataloge.de.seiten.einstellungen.abschnitte.eingriffe,
    )
    await userEvent.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: worte.freigeben }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: worte.tun }),
    )
    expect(await screen.findByText(/vor-dem-neuanfang\.db/)).toBeInTheDocument()

    await zeigtNurDenNeuenStand(gehalten)
  })

  // **Und ein zurückgehaltener Fehler** (C05-Rest, #2284): eine alte Antwort
  // ist nicht nur ein alter Stand, sie kann auch scheitern — und ihre Meldung
  // gehört ebenso wenig in den neuen.
  it('meldet nach einem Profilwechsel keinen alten Fehler', async () => {
    const { gehalten } = belege()
    render(<App />)

    await userEvent.click(await leisteneintrag('decks'))
    await vi.waitFor(() => {
      expect(gehalten.length).toBeGreaterThan(1)
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )
    await screen.findByRole('heading', { level: 1, name: /MagicSetEditor/ })

    await zeigtNurDenNeuenStand(gehalten, { decks: [], fehler: altfehler })
  })

  // **Ein abgebrochenes Zurückspielen ist kein Übergang** (C05-Rest, #2284):
  // die Antwort, die vorher gefragt wurde, ist dann die gültige — der Schutz
  // darf sie nicht verwerfen.
  it('lässt nach einem abgebrochenen Zurückspielen die gefragte Antwort gelten', async () => {
    const { gehalten } = belege(true)
    const worte = kataloge.de.seiten.bestandspflege
    render(<App />)

    await userEvent.click(await leisteneintrag('decks'))
    await vi.waitFor(() => {
      expect(gehalten.length).toBeGreaterThan(1)
    })

    await userEvent.click(await leisteneintrag('einstellungen'))
    await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.waehlen }),
    )
    await userEvent.click(await leisteneintrag('decks'))

    for (const freigeben of gehalten) freigeben({ decks: [alt], fehler: null })

    expect(await screen.findByText(alt.name)).toBeInTheDocument()
    expect(screen.queryByText(neu.name)).toBeNull()
  })
})
