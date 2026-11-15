// **Die Fläche der Tischausgabe** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Wogegen hier geprüft wird
//
// **Dass der Weg ankommt** — die Rechnung steht in `tischabzug.test.ts`.
// Hier: dass die Vorschau entsteht, bevor gesichert wird, dass ein Befund
// dasteht statt eines leeren Bildes, und dass der Abzug **so** hinausgeht,
// wie er entstanden ist.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Tischabzugfeld } from './Tischabzugfeld'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { einePartie } from '../seiten/partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { tischplan } from '../tisch/tischplan'
import { zweiSitze } from '../test/planprobe'

import type { Abzug, Abzugsfehler } from '../tisch/tischabzug'

const eigene = kataloge.de.seiten.spielen

const plan = tischplan(einePartie(zweiSitze))

afterEach(() => {
  raeumeApi()
})

// zeige zeichnet die Fläche über einem Abzugsweg, den der Lauf bestimmt.
function zeige(
  abzug: () => Promise<Abzug | { fehler: Abzugsfehler }> = () =>
    Promise.resolve({
      bild: 'data:image/png;base64,AAA',
      breite: 12,
      hoehe: 8,
    }),
) {
  render(
    <Sprachraum sprache="de">
      <Tischabzugfeld plan={plan} abzug={abzug} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Erst die Vorschau, dann das Bild — und die Maße stehen dabei.**
it('zeigt die Vorschau samt ihren Maßen', async () => {
  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  expect(await screen.findByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,AAA',
  )

  expect(
    screen.getByText(
      einsetzen(eigene.abzugMasse, { breite: '12', hoehe: '8' }),
    ),
  ).toBeInTheDocument()
})

// **Gesichert wird, was in der Vorschau steht** — nicht ein zweiter Abzug,
// der zwischendurch anders aussähe (AK-04, „konsistenter Snapshot").
it('sichert genau das Bild der Vorschau', async () => {
  const gesichert = vi.fn(() =>
    Promise.resolve({
      pfad: '/tmp/tisch.png',
      abgebrochen: false,
      fehler: null,
    }),
  )

  belegeApi({ TischabzugSpeichern: gesichert })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  await screen.findByAltText(eigene.abzugVorschau)

  await hand.click(screen.getByRole('button', { name: eigene.abzugSichern }))

  expect(gesichert).toHaveBeenCalledWith('data:image/png;base64,AAA')

  expect(
    await screen.findByText(
      einsetzen(eigene.abzugGesichert, { pfad: '/tmp/tisch.png' }),
    ),
  ).toBeInTheDocument()
})

// **Ein weggeklickter Dialog lässt die Vorschau stehen** — er ist eine
// Entscheidung und kein Fehler, und ein „Gesichert unter …" wäre gelogen.
it('meldet nichts, wo niemand einen Ort gewählt hat', async () => {
  belegeApi({
    TischabzugSpeichern: () =>
      Promise.resolve({ pfad: '', abgebrochen: true, fehler: null }),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  await screen.findByAltText(eigene.abzugVorschau)

  await hand.click(screen.getByRole('button', { name: eigene.abzugSichern }))
  await antwortAbwarten()

  expect(
    screen.queryByText(new RegExp(eigene.abzugGesichert.split(' {')[0] ?? '')),
  ).not.toBeInTheDocument()

  // **Und das Bild steht noch da** — ohne diese Zeile wäre der Lauf auch
  // grün, wenn das Wegklicken die Vorschau mitgenommen hätte (§10).
  expect(screen.getByAltText(eigene.abzugVorschau)).toBeInTheDocument()
})

// **Gesichert wird nur, was es gibt.** Ohne Vorschau steht der Knopf nicht da
// — ein Knopf ohne Wirkung ist eine Zusage, die niemand einlöst (§10).
it('bietet das Sichern erst nach der Vorschau an', () => {
  zeige()

  expect(
    screen.queryByRole('button', { name: eigene.abzugSichern }),
  ).not.toBeInTheDocument()
})

// **Ein Befund steht als Satz da und nicht als leeres Bild.**
it('nennt den Grund, wenn es kein Bild gibt', async () => {
  const hand = zeige(() => Promise.resolve({ fehler: 'zuGross' as const }))

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    eigene.abzugZuGross,
  )

  expect(screen.queryByAltText(eigene.abzugVorschau)).not.toBeInTheDocument()
})

// **Die Zonenwahl steht nur da, wo sie gemeint ist** — drei Bedienelemente
// für einen Ausschnitt, der keine Zone ist, wären Platz für nichts (#1771).
it('fragt die Zone erst, wenn eine gemeint ist', async () => {
  const hand = zeige()

  expect(
    screen.queryByRole('combobox', { name: eigene.abzugZonenwahl }),
  ).not.toBeInTheDocument()

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugAusschnitt }),
    eigene.abzugEineZone,
  )

  expect(
    screen.getByRole('combobox', { name: eigene.abzugZonenwahl }),
  ).toBeInTheDocument()
})

