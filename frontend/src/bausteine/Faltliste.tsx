// **Eine lange Liste, die vorn zeigt und den Rest stückweise** — UXR-08
// (#1916).
//
// # Wofür
//
// **Für Listen, deren Länge der Bestand bestimmt und nicht die Seite.** Die
// Verteilung eines Felds hat so viele Balken, wie das Feld Werte hat:
// gemessen an `White.mse-set` (3.761 Karten, aus der Setdatei gezählt am
// 2026-10-02) sind das für `name` **3.756**, für `illustrator` 586, für
// `sub type` 378. Ein Turnierpack aus `magic.mse-game` zieht **76** Karten.
// Alle stünden untereinander in einer Schublade von 460 Punkten Breite — R4
// aus `docs/UIREVIEW.md` erlaubt eine Rolltiefe von 2,5.
//
// # Was es zusagt
//
//   - **Vorn steht, was am meisten zählt** — die Reihenfolge bestimmt der
//     Aufrufer; Belegung, Verteilung und Auszählung kommen absteigend.
//   - **Gedeckelt wird sichtbar** (§10, wie bei der `Namensliste`): neben dem
//     Knopf steht, wie viele von wie vielen dastehen. Eine Liste, die einfach
//     aufhört, liest sich wie eine vollständige.
//   - **Stückweise und nicht alles auf einmal**, wie die `Listenansicht`: ein
//     Klick zeigt die nächsten, höchstens `schritt`.
//   - **Der Fokus geht nicht verloren.** Solange es weitere gibt, bleibt er
//     auf dem Knopf; verschwindet der Knopf mit dem letzten Stück, geht er auf
//     die Zahl daneben — sonst stünde er auf dem Dokument, und der nächste
//     Tabulator begänne oben.
//
// **Die Worte sind die der `Listenansicht`** — „Weitere {anzahl} zeigen" und
// „{gezeigt} von {gesamt}": derselbe Satz für dieselbe Sache.
import { useRef, useState, type ReactNode } from 'react'

import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface FaltlisteEigenschaften<T> {
  eintraege: readonly T[]

  // vorn ist, wie viele zu Anfang dastehen.
  vorn: number

  // schritt ist, wie viele ein Klick dazu zeigt.
  schritt: number

  // zeile zeichnet einen Eintrag — als `li` mit Schlüssel.
  zeile: (eintrag: T, stelle: number) => ReactNode

  // geordnet macht aus der Liste eine nummerierte (`ol`).
  geordnet?: boolean

  // immer nennt Einträge, die auch hinter der Falte dastehen — etwa die
  // offene Zeile einer Liste, die sonst mit ihrem Inhalt verschwände (#2035).
  // Sie stehen an ihrer Stelle in der Reihenfolge und zählen als gezeigt.
  immer?: ((eintrag: T) => boolean) | undefined

  // bezeichnung nennt die Liste für ein Vorleseprogramm — `aria-label`,
  // wo keine Überschrift sie benennt (das Journal im Kasten, UXR-15).
  bezeichnung?: string | undefined

  klasse: string
}

export function Faltliste<T>({
  eintraege,
  vorn,
  schritt,
  zeile,
  geordnet = false,
  immer,
  bezeichnung,
  klasse,
}: FaltlisteEigenschaften<T>) {
  const [zahl, setzeZahl] = useState(vorn)

  const gezeigt = eintraege.filter(
    (eintrag, stelle) => stelle < zahl || immer?.(eintrag) === true,
  )
  const Liste = geordnet ? 'ol' : 'ul'

  return (
    <>
      <Liste aria-label={bezeichnung} className={klasse}>
        {gezeigt.map(zeile)}
      </Liste>

      {eintraege.length > vorn && (
        <Faltfuss
          gezeigt={gezeigt.length}
          gesamt={eintraege.length}
          schritt={schritt}
          weiter={() => {
            setzeZahl(zahl + schritt)
          }}
        />
      )}
    </>
  )
}

// Faltfuss steht unter einer gefalteten Liste: wie viele von wie vielen, und
// der Knopf für die nächsten.
//
// **Eigens ausgeführt** (UXR-06, #1914), weil der Verlauf der Kartenseite
// dieselbe Falte braucht und eine Tabelle ist, keine Liste
// (`Verlaufsliste.tsx`). Derselbe Satz für dieselbe Sache — und dieselbe
// Zusage für den Fokus.
export function Faltfuss({
  gezeigt,
  gesamt,
  schritt,
  weiter,
}: {
  gezeigt: number
  gesamt: number
  schritt: number
  weiter: () => void
}) {
  const worte = useTexte().bausteine.listenansicht
  const stand = useRef<HTMLParagraphElement>(null)

  const rest = gesamt - gezeigt

  return (
    <div className="mt-2 flex flex-wrap items-center gap-3">
      <p ref={stand} tabIndex={-1} role="status" className="text-sm text-leise">
        {einsetzen(worte.zahl, {
          gezeigt: String(gezeigt),
          gesamt: String(gesamt),
        })}
      </p>

      {rest > 0 && (
        <Knopf
          aufKlick={() => {
            if (rest <= schritt) {
              stand.current?.focus()
            }
            weiter()
          }}
        >
          {einsetzen(worte.weitere, {
            anzahl: String(Math.min(schritt, rest)),
          })}
        </Knopf>
      )}
    </div>
  )
}
