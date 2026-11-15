// **Die Anbindung an Wails 3** — ENG-02
// (#1727).
//
// # Warum sie an die alten zwei Stellen hängt
//
// Unter Wails 2 legte die Hülle die Bindungen unter `window.go.app.Bindungen`
// und die Laufzeit unter `window.runtime` ab; `fassade.ts` und `fenster.ts`
// lesen dort, und jeder Lauf legt seine Vorrichtung genau dorthin. Wails 3
// legt nichts mehr ab — es kommt als Modul (`@wailsio/runtime`), gerufen wird
// über `Call.ByName`, gehört über `Events.On`.
//
// **Die Anbindung stellt die alte Gestalt wieder her**, statt Fassade,
// Fenster und sechsundzwanzig Läufe umzubauen: ein Frameworkwechsel ist eine
// eigene Migration und keine Neugestaltung (§5). Was sich ändert, steht hier
// und nirgends sonst.
import type { Api } from './api'
import { Bindungsfehler } from './bindungsfehler'
import { methoden } from './fassade'

import type { FehlerDTO } from './formen'
import { neueKennung } from './kennung'

// Vorsatz ist der volle Name der Fassade in Go — Paketpfad und Typ.
//
// **So bildet Wails 3 ihn** (`bindings.go`, v3.0.0-beta.23):
// `<Paketpfad>.<Typ>.<Methode>`. Ein Tippfehler hier gäbe an jeder Bindung
// „unbekannte Methode" — der Lauf daneben hält den Vorsatz gegen Go.
export const vorsatz =
  'github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.'

// EreignisAbgelegt ist das Ereignis, mit dem Go abgelegte Dateien meldet —
// dasselbe Wort wie `fenster.EreignisAbgelegt` in Go.
export const ereignisAbgelegt = 'datei:abgelegt'

// Rufer ist der Teil von `Call`, den die Anbindung braucht.
export interface Rufer {
  ByName: (name: string, ...argumente: unknown[]) => Promise<unknown>
}

// Hoerer ist der Teil von `Events`, den die Anbindung braucht.
export interface Hoerer {
  On: (
    name: string,
    rueckruf: (ereignis: { data: unknown }) => void,
  ) => () => void
}

// Ablage ist, was Go über ein Ablegen meldet (`fenster.Ablage`).
interface Ablage {
  x: number
  y: number
  pfade: string[]
}

// fehlermarke steht vor dem Fehler, den ein Tor in Go als Panik wirft —
// dasselbe Wort wie `app.Fehlermarke`, und ein Lauf in Go hält beide gleich.
export const fehlermarke = 'mse4mac-fehler:'

// ausgepackt macht aus einer Abweisung des Tors einen Bindungsfehler — OPS-10
// (#1908).
//
// **Wails gibt eine Panik nur als Satz weiter** (`<Methode>: panic: …`), und
// ein abgeschalteter Bereich soll im Fenster dieselbe Abweisung sein wie an
// den drei Flächen ohne Fenster.
//
// **Jede andere Abweisung ist fremd** und wird ein Fehler »unbekannt« mit
// Kennung — F03 (#2342).
// Keine Bindung gibt einen `error` zurück; was dennoch abweist, ist eine
// Panik (`<Paketpfad>.Bindungen.<Methode>: panic: <Wert>`, `bindings.go`)
// oder Wails' eigener Weg für eine Antwort ohne JSON —
// `new Error(await response.text())` (`runtime.ts`, beide v3.0.0-beta.27).
// Bis zum Zweitblick über Phase 1 von 0.9.0 (Befund 6) ging nur die Panik
// diesen Weg, und der Rest zeigte seinen Rohtext im Fenster. Paketpfad,
// Laufzeitsatz und Antworttext gehören ins Protokoll; gemeldet wird über
// denselben Rufer, mit der Methode als Bereich.
function ausgepackt(grund: unknown, ruf: Rufer, methode: string): Error {
  const text = grund instanceof Error ? grund.message : String(grund)

  const stelle = text.indexOf(fehlermarke)
  if (stelle >= 0) {
    try {
      const fehler = JSON.parse(
        text.slice(stelle + fehlermarke.length),
      ) as FehlerDTO

      return new Bindungsfehler(fehler)
    } catch {
      // **Eine Marke, die sich nicht lesen lässt, ist fremd wie jede andere.**
    }
  }

  return fremdAlsFehler(text, ruf, methode)
}

