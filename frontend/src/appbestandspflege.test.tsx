// Das Zurückspielen einer Sicherung, von der Seite aus
// (#940).
//
// **Über die Einstellungsseite und nicht über den Haken unmittelbar**: geprüft
// werden soll der Weg, den jemand geht — wählen, sehen, was drin steht,
// bestätigen. Ein Haken für sich sagt darüber nichts.
import type { Bilderlage, Sicherungsauswahl } from './api'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { zuGebiet } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.bestandspflege

// ersetzt ist ein geglücktes Zurückspielen, wie Go es meldet.
const ersetzt = {
  ersetzt: true,
  beiseite: '/Users/probe/vorher.db',
  unveraendert: false,
}

// eineSicherung ist die geprüfte Auswahl, wie Go sie liefert.
const eineSicherung: Sicherungsauswahl = {
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

// zurEinstellungsseite öffnet die Anwendung und wechselt dorthin.
//
// **Über die Navigationsleiste** und nicht über einen gesetzten Zustand: der
// Weg dorthin gehört zu dem, was hier geprüft wird.
async function zurEinstellungsseite() {
  const nutzer = userEvent.setup()

  render(<App />)

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )

  // **Die Bestandspflege liegt in ihrem Gebiet** — Werkstück I von Phase
  // 4.10 (#1325).
  await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)

  await screen.findByRole('heading', { name: worte.titel })

  return nutzer
}

describe('Bestandspflege', () => {
  it('zeigt, was in der gewählten Sicherung steht, bevor sie fragt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))

    // **Die Zahlen stehen da, bevor jemand bestätigt.** „Wirklich
    // zurückspielen?" ist keine Frage, die sich beantworten lässt.
    const rueckfrage = await screen.findByText(worte.rueckfrage)

    expect(rueckfrage).toBeInTheDocument()
    expect(screen.getByText(eineSicherung.pfad)).toBeInTheDocument()
    expect(screen.getByText('107')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()

    // **Ohne Pakete steht „keine" da** und kein leeres Feld.
    expect(screen.getAllByText(worte.keine)).toHaveLength(2)
  })

  it('spielt erst zurück, wenn jemand bestätigt', async () => {
    const SicherungZurueckspielen = vi.fn(() => Promise.resolve(ersetzt))

    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen,
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByText(worte.rueckfrage)

    // **Bis hierher ist nichts geschehen** — und das ist die Zusage. Der Beleg
    // dafür, dass jemand zuhört, steht darunter: derselbe Weg löst gleich aus.
    expect(SicherungZurueckspielen).not.toHaveBeenCalled()

    await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

    await waitFor(() => {
      expect(SicherungZurueckspielen).toHaveBeenCalledWith(
        eineSicherung.pfad,
        '',
      )
    })

    // Danach ist die Rückfrage fort und der Ort des alten Bestands steht da.
    await waitFor(() => {
      expect(screen.queryByText(worte.rueckfrage)).not.toBeInTheDocument()
    })

    // **Über den Text und nicht über die Rolle**: das Meldungsband des
    // Fensters trägt dieselbe, und `getByRole('status')` fände beide.
    expect(screen.getByText(/vorher\.db/)).toBeInTheDocument()
  })

  it('lässt die Rückfrage wieder abbrechen', async () => {
    const SicherungZurueckspielen = vi.fn()

    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen,
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByText(worte.rueckfrage)

    await nutzer.click(screen.getByRole('button', { name: worte.abbrechen }))

    await waitFor(() => {
      expect(screen.queryByText(worte.rueckfrage)).not.toBeInTheDocument()
    })

    expect(SicherungZurueckspielen).not.toHaveBeenCalled()
  })

  it('hinterlässt nach einem Abbruch im Dateidialog nichts', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () =>
        Promise.resolve({ ...eineSicherung, pfad: '', abgebrochen: true }),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))

    await waitFor(() => {
      expect(screen.queryByText(worte.rueckfrage)).not.toBeInTheDocument()
    })
  })

  // **Die Meldung nennt die Migration, die diese Fassung nicht kennt** — und
  // das geht nur, weil der Fehler seine Werte mitbringt (#940).
  it('nennt bei einer neueren Fassung, woran es liegt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () =>
        Promise.resolve({
          ...eineSicherung,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'sicherungNeuereFassung',
            meldung: 'der deutsche Satz aus Go',
            werte: {
              unbekannt: '9999_aus_der_zukunft.sql',
              stand: '0012_paketkurzname.sql',
            },
          },
        }),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))

    const meldung = await screen.findByRole('alert')

    expect(meldung).toHaveTextContent('9999_aus_der_zukunft.sql')
    expect(meldung).toHaveTextContent('0012_paketkurzname.sql')

    // **Und nicht den deutschen Satz aus Go** — der ist der Rückfall für einen
    // Schlüssel, den der Katalog nicht kennt.
    expect(meldung).not.toHaveTextContent('der deutsche Satz aus Go')
  })

  it('meldet einen Fehler beim Zurückspielen und behält die Auswahl', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen: () =>
        Promise.resolve({
          ersetzt: false,
          beiseite: '',
          unveraendert: true,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'sicherungUnbrauchbar',
            meldung: 'roh',
          },
        }),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByText(worte.rueckfrage)
    await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

    await screen.findByRole('alert')

    // **Geschrieben ist geschrieben** (§7, Regel 3): die Auswahl bleibt stehen,
    // damit niemand die Datei neu suchen muss.
    expect(screen.getByText(worte.rueckfrage)).toBeInTheDocument()
  })

  // **Kennt die Sicherung eine Tabelle nicht, steht dort ein Strich.**
  //
  // −1 und 0 sind zwei Auskünfte; eine Oberfläche, die beide als „0" zeigte,
  // behauptete etwas über eine Datei, das sie nicht weiß.
  it('zeigt für eine unbekannte Zahl einen Strich', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () =>
        Promise.resolve({ ...eineSicherung, sets: -1, karten: -1 }),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByText(worte.rueckfrage)

    expect(screen.getAllByText(worte.unbekannt)).toHaveLength(2)
  })

  // **Nach einem Zurückspielen liest die Oberfläche neu** — der ganze Bestand
  // ist ein anderer (Zweitblick über Phase S).
  //
  // Ohne das stünden die Einstellungen aus dem ersetzten Stand da, und die
  // Set- und Kartenseiten zeigten Kennungen, die es nicht mehr gibt. Das ist
  // Regel 1 aus §7 in ihrer größten Form: das Bezugsobjekt hat gewechselt.
  it('lädt nach dem Zurückspielen alles neu', async () => {
    const Einstellungen = vi.fn(() =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'sprache',
            wert: 'de',
            wirksam: 'de',
            werte: ['de', 'en'],
            bereich: 'anwendung',
          },
        ],
      }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Einstellungen,
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen: () => Promise.resolve(ersetzt),
    })

    const nutzer = await zurEinstellungsseite()

    await waitFor(() => {
      expect(Einstellungen).toHaveBeenCalled()
    })

    const vorher = Einstellungen.mock.calls.length

    // **Auf den Knopf gewartet, nicht auf den Ruf** (#1809): er steht erst,
    // wenn die Antwort der Einstellungen gezeichnet ist.
    await nutzer.click(
      await screen.findByRole('button', { name: worte.waehlen }),
    )
    await screen.findByText(worte.rueckfrage)
    await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

    await waitFor(() => {
      expect(Einstellungen.mock.calls.length).toBeGreaterThan(vorher)
    })
  })

  // **Und die Seiten daneben laden ebenfalls neu** — weil sie beim Wechsel
  // ohnehin neu entstehen.
  //
  // Der Nachlauf zum Zweitblick hat gefragt, ob eine offene Set- oder
  // Paketseite ihre Liste aus dem ersetzten Bestand behält. **Sie kann nicht
  // offen sein**: das Zurückspielen steht allein in den Einstellungen, und
  // `Anwendungsinhalt` wählt seine Seite über ein `switch` — wer wechselt,
  // hängt die vorige ab und die nächste neu ein.
  //
  // Der Test steht hier, weil diese Begründung sonst eine Behauptung wäre. Er
  // hält sie fest: käme jemals ein zweiter Weg zum Zurückspielen dazu oder
  // bliebe eine Seite über den Wechsel stehen, wird er rot.
  it('liest die Setliste nach dem Zurückspielen neu', async () => {
    const Sets = vi.fn(() => Promise.resolve({ sets: [] }))

    belegeApi({
      ...belegeFenster('de'),
      Sets,
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen: () => Promise.resolve(ersetzt),
    })

    const nutzer = await zurEinstellungsseite()

    // Einmal auf die Setseite und zurück — damit sie ihre Liste **hatte**.
    await nutzer.click(await leisteneintrag('sets'))

    await waitFor(() => {
      expect(Sets).toHaveBeenCalled()
    })

    // **Über den Eintrag der Seitenleiste** — derselbe Weg wie oben zu den
    // Sets, und einer, der wartet, bis er dasteht (#1809).
    await nutzer.click(await leisteneintrag('einstellungen'))

    // **Wieder zum Gebiet**: die Seite kommt beim ersten Gebiet an, und die
    // Bestandspflege ist das vierte.
    await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)

    await screen.findByRole('heading', { name: worte.titel })

    const vorher = Sets.mock.calls.length

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByText(worte.rueckfrage)
    await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

    await waitFor(() => {
      expect(screen.queryByText(worte.rueckfrage)).not.toBeInTheDocument()
    })

    await nutzer.click(await leisteneintrag('sets'))

    await waitFor(() => {
      expect(Sets.mock.calls.length).toBeGreaterThan(vorher)
    })
  })
})

