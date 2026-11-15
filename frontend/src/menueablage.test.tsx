// **Die Ablage und das Hilfemenü, durch die ganze Anwendung** —
// #891,
// Werkstück 2 von Phase Y.
//
// # Was vorher war
//
// ⌘O, ⌘W und ⌘P taten nichts, und ein Hilfemenü gab es nicht. Alle drei sind
// unter macOS gewohnte Bedeutungen (§6).
//
// # Warum ein Lauf über `App` und nicht über den Rahmen
//
// Der Weg geht über **vier** Stellen: das Ereignis aus Go, den Horcher in
// `fenster.ts`, den Rückruf in `Anwendung` und die Seite, die danach dasteht.
// Ein Lauf gegen den Rahmen sagte über die Naht dazwischen nichts.
import { act, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { effekteNachholen } from './test/effekte'
import { kataloge } from './sprache/kataloge'

interface MitLaufzeit {
  runtime?: {
    EventsOn: (
      name: string,
      rueckruf: (...daten: unknown[]) => void,
    ) => () => void
    OnFileDrop?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
      ueberZiel?: boolean,
    ) => void
    OnFileDropOff?: () => void
  }
}

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
  raeumeApi()
})

const texte = kataloge.de.seiten

// spielpaket ist das eine installierte Spiel dieser Vorrichtung.
//
// **Ohne es steht das Anlegeformular nicht** — `SetNeu` zeigt dann den leeren
// Zustand, und wer auf sein Namensfeld zusichert, prüft eine Seite, die es im
// Betrieb so nicht gibt.
const spielpaket = {
  art: 'mse-game',
  name: 'probe',
  anzeigename: 'Probe',
  version: '',
  pfad: '',
  herkunft: '',
  pruefsumme: '',
  braucht: [],
  gruppe: '',
}

// stilpaket ist sein Stil. **Ein Spiel ohne Stil ist keines, aus dem sich ein
// Set machen lässt** (UX-05.AK-04), und die Spielauswahl entfällt ohnehin,
// wo es nur eines gibt — der Eintrag, auf den dieser Lauf wartet, ist deshalb
// der des Stils.
const stilpaket = {
  ...spielpaket,
  art: 'mse-style',
  name: 'probe-schlicht',
  anzeigename: 'Probe-schlicht',
}

// Kartenzeilen ist der Beleg dafür, **welches** Set aufgeschlagen ist: die
// Kartenseite lädt seine Zeilen, sobald sie dasteht.
let Kartenzeilen = vi.fn()

// bereit stellt eine Laufzeit, zeichnet und gibt die angemeldeten Rückrufe.
async function bereit(
  zusatz: {
    SetAusDatei?: unknown
    ablegen?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
    ) => void
  } = {},
) {
  const horcher = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      horcher.set(name, rueckruf)

      return () => undefined
    },

    // **Auch das Ablegen gehört zur Laufzeit** (#891).
    OnFileDrop: zusatz.ablegen ?? (() => undefined),
    OnFileDropOff: () => undefined,
  }

  // **Wessen Karten geladen werden, sagt, welches Set offen ist.**
  Kartenzeilen = vi.fn().mockResolvedValue({ zeilen: [], fehler: null })

  // **Öffnen geht über die Vorschau** (FMT-03, #1692) — ein abgebrochener
  // Dialog gibt keine.
  const SetVorschauWaehlen = vi.fn().mockResolvedValue({
    abgebrochen: true,
    vorschau: null,
    fehler: null,
  })

  belegeApi({
    ...belegeFenster('de'),
    Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    // **Ein Spielpaket, und das ist keine Kosmetik**
    // (#1252).
    //
    // Hier stand eine leere Liste. `SetNeu` zeigt dann — **sobald die Antwort
    // da ist** — „kein Spiel installiert" statt des Formulars, und das
    // Namensfeld verschwindet. Der Fokustest weiter unten war deshalb grün,
    // solange er **vor** der Antwort zusicherte, und rot, sobald sie ihn
    // überholte: genau die Gestalt aus §3 („eine Zusicherung, die schon vor
    // der Antwort zutrifft"), nur andersherum.
    Pakete: vi
      .fn()
      .mockResolvedValue({ pakete: [spielpaket, stilpaket], fehler: null }),
    Karten: vi.fn().mockResolvedValue({ karten: [], fehler: null }),
    Kartenfelder: vi.fn().mockResolvedValue({ felder: [], fehler: null }),
    Setangaben: vi
      .fn()
      .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    SetVorschauWaehlen,
    Kartenzeilen,
    ...(zusatz.SetAusDatei === undefined
      ? {}
      : { SetAusDatei: zusatz.SetAusDatei as never }),
  })

  render(<App />)

  await screen.findByRole('heading', { name: texte.start.ueberschrift })
  await effekteNachholen()

  return { horcher, SetVorschauWaehlen }
}

