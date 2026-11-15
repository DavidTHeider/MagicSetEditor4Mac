// **Eine gewählte Teilmenge muss auch hinausgehen** —
// #1170,
// Befund 3 des ersten Zweitblicks über die Freigabe 0.3.0.
//
// # Der Befund — und was an ihm nicht stimmte
//
// `ausfuehren` liest `auswahl`, führte sie aber nicht in den Abhängigkeiten
// von `useCallback`. Der Zweitblick schloss daraus: die gemerkte Fassung
// behalte die **erste** — die leere —, und leer heißt nach Vertrag „das ganze
// Set"; wer drei von dreihundert Karten wähle, bekomme alle dreihundert.
//
// **Das trifft heute nicht zu.** `useVorgang` gibt bei jedem Zeichnen ein
// **neues** Objekt (`return { laeuft, starte }`), und `vorgang` steht in den
// Abhängigkeiten. Damit wird der Rückruf ohnehin bei jedem Zeichnen neu
// gebaut, und `auswahl` ist stets die aktuelle. Die Gegenprobe belegt es: die
// Abhängigkeit entfernt, Lauf **grün**.
//
// **Nachgestellt wurde es trotzdem** — mit `useMemo` um die Rückgabe von
// `useVorgang` **und** entfernter Abhängigkeit kam `[]` an. Der Befund
// beschreibt also keine heutige Wirkung, sondern eine Lage, die einen
// Handgriff entfernt ist: wer `useVorgang` merkt — und das ist der nächste
// naheliegende Schritt gegen unnötige Zeichenvorgänge —, hat den Fehler.
//
// **ESLint hat die Abhängigkeit beim Namen genannt**, als **Warnung**, und
// Warnungen ließen die Stufe grün. Seit demselben Tag ist die Regel ein
// Fehler.
//
// # Warum die Läufe daneben es nicht sahen
//
// `vorlagenlauf.test.tsx` prüft den Aufruf mit `[]` — dem Wert, der auch im
// kaputten Stand ankommt. Ein Lauf über den leeren Fall ist blind für genau
// diesen Fehler.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.bausteine.ausgabeauswahl

const satz = {
  id: '7',
  name: 'Probeset',
  spiel: 'probe',
  stil: 'probe-schlicht',
  karten: 2,
  herkunft: '',
  herkunftFehlt: false,
}

const vorlage = {
  paket: 'probe-ausfuhr.mse-export-template',
  name: 'Probeausfuhr',
  dateityp: '*.txt|*.txt',
  optionen: [],
}

it('reicht die gewählten Karten an die Bindung — nicht die leere Liste', async () => {
  const VorlageAusfuehren = vi.fn(() =>
    Promise.resolve({ pfad: '/pfad/aus.txt', befunde: [], abgebrochen: false }),
  )

  belegeApi({
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [
          { id: '11', nummer: 0, name: 'Erste', felder: 2 },
          { id: '12', nummer: 1, name: 'Zweite', felder: 2 },
        ],
        fehler: null,
      }),
    Exportvorlagen: () => Promise.resolve({ vorlagen: [vorlage], befunde: [] }),
    VorlageAusfuehren,
  })

  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={satz} schliessen={vi.fn()} />
    </Sprachraum>,
  )

  // **Erst auf „nur ausgewählte" umstellen** — das wählt die erste Karte vor.
  //
  // **Über den eingesetzten Satz und nicht über ein Muster**: `einsetzen`
  // füllt `{zahl}`, und ein Regex über den rohen Katalogsatz stolperte über
  // seine Klammern.
  //
  // **Erst, wenn er frei ist** — gesperrt bleibt er, bis die Kartenzeilen da
  // sind, und `findByRole` findet auch einen gesperrten. Unter der Last der
  // Pipeline ging der Klick ins Leere (zweimal am 2026-10-19), allein lief
  // der Lauf grün: gewartet wird auf die Wirkung, das Laden (§3).
  const nurAusgewaehlte = await screen.findByRole('radio', {
    name: einsetzen(eigene.teilmenge, { zahl: '0' }),
  })
  await vi.waitFor(() => {
    expect(nurAusgewaehlte).toBeEnabled()
  })
  await userEvent.click(nurAusgewaehlte)

  // **Und dann die zweite dazu.** Zwei von zwei wären wieder das ganze Set;
  // eine von zwei ist die Aussage.
  expect(await screen.findByRole('checkbox', { name: 'Erste' })).toBeChecked()

  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.seiten.sets.vorlageAusfuehren,
    }),
  )

  expect(VorlageAusfuehren).toHaveBeenCalledWith(
    '7',
    vorlage.paket,
    vorlage.dateityp,
    {},
    ['11'],
  )
})