// **Eine gewählte Zone geht mit ihrem Sitz hinaus** — jeder Sitz hat seinen
// eigenen Friedhof, und „der Friedhof" allein meinte vier verschiedene.
it('reicht die gewählte Zone samt ihrem Sitz weiter', async () => {
  const abzug = vi.fn(() =>
    Promise.resolve({ bild: 'data:image/png;base64,AAA', breite: 1, hoehe: 1 }),
  )

  const hand = zeige(abzug)

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugAusschnitt }),
    eigene.abzugEineZone,
  )

  // **Das zweite Feld des Plans**, damit die Wahl wirklich eine ist: bei der
  // Vorgabe wäre der Lauf auch grün, wenn niemand zuhört (§3).
  const zweites = plan.felder[1]

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugZonenwahl }),
    '1',
  )

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  expect(abzug).toHaveBeenCalledWith(
    { art: 'zone', zone: zweites?.zone, sitz: zweites?.sitz },
    150,
  )
})

// **Der gemeinsame Stapel steht ohne Sitz da** — CR 405.1, und er gehört
// keinem. Ein Eintrag mit einer leeren Kennung daneben sähe aus wie ein
// Fehler.
it('nennt den gemeinsamen Stapel ohne Sitz', async () => {
  const hand = zeige()

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugAusschnitt }),
    eigene.abzugEineZone,
  )

  const eintraege = [
    ...screen
      .getByRole('combobox', { name: eigene.abzugZonenwahl })
      .querySelectorAll('option'),
  ].map((eines) => eines.textContent)

  expect(eintraege).toContain(eigene.zoneStapel)
})

// **Und der gewählte Ausschnitt geht wirklich hinaus.**
it('reicht Ausschnitt und Auflösung weiter', async () => {
  const abzug = vi.fn(() =>
    Promise.resolve({ bild: 'data:image/png;base64,AAA', breite: 1, hoehe: 1 }),
  )

  const hand = zeige(abzug)

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugAusschnitt }),
    eigene.abzugGanzerTisch,
  )

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugAufloesung }),
    '300',
  )

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  expect(abzug).toHaveBeenCalledWith({ art: 'tisch' }, 300)
})

// **Dasselbe Bild auf einem Blatt Papier** — PLAY-05.AK-03
// (#1713).
it('sichert die Vorschau als PDF, mit dem gewählten Blatt', async () => {
  const gesichert = vi.fn(() =>
    Promise.resolve({
      pfad: '/tmp/tisch.pdf',
      abgebrochen: false,
      fehler: null,
    }),
  )

  belegeApi({ TischabzugPDFSpeichern: gesichert })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  await screen.findByAltText(eigene.abzugVorschau)

  await hand.selectOptions(
    screen.getByRole('combobox', { name: eigene.abzugBlatt }),
    eigene.abzugBlattLetter,
  )

  await hand.click(screen.getByRole('button', { name: eigene.abzugPDF }))

  expect(gesichert).toHaveBeenCalledWith('data:image/png;base64,AAA', 'letter')

  expect(
    await screen.findByText(
      einsetzen(eigene.abzugGesichert, { pfad: '/tmp/tisch.pdf' }),
    ),
  ).toBeInTheDocument()
})

// **Das Seitenformat steht nur da, wo es etwas zu sichern gibt** — ohne
// Vorschau wäre es eine Wahl ohne Gegenstand (#1771).
it('fragt das Seitenformat erst mit einer Vorschau', () => {
  zeige()

  expect(
    screen.queryByRole('combobox', { name: eigene.abzugBlatt }),
  ).not.toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: eigene.abzugPDF }),
  ).not.toBeInTheDocument()
})
