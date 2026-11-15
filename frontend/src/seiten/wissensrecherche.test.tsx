// **Die Recherche im Sucharbeitsplatz** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass eine leere Antwort ihren Grund nennt.** „Nichts gefunden" heißt
// sechs Dinge, und nur eines davon heißt, dass es die Antwort nicht gibt.
//
// **Und dass die Herkunft am Treffer steht** — drei Angaben, weil das
// Abrufdatum kein Gültigkeitsdatum ist.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Wissensrecherche } from './Wissensrecherche'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.recherche

// einTreffer ist ein Fund samt Herkunft.
function einTreffer(anders: Record<string, unknown> = {}) {
  return {
    rang: 1,
    kennung: 'scryfall:o1:0',
    quelle: 'oracle',
    kennungssystem: 'scryfall',
    fachkennung: 'o1',
    semantik: 'konzept',
    eigeneKarteID: 0,
    titel: 'Funkenschlag',
    ausschnitt: '…3 Schadenspunkte…',
    text: 'Funkenschlag fügt 3 Schadenspunkte zu.',
    sprache: 'de',
    regelnummer: '',
    urheber: '',
    abschnitt: 0,
    von: 0,
    bis: 40,
    bezugsadresse: 'https://beispiel.invalid/o1',
    quellrevision: '2026-09-01',
    quelldatum: '2026-09-01',
    abgerufen: '2026-09-11',
    weitere: [],
    ...anders,
  }
}

// eineAntwort baut ein vollständiges Ergebnis.
function eineAntwort(anders: Record<string, unknown> = {}) {
  return {
    treffer: [einTreffer()],
    modus: 'hybrid',
    leere: '',
    grund: '',
    modell: 'all-minilm',
    modellrevision: 'aaaa',
    suchkonfiguration: 'rrf-60',
    indexfingerabdruck: 'abc',
    raumabdruck: 'def',
    ausgangsstand: '',
    korpusstaende: [
      { quelle: 'oracle', revision: '2026-09-01', datum: '2026-09-01' },
    ],
    einschraenkungen: [],
    fortsetzung: '',
    fehler: null,
    ...anders,
  }
}

// zeige rendert den Bereich im deutschen Sprachraum.
function zeige(oeffnen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Wissensrecherche oeffnen={oeffnen} />
    </Sprachraum>,
  )

  return oeffnen
}

// frage tippt eine Frage und schlägt nach.
async function frage(text = 'blitz') {
  const nutzer = userEvent.setup()

  await nutzer.type(screen.getByLabelText(texte.rechercheFrage), text)
  await nutzer.click(
    screen.getByRole('button', { name: texte.rechercheSuchen }),
  )

  return nutzer
}

