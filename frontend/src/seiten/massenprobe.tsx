// Die Vorrichtung der Massenänderung.
//
// **Eine eigene Datei**, weil zwei Prüfdateien sie brauchen und weil
// `massenaenderung.test.tsx` sonst über die 400 Zeilen aus §2 liefe. Der
// Schnitt folgt der Frage: dort steht, was die Massenänderung **zusagt**, in
// `massenlaeufe.test.tsx`, was sie bei **nebenläufigen** Handgriffen tut (§7),
// und hier, **wie** sie für einen Test entsteht.
//
// Getrennt am 2026-09-02.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, vi } from 'vitest'

import type {
  FeldbeschreibungDTO,
  KartenzeileDTO,
  MassenvorschauAntwort,
} from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Massenaenderung } from './Massenaenderung'

export const texte = kataloge.de.seiten.karten.massen

export const karten: KartenzeileDTO[] = [
  { id: '1', nummer: 0, name: 'Feuerball', felder: 1 },
  { id: '2', nummer: 1, name: 'Eisblitz', felder: 2 },
]

// **Der Wert, um den es geht, kommt aus einer eigenen Bindung** (#891). Bis
// zum 2026-08-25 trug ihn die Auswahl mit; die Liste bringt keine Feldwerte
// mehr, also wird das eine Feld geholt, das gewählt ist.
export const feldwerte = {
  // **Ohne Fallunterscheidung**: `Kartenfelder` bietet genau ein Feld an, und
  // eine Verzweigung auf ein zweites wäre ein Zweig, den kein Test betreten
  // kann.
  Kartenfeldwerte: () =>
    Promise.resolve({ werte: [{ id: '2', wert: 'Zauber' }] }),

  // **Die Felder der Vorlage** (#1166, Werkstück 5). Seit das Feld gewählt
  // statt getippt wird, gehört die Liste zur Vorrichtung: ohne sie stünde eine
  // leere Auswahl da, und das ist genau der Zustand, den ein Mensch im Fenster
  // hätte, wenn die Vorlage nicht lesbar wäre.
  Kartenfelder: () =>
    Promise.resolve({
      felder: [feldbeschreibung('typ', 'Kartentyp')],
      fehler: null,
    }),
}

// feldbeschreibung baut ein Textfeld für die Vorrichtung.
export function feldbeschreibung(
  name: string,
  aufschrift: string,
): FeldbeschreibungDTO {
  return {
    name,
    aufschrift,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    editierbar: true,
    gespeichert: true,
    teilfelder: [],
    reihe: '',
    stelle: 0,
  }
}

export function zeige(auswahl = karten, nachladen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Massenaenderung setID="1" karten={auswahl} nachladen={nachladen} />
    </Sprachraum>,
  )

  return nachladen
}

// waehleFeld wählt ein Feld — **nachdem** die Vorlage geantwortet hat.
//
// # Warum das eine eigene Funktion ist
//
// `findByLabelText(texte.feld)` findet die Auswahlliste sofort: sie steht mit
// ihrem Platzhalter „— Feld wählen —" da, **bevor** die Felder der Vorlage
// eintreffen. Ein `selectOptions` darauf scheitert mit „Value \"typ\" not
// found in options" — oder, wenn die Antwort schnell genug war, eben nicht.
//
// **Bis zum 2026-09-02 war sie immer schnell genug**, weil `test/api.ts`
// synchron antwortete. Seither kommt jede Antwort eine Makroaufgabe später,
// so wie über die IPC — und damit war dieser Griff hier nur noch manchmal
// richtig (§3, „eine Vorrichtung, die großzügiger ist als der Betrieb").
export async function waehleFeld(feld: string) {
  const liste = await screen.findByLabelText(texte.feld)

  // **Gewartet wird auf die Möglichkeit, nicht auf die Liste.** Die Zahl ist
  // die Zusicherung: mehr als der Platzhalter.
  await waitFor(() => {
    expect(within(liste).getAllByRole('option').length).toBeGreaterThan(1)
  })

  await userEvent.selectOptions(liste, feld)
}

// vorschauAntwort ist die Vorschau, wie sie der Go-Teil für die zwei
// Messkarten gäbe: die erste ändert sich, die zweite steht schon so.
export function vorschauAntwort(wert = 'Artefakt'): MassenvorschauAntwort {
  return {
    feld: 'typ',
    wert,
    zeilen: [
      {
        karteID: '1',
        name: 'Feuerball',
        alt: '',
        gabAlt: false,
        neu: wert,
        ausschluss: '',
      },
      {
        karteID: '2',
        name: 'Eisblitz',
        alt: 'Zauber',
        gabAlt: true,
        neu: wert,
        ausschluss: '',
      },
    ],
  }
}

// mitVorschau belegt die Vorschau — die Grundlage jedes Anwendens.
export const mitVorschau = {
  ...feldwerte,
  FelderVorschau: () => Promise.resolve(vorschauAntwort()),
}

// vorschauen wählt Feld und Wert und holt die Vorschau.
export async function vorschauen(feld = 'typ', wert = 'Artefakt') {
  await waehleFeld(feld)
  await userEvent.type(screen.getByLabelText(texte.wert), wert)
  await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))
  await screen.findByRole('table')
}

// fuelleAus geht den ganzen Weg: Vorschau, dann Anwenden.
export async function fuelleAus(feld = 'typ', wert = 'Artefakt') {
  await vorschauen(feld, wert)
  await userEvent.click(screen.getByRole('button', { name: texte.anwenden }))
}
