// **Was neben dem Spiel läuft** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum ein Haken für beide Seiten
//
// **Weil am Tisch geredet wird und nicht in eine Richtung.** Wirt und Gast
// bekommen dieselben Meldungen über dieselbe Naht und schicken über dieselben
// zwei Bindungen; zwei Haken liefen beim ersten Unterschied auseinander (§16).
//
// # Warum das Gespräch wächst und die Zeiger nicht
//
// **Weil sie Verschiedenes sind.** Ein Satz bleibt stehen; ein Zeiger ist der
// **letzte** Ort eines Sitzes und ersetzt den vorigen. Der erste Anlauf legte
// beides in dieselbe Liste — bei zwanzig Meldungen je Sekunde sind das 1.200
// Einträge in der Minute, und niemand sieht einen davon.
//
// # Und warum der Ablauf hier steht und nicht in Go
//
// **Weil er eine Aussage über die Wanduhr ist und über das, was jemand
// sieht.** Das Issue verlangt „monotonen Empfängerstand **und Ablauf**": den
// Stand hält Go, denn er ist eine Zusage gegen das andere Ende der Leitung —
// eine überholte Präsenz kommt gar nicht erst an. Der Ablauf gehört dorthin,
// wo er sichtbar wird.
//
// **Was er nicht ist: der Beleg dafür, dass jemand da ist.** Wer am Tisch
// sitzt, sagt die Sitzung. Läuft ein Zeiger ab, heißt das nur, dass jemand die
// Maus hat stehen lassen.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api, auspacken, type Nebenbei } from '../api'
import { horcheNachholend } from '../nachholen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'

// ereignisNebenbei ist die Naht, über die Chat und Präsenz hereinkommen.
//
// **Wörtlich dieselbe wie in `internal/app/partiebeitritt.go`.**
export const ereignisNebenbei = 'partie:nebenbei'

// Praesenzfrist ist, wie lange ein Zeiger gilt, wenn nichts nachkommt.
//
// **Fünfzehn Sekunden, und die Zahl ist eine Setzung.** Kürzer flackerte ein
// Zeiger weg, sobald jemand nachdenkt; länger stünde er da, während der andere
// längst am Kühlschrank ist. **Der verlässliche Weg, einen Zeiger loszuwerden,
// ist das Ende der Sitzung** — die Frist fängt nur den Fall, dass die Leitung
// steht und der Mensch nicht.
export const Praesenzfrist = 15_000

// Gespraechslaenge ist, wie viele Sätze das Fenster behält.
//
// **Hundert, und der Rest fällt vorn heraus.** Eine Liste ohne Grenze ist in
// einer langen Partie ein Speicherleck mit Text darin.
export const Gespraechslaenge = 100

// Zeigertakt ist, wie oft nachgesehen wird, ob etwas abgelaufen ist.
export const Zeigertakt = 1_000

export interface Chatzeile {
  // nummer ist der Schlüssel der Zeile — zwei gleiche Sätze bleiben zwei.
  nummer: number

  von: string
  text: string
}

export interface Zeigerstand {
  von: string
  art: string
  objekt: string
  x: number
  y: number

  // seit ist die Wanduhr dieses Fensters beim Eintreffen.
  //
  // **Nicht die des Absenders.** Zwei Rechner gehen verschieden, und ein
  // Ablauf gegen eine fremde Uhr liefe mal zu früh und mal nie.
  seit: number
}

export interface Nebenlauf {
  gespraech: Chatzeile[]

  // zeiger sind die **gültigen** Zeiger, je Sitz einer.
  zeiger: Zeigerstand[]

  sagen: (text: string) => void
  zeigen: (art: string, objekt: string, x: number, y: number) => void

  fehler: string | undefined
}

