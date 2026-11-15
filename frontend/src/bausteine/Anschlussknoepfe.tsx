// **Wohin ein Exemplar weiterreisen kann** — UX-03.AK-01
// (#1675).
//
// # Wo sie stehen, und warum dort
//
// Der Entwurf zeichnet sie **unten im Inspector der Sammlung**, nebeneinander
// und an den Fuß gedrückt (`docs/prototyp/project/Sammlung.dc.html`,
// `margin-top: auto`): *Zum Tausch* und *Ins Deck*. Sie stehen am Fuß, weil
// sie die Fläche **verlassen** — darüber steht, was man am Exemplar selbst
// tut.
//
// **`mt-auto` steht hier nicht, und das ist eine Abweichung mit Grund.** Der
// Inspector des `Arbeitsplatz` ist `items-start` und damit so hoch wie sein
// Inhalt; ein `margin-top: auto` fände dort keinen Platz zum Wachsen und
// täte nichts. Was den Fuß kenntlich macht, ist stattdessen eine Trennlinie
// — und die hält 3:1 in beiden Erscheinungsbildern (`docs/KONTRASTE.md`).
//
// # Warum ein eigenes Bauteil
//
// **Weil der Inspector bei 301 Zeilen steht** (§2) und weil die Frage eine
// andere ist: dort wird ein Exemplar geändert, hier wird es weitergereicht.
//
// # Und warum der Zustand unverändert mitreist
//
// **Weil `Handelsposition.Guete` ihn nicht übersetzt.** Nachgelesen am
// 2026-09-21: kein Dienst und keine Naht vergleicht das Feld gegen eine
// Aufzählung — `internal/app/handelsvorgang.go` reicht es durch,
// `internal/mcpnaehte/handelsstand.go` ebenso. Eine Umrechnungstabelle
// `nearMint` → `NM` wäre eine Zuordnung ohne Quelle (§11), und beides ist
// dieselbe Cardmarket-Skala.
import { insDeck, zumTausch } from './exemplaruebergabe'
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { useTexte } from '../sprache/kontext'

import type { ExemplarDTO } from '../api'

export function Anschlussknoepfe({ exemplar }: { exemplar: ExemplarDTO }) {
  const eigene = useTexte().seiten.sammlung
  const anschluss = useAnschluss()

  return (
    <div className="mt-2 flex flex-wrap gap-2 border-t border-kante pt-3">
      <Knopf
        aufKlick={() => {
          anschluss.uebergib(zumTausch(exemplar))
        }}
      >
        {eigene.zumTausch}
      </Knopf>

      <Knopf
        aufKlick={() => {
          anschluss.uebergib(insDeck(exemplar))
        }}
      >
        {eigene.insDeck}
      </Knopf>
    </div>
  )
}