// **Die Rückfrage sagt, was mit den eingesetzten Bildern geschieht**
// (#1162).
//
// # Der Befund
//
// Sie sagte für jede Sicherung dasselbe. Bei einer nackten `.db` aus 0.1.0
// oder 0.2.0 sagte sie damit das **Gegenteil** dessen, was geschieht: die
// vorhandenen Bilder bleiben stehen, sie werden nicht ersetzt. Genau dafür
// verlangt #1162 einen Hinweis, und die Oberfläche sagte das Gegenteil.
//
// Angestrichen vom Nachlauf des Zweitblicks, Befund 2.
describe('Die Rückfrage und die Bilder', () => {
  const faelle: [Bilderlage, string][] = [
    ['ersetzt', worte.bilderErsetzt],
    ['geleert', worte.bilderGeleert],
    ['unberuehrt', worte.bilderUnberuehrt],
  ]

  it.each(faelle)('sagt bei %s, was geschieht', async (lage, satz) => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () =>
        Promise.resolve({ ...eineSicherung, bilder: lage }),
    })

    const nutzer = await zurEinstellungsseite()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))

    expect(await screen.findByText(satz)).toBeInTheDocument()

    // **Und nicht einen der anderen beiden.** Ohne diese Zusicherung wäre der
    // Lauf auch dann grün, wenn alle drei Sätze zugleich dastünden.
    for (const [andere, andererSatz] of faelle) {
      if (andere !== lage) {
        expect(screen.queryByText(andererSatz)).not.toBeInTheDocument()
      }
    }
  })
})
