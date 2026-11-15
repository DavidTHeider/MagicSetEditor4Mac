// **Ein Entwurf merkt sich, wovon er ausging** — U01 bis U06
// (#2119).
//
// # Der Befund
//
// Die Formulare stehen auf `defaultValue` und wurden bei jedem Verlaufsstand
// von außen neu gebaut, damit nach ⌘Z kein zurückgenommener Text im Feld
// stand. Der Preis: jede Änderung irgendwo — ein anderes Feld, ein anderes
// Programm — nahm dem, der gerade tippte, Entwurf und Fokus. Und änderte ein
// anderer **dasselbe** Feld, schrieb das nächste Verlassen still darüber.
//
// # Der Vertrag
//
// - **`schluessel`** gehört an die Eingabe, und er wechselt nur, wenn sie
//   etwas anderes zeigen muss als das, was in ihr steht: ein fremder Wert
//   ohne Entwurf, oder ein verworfener Entwurf. Solange ein Entwurf lebt,
//   bleibt er — Text, Schreibmarke und Fokus bleiben. **Und kommt der eigene
//   Entwurf im Bestand an, bleibt er auch**: die Eingabe zeigt ihn schon,
//   und ein Neuaufbau nähme dem, der nach einem Einschub weitertippt, den
//   Fokus.
// - **Konflikt** heißt: der Bestand zeigt einen Wert, den der Entwurf nicht
//   kennt — weder den Ausgangswert noch einen, den er selbst geschickt oder
//   gesehen hat, noch sich selbst. Dann schreibt Verlassen nicht; `behalten`
//   schreibt ausdrücklich, `verwerfen` nimmt den fremden Wert. **Geschickte
//   Werte zählen mit**, denn wer verlässt, weitertippt und wieder verlässt,
//   sieht danach seinen ersten Text im Bestand ankommen.
// - **Geschickt ist nicht geschrieben.** Ein Entwurf bleibt nach dem
//   Schreiben stehen, bis der Bestand ihn zeigt; scheitert es, ist er wieder
//   ein gewöhnlicher Entwurf **mit den Werten von damals** — und eine
//   fremde Änderung danach ist ein Konflikt, kein stilles Überschreiben.
// - **Und `schreiben` bekommt die bekannten Werte mit** (U02b, #2179): ein
//   fremder Wert, den das Fenster noch **nicht gezeigt** hat, ist sonst
//   unsichtbar — der Bestand prüft die Liste in der Transaktion, die
//   schreibt, und antwortet mit `fremderWert`.
// - **Sagt `schreiben`, dass es geglückt ist, ist der Entwurf erledigt,
//   sobald sich der Bestand ändert** — auch wenn er dort nicht wörtlich
//   ankommt (Gesamtreview über 0.9.0, Befund 1): seit #2396 speichert der
//   Bestand den gerechneten Wert, aus „Glut" wird „[Glut]". Die Eingabe wird
//   dann neu aufgebaut, denn sie hält den getippten. Ein Schreiben, das
//   nichts zurückgibt, bleibt beim wörtlichen Vergleich.
// - **Sagt es, was es gespeichert hat, gilt genau das** (#2535): erledigt ist
//   der Entwurf, wenn der Bestand **diesen** Wert zeigt — auch wenn er der
//   Ausgangswert ist —, und ein anderer ist ein Konflikt. „Irgendeine
//   Änderung" nahm sonst einen fremden Wert als eigenen, der zwischen
//   Speichern und Nachladen kam.
import { useState } from 'react'

interface Entwurf {
  bekannt: readonly string[]
  wert: string
  abgeschickt: boolean

  // vorher ist der Bestand beim Abschicken; bestaetigt sagt, dass das
  // Schreiben geglückt ist, gespeichert, was es gespeichert hat.
  vorher?: string
  bestaetigt?: boolean
  gespeichert?: string
}

// gespeichertAus liest den gespeicherten Wert aus der Antwort von `schreiben`.
function gespeichertAus(gelungen: unknown): string | undefined {
  if (typeof gelungen !== 'object' || gelungen === null) {
    return undefined
  }

  const wert: unknown = (gelungen as { gespeichert?: unknown }).gespeichert

  return typeof wert === 'string' ? wert : undefined
}

export interface Entwurfsgriff {
  schluessel: string
  konflikt: boolean
  aendern: (neu: string) => void
  verlassen: (neu: string) => void
  behalten: () => void
  verwerfen: () => void
  uebernehmen: (neu: string) => void

  // bekannt sind die Werte, die dieses Feld kennt — für ein Schreiben **neben**
  // dem Entwurf, das seinen Wert selbst rechnet (die Wortwahl): im Konflikt
  // ohne den fremden, damit der Bestand die Wahl abweist (Befund 1 des
  // Nachlaufs über Phase 0.7.0-I).
  bekannt: readonly string[]
}