// fremdAlsFehler meldet den Rohtext unter einer neuen Kennung und gibt den
// Fehler, den das Fenster zeigt. **Die Meldung ist Bestenfalls-Arbeit.**
function fremdAlsFehler(text: string, ruf: Rufer, methode: string): Error {
  const kennung = neueKennung()

  ruf
    .ByName(`${vorsatz}FensterfehlerMelden`, {
      kennung,
      ebene: 'ruf',
      bereich: methode.toLowerCase().slice(0, 40),
      meldung: text,
    })
    .catch(() => undefined)

  return new Bindungsfehler({
    code: 'unbekannt',
    schluessel: 'unbekanntMitVorgang',
    meldung: '',
    werte: { vorgang: kennung },
  })
}

// wailsBindungen baut für jede Methode der Fassade einen Ruf nach Go.
export function wailsBindungen(ruf: Rufer): Partial<Api> {
  const gebaut: Record<string, (...argumente: unknown[]) => Promise<unknown>> =
    {}

  for (const name of methoden) {
    gebaut[name] = (...argumente) =>
      ruf
        .ByName(vorsatz + name, ...argumente)
        .catch((grund: unknown) => Promise.reject(ausgepackt(grund, ruf, name)))
  }

  return gebaut
}

// werte liest die Werte eines Ereignisses — **immer eine Liste**, denn Go
// schickt sie so (`fenster.nutzlast`). Alles andere ist kein Wert.
function werte(daten: unknown): unknown[] {
  return Array.isArray(daten) ? daten : []
}

// Fensterlaufzeit ist der Teil von Wails' `Window`, den die Oberfläche
// braucht — der Titel (AP-7, #2364).
export interface Fensterlaufzeit {
  SetTitle: (titel: string) => Promise<void>
}

// wailsLaufzeit gibt die Laufzeit in der Gestalt, die `fenster.ts` liest.
export function wailsLaufzeit(hoeren: Hoerer, fenster?: Fensterlaufzeit) {
  let ablageAbmelden: (() => void) | undefined

  return {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) =>
      hoeren.On(name, (ereignis) => {
        rueckruf(...werte(ereignis.data))
      }),

    // **Der zweite Parameter bleibt ungenutzt** — wie unter Wails 2 steht er
    // auf `false`; welches Feld gemeint ist, weiß die Oberfläche selbst.
    OnFileDrop: (rueckruf: (x: number, y: number, pfade: string[]) => void) => {
      ablageAbmelden?.()
      ablageAbmelden = hoeren.On(ereignisAbgelegt, (ereignis) => {
        const ablage = ereignis.data as Ablage
        rueckruf(ablage.x, ablage.y, ablage.pfade)
      })
    },

    OnFileDropOff: () => {
      ablageAbmelden?.()
      ablageAbmelden = undefined
    },

    // **Der Titel über Wails' Fenster**, ohne eigenen Ruf nach Go — er
    // schreibt nichts und gehört der Hülle (AP-7, #2364).
    TitelSetzen: (titel: string) => {
      void fenster?.SetTitle(titel)
    },
  }
}

// anbinden legt Bindungen und Laufzeit dort ab, wo Fassade und Fenster sie
// suchen.
export function anbinden(
  ruf: Rufer,
  hoeren: Hoerer,
  ziel: object,
  fenster?: Fensterlaufzeit,
): void {
  Object.assign(ziel, {
    go: { app: { Bindungen: wailsBindungen(ruf) } },
    runtime: wailsLaufzeit(hoeren, fenster),
  })
}
