// **Ein Horcher, der nachholt, was er verpasst hat** — M9 (b) aus dem
// Zweitblick über Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// # Wofür es das gibt
//
// **`<Activity mode="hidden">` baut die Effekte einer verborgenen Seite ab**
// (`seiten/Seiteninhalt.tsx`), und Wails meldet nicht nach. Wer am fremden
// Tisch saß und kurz zu den Karten wechselte, verpasste Tischbild, Züge und
// Gespräch — und sah danach einen Tisch, an dem inzwischen gespielt worden
// war.
//
// # Wie
//
// **Ein Sammler je Ereignis hängt einmal an der Laufzeit und bleibt** —
// gestartet beim ersten Horcher, nicht beim Einblenden. Er puffert, was
// kommt, bis zur [Nachholgrenze]. Jede **Stelle** (der Name des Horchers)
// merkt sich, bis wohin sie gesehen hat; hängt sie wieder ein, bekommt sie
// zuerst, was dazwischen kam, und danach, was kommt.
//
// # Was es nicht tut
//
// **Einer neuen Stelle nachspielen.** Wer zum ersten Mal horcht, fängt jetzt
// an — sonst bekäme eine Seite beim ersten Öffnen die Züge einer Partie, an
// der sie nie saß. **Und die Zeit nachholen**: ein Zeiger, der während des
// Verbergens kam, gilt nach dem Einblenden so, als käme er gerade.
import { laufzeit } from './fenster'

// Nachholgrenze ist, wie viele Meldungen je Ereignis höchstens warten.
//
// **Fünfhundert** — mehr Züge, als eine Partie in der Zeit hat, in der
// jemand eine andere Seite ansieht, und wenig genug, dass eine lange Partie
// kein Speicherleck wird.
export const Nachholgrenze = 500

interface Eintrag {
  nummer: number
  daten: unknown[]

  // eingang ist, wann die Meldung kam — **nicht** wann sie nachgeholt wird.
  // Wer eine Frist an ihr misst, misst sonst ab dem Einblenden (Z2 aus dem
  // Zweitblick über Phase 6.11).
  eingang: number
}

interface Strom {
  quelle: ReturnType<typeof laufzeit>
  eintraege: Eintrag[]
  horcher: Set<(eintrag: Eintrag) => void>
  abmelden: (() => void) | undefined

  // verworfen ist die Nummer der jüngsten Meldung, die über der Grenze
  // wegfiel — 0, solange keine.
  verworfen: number
}

const stroeme = new Map<string, Strom>()
const gesehen = new Map<string, number>()
let zaehler = 0

// strom gibt den Sammler eines Ereignisses — **neu, wenn die Laufzeit eine
// andere ist** als die, an der er hängt, und dann lässt er die alte los:
// sonst hielte jede abgelöste Laufzeit Puffer und Horcher fest.
function strom(ereignis: string): Strom {
  const quelle = laufzeit()
  const bisher = stroeme.get(ereignis)
  if (bisher !== undefined && bisher.quelle === quelle) {
    return bisher
  }

  bisher?.abmelden?.()

  const neu: Strom = {
    quelle,
    eintraege: [],
    horcher: new Set(),
    abmelden: undefined,
    verworfen: 0,
  }
  stroeme.set(ereignis, neu)

  neu.abmelden = quelle?.EventsOn?.(ereignis, (...daten) => {
    zaehler += 1
    const eintrag = { nummer: zaehler, daten, eingang: Date.now() }

    neu.eintraege.push(eintrag)
    for (const weg of neu.eintraege.splice(
      0,
      neu.eintraege.length - Nachholgrenze,
    )) {
      neu.verworfen = weg.nummer
    }

    for (const horcher of neu.horcher) {
      horcher(eintrag)
    }
  })

  return neu
}

// horcheNachholend meldet einen Horcher an — und holt für `stelle` nach,
// was seit ihrem letzten Abmelden kam.
//
// **Und sagt `luecke`, wenn davon etwas fehlt** (Z3 aus dem Zweitblick über
// Phase 6.11): war die Seite so lange verborgen, dass über der
// [Nachholgrenze] Meldungen wegfielen, die diese Stelle nie gesehen hat. Eine
// Stelle, die zum ersten Mal horcht, hat nichts verpasst.
export function horcheNachholend(
  ereignis: string,
  stelle: string,
  nimm: (daten: unknown[], eingang: number) => void,
  luecke?: () => void,
): () => void {
  const sammler = strom(ereignis)
  const schluessel = `${stelle}\u0000${ereignis}`
  const bekannt = gesehen.get(schluessel)
  const zuletzt = bekannt ?? zaehler

  if (bekannt !== undefined && sammler.verworfen > bekannt) {
    luecke?.()
  }

  for (const eintrag of sammler.eintraege) {
    if (eintrag.nummer > zuletzt) {
      nimm(eintrag.daten, eintrag.eingang)
    }
  }

  gesehen.set(schluessel, zaehler)

  const horcher = (eintrag: Eintrag) => {
    gesehen.set(schluessel, eintrag.nummer)
    nimm(eintrag.daten, eintrag.eingang)
  }

  sammler.horcher.add(horcher)

  return () => {
    sammler.horcher.delete(horcher)
  }
}
