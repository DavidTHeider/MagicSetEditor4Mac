// **Die Zwecke und die Sperre eines Gegenübers** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine zweite Datei**, weil `peerkontakte.test.tsx` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, wie ein
// Gegenüber **entsteht** und **geprüft** wird — hier, was es danach **darf**.
import { cleanup, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { dieZwecke, einKontakt, eineListe, zeige } from './peerkontaktprobe'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

// **Ein Zweck ist ein Schalter** (UXR-16, #1924; R7 in `docs/UIREVIEW.md`):
// was sofort wirkt und zwei Werte hat, ist ein Schalter. Bis dahin stand je
// Zweck ein Knopf „Freigeben" oder „Zurücknehmen", und ob ein Zweck galt, las
// man aus dem Wort, das ihn umkehrt — die Läufe darunter suchten den Knopf.

// **Eine Freigabe geht mit ihrem Zweck hinaus.**
it('gibt einen einzelnen Zweck frei', async () => {
  const gerufen = vi.fn(() =>
    Promise.resolve(
      eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
    ),
  )

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(eineListe([einKontakt({ stand: 'bestaetigt' })])),
    PeerzweckErlauben: gerufen,
  })

  const hand = await zeige()

  await screen.findByText(eigene.kontakteStandBestaetigt)

  // **Der Schalter für `handel`** — gesucht wird über seinen Namen und nicht
  // über eine Stelle in der Liste: eine Stelle wäre beim fünften Zweck eine
  // andere, und der Lauf prüfte dann etwas anderes, ohne rot zu werden.
  const schalter = screen.getByRole('switch', { name: eigene.zweckHandel })

  expect(schalter).not.toBeChecked()

  await hand.click(schalter)

  expect(gerufen).toHaveBeenCalledWith('bgegenueber', 'handel')
})

// **Ein gesperrtes Gegenüber lässt sich entsperren und nicht freigeben.**
it('zeigt bei einer Sperre nur den Weg zurück', async () => {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'gesperrt', rechte: [] })]),
      ),
  })

  await zeige()

  expect(
    await screen.findByRole('button', { name: eigene.kontakteEntsperren }),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: eigene.kontakteSperren }),
  ).not.toBeInTheDocument()

  expect(screen.queryByRole('switch')).not.toBeInTheDocument()
})

// **Wer niemandem begegnet ist, liest, was als Nächstes geschieht** (§6).
it('erklärt den leeren Zustand', async () => {
  belegeApi({ Peerkontakte: () => Promise.resolve(eineListe([])) })

  await zeige()

  expect(await screen.findByText(eigene.kontakteLeer)).toBeInTheDocument()
})

// **Ein Fehler beim Schreiben steht am Fenster und nicht im Protokoll.**
it('zeigt einen Fehler des Bestätigens', async () => {
  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe()),
    PeerkontaktBestaetigen: () =>
      Promise.resolve(
        eineListe([], {
          code: 'validierung',
          schluessel: '',
          meldung: 'Der Abdruck passt nicht.',
        }),
      ),
  })

  const hand = await zeige()

  await hand.type(
    await screen.findByLabelText(eigene.kontakteVergleichen),
    'falsch',
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.kontakteBestaetigen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Abdruck passt nicht.',
  )
})

// **Und ein Fehler beim Lesen ebenso.**
it('zeigt einen Fehler des Lesens', async () => {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([], {
          code: 'unbekannt',
          schluessel: '',
          meldung: 'Kein Bestand.',
        }),
      ),
  })

  await zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent('Kein Bestand.')
})

// **Eine Sperre geht mit ihrer Kennung hinaus.**
it('sperrt ein Gegenüber', async () => {
  const gerufen = vi.fn(() =>
    Promise.resolve(eineListe([einKontakt({ stand: 'gesperrt' })])),
  )

  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe()),
    PeerkontaktSperren: gerufen,
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kontakteSperren }),
  )

  expect(gerufen).toHaveBeenCalledWith('bgegenueber')
})

// **Und ein Entsperren ebenso.**
it('entsperrt ein Gegenüber', async () => {
  const gerufen = vi.fn(() => Promise.resolve(eineListe()))

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(eineListe([einKontakt({ stand: 'gesperrt' })])),
    PeerkontaktEntsperren: gerufen,
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kontakteEntsperren }),
  )

  expect(gerufen).toHaveBeenCalledWith('bgegenueber')
})