// useEntwurf hält den Entwurf **eines** Feldes gegen seinen gespeicherten
// `wert`. `getippt` meldet ihn einer Vorschau; `undefined` verwirft ihn dort.
export function useEntwurf(
  wert: string,
  // **Gibt `schreiben` ein Versprechen zurück, sagt es, ob das Schreiben
  // glückte** — `true` bestätigt den Entwurf, `{ gespeichert }` bestätigt ihn
  // mit dem Wert, der im Bestand steht (#2535). Alles andere bleibt beim
  // wörtlichen Vergleich.
  schreiben: (neu: string, bekannt: readonly string[]) => unknown,
  getippt: (neu: string | undefined) => void = () => undefined,
): Entwurfsgriff {
  const [entwurf, setzeEntwurf] = useState<Entwurf>()

  // anzeige sagt, welcher Wert in der Eingabe der Fassung `fassung` steht,
  // wenn kein Entwurf lebt.
  const [anzeige, setzeAnzeige] = useState({ fassung: 0, wert })

  // ruhen beendet den Entwurf; die Eingabe zeigt danach `gezeigt`.
  const ruhen = (gezeigt: string, neuAufbauen = false) => {
    setzeEntwurf(undefined)
    setzeAnzeige((bisher) => ({
      fassung: bisher.fassung + (neuAufbauen ? 1 : 0),
      wert: gezeigt,
    }))
  }

  // **Erledigt ist er, wenn der Bestand ihn zeigt** — und dann wird er
  // vergessen, nicht nur übergangen: sonst lebte er bei der nächsten fremden
  // Änderung als Konflikt wieder auf. **Ohne Entwurf folgt die Eingabe dem
  // Bestand.** Beides ist Zustand, der aus den Eigenschaften folgt, und wird
  // während des Zeichnens nachgeführt (React: „Adjusting some state when a
  // prop changes").
  const erledigt =
    entwurf?.abgeschickt === true &&
    (wert === entwurf.wert ||
      (entwurf.gespeichert === undefined
        ? entwurf.bestaetigt === true && wert !== entwurf.vorher
        : wert === entwurf.gespeichert))
  if (erledigt) {
    ruhen(wert, wert !== entwurf.wert)
  } else if (entwurf === undefined && wert !== anzeige.wert) {
    ruhen(wert, true)
  }
  const lebt = erledigt ? undefined : entwurf

  const konflikt =
    lebt !== undefined && wert !== lebt.wert && !lebt.bekannt.includes(wert)

  // geschickt sagt, ob dieser Entwurf schon geschrieben hat — `abschicken`
  // nimmt dabei den Wert unter die bekannten auf, neben den Ausgangswert.
  const geschickt = lebt !== undefined && lebt.bekannt.length > 1

  // frisch ist der lebende Entwurf — oder einer, der eben beginnt.
  const frisch: Entwurf = lebt ?? {
    bekannt: [wert],
    wert,
    abgeschickt: false,
  }

  // **Im Zustand ungekürzt, hinaus ohne Doppel**: `geschickt` zählt die
  // Länge, und eine entdoppelte Liste machte aus einem geschickten Einschub
  // gleichen Werts einen ungeschickten.
  const abschicken = (neu: string) => {
    const bekannt = [...frisch.bekannt, wert, neu]
    const geschickt: Entwurf = {
      ...frisch,
      bekannt,
      wert: neu,
      abgeschickt: true,
      vorher: wert,
      bestaetigt: false,
    }
    setzeEntwurf(geschickt)

    // **Nur dieser Entwurf wird bestätigt** — wer danach weitertippt, hat
    // einen neuen, und den bestätigt dieses Schreiben nicht.
    const ergebnis = schreiben(neu, [...new Set(bekannt)])
    if (ergebnis instanceof Promise) {
      void ergebnis.then((gelungen: unknown) => {
        const gespeichert = gespeichertAus(gelungen)
        if (gelungen === true || gespeichert !== undefined) {
          setzeEntwurf((jetzt) =>
            jetzt === geschickt
              ? {
                  ...jetzt,
                  bestaetigt: true,
                  ...(gespeichert === undefined ? {} : { gespeichert }),
                }
              : jetzt,
          )
        }
      })
    }
  }

  const aendern = (neu: string) => {
    setzeEntwurf({ ...frisch, wert: neu, abgeschickt: false })
    getippt(neu)
  }

  return {
    schluessel: String(anzeige.fassung),
    konflikt,
    bekannt: [
      ...new Set(konflikt ? frisch.bekannt : [...frisch.bekannt, wert]),
    ],
    aendern,
    verlassen: (neu) => {
      if (konflikt) {
        return
      }

      // **Ein unveränderter Wert wird nicht geschrieben.** Wer ein Feld
      // anklickt und wieder verlässt, legte sonst einen Schritt im Verlauf
      // an, der nichts ändert — und ⌘Z nähme ihn dann zurück, ohne dass
      // etwas anders aussieht. **Außer, dieser Entwurf hat schon geschickt**:
      // dann bekommt der Bestand gleich einen anderen Text, und wer zum
      // Ausgangswert zurückkehrt, will ihn dort sehen.
      if (neu === wert && !geschickt) {
        ruhen(wert)

        return
      }

      abschicken(neu)
    },
    // **Der fremde Wert ist gesehen** — `abschicken` nimmt ihn unter die
    // bekannten auf. Gerufen wird es nur im Konflikt, also mit lebendem
    // Entwurf; `frisch` ist dann er.
    behalten: () => {
      abschicken(frisch.wert)
    },
    verwerfen: () => {
      ruhen(wert, true)
      getippt(undefined)
    },
    // **Ein Einschub aus einem Menü ist Tippen und Verlassen in einem** —
    // im Konflikt nur Tippen: dann entscheidet erst `behalten`.
    uebernehmen: (neu) => {
      if (konflikt) {
        aendern(neu)

        return
      }

      abschicken(neu)
    },
  }
}
