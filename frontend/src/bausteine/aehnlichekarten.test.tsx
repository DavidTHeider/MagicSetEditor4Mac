// **Ähnliche Karten im Editor** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass kein Stand hineingeht und einer herauskommt.** Den gerechneten Stand
// einer Karte rechnet der Bestand; das Fenster kennt ihn erst mit der Antwort
// — und nur deshalb kann es die zweite Seite überhaupt anfordern.
//
// **Und dass kein Weg „zur Karte" angeboten wird**: die Karte, zu der gesucht
// wurde, ist schon offen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { AehnlicheKarten } from './AehnlicheKarten'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.recherche

// eineAntwort baut ein Ergebnis mit einem eigenen Treffer.
function eineAntwort(anders: Record<string, unknown> = {}) {
  return {
    treffer: [
      {
        rang: 1,
        kennung: 'eigenekarte:9:0',
        quelle: 'eigenekarte',
        kennungssystem: '',
        fachkennung: '',
        semantik: '',
        eigeneKarteID: 9,
        titel: 'Zweitkarte',
        ausschnitt: '…3 Schaden…',
        text: 'Zweitkarte fügt 3 Schaden zu.',
        sprache: 'de',
        regelnummer: '',
        urheber: '',
        abschnitt: 0,
        von: 0,
        bis: 20,
        bezugsadresse: '',
        quellrevision: '1',
        quelldatum: '',
        abgerufen: '2026-09-11',
        weitere: [],
      },
    ],
    modus: 'hybrid',
    leere: '',
    grund: '',
    modell: 'all-minilm',
    modellrevision: 'aaaa',
    suchkonfiguration: 'rrf-60',
    indexfingerabdruck: 'abc',
    raumabdruck: 'def',
    ausgangsstand: 'stand-von-jetzt',
    korpusstaende: [],
    einschraenkungen: [],
    fortsetzung: '',
    fehler: null,
    ...anders,
  }
}

// zeige rendert den Bereich im deutschen Sprachraum.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <AehnlicheKarten karteID="7" />
    </Sprachraum>,
  )
}