// **Ein Entzug geht mit seinem Zweck hinaus.**
it('nimmt einen einzelnen Zweck zurück', async () => {
  const gerufen = vi.fn(() =>
    Promise.resolve(eineListe([einKontakt({ stand: 'bestaetigt' })])),
  )

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['partie'] })]),
      ),
    PeerzweckEntziehen: gerufen,
  })

  const hand = await zeige()

  const schalter = await screen.findByRole('switch', {
    name: eigene.zweckPartie,
  })

  expect(schalter).toBeChecked()

  await hand.click(schalter)

  expect(gerufen).toHaveBeenCalledWith('bgegenueber', 'partie')
})

// **Ein eingetragenes Gegenüber geht mit Kennung und Namen hinaus.**
//
// # Warum dieser Weg überhaupt am Fenster steht
//
// **Weil es die Begegnung noch nicht gibt** — der Transport kommt mit N2 und
// N3. Bis dahin ist der Austausch der Kennung von Hand der einzige Weg, und
// was dabei entsteht, darf nichts.
it('trägt ein Gegenüber von Hand ein', async () => {
  const gerufen = vi.fn(() => Promise.resolve(eineListe()))

  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe([])),
    PeerkontaktAufnehmen: gerufen,
  })

  const hand = await zeige()

  // **Die Kennung von Hand steht hinter einem Kasten** (NET-01.AK-01): sie
  // ist eine technische Angabe, und vorn steht der Einladungstausch.
  await hand.click(
    await screen.findByRole('button', { name: eigene.einladung.vonHand }),
  )
  await hand.type(
    await screen.findByLabelText(eigene.kontakteAufnehmenKennung),
    'bgegenueber',
  )
  await hand.type(
    screen.getByLabelText(eigene.kontakteAufnehmenName),
    'Annas Rechner',
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.kontakteAufnehmen }),
  )

  expect(gerufen).toHaveBeenCalledWith('bgegenueber', 'Annas Rechner')
})

// **Ohne Kennung geht nichts hinaus.**
//
// **Der Knopf steht angehalten da und nicht gar nicht**: wer ihn sucht, soll
// ihn sehen — und dann sehen, dass ein Feld fehlt (§6).
it('lässt ohne Kennung nichts hinaus', async () => {
  const gerufen = vi.fn(() => Promise.resolve(eineListe()))

  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe([])),
    PeerkontaktAufnehmen: gerufen,
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.einladung.vonHand }),
  )

  expect(
    await screen.findByRole('button', { name: eigene.kontakteAufnehmen }),
  ).toBeDisabled()

  expect(gerufen).not.toHaveBeenCalled()
})

// **Ein Stand und ein Zweck, die dieses Fenster noch nicht kennt, stehen
// trotzdem da.**
//
// # Warum der Rückfall der Wert selbst ist
//
// **Weil er immer noch eine Auskunft ist.** Go entscheidet, welche Stände und
// welche Zwecke es gibt; ein Fenster, das eine ältere Fassung ist, kennt einen
// neuen nicht. Eine leere Zeile wäre dann keine Auskunft, sondern eine Lücke —
// und niemand wüsste, dass es sie gibt.
it('zeigt einen unbekannten Stand und Zweck im Klartext', async () => {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve({
        kontakte: [einKontakt({ stand: 'vielleicht' })],
        zwecke: [...dieZwecke, 'leihgabe'],
        fehler: null,
      }),
  })

  await zeige()

  expect(await screen.findByText('vielleicht')).toBeInTheDocument()

  // **Der Zweck steht nur bei einem bestätigten Gegenüber** — deshalb ein
  // zweiter Lauf mit demselben Katalog, sauber getrennt vom ersten.
  cleanup()
  raeumeApi()

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve({
        kontakte: [einKontakt({ stand: 'bestaetigt' })],
        zwecke: [...dieZwecke, 'leihgabe'],
        fehler: null,
      }),
  })

  await zeige()

  expect(await screen.findByText('leihgabe')).toBeInTheDocument()
})
