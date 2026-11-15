// **Ein Kasten, der zu ist, bis jemand ihn öffnet.**
//
// # Wozu
//
// Werkstück C von Phase 4.10
// (#1325).
// Auf der Kartenlasche standen zwei Werkzeugkästen — *Karten aus einem
// Skript* und *Autoersetzung* — **über** der Kartenliste und schoben sie mit
// 430 Punkten unter die Falz: gemessen am 2026-09-03 lagen dort **2 von 19**
// Textfeldern über 800 Punkten, und die erste Karte begann bei 755.
//
// `AGENTS.md` §6 sagt: *„Die Karte ist das Hauptobjekt."* Beides sind
// Werkzeuge, die man **einmal** einstellt; sie standen an der Stelle der
// Sache, die man **immer** braucht.
//
// # Warum ein eigenes Bauteil und kein `<details>`
//
// `<details>` bringt seinen eigenen Zustand mit und lässt sich von außen
// nicht schließen — beim Wechsel des Bezugsobjekts (§7, Regel 1) wäre das
// eine Zusage, die niemand einlösen kann. Und sein Dreieck ist in jedem
// Browser anders.
//
// **Gezeichnet wird er erst beim ersten Öffnen — danach wird er nur noch
// verborgen.** Zu Anfang steht nichts im Baum: ein Inhalt, der den Bestand
// fragt (`Gegenueberkasten`), fragt erst, wenn jemand ihn sehen will.
// Einmal gezeichnet, bleibt er stehen und bekommt beim Zuklappen `hidden` —
// nicht sichtbar, nicht im Tabweg, nicht im Zugänglichkeitsbaum.
//
// **Bis zum 2026-09-25 wurde er beim Zuklappen ausgehängt**, und mit ihm
// ging verloren, was darin eingegeben war: der Name eines halb angelegten
// Decks, die Kennung eines Gegenübers. UX-08 S11.NAT-03
// (#1680)
// sagt: *„Aufklappen verändert keine Eingaben."* **`aria-controls` steht
// weiter nur, solange er offen ist** — ein Verweis auf einen verborgenen
// Kasten zeigte auf etwas, das niemand sieht.
import { useEffect, useId, useRef, useState, type ReactNode } from 'react'

import { Knopf } from './Knopf'

interface KastenEigenschaften {
  // titel beschriftet den Schalter — er ist zugleich die Überschrift des
  // Kastens, denn zwei Namen für dieselbe Sache sind einer zu viel.
  titel: string

  // offenGehalten hält den Kasten offen, solange es stimmt — die Suche der
  // Einstellungen setzt es, wenn ein Treffer darin liegt (UX-04.AK-03,
  // #1676): „führt direkt zum Feld" hieße sonst, zu einem Feld zu führen,
  // das gar nicht gezeichnet ist.
  offenGehalten?: boolean | undefined

  // schreibt sperrt den Schalter für einen Leser — G10 (#2276): ein Kasten,
  // der nur ein Formular zum Anlegen hält, öffnet sich dann gar nicht erst,
  // statt eines zu zeigen, das am Ende nicht abzuschicken ist.
  schreibt?: boolean

  // aufschlagen ist ein Wunsch, den Kasten zu öffnen und seinem Schalter den
  // Fokus zu geben — **je neuem Wert einmal** (AP-10, #2367): wer ihn danach
  // schließt, schließt ihn.
  aufschlagen?: string | undefined

  children: ReactNode
}

export function Aufklappkasten({
  titel,
  offenGehalten = false,
  schreibt = false,
  aufschlagen,
  children,
}: KastenEigenschaften) {
  const [offenGeklickt, setzeOffen] = useState(false)

  // **Beim Zeichnen gesetzt**, wie `gezeichnet` unten — ein Effekt öffnete
  // ihn einen Rahmen zu spät, und der Fokus fände den Inhalt noch nicht.
  const [gesehen, setzeGesehen] = useState<string>()
  if (aufschlagen !== gesehen) {
    setzeGesehen(aufschlagen)
    if (aufschlagen !== undefined) setzeOffen(true)
  }
  const griff = useRef<HTMLButtonElement | null>(null)
  useEffect(() => {
    if (aufschlagen !== undefined) griff.current?.focus()
  }, [aufschlagen])
  const offen = offenGeklickt || offenGehalten

  // **Einmal offen, bleibt gezeichnet** — beim Zeichnen gesetzt und nicht in
  // einem Effekt, damit der Inhalt schon beim ersten Öffnen dasteht.
  const [gezeichnet, setzeGezeichnet] = useState(offen)
  if (offen && !gezeichnet) {
    setzeGezeichnet(true)
  }
  const kastenID = useId()

  return (
    <section className="flex flex-col gap-2">
      {/* **`steuert` steht nur, wenn der Kasten steht.** `Knopf` setzt
          `aria-controls` ohnehin nur bei `ausgeklappt`; die Kennung hier
          trotzdem leer zu lassen wäre eine zweite Wahrheit über denselben
          Zustand. */}
      <Knopf
        gewicht="leise"
        griff={griff}
        aufKlick={() => {
          setzeOffen((war) => !war)
        }}
        ausgeklappt={offen}
        steuert={kastenID}
        schreibt={schreibt}
      >
        {titel}
      </Knopf>

      {gezeichnet && (
        <div
          id={kastenID}
          hidden={!offen}
          className="rounded-sm border border-kante p-3"
        >
          {children}
        </div>
      )}
    </section>
  )
}