// usePartienebenbei hört zu und lässt reden — für beide Seiten des Tisches.
//
// **`eigenerSitz` kommt von außen**, weil nur der Aufrufer ihn kennt: der Wirt
// aus seiner Freigabe, der Gast aus dem Tischbild (`sicht.fuer`). Ein Haken,
// der ihn selbst holte, holte ihn zweimal verschieden.
//
// **`partie` ebenso** — Zweitblick über Phase 6.11 (#1672, zu Z1): Wirt und
// Gast hören dasselbe Ereignis, und dieselbe Anwendung kann zugleich beides
// sein. Genommen wird nur, was zu `partie` gehört; solange sie leer ist,
// nichts.
export function usePartienebenbei(
  eigenerSitz: string,
  stelle: string,
  partie: string,
): Nebenlauf {
  const texte = useTexte()

  const [gespraech, setzeGespraech] = useState<Chatzeile[]>([])
  const [zeiger, setzeZeiger] = useState<Zeigerstand[]>([])
  const [fehler, setzeFehler] = useState<string>()

  const naechste = useRef(0)
  const meine = useRef(partie)

  // **Ein Wechsel der Partie räumt, was zur vorigen gehörte** (§7, Regel 1;
  // N4 aus dem Nachlauf über 6.11): der Haken bleibt eingehängt, und die
  // Sätze von A stünden sonst im Gespräch von B.
  const [fuer, setzeFuer] = useState(partie)
  if (fuer !== partie) {
    setzeFuer(partie)
    setzeGespraech([])
    setzeZeiger([])
    setzeFehler(undefined)
  }

  useEffect(() => {
    meine.current = partie
  })

  // **Der Horcher hängt am Fenster und nicht am Tisch.** Einer, der erst beim
  // Beitreten angemeldet würde, verpasste das erste Wort — `EventsOn` meldet
  // nicht nach.
  // **Und er holt nach** (M9 b): verborgen baut `<Activity>` ihn ab, und
  // was in der Zeit gesagt wurde, bekommt er beim Einblenden — je `stelle`,
  // denn Wirt und Gast lesen dasselbe Ereignis.
  useEffect(
    () =>
      horcheNachholend(ereignisNebenbei, stelle, (daten, eingang) => {
        const [von, eines] = daten
        if (meine.current === '' || von !== meine.current) return

        nimmNebenbei(
          eines as Nebenbei,
          eingang,
          naechste,
          setzeGespraech,
          setzeZeiger,
        )
      }),
    [stelle],
  )

  // **Der Ablauf** — und der Takt läuft nur, solange es etwas abzulaufen gibt.
  useEffect(() => {
    if (zeiger.length === 0) return undefined

    const takt = setInterval(() => {
      setzeZeiger(nurGueltige)
    }, Zeigertakt)

    return () => {
      clearInterval(takt)
    }
  }, [zeiger.length])

  const sagen = useCallback(
    (text: string) => {
      void (async () => {
        setzeFehler(undefined)

        // **Eine Antwort gehört zur Partie, für die gesagt wurde** (§7,
        // Regel 2).
        const fuerPartie = meine.current

        try {
          auspacken(await api().PartieSagen(text), () => undefined)
          if (meine.current !== fuerPartie) return

          // **Das eigene Wort steht danach da und nicht davor** (§7,
          // Regel 3): der Rundruf geht an alle **außer** diesem Fenster, und
          // was nicht hinausging, soll auch nicht dastehen.
          naechste.current += 1

          setzeGespraech((bisher) =>
            angehaengt(bisher, {
              nummer: naechste.current,
              von: eigenerSitz,
              text,
            }),
          )
        } catch (ursache: unknown) {
          if (meine.current === fuerPartie) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })()
    },
    [texte, eigenerSitz],
  )

  const zeigen = useCallback(
    (art: string, objekt: string, x: number, y: number) => {
      void (async () => {
        try {
          auspacken(
            await api().PartieZeigen(art, objekt, x, y),
            () => undefined,
          )
        } catch {
          // **Ein verlorener Zeiger ist kein Befund.** Er ist flüchtig, der
          // nächste kommt in Millisekunden, und eine Meldung dafür stünde
          // dauerhaft im Fenster für etwas, das niemand beheben kann.
        }
      })()
    },
    [],
  )

  return { gespraech, zeiger, sagen, zeigen, fehler }
}

// nimmNebenbei trennt, was hereinkommt: ein Satz wächst an, ein Zeiger ersetzt.
function nimmNebenbei(
  eines: Nebenbei,
  eingang: number,
  naechste: { current: number },
  setzeGespraech: (fort: (bisher: Chatzeile[]) => Chatzeile[]) => void,
  setzeZeiger: (fort: (bisher: Zeigerstand[]) => Zeigerstand[]) => void,
): void {
  if (eines.art === 'chat') {
    naechste.current += 1

    const zeile = { nummer: naechste.current, von: eines.von, text: eines.text }

    setzeGespraech((bisher) => angehaengt(bisher, zeile))

    return
  }

  const stand: Zeigerstand = {
    von: eines.von,
    art: eines.zeigerart,
    objekt: eines.objekt,
    x: eines.x,
    y: eines.y,
    seit: eingang,
  }

  // **Abgelaufen kommt er nicht wieder** (Z2): nachgeholt trägt er seinen
  // Eingang, und war die Frist im Verborgenen um, zeigt er nichts mehr.
  if (Date.now() - eingang >= Praesenzfrist) return

  setzeZeiger((bisher) => [
    ...nurGueltige(bisher).filter((einer) => einer.von !== stand.von),
    stand,
  ])
}

// angehaengt hängt eine Zeile an und wirft vorn weg, was zu viel ist.
function angehaengt(bisher: Chatzeile[], zeile: Chatzeile): Chatzeile[] {
  const alle = [...bisher, zeile]

  return alle.length > Gespraechslaenge
    ? alle.slice(alle.length - Gespraechslaenge)
    : alle
}

// nurGueltige lässt fallen, was abgelaufen ist — und gibt **dieselbe** Liste
// zurück, wenn nichts fiel.
//
// **Sonst zeichnete das Fenster je Sekunde neu**, ohne dass sich etwas ändert:
// eine neue Liste ist für React ein neuer Wert, auch wenn dasselbe darin steht.
function nurGueltige(bisher: Zeigerstand[]): Zeigerstand[] {
  const jetzt = Date.now()
  const gueltig = bisher.filter((einer) => jetzt - einer.seit < Praesenzfrist)

  return gueltig.length === bisher.length ? bisher : gueltig
}
