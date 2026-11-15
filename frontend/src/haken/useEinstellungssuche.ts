// **Was die Suche der Einstellungsseite tut** — UX-04.AK-03
// (#1676).
//
// # Warum ein Haken und nicht die Seite
//
// **Weil `Einstellungen.tsx` mit der Trefferanzeige über die Zeilengrenze
// aus §2 ging** — 478 Zeilen. Der Schnitt folgt der Frage: dort steht, **wie
// die Seite aussieht** — neun Gruppen, was in welcher liegt, was daneben
// steht —, hier, **was die Suche damit macht**: wer getroffen ist, wo das
// liegt, wohin die Liste springt und wohin der Fokus.
//
// # Was er nicht tut
//
// **Er vergleicht nicht selbst.** Wonach verglichen wird, steht in
// [einstellungstreffer.ts](../bausteine/einstellungstreffer.ts) — dieselbe
// Regel benutzt auch, wer nur zählen will.
import { useEffect, useState } from 'react'

import { passtZurFrage } from '../bausteine/einstellungstreffer'

import type { EinstellungDTO } from '../api'
import type { Einstellungsgruppe as Gruppe } from '../seiten/einstellungsgruppen'
import type { Katalog } from '../sprache/kataloge'
import { useSprache } from '../sprache/kontext'

export interface Suchstand {
  // trefferIn gibt die getroffenen Schlüssel einer Gruppe — nie `undefined`.
  trefferIn: (gruppe: Gruppe) => ReadonlySet<string>

  // ersteMitTreffer ist die erste Gruppe, in der etwas liegt.
  ersteMitTreffer: Gruppe | undefined

  // aufSprung läuft bei der Eingabetaste und setzt den Fokus auf das erste
  // getroffene Feld.
  aufSprung: () => void
}

export function useEinstellungssuche({
  gezeigte,
  frage,
  texte,
  gruppen,
  gruppeFuer,
  offeneGruppe,
  zumGebiet,
}: {
  gezeigte: readonly EinstellungDTO[]
  frage: string
  texte: Katalog
  gruppen: readonly Gruppe[]
  gruppeFuer: (schluessel: string) => Gruppe
  offeneGruppe: Gruppe
  zumGebiet: (schluessel: string) => void
}): Suchstand {
  const sprache = useSprache()

  // **Ein Wunsch zu springen.** Die Eingabetaste zählt ihn hoch, der Effekt
  // unten setzt den Fokus.
  //
  // **Eine Zahl und kein Schalter**: wer zweimal hintereinander die
  // Eingabetaste drückt, meint es zweimal — und ein Schalter, den der Effekt
  // zurücknähme, wäre ein `setState` im Effekt.
  const [sprungwunsch, setzeSprungwunsch] = useState(0)

  // **Die zuletzt gesehene Frage** — daran hängt, ob die Liste umschaltet.
  const [letzteFrage, setzeLetzteFrage] = useState('')

  // **Ein Verbund und keine `Map`**: eine `Map` gäbe bei `get` ein
  // `undefined`, das es hier nie gibt — und ein `?? new Set()` daneben wäre
  // ein Zweig, den kein Aufruf erreicht (§3).
  const trefferJeGruppe = Object.fromEntries(
    gruppen.map((gruppe) => [gruppe, new Set<string>()]),
  ) as Record<Gruppe, Set<string>>

  for (const eintrag of gezeigte) {
    if (passtZurFrage(eintrag, frage, texte, sprache)) {
      trefferJeGruppe[gruppeFuer(eintrag.schluessel)].add(eintrag.schluessel)
    }
  }

  const trefferIn = (gruppe: Gruppe) => trefferJeGruppe[gruppe]

  // **Wo der erste Treffer liegt** — dorthin springt die Gruppenliste, wenn
  // die offene Gruppe keinen hat, und dorthin führt die Eingabetaste.
  const ersteMitTreffer = gruppen.find((gruppe) => trefferIn(gruppe).size > 0)

  // **Die Gruppe kommt mit.** Wer „proxy" tippt, soll nicht raten müssen, wo
  // der Treffer liegt.
  //
  // **Ohne Fokusraub**: der Wunsch öffnet nur; wer tippt, tippt weiter. Erst
  // die Eingabetaste setzt den Fokus (§6 — jede Aktion ist per Tastatur
  // erreichbar, aber keine nimmt sie einem weg).
  //
  // **Beim Zeichnen und nicht im Effekt.** React nennt das „Zustand während
  // des Zeichnens anpassen"; dieselbe Gestalt steht in
  // [useGebietswahl](./useGebietswahl.ts). Ein Effekt, der `setState` ruft,
  // zeichnete die Seite ein zweites Mal — und der Linter weist ihn zu Recht
  // ab.
  if (frage !== letzteFrage) {
    setzeLetzteFrage(frage)

    if (ersteMitTreffer !== undefined && trefferIn(offeneGruppe).size === 0) {
      zumGebiet(ersteMitTreffer)
    }
  }

  // **Und der Sprung auf das Feld** — eine Zeichnung, nachdem die Gruppe
  // offen ist, sonst gibt es das Feld noch nicht.
  //
  // **Über den DOM und nicht über eine Referenz**: das Feld kann in einem
  // Bauteil stehen, das die Seite gar nicht kennt — ein `Dienstschalter`
  // etwa. Was sie kennt, ist der Schlüssel, und den trägt jedes
  // Bedienelement als `data-einstellung` (#1768).
  useEffect(() => {
    if (sprungwunsch === 0 || ersteMitTreffer === undefined) {
      return
    }

    // **Eine Schleife mit `break` und kein `[0] ?? ''`**: die Menge ist nicht
    // leer — das entscheidet `ersteMitTreffer` —, und ein Rückfall daneben
    // wäre ein Zweig, den kein Aufruf betritt (§3).
    for (const erster of trefferIn(ersteMitTreffer)) {
      const feld = document.querySelector<HTMLElement>(
        `[data-einstellung="${erster}"]`,
      )

      // **Eine Segmentreihe ist selbst keine Tabstelle** — UXR-F2
      // (#2008):
      // ihr Schlüssel steht an der Gruppe, und eine Gruppe nimmt den Fokus
      // nicht an. Er gehört dem gewählten Segment — oder dem ersten, wo
      // keines gewählt ist —, denn das ist die eine Tabstelle der Reihe.
      const ziel =
        feld?.getAttribute('role') === 'radiogroup'
          ? (feld.querySelector<HTMLElement>('input:checked') ??
            feld.querySelector<HTMLElement>('input'))
          : feld

      ziel?.focus()

      break
    }
    // **Nur am Wunsch**: die Trefferliste entsteht bei jeder Zeichnung neu,
    // und mit ihr in der Liste spränge der Fokus bei jedem Anschlag.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sprungwunsch])

  return {
    trefferIn,
    ersteMitTreffer,
    aufSprung: () => {
      setzeSprungwunsch(sprungwunsch + 1)
    },
  }
}