describe('Ähnliche Karten im Editor', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sagt, dass noch nichts gesucht wurde', () => {
    belegeApi({})

    zeige()

    expect(screen.getByText(texte.aehnlichLeer)).toBeInTheDocument()
  })

  // **Der Satz, für den es diesen Bereich gibt.**
  it('fragt ohne Stand — den rechnet der Bestand', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ AehnlicheKarten: gerufen })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    expect(await screen.findByText('Zweitkarte')).toBeInTheDocument()
    expect(gerufen).toHaveBeenCalledWith(
      7,
      '',
      10,

      // **Ohne Freigabe** (A13.4, #1452): der eigene Bestand bleibt außen,
      // solange niemand das Kreuz setzt.
      { sprache: '', korpusstand: '', freigegeben: [] },
      '',
    )
  })

  // **Kein Weg „zur Karte"**: sie ist schon offen.
  it('bietet keinen Weg zur Karte an', async () => {
    belegeApi({ AehnlicheKarten: () => Promise.resolve(eineAntwort()) })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    await screen.findByText('Zweitkarte')

    expect(
      screen.queryByRole('button', { name: texte.rechercheOeffnen }),
    ).not.toBeInTheDocument()
  })

  // **Ohne Einbettungsmodell gibt es keine Ähnlichkeit** — und das steht da,
  // statt als „nichts gefunden" durchzugehen.
  it('sagt, wenn das Einbettungsmodell fehlt', async () => {
    belegeApi({
      AehnlicheKarten: () =>
        Promise.resolve(
          eineAntwort({
            treffer: [],
            modus: 'lexikalisch',
            leere: 'ohne-modell',
            grund: 'Das Modell ist nicht installiert.',
          }),
        ),
    })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    expect(
      await screen.findByText(
        `Das Modell ist nicht installiert. ${texte.rechercheLeereOhneModell}`,
      ),
    ).toBeInTheDocument()
  })

  it('zeigt einen Fehler des Dienstes', async () => {
    belegeApi({
      AehnlicheKarten: () =>
        Promise.resolve(
          eineAntwort({
            fehler: { code: 'konflikt', meldung: 'Die Karte änderte sich.' },
          }),
        ),
    })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Karte änderte sich.',
    )
  })

  it('zeigt einen Sturz der Naht', async () => {
    belegeApi({
      AehnlicheKarten: () => Promise.reject(new Error('kaputt')),
    })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    // **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742).
    expect(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.kieinrichtung.oeffnen,
      }),
    ).toBeInTheDocument()
  })

  // **Der eigene Bestand geht nur mit Kreuz mit** — Werkstück A13.4 von
  // Phase 5.3 (#1452).
  //
  // # Was hier gemessen wird
  //
  // **Dass das Kreuz die Anfrage ändert.** Ein Kasten, der die Wahl zeigte
  // und sie nicht mitschickte, sähe für jemanden, der ihn setzt, genauso aus
  // wie einer, der sie einlöst.
  it('sucht die eigenen Karten erst nach dem Kreuz mit', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ AehnlicheKarten: gerufen })

    zeige()

    const hand = userEvent.setup()

    await hand.click(screen.getByLabelText(texte.freigabeEigeneKarten))
    await hand.click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    await screen.findByText('Zweitkarte')

    expect(gerufen).toHaveBeenCalledWith(
      7,
      '',
      10,
      { sprache: '', korpusstand: '', freigegeben: ['eigenekarte'] },
      '',
    )
  })

  // **Und ein zweites Kreuz nimmt sie wieder heraus.**
  //
  // **Die Gegenprobe** — ohne sie sagte der Lauf darüber nur, dass ein Kreuz
  // etwas hinzufügt, und nicht, dass es eine Wahl ist.
  it('nimmt die eigenen Karten beim Abwählen wieder heraus', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ AehnlicheKarten: gerufen })

    zeige()

    const hand = userEvent.setup()
    const kreuz = screen.getByLabelText(texte.freigabeEigeneKarten)

    await hand.click(kreuz)
    await hand.click(kreuz)
    await hand.click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    await screen.findByText('Zweitkarte')

    expect(gerufen).toHaveBeenCalledWith(
      7,
      '',
      10,
      { sprache: '', korpusstand: '', freigegeben: [] },
      '',
    )
  })

  // **Ein Wechsel der Freigabe räumt das Ergebnis mit** (§7, Regel 1).
  //
  // **Weil die Liste sonst zu einer anderen Frage gehörte**: eine Antwort
  // ohne den eigenen Bestand unter einem gesetzten Kreuz behauptet, er sei
  // dabei gewesen.
  it('räumt das Ergebnis, wenn die Freigabe wechselt', async () => {
    belegeApi({ AehnlicheKarten: () => Promise.resolve(eineAntwort()) })

    zeige()

    const hand = userEvent.setup()

    await hand.click(screen.getByRole('button', { name: texte.aehnlichSuchen }))

    expect(await screen.findByText('Zweitkarte')).toBeInTheDocument()

    await hand.click(screen.getByLabelText(texte.freigabeEigeneKarten))

    expect(screen.queryByText('Zweitkarte')).not.toBeInTheDocument()
    expect(screen.getByText(texte.aehnlichLeer)).toBeInTheDocument()
  })

  // **Die Gegenprobe am Schaden** — #1553.
  //
  // `keinefestekennung.test.ts` liest den Quelltext und sagt, dass keine feste
  // Kennung mehr dasteht. Was er **nicht** sagt, ist, dass das etwas ändert.
  // Dieser Fall stellt denselben Kasten zweimal auf: mit
  // `id="aehnlich-ueberschrift"` stünde die Kennung zweimal im Dokument, und
  // `aria-labelledby` zeigte für beide Kästen auf denselben Knoten.
  it('gibt zweimal aufgestellt keine doppelte Kennung', () => {
    render(
      <Sprachraum sprache="de">
        <AehnlicheKarten karteID="7" />
        <AehnlicheKarten karteID="8" />
      </Sprachraum>,
    )

    const kennungen = [...document.querySelectorAll('[id]')].map(
      (knoten) => knoten.id,
    )

    expect(kennungen.length).toBeGreaterThan(1)
    expect(kennungen).toEqual([...new Set(kennungen)])
  })
})
