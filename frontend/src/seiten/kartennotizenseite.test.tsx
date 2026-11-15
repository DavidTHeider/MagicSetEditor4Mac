// **Die Notizen an der Kartenlasche** — Werkstück 3 von Phase 3.5
// (#1166).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `kartennotizen.test.tsx` prüft, was der **Kasten** tut; hier steht, dass er
// an der Seite hängt und dass die Karte nach dem Schreiben neu geladen wird.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { EinzelkarteAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const texte = kataloge.de.seiten.karten.notizen

afterEach(() => {
  raeumeApi()
})

// **Und der Weg durch die ganze Seite** — Werkstück 3 von Phase 3.5.
//
// # Was hier steht und nicht oben
//
// Oben steht, was der Kasten **tut**; hier, dass er an der Kartenlasche hängt
// und dass die Karte nach dem Schreiben **neu geladen** wird. Ohne das zweite
// stünde im Kasten weiter der alte Text, sobald jemand die Karte wechselt und
// zurückkommt — und §7 Regel 3 sagt, dass nach einem geglückten Schreiben
// nachgeladen wird.
describe('die Notizen an der Kartenlasche', () => {
  it('schreibt und lädt die Karte nach', async () => {
    const gesetzt = vi.fn().mockResolvedValue({ fehler: null })
    const gelesen = vi.fn().mockResolvedValue({
      karte: {
        id: '5',
        nummer: 0,
        werte: [{ feld: 'name', wert: 'Feuerball' }],
      },
      fehler: null,
    })

    belegeApi({
      // **Der Verlauf antwortet leer** — der Anhang zeigt ihn seit U6 immer,

      // und unbelegt stünde hier eine zweite Meldung neben der geprüften.

      Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 1 }],
        }),
      Karte: gelesen,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),

      // **Der Verweiskasten wird mitgeladen** (#1167). Ohne diesen Beleg
      // meldete er selbst „nicht belegt" — und die Prüfung darunter suchte
      // **eine** Meldung und fände zwei.
      KarteVerweis: () =>
        Promise.resolve({ verweis: null, befunde: [], fehler: null }),

      FeldSetzen: gesetzt,
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const kasten = await screen.findByLabelText(texte.ueberschrift)

    // **Wie oft die Karte vor dem Schreiben gelesen wurde** — die Zahl ist
    // der Bezugspunkt, nicht null: die Seite lädt sie beim Aufschlagen.
    const vorher = gelesen.mock.calls.length

    await userEvent.type(kasten, 'meine Notiz')
    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith(
        '1',
        '5',
        'notes',
        'meine Notiz',
        expect.any(Array),
      )
    })

    // **Und danach wurde sie neu gelesen** (§7, Regel 3).
    await waitFor(() => {
      expect(gelesen.mock.calls.length).toBeGreaterThan(vorher)
    })
  })
})

// **Und beim Kartenwechsel stehen die Notizen der neuen Karte da.**
//
// # Der Fehler, den dieser Lauf gefangen hat
//
// Ein Klick in die Liste ändert die gewählte Kennung **sofort** — sie kommt
// aus den Zeilen, die schon dastehen. Die **Werte** der neuen Karte kommen
// danach. Der Kasten baut sich beim Wechsel neu auf und nimmt seinen
// Anfangstext aus den Werten, die es in diesem Augenblick gibt: denen der
// **vorigen** Karte, oder gar keinen.
//
// Danach ändert sich sein Schlüssel nicht mehr, und der Anfangstext bleibt.
// Im Fenster stand deshalb nach jedem Kartenwechsel ein leerer Kasten, obwohl
// jede Karte des kanonischen `Vanguard` ein `notes` trägt — am 2026-08-27
// nachgesehen.
//
// **Kein Test hat es gesehen**, weil keiner die Karte wechselte, während die
// Werte noch unterwegs waren. §3 nennt die Klasse: eine Zusicherung, die schon
// vor der Antwort zutrifft.
describe('die Notizen nach einem Kartenwechsel', () => {
  it('zeigen die Notizen der neuen Karte', async () => {
    const karten: Record<string, { feld: string; wert: string }[]> = {
      '5': [
        { feld: 'name', wert: 'Feuerball' },
        { feld: 'notes', wert: 'die erste Notiz' },
      ],
      '6': [
        { feld: 'name', wert: 'Eisblitz' },
        { feld: 'notes', wert: 'die zweite Notiz' },
      ],
    }

    belegeApi({
      // **Der Verlauf antwortet leer** — der Anhang zeigt ihn seit U6 immer,

      // und unbelegt stünde hier eine zweite Meldung neben der geprüften.

      Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [
            { id: '5', nummer: 0, name: 'Feuerball', felder: 2 },
            { id: '6', nummer: 1, name: 'Eisblitz', felder: 2 },
          ],
        }),
      Karte: (karteID: string) =>
        Promise.resolve({
          karte: { id: karteID, nummer: 0, werte: karten[karteID] ?? [] },
          fehler: null,
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await waitFor(async () => {
      expect(await screen.findByLabelText(texte.ueberschrift)).toHaveValue(
        'die erste Notiz',
      )
    })

    await userEvent.click(screen.getByRole('button', { name: /Eisblitz/ }))

    await waitFor(() => {
      expect(screen.getByLabelText(texte.ueberschrift)).toHaveValue(
        'die zweite Notiz',
      )
    })
  })
})

// **Und die Notizen stehen da, obwohl sie später kommen.**
//
// # Der Fehler, den dieser Lauf gefangen hat
//
// Die Kartenlasche lädt die **Zeilen** sofort und die **Karte samt Werten**
// danach. Der Kasten nimmt seinen Anfangstext beim Aufbau aus den Werten — und
// beim Aufbau gab es sie noch nicht. Im Fenster stand deshalb an jeder Karte
// des kanonischen `Vanguard` ein leerer Kasten, obwohl jede ein `notes` trägt.
//
// **Kein Test hat es gesehen**, weil jeder die Werte **gleich** mitgab. §3
// nennt genau das: eine Zusicherung, die schon vor der Antwort zutrifft. Hier
// kommt die Karte deshalb aus einer Zusage, die erst später erfüllt wird.
describe('die Notizen einer später geladenen Karte', () => {
  it('zeigen den Wert, der nach dem Aufbau ankommt', async () => {
    let liefere: (antwort: EinzelkarteAntwort) => void = () => undefined

    const spaeter = new Promise<EinzelkarteAntwort>((aufloesen) => {
      liefere = aufloesen
    })

    belegeApi({
      // **Der Verlauf antwortet leer** — der Anhang zeigt ihn seit U6 immer,

      // und unbelegt stünde hier eine zweite Meldung neben der geprüften.

      Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 1 }],
        }),
      Karte: () => spaeter,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    // **Erst jetzt kommt die Karte.** Der Kasten steht zu diesem Zeitpunkt
    // schon oder noch nicht — beides ist erlaubt; was gilt, ist sein Inhalt
    // danach.
    liefere({
      karte: {
        id: '5',
        nummer: 0,
        werte: [
          { feld: 'name', wert: 'Feuerball' },
          { feld: 'notes', wert: 'Sets: [pmoa];' },
        ],
      },
      fehler: null,
    })

    await waitFor(async () => {
      expect(await screen.findByLabelText(texte.ueberschrift)).toHaveValue(
        'Sets: [pmoa];',
      )
    })
  })
})

