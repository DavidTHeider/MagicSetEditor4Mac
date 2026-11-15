// **Was an ein Modell ginge** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass der Ort dasteht.** „Gerechnet wird auf diesem Rechner" und
// „gerechnet wird bei einem fremden Anbieter" sind zwei verschiedene
// Entscheidungen, und die Zahlen daneben sagen dazu nichts.
//
// **Und dass der Ausschnitt dasteht und nicht der ganze Abschnitt** — was
// hier steht, ist genau das, was ein Modell bekäme.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Kontextfreigabe } from './Kontextfreigabe'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einBeleg, einPaket } from './kontextpaketprobe'

const texte = kataloge.de.seiten.recherche

// zeige rendert den Bereich im deutschen Sprachraum.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Kontextfreigabe freigegeben={[]} fall="regeln" frage="trampelschaden" />
    </Sprachraum>,
  )
}

// stelleZusammen tippt eine Aufgabe und drückt den Knopf.
async function stelleZusammen() {
  const nutzer = userEvent.setup()

  await nutzer.type(
    screen.getByLabelText(texte.freigabeAufgabe),
    'Erkläre Trampelschaden',
  )
  await nutzer.click(screen.getByRole('button', { name: texte.freigabeZeigen }))

  return nutzer
}

describe('Was an ein Modell ginge', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('zeigt Ort, Budget, Kategorien und den Ausschnitt', async () => {
    belegeApi({
      KontextZusammenstellen: () => Promise.resolve(einPaket()),
    })

    zeige()
    await stelleZusammen()

    expect(await screen.findByText(texte.freigabeOrtLokal)).toBeInTheDocument()
    expect(screen.getByText(/120 von 4096 Tokens/)).toBeInTheDocument()
    expect(screen.getByText(/aufgabe · oeffentlich/)).toBeInTheDocument()
    expect(
      screen.getByText('702.19a Trampelschaden ändert die Zuweisung.'),
    ).toBeInTheDocument()
  })

  // **Der Satz, für den es diesen Bereich gibt.**
  it('sagt, wenn bei einem fremden Anbieter gerechnet würde', async () => {
    belegeApi({
      KontextZusammenstellen: () =>
        Promise.resolve(
          einPaket({
            freigabe: {
              ...einPaket().freigabe,
              ort: 'cloud',
              zielanbieter: 'anthropic',
              privat: true,
            },
          }),
        ),
    })

    zeige()
    await stelleZusammen()

    expect(await screen.findByText(texte.freigabeOrtCloud)).toBeInTheDocument()
    expect(screen.getByText(/Eigener Inhalt ist dabei/)).toBeInTheDocument()
  })

  // **Die Adresse steht da, wo eine eingestellt ist** (A18, #1533) — hinter
  // einem Anbieternamen kann ein Gateway stehen.
  it('nennt die Zieladresse, wo eine eingestellt ist', async () => {
    belegeApi({
      KontextZusammenstellen: () =>
        Promise.resolve(
          einPaket({
            freigabe: {
              ...einPaket().freigabe,
              zieladresse: 'https://gateway.firma.intern/v1',
            },
          }),
        ),
    })

    zeige()
    await stelleZusammen()

    // **Der ganze Satz und kein Ausdruck**: `{name}: {wert}` sind zwei
    // Textknoten, und ein loser Ausdruck findet keinen von beiden.
    expect(
      await screen.findByText(
        `${texte.freigabeZieladresse}: https://gateway.firma.intern/v1`,
      ),
    ).toBeInTheDocument()
  })

  // **Und sie fehlt, wo der Anbieter selbst gemeint ist** — eine leere Zeile
  // „Zieladresse: " wäre eine Frage ohne Antwort.
  it('zeigt keine Zieladresse, wo die Vorgabe gilt', async () => {
    belegeApi({ KontextZusammenstellen: () => Promise.resolve(einPaket()) })

    zeige()
    await stelleZusammen()

    await screen.findByText(texte.freigabeOrtLokal)

    expect(
      screen.queryByText(new RegExp(texte.freigabeZieladresse)),
    ).not.toBeInTheDocument()
  })

  it('sagt, wenn ein Beleg nicht mehr dem aktiven Stand entspricht', async () => {
    belegeApi({
      KontextZusammenstellen: () =>
        Promise.resolve(einPaket({ belege: [einBeleg({ veraltet: true })] })),
    })

    zeige()
    await stelleZusammen()

    expect(
      await screen.findByText(new RegExp(texte.freigabeVeraltet)),
    ).toBeInTheDocument()
  })

  it('sagt, wenn gar nichts zusammenkam', async () => {
    belegeApi({
      KontextZusammenstellen: () => Promise.resolve(einPaket({ belege: [] })),
    })

    zeige()
    await stelleZusammen()

    expect(
      await screen.findByText(texte.freigabeKeineBelege),
    ).toBeInTheDocument()
  })

  // **Ein anderes Zielmodell ist ein anderes Budget** (§7, Regel 1).
  it('räumt das Paket, wenn der Anbieter wechselt', async () => {
    belegeApi({
      KontextZusammenstellen: () => Promise.resolve(einPaket()),
    })

    zeige()
    const nutzer = await stelleZusammen()

    await screen.findByText(texte.freigabeOrtLokal)

    await nutzer.selectOptions(
      screen.getByLabelText(texte.freigabeAnbieter),
      'anthropic',
    )

    expect(screen.queryByText(texte.freigabeOrtLokal)).not.toBeInTheDocument()
  })

  // **Der Anbieter wird gewählt** — UX-06.AK-01
  // (#1678):
  // es gibt vier, und ein Tippfehler darin ergab ein Paket für einen
  // Anbieter, den es nicht gibt.
  it('wählt den Anbieter aus den vieren, die es gibt', () => {
    belegeApi({})

    zeige()

    const anbieter = screen.getByLabelText(texte.freigabeAnbieter)

    expect(anbieter.tagName).toBe('SELECT')
    expect(
      [...(anbieter as HTMLSelectElement).options].map((o) => o.value),
    ).toEqual(['ollama', 'anthropic', 'openai', 'xai'])
  })

  it('räumt das Paket, wenn das Modell wechselt', async () => {
    belegeApi({
      KontextZusammenstellen: () => Promise.resolve(einPaket()),
    })

    zeige()
    const nutzer = await stelleZusammen()

    await screen.findByText(texte.freigabeOrtLokal)

    await nutzer.type(screen.getByLabelText(texte.freigabeModell), 'g')

    expect(screen.queryByText(texte.freigabeOrtLokal)).not.toBeInTheDocument()
  })

  // **Ein anderes Budget ist ein anderes Paket** (§7, Regel 1) — für beide
  // Zahlen, und nicht nur für die erste.
  it.each([
    ['Kontextfenster', texte.freigabeFenster],
    ['Ausgabereserve', texte.freigabeReserve],
  ])('räumt das Paket, wenn das %s wechselt', async (_name, etikett) => {
    belegeApi({
      KontextZusammenstellen: () => Promise.resolve(einPaket()),
    })

    zeige()
    const nutzer = await stelleZusammen()

    await screen.findByText(texte.freigabeOrtLokal)

    await nutzer.type(screen.getByLabelText(etikett), '0')

    expect(screen.queryByText(texte.freigabeOrtLokal)).not.toBeInTheDocument()
  })

  // **Und die Zahl kommt beim Dienst an** — ein Feld, das nur aussieht wie
  // eine Eingabe, wäre eine Vorbelegung mit zusätzlichen Schritten.
  it('reicht Fenster und Ausgabereserve durch', async () => {
    const gerufen = vi.fn(() => Promise.resolve(einPaket()))

    belegeApi({ KontextZusammenstellen: gerufen })

    zeige()

    const nutzer = userEvent.setup()

    await nutzer.clear(screen.getByLabelText(texte.freigabeFenster))
    await nutzer.type(screen.getByLabelText(texte.freigabeFenster), '2048')
    await nutzer.clear(screen.getByLabelText(texte.freigabeReserve))
    await nutzer.type(screen.getByLabelText(texte.freigabeReserve), '256')
    await stelleZusammen()

    expect(gerufen).toHaveBeenCalledWith(
      'Erkläre Trampelschaden',
      '',
      'regeln',
      'trampelschaden',
      20,
      'ollama',
      '',
      2048,
      256,
      0,

      // **Und die Freigabe reist mit** (A13.4, #1452) — hier leer, weil
      // dieser Lauf keine gesetzt hat. Ein Paket, das ohne sie gezeigt und
      // mit ihr gesendet würde, zeigte etwas anderes, als es schickt.
      [],
    )
  })

  it('zeigt einen Fehler des Dienstes', async () => {
    belegeApi({
      KontextZusammenstellen: () =>
        Promise.resolve(
          einPaket({
            fehler: { code: 'validierung', meldung: 'Ohne Aufgabe nichts.' },
          }),
        ),
    })

    zeige()
    await stelleZusammen()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Ohne Aufgabe nichts.',
    )
  })

  it('zeigt einen Sturz der Naht', async () => {
    belegeApi({
      KontextZusammenstellen: () => Promise.reject(new Error('kaputt')),
    })

    zeige()
    await stelleZusammen()

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
