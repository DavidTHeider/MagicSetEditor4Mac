// **Wie das Fenster dasteht, soweit die Menüleiste es wissen muss** — UXR-F14
// (#2029).
//
// # Wozu
//
// Das Menü *Darstellung* blendet Seitenleiste, Navigator und Inspector ein und
// aus, und seine Einträge heißen nach dem Stand: „einblenden", wo das Teil zu
// ist, „ausblenden", wo es steht — und sie sind gegraut, wo die Seite das Teil
// nicht hat (Apple, Human Interface Guidelines, „The menu bar"). Die Leiste ist
// nativ und liegt in Go; **was steht, weiß nur das Fenster**.
//
// # Wer meldet, und wohin
//
// Jedes Teil meldet sich **dort, wo es lebt** (`useFensterteilmeldung`): die
// Seitenleiste in `useSeitenleiste`, die Spalten im `Arbeitsplatz`. Der
// `Darstellungsstandraum` sammelt die Meldungen und gibt den ganzen Stand an
// `DarstellungsstandMelden` — **ein** Weg für alle drei, und ein Eintrag, der
// etwas umlegt, geht den umgekehrten (Ereignis ins Fenster, neuer Stand
// zurück). Go führt keinen eigenen Stand; es merkt sich den letzten gemeldeten
// (`internal/app/menuenaht.go`).
//
// # Was ein Teil meldet
//
// `vorhanden`: es gibt dort etwas umzulegen — die Seite hat den Arbeitsplatz
// und die Spalte, die Leiste steht links statt oben, keine Schublade fängt die
// Tasten. `offen`: es steht (bei Enge: es liegt über dem Zentrum). **Ein Teil,
// das nicht meldet, fehlt** — der Nullwert ist der Anfang, und eine Seite ohne
// Arbeitsplatz muss dafür nichts tun.
import { createContext, use, useEffect } from 'react'

// Fensterteil ist der Stand eines Teils — dieselben zwei Angaben wie
// `app.Fensterteil` in Go.
export interface Fensterteil {
  vorhanden: boolean
  offen: boolean
}

// Darstellungsstand ist, was `DarstellungsstandMelden` entgegennimmt.
export interface Darstellungsstand {
  seitenleiste: Fensterteil
  navigator: Fensterteil
  inspector: Fensterteil
}

export type Fensterteilname = keyof Darstellungsstand

// fehlt ist ein Teil, das niemand meldet.
export const fehlt: Fensterteil = { vorhanden: false, offen: false }

// anfangsstand ist der Nullwert aus Go: nichts ist da, alles ist zu.
export const anfangsstand: Darstellungsstand = {
  seitenleiste: fehlt,
  navigator: fehlt,
  inspector: fehlt,
}

// Teilmeldung ist eine Zeile an den Raum: dieses Teil steht jetzt so da.
export interface Teilmeldung {
  teil: Fensterteilname
  wert: Fensterteil
}

export type Melder = (meldung: Teilmeldung) => void

// Darstellungsmeldung ist der Weg vom Teil zum Raum.
//
// **Ohne Raum gibt es keinen Melder, und das ist kein Fehler**: ein Bauteil,
// das außerhalb der Anwendung geprüft wird — jeder Test des Arbeitsplatzes —,
// meldet dann nichts, genau wie `Verdeckung` ohne Schublade falsch ist.
export const Darstellungsmeldung = createContext<Melder | undefined>(undefined)

// gleichesTeil sagt, ob zwei Teile gleich dastehen.
export function gleichesTeil(a: Fensterteil, b: Fensterteil): boolean {
  return a.vorhanden === b.vorhanden && a.offen === b.offen
}

// gleicherStand sagt, ob zwei Stände gleich sind.
export function gleicherStand(
  a: Darstellungsstand,
  b: Darstellungsstand,
): boolean {
  return (
    gleichesTeil(a.seitenleiste, b.seitenleiste) &&
    gleichesTeil(a.navigator, b.navigator) &&
    gleichesTeil(a.inspector, b.inspector)
  )
}

// zusammenfuehren trägt eine Meldung in den Stand ein.
//
// **Gleiches gibt denselben Stand zurück**, nicht einen gleichen: React
// zeichnet dann nicht neu, und der Raum sendet nichts. Ein Teil, das beim
// Wechsel der Seite erst abmeldet und gleich wieder meldet, ändert so nichts.
export function zusammenfuehren(
  stand: Darstellungsstand,
  meldung: Teilmeldung,
): Darstellungsstand {
  if (gleichesTeil(stand[meldung.teil], meldung.wert)) {
    return stand
  }

  return { ...stand, [meldung.teil]: meldung.wert }
}

// useFensterteilmeldung meldet, wie ein Teil dasteht — und nimmt es zurück,
// wenn das Bauteil geht.
//
// **Abmelden heißt „fehlt"**: der Arbeitsplatz einer Seite, die man verlässt,
// ist weg, und die Menüeinträge darauf sind es auch. Beim Wechsel zwischen zwei
// Arbeitsplätzen laufen Abmelden und Anmelden im selben Zug und heben sich
// auf (`zusammenfuehren`).
export function useFensterteilmeldung(
  teil: Fensterteilname,
  vorhanden: boolean,
  offen: boolean,
): void {
  const melde = use(Darstellungsmeldung)

  useEffect(() => {
    melde?.({ teil, wert: { vorhanden, offen } })

    return () => {
      melde?.({ teil, wert: fehlt })
    }
  }, [melde, teil, vorhanden, offen])
}