// **Ein Schreibfehler überlebt den Kartenwechsel** — Befund 3 des Nachlaufs
// über Phase 3.5.
//
// # Der Befund
//
// Der Notizkasten wird beim Kartenwechsel abgebaut (der `key` in
// `Kartenspalte`). Ein Schreibvorgang, den das Verlassen ausgelöst hat, lebt
// in seiner Closure weiter — und meldete seinen Fehler an eine Komponente, die
// es nicht mehr gibt. Der gewöhnliche Weg »Notiz tippen und gleich die nächste
// Karte anklicken« verlor damit bei einem Schreibfehler die Änderung **ohne
// sichtbare Auskunft**.
//
// **Die Spalte überlebt den Wechsel**, und deshalb hält sie die Meldung.
//
// # Dieser Lauf hat am 2026-09-16 einmal geflackert, und was gemessen wurde
//
// **Einmal in rund fünfzehn vollen Läufen** lief er in seine Frist
// (`15212 ms`) statt zuzusichern. Was daraufhin gemessen wurde, schließt die
// naheliegende Erklärung aus — er ist **nicht zu langsam**:
//
// | Bedingung | Dauer | Ausgang |
// | --- | --- | --- |
// | acht volle Läufe mit Abdeckung | — | achtmal grün |
// | voller Lauf, Maschine frei | 184 ms | grün |
// | voller Lauf, zwanzig Endlosschleifen auf zehn Kernen (`load` 65) | 372 ms | grün |
//
// **Last, die die Laufzeit verdoppelt, bringt ihn nicht auf fünfzehn
// Sekunden.** Und ein falscher Anker hätte nach einer Sekunde rot gemeldet —
// `findByRole` läuft dann ab. Die Frist bedeutet also einen **Hänger**.
//
// **Der Verdacht, unbelegt:** `userEvent` wickelt jede Handlung in `act`. Die
// `FeldSetzen`-Zusage hängt hier absichtlich über den Kartenwechsel hinweg —
// das ist der Gegenstand des Laufs —, und in einer seltenen Verschränkung
// wartet `act` auf sie, statt sie liegenzulassen.
//
// **Warum hier trotzdem nichts geändert wurde:** wer nicht nachstellen kann,
// kann auch nicht belegen, dass er etwas behoben hat (§10). Was belegt ist,
// ist der **negative** Befund — und der sagt dem Nächsten, wo er nicht
// suchen muss.
describe('ein Notizfehler beim Kartenwechsel', () => {
  it('steht auch dann da, wenn der Kasten schon abgebaut ist', async () => {
    let scheitern: (grund: Error) => void = () => undefined

    const gesetzt = vi.fn(
      () =>
        new Promise<never>((_, ablehnen) => {
          scheitern = ablehnen
        }),
    )

    belegeApi({
      // **Der Verlauf antwortet leer** — der Anhang zeigt ihn seit U6 immer,

      // und unbelegt stünde hier eine zweite Meldung neben der geprüften.

      Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [
            { id: '5', nummer: 0, name: 'Feuerball', felder: 1 },
            { id: '6', nummer: 1, name: 'Eisblitz', felder: 1 },
          ],
        }),
      Karte: (karteID: string) =>
        Promise.resolve({
          karte: { id: karteID, nummer: 0, werte: [] },
          fehler: null,
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),

      // **Der Verweiskasten wird mitgeladen** (#1167). Ohne diesen Beleg
      // meldete er selbst „nicht belegt" — und die Prüfung darunter suchte
      // **eine** Meldung und fände zwei.
      KarteVerweis: () =>
        Promise.resolve({ verweis: null, befunde: [], fehler: null }),

      FeldSetzen: gesetzt,
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const kasten = await screen.findByLabelText(texte.ueberschrift)

    await userEvent.type(kasten, 'geht schief')
    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalled()
    })

    // **Jetzt wechselt die Karte** — der Kasten wird abgebaut.
    await userEvent.click(screen.getByRole('button', { name: /Eisblitz/ }))

    // **Und erst danach scheitert das Schreiben.**
    scheitern(new Error('der Bestand ist belegt'))

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
