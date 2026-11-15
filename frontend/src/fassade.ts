// **Die Fassade zur Wails-Bindung** — und was ohne sie geschieht.
//
// Eine eigene Datei, weil `api.ts` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was** die Anwendung anbietet, hier,
// **wie** man dorthin kommt.
import type { Api } from './api'
import vertrag from './bindungsvertrag.json'
import { Katalogfehler } from './katalogfehler'

// **Der Pfad ist `go.app.Bindungen`, nicht `go.app.App`.**
//
// Wails legt die gebundene Struktur unter ihrem **Typnamen** ab. Solange `*App`
// gebunden war, hieß sie `App`; seit die Oberfläche eine eigene Fassade ist
// (`internal/app/bindungen.go`), heißt sie `Bindungen`. Der alte Pfad fand
// nichts — und die Fassade legte klaglos ihre Platzhalter darüber, also genau
// die Form, die aussieht, als wäre nichts passiert.
//
// Gesehen hat das erst der Aufruf in der laufenden Anwendung. Kein Test kann
// es sehen: die Vorrichtung legt selbst an, was hier gesucht wird.
interface MitGo {
  go?: { app?: { Bindungen?: Partial<Api> } }
}

// methoden sind die Namen aus dem **quittierten Bindungsvertrag** —
// #1825.
//
// # Warum hier keine eigene Liste steht
//
// **Eine Tatsache, eine Liste**: `bindungsvertrag.json` führt die Namen. Er
// sagt mehr als sie (den Abdruck jeder Signatur), und ein Wächter hält ihn
// gegen Go (`internal/werkzeuge/bindungsvertrag_test.go`).
//
// # Was die Gleichheit mit `Api` hält
//
// **Der Übersetzer, in beide Richtungen.** `abdruecke` verlangt für jeden
// Namen aus `Api` einen Eintrag im Vertrag; die Zuweisung an `methoden`
// verlangt, dass jeder Vertragsname in `Api` steht. Fehlt einer auf einer
// Seite, bricht `tsc` — nicht erst das Fenster.
const abdruecke: Record<keyof Api, string> = vertrag.methoden

export const methoden: readonly (keyof Api)[] = Object.keys(
  abdruecke,
) as (keyof typeof vertrag.methoden)[]

// api gibt die Fassade — echt, wenn Wails läuft, sonst eine, die sich meldet.
export function api(): Api {
  return fassade((globalThis as unknown as MitGo).go?.app?.Bindungen)
}

// fassade füllt jede fehlende Methode mit einer, die ihren Namen nennt.
//
// **Nicht mit `undefined`.** Eine Fassade, die halb belegt ist und im Übrigen
// nichts liefert, führt zu einer leeren Seite ohne Meldung — und dann sucht
// jemand im Zustand statt in der Vorrichtung.
export function fassade(teile: Partial<Api> | undefined): Api {
  const gebaut = {} as Record<keyof Api, unknown>

  for (const name of methoden) {
    const vorhanden = teile?.[name]

    // **Abgelehnt, nicht geworfen.** Ein Platzhalter, der synchron wirft,
    // bricht die Zusage `Promise<T>`: ein `.catch()` am Aufrufer fängt ihn
    // nicht, und der Fehler landet als unbehandelte Ausnahme irgendwo, wo
    // niemand ihn dem Aufruf zuordnet.
    gebaut[name] =
      vorhanden ??
      (() => Promise.reject(new Katalogfehler('methodeFehlt', { name })))
  }

  return gebaut as unknown as Api
}

// rufAbbrechen bricht einen laufenden Ruf ab — BK-07d
// (#2118).
//
// **Über das `cancel` der Wails-Laufzeit**: jeder Ruf ist dort ein
// `CancellablePromise`, und sein Abbruch beendet den Rufkontext in Go
// (`@wailsio/runtime` 3.0.0-beta.23, `cancellable.js` und `calls.js`,
// nachgelesen am 2026-10-07; `.catch` in `anbindung.ts` reicht ihn weiter).
// Ein Ruf ohne `cancel` — eine Vorrichtung, ein schon erledigter — bleibt
// unberührt.
export function rufAbbrechen(lauf: Promise<unknown> | undefined): void {
  const abbrechbar = lauf as { cancel?: () => unknown } | undefined

  void abbrechbar?.cancel?.()
}
