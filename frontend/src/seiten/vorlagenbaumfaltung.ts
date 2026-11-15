// **Wann der Vorlagenbaum faltet** — UXR-18
// (#1926),
// Regel R6 aus `docs/UIREVIEW.md`.
//
// # Der Anlass
//
// Gemessen am 2026-10-02 am nativen Fenster (Hüllenprobe `uirevue`, Commit
// ddadb180): mit den 526 Paketen des Full-Magic-Pack zeichnete die Werkstatt
// **35.769 Bedienelemente**, 15 davon im Bild. Fast alle waren Beiwerk — rund
// 33.000 Bilder und Schriften, jedes ein Knopf, und alles aufgeklappt
// (`service.Textendungen` zählt dasselbe Pack). Jeder Knopf kostet
// Webprozess-Speicher (#1852).
//
// # Die Regel
//
// - **Ein Baum, der aufgeklappt höchstens `Faltgrenze` Zeilen zeichnet,
//   steht wie bisher ganz offen**, das Beiwerk in der Reihe. Der Kernweg
//   `paket` bleibt damit bei fünf Handlungen (UX-ARCHITEKTUR).
// - **Ein größerer beginnt zugeklappt.** Offen steht das Paket der Wahl und,
//   was jemand selbst aufgeklappt hat; das Beiwerk eines offenen Pakets steht
//   hinter einer eigenen Zeile.
// - **Keine Liste zeigt mehr als `Stueck` Einträge auf einmal**; der Rest
//   steht hinter „Weitere zeigen". Die Wahl steht immer da.
//
// **Gerechnet wird über das, was der Filter übrig lässt**: wer „rahmen"
// tippt, sieht seine Treffer aufgeklappt, solange es wenige sind — derselbe
// Weg wie vorher.
//
// # Warum falten und nicht nur zeichnen, was zu sehen ist
//
// `useSichtfenster` rechnet mit **einer** Zeilenhöhe. Der Baum hat Zeilen
// verschiedener Höhe — eine Meldung am Zweig, das Angebot einer Kopie —, und
// ein Baum, der sie abschnitte, verlöre den Satz, der sagt, was zu tun ist.
import type { Zweig } from '../haken/useVorlagenbaum'

// Faltgrenze ist R6: über 200 gezeichnete Bedienelemente wird gefaltet.
export const Faltgrenze = 200

// Stueck ist, wie viele Einträge eine Liste auf einmal zeigt — die Hälfte
// der Grenze, damit neben einer vollen Paketliste ein offenes Paket Platz hat.
export const Stueck = 100

// Klappwahl hält, was jemand von Hand auf- oder zugeklappt hat — `p:` vor
// einem Paket, `b:` vor seinem Beiwerk.
export type Klappwahl = Readonly<Record<string, boolean>>

// Stuecke hält je Liste, wie weit „Weitere zeigen" sie geöffnet hat —
// `pakete`, `d:` vor den Dateien eines Pakets, `b:` vor seinem Beiwerk.
export type Stuecke = Readonly<Record<string, number>>

export interface Zweigplan {
  zweig: Zweig
  offen: boolean

  // dateien und beiwerk sind, was gezeichnet wird. naechsteDatei und
  // naechstesBild sind der erste Eintrag hinter „Weitere zeigen" — der den
  // Fokus bekommt — und undefined, wenn alles dasteht.
  dateien: string[]
  naechsteDatei: string | undefined

  // beiwerkZeile sagt, dass das Beiwerk hinter einer eigenen Zeile steht.
  beiwerkZeile: boolean
  beiwerkOffen: boolean
  beiwerk: string[]
  naechstesBild: string | undefined
}

export interface Baumplan {
  zweige: Zweigplan[]
  naechstesPaket: string | undefined
}

interface Planangaben {
  gewaehlt: { paket: string; datei: string } | null
  klappwahl: Klappwahl
  stuecke: Stuecke
}

// stueck gibt, wie viele Einträge einer Liste gezeichnet werden — so viele,
// wie geöffnet sind, und wenigstens bis zur Wahl.
function stueck(gesamt: number, geoeffnet: number, wahlstelle: number): number {
  return Math.min(gesamt, Math.max(geoeffnet, wahlstelle + 1))
}

// zeilen zählt, was ein Zweig aufgeklappt zeichnete.
function zeilen(zweig: Zweig): number {
  return 1 + zweig.dateien.length + zweig.beiwerk.length
}

// plane rechnet, was der Baum zeichnet.
export function plane(
  zweige: readonly Zweig[],
  { gewaehlt, klappwahl, stuecke }: Planangaben,
): Baumplan {
  const gross =
    zweige.reduce((summe, zweig) => summe + zeilen(zweig), 0) > Faltgrenze
  const wahlpaket = zweige.findIndex((zweig) => zweig.paket === gewaehlt?.paket)
  const anzahl = stueck(zweige.length, stuecke.pakete ?? Stueck, wahlpaket)

  const geplant = zweige.slice(0, anzahl).map((zweig): Zweigplan => {
    const hier = zweig.paket === gewaehlt?.paket ? gewaehlt.datei : undefined
    const textwahl = hier === undefined ? -1 : zweig.dateien.indexOf(hier)
    const bildwahl = hier === undefined ? -1 : zweig.beiwerk.indexOf(hier)

    const offen =
      klappwahl[`p:${zweig.paket}`] ?? (!gross || hier !== undefined)
    const beiwerkZeile = gross && zweig.beiwerk.length > 0
    const beiwerkOffen =
      !beiwerkZeile || (klappwahl[`b:${zweig.paket}`] ?? bildwahl >= 0)

    const dateizahl = offen
      ? stueck(
          zweig.dateien.length,
          stuecke[`d:${zweig.paket}`] ?? Stueck,
          textwahl,
        )
      : 0
    const bildzahl =
      offen && beiwerkOffen
        ? stueck(
            zweig.beiwerk.length,
            stuecke[`b:${zweig.paket}`] ?? Stueck,
            bildwahl,
          )
        : 0

    return {
      zweig,
      offen,
      dateien: zweig.dateien.slice(0, dateizahl),
      naechsteDatei: offen ? zweig.dateien[dateizahl] : undefined,
      beiwerkZeile,
      beiwerkOffen,
      beiwerk: zweig.beiwerk.slice(0, bildzahl),
      naechstesBild:
        offen && beiwerkOffen ? zweig.beiwerk[bildzahl] : undefined,
    }
  })

  return { zweige: geplant, naechstesPaket: zweige[anzahl]?.paket }
}