// ausloesen ruft einen angemeldeten Horcher — und wird laut, wenn keiner hängt.
// ausloesen feuert ein Menüereignis — **in `act`**.
//
// # Warum das dazugehört
//
// Ein Menübefehl ist ein Ereignis von außen, und er ändert Zustand. React
// warnt seit je laut („An update to Anwendung inside a test was not wrapped in
// act"), und solange kein Zustand am Ergebnis hing, war die Warnung folgenlos.
//
// **Seit die Einführung ein Fenster ist** (#1325), hängt einer daran: der
// Befehl schließt sie. Ohne `act` wurde die Änderung nie geflusst — der
// Dialog blieb stehen, sein `aria-modal` nahm den Rest aus dem
// Zugänglichkeitsbaum, und `findByRole` fand die Überschrift nicht, die im
// DOM stand. **Gemessen**: mit einem `setTimeout` von 50 ms davor stand der
// Dialog bei null und „Sets" da.
//
// Das ist dieselbe Klasse wie §3, „der Beleg geht ins Leere, wenn er zu früh
// kommt" — nur andersherum: hier fehlte das Nachfassen **nach** dem Ereignis.
function ausloesen(
  horcher: Map<string, (...daten: unknown[]) => void>,
  name: string,
) {
  const rueckruf = horcher.get(name)
  if (rueckruf === undefined) {
    throw new Error(`kein Horcher für ${name}`)
  }

  act(() => {
    rueckruf()
  })
}