describe('Die Recherche im Sucharbeitsplatz', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('zeigt Treffer samt Herkunft und dem Modus', async () => {
    belegeApi({ Recherchieren: () => Promise.resolve(eineAntwort()) })

    zeige()
    await frage()

    expect(await screen.findByText('Funkenschlag')).toBeInTheDocument()
    expect(
      screen.getByText(/oracle · Stand 2026-09-01, gültig 2026-09-01/),
    ).toBeInTheDocument()
    // **Über den Treffern und nicht in der Erklärung**: derselbe Satz steht
    // im Einleitungstext, und ein loser Ausdruck fände beide.
    expect(
      screen.getByText(
        `${texte.rechercheModus}: ${texte.rechercheModusHybrid}`,
      ),
    ).toBeInTheDocument()
  })

  // **Der Satz, für den es diesen Bereich gibt.**
  it('nennt den Grund einer leeren Antwort statt nur „nichts gefunden"', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({
            treffer: [],
            modus: 'ohne-index',
            leere: 'ohne-index',
            grund: 'Es ist kein Suchindex aktiv.',
          }),
        ),
    })

    zeige()
    await frage()

    // **Lage und Rat, in dieser Reihenfolge und ohne Doppelung.**
    expect(
      await screen.findByText(
        `Es ist kein Suchindex aktiv. ${texte.rechercheLeereOhneIndex}`,
      ),
    ).toBeInTheDocument()
  })

  // **Ein unbekannter Grund fällt auf den Satz des Dienstes zurück** — und
  // nicht auf „undefined".
  it('zeigt den Satz des Dienstes, wenn der Grund unbekannt ist', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({
            treffer: [],
            leere: 'erfunden',
            grund: 'Etwas Neues.',
          }),
        ),
    })

    zeige()
    await frage()

    expect(await screen.findByText('Etwas Neues.')).toBeInTheDocument()
  })

  it('zeigt, was nicht getan wurde', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({ einschraenkungen: ['ohne Vektoren gesucht'] }),
        ),
    })

    zeige()
    await frage()

    expect(await screen.findByText(/ohne Vektoren gesucht/)).toBeInTheDocument()
  })

  it('nennt die gebündelten Abschnitte derselben Karte', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({
            treffer: [einTreffer({ weitere: ['scryfall:o1:1'] })],
          }),
        ),
    })

    zeige()
    await frage()

    expect(await screen.findByText(/1 weitere Abschnitte/)).toBeInTheDocument()
  })

  // **Nur eine eigene Karte lässt sich öffnen.** Eine fremde Adresse wird
  // nicht angeklickt — eine Trefferaktion lädt keine fremden URLs.
  it('öffnet nur einen Treffer aus diesem Bestand', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({ treffer: [einTreffer({ eigeneKarteID: 42 })] }),
        ),
    })

    const oeffnen = zeige()
    const nutzer = await frage()

    await nutzer.click(
      await screen.findByRole('button', { name: texte.rechercheOeffnen }),
    )

    expect(oeffnen).toHaveBeenCalledWith('', '42')
  })

  it('zeigt keinen Weg zu einer fremden Quelle', async () => {
    belegeApi({ Recherchieren: () => Promise.resolve(eineAntwort()) })

    zeige()
    await frage()

    await screen.findByText('Funkenschlag')

    expect(
      screen.queryByRole('button', { name: texte.rechercheOeffnen }),
    ).not.toBeInTheDocument()
  })

  // **Ein Wechsel räumt das Ergebnis mit** (§7, Regel 1).
  it('räumt die Treffer, wenn der Fall wechselt', async () => {
    belegeApi({ Recherchieren: () => Promise.resolve(eineAntwort()) })

    zeige()
    const nutzer = await frage()

    await screen.findByText('Funkenschlag')

    await nutzer.selectOptions(
      screen.getByLabelText(texte.rechercheFall),
      'regeln',
    )

    expect(screen.queryByText('Funkenschlag')).not.toBeInTheDocument()
  })

  it('sucht auch über die Eingabetaste', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ Recherchieren: gerufen })

    zeige()

    const nutzer = userEvent.setup()

    await nutzer.type(
      screen.getByLabelText(texte.rechercheFrage),
      'blitz{Enter}',
    )

    await waitFor(() => {
      expect(gerufen).toHaveBeenCalled()
    })
  })

  // **Ein unbekannter Modus kommt roh durch** — und nicht als leere Zeile.
  // Er ist die Gestalt, in der eine neue Fassung des Dienstes hier ankäme.
  it('zeigt einen unbekannten Modus, wie er kam', async () => {
    belegeApi({
      Recherchieren: () => Promise.resolve(eineAntwort({ modus: 'neuartig' })),
    })

    zeige()
    await frage()

    expect(
      await screen.findByText(`${texte.rechercheModus}: neuartig`),
    ).toBeInTheDocument()
  })

  // **Ohne Ausschnitt steht der ganze Abschnitt da** — eine leere Zeile wäre
  // ein Treffer, den niemand lesen kann.
  it('zeigt den Abschnitt, wo es keinen Ausschnitt gibt', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({ treffer: [einTreffer({ ausschnitt: '' })] }),
        ),
    })

    zeige()
    await frage()

    expect(
      await screen.findByText('Funkenschlag fügt 3 Schadenspunkte zu.'),
    ).toBeInTheDocument()
  })

  it('zeigt einen Fehler des Dienstes', async () => {
    belegeApi({
      Recherchieren: () =>
        Promise.resolve(
          eineAntwort({
            fehler: { code: 'konflikt', meldung: 'Die Karte änderte sich.' },
          }),
        ),
    })

    zeige()
    await frage()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Karte änderte sich.',
    )
  })

  it('zeigt einen Sturz der Naht', async () => {
    belegeApi({ Recherchieren: () => Promise.reject(new Error('kaputt')) })

    zeige()
    await frage()

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  // **Das Kreuz steht nur beim Kartenfall** — Werkstück A13.4 von Phase 5.3
  // (#1452).
  //
  // # Warum das eine Aussage ist und keine Aufräumarbeit
  //
  // **Weil eine Regelrecherche gar nicht im eigenen Bestand sucht.** Ein
  // Kreuz dort bedeutete nichts — und `domain` weist eine solche Freigabe
  // ausdrücklich ab. Wer es stehen ließe, böte eine Entscheidung an, die zu
  // einem Fehler führt.
  it('bietet die Freigabe nur an, wo sie etwas bewirkt', async () => {
    belegeApi({ Recherchieren: () => Promise.resolve(eineAntwort()) })

    zeige()

    expect(
      screen.getByLabelText(texte.freigabeEigeneKarten),
    ).toBeInTheDocument()

    await userEvent
      .setup()
      .selectOptions(screen.getByLabelText(texte.rechercheFall), 'regeln')

    expect(
      screen.queryByLabelText(texte.freigabeEigeneKarten),
    ).not.toBeInTheDocument()
  })

  // **Und das gesetzte Kreuz reist mit.**
  it('schickt die Freigabe mit, wenn sie gesetzt ist', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ Recherchieren: gerufen })

    zeige()

    await userEvent
      .setup()
      .click(screen.getByLabelText(texte.freigabeEigeneKarten))

    await frage()

    await screen.findByText('Funkenschlag')

    expect(gerufen).toHaveBeenCalledWith(
      'karten',
      'blitz',
      10,
      { sprache: '', korpusstand: '', freigegeben: ['eigenekarte'] },
      '',
    )
  })
})