describe('Die Einträge der Ablage', () => {
  it('führen auf ihre Seite', async () => {
    const { horcher } = await bereit()

    // **Ohne offenes Set führt der Druckbogen ins Leere** — er ist eine
    // Setseite. Erst ein Set, dann die Seite; das ist dieselbe Ordnung wie
    // bei ⌘S.
    horcher.get('datei:geoeffnet')?.('7')

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.karten.anlegen }),
      ).toBeInTheDocument()
    })

    // **Drucken führt zum Druckbogen.**
    ausloesen(horcher, 'menue:drucken')

    expect(
      await screen.findByRole('heading', { name: texte.druckbogen.titel }),
    ).toBeInTheDocument()

    // **Die Hilfe zur Hilfeseite.**
    ausloesen(horcher, 'menue:hilfe')

    expect(
      await screen.findByRole('heading', { name: texte.hilfe.titel }),
    ).toBeInTheDocument()

    // **Und Schließen zur Setliste** — das Set, nicht das Fenster.
    ausloesen(horcher, 'menue:set-schliessen')

    expect(
      await screen.findByRole('heading', { name: texte.sets.titel }),
    ).toBeInTheDocument()
  })

  // **Öffnen führt zur Setliste und stößt dort den Dialog an.**
  //
  // Nicht eine zweite Fassung derselben Frage: ⌘O und der Knopf „Set öffnen …"
  // landen in derselben Funktion.
  it('stoßen den Dateidialog der Setliste an', async () => {
    const { horcher, SetVorschauWaehlen } = await bereit()

    ausloesen(horcher, 'menue:set-oeffnen')

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    // `toHaveBeenCalled` trifft zu, sobald gerufen wurde — die Antwort ist
    // dann noch unterwegs, und was sie bewirkt, steht noch nicht da.
    expect(
      await screen.findByRole('heading', { name: texte.sets.titel }),
    ).toBeInTheDocument()

    expect(SetVorschauWaehlen).toHaveBeenCalled()
  })

  // **⌘N führt zur Setliste und setzt den Fokus ins Namensfeld** (#891).
  //
  // Der Weg ist derselbe wie bei ⌘O; geprüft wird, dass er ein **anderes**
  // Ziel hat — sonst täten beide dasselbe, und niemand sähe es.
  it('setzen den Fokus für ein neues Set', async () => {
    const { horcher } = await bereit()

    ausloesen(horcher, 'menue:set-neu')

    expect(
      await screen.findByRole('heading', { name: texte.sets.titel }),
    ).toBeInTheDocument()

    // **Erst das Abschlusssignal, dann die Zusicherung** (§3,
    // #1252).
    //
    // Das Namensfeld steht erst da, wenn der Paketkatalog geantwortet hat —
    // seit dem Assistenten (UX-05, #1677) springt ⌘N zum Schritt „Name".
    // Ohne dieses Warten sicherte der Test gegen den **Ladezustand** zu, und
    // ob er grün wurde, entschied die Maschine.
    expect(await screen.findByLabelText(texte.sets.neuName)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByLabelText(texte.sets.neuName)).toHaveFocus()
    })
  })

  // **„Zuletzt benutzt" schlägt sein Set auf** (#891).
  //
  // Die Nutzlast wird geprüft, nicht geglaubt: sie kommt über die
  // Ereignisnaht als `unknown`.
  it('schlagen ein zuletzt benutztes Set auf', async () => {
    const { horcher } = await bereit()

    const zuletzt = horcher.get('menue:set-zuletzt')
    if (zuletzt === undefined) {
      throw new Error('kein Horcher für menue:set-zuletzt')
    }

    // **Eine unpassende Nutzlast tut nichts** — die Startseite bleibt.
    zuletzt(42)
    zuletzt('')

    await effekteNachholen()

    expect(
      screen.getByRole('heading', { name: texte.start.ueberschrift }),
    ).toBeInTheDocument()

    zuletzt('7')

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.karten.anlegen }),
      ).toBeInTheDocument()
    })
  })

  // **Eine gezogene `.mse-set` wird geöffnet** (#891).
  //
  // Und sie landet **nicht** in einem Bildfeld: bis zum 2026-08-25 ging jeder
  // Pfad ins Feld darunter.
  it('öffnen ein gezogenes Setpaket', async () => {
    const SetAusDatei = vi.fn().mockResolvedValue({
      abgebrochen: false,
      setID: '7',
      karten: 3,
      befunde: [],
      angelegt: ['7'],
      unvollstaendig: false,
      schonOffen: false,
      fehler: null,
    })

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/Users/david/Vanguard.mse-set'])

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalledWith('/Users/david/Vanguard.mse-set')
    })

    // **Und es wird aufgeschlagen** — wer ein Set zieht, will hinein.
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.karten.anlegen }),
      ).toBeInTheDocument()
    })
  })

  // **Ein Setpaket, das sich nicht lesen lässt, sagt warum.**
  it('sagen es, wenn ein gezogenes Setpaket nicht aufgeht', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei: vi.fn().mockRejectedValue(new Error('Die Datei ist leer.')),
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/Users/david/Kaputt.mse-set'])

    expect(await screen.findByText(/Kaputt\.mse-set/)).toBeInTheDocument()
  })

  // **Und ein zweiter Anstoß öffnet ihn wieder.**
  //
  // Ein Wahrheitswert wäre nach dem ersten Mal gesetzt und beim zweiten
  // unverändert — der Dialog käme nicht wieder, und niemand könnte sagen,
  // warum. Deshalb zählt der Anstoß hoch.
  it('öffnen ihn auch beim zweiten Mal', async () => {
    const { horcher, SetVorschauWaehlen } = await bereit()

    ausloesen(horcher, 'menue:set-oeffnen')

    await waitFor(() => {
      expect(SetVorschauWaehlen).toHaveBeenCalledTimes(1)
    })

    ausloesen(horcher, 'menue:set-oeffnen')

    await waitFor(() => {
      expect(SetVorschauWaehlen).toHaveBeenCalledTimes(2)
    })
  })
})
