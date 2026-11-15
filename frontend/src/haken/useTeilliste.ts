// **Gruppen und Reihenfolge in der Teilliste** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// # Warum ein Haken
//
// Dieselbe Begründung wie bei `useSymmetrie` daneben: `Symbol.tsx` steht an
// der Zeilengrenze aus §2, und der Schnitt folgt der Frage — *was* mit der
// Teilliste geschieht, steht hier; *wie* sie aussieht, in
// `Symbolteilliste.tsx`.
//
// # Die Regeln kommen von der Go-Seite
//
// Wo die Gruppe landet, in welcher Reihenfolge ihre Kinder stehen und wie die
// Zielstelle beim Verschieben zählt — das sind drei Rechnungen von MSE, und
// sie stehen **einmal** in `internal/service/symbolteile.go`. Eine zweite
// Fassung hier liefe mit jener auseinander.

import { useCallback, useState } from 'react'

import type { SymbolAntwort, SymbolteilDTO } from '../antworten'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { gewaehltNach, nachEinemWegfall, type Wahl } from '../seiten/teilwahl'
import type { Handgriff } from './useZeichenverlauf'
import type { Zeichenschutz } from './useZeichenschutz'

interface TeillisteEigenschaften {
  teile: SymbolteilDTO[]
  eintragen: (teile: SymbolteilDTO[], handgriff: Handgriff) => void
  klagen: (ursache: unknown) => void

  // schutz hält §7 Regeln 2 und 4 über der Zeichnung — Befund 2 des
  // Zweitblicks über Phase 4.11. Die Begründung steht am Haken.
  schutz: Zeichenschutz
}

export interface Teillistenwege {
  // wahl sind die gewählten Stellen, in Listenreihenfolge — nie leer.
  wahl: Wahl

  // gewaehlt ist die **eine** Stelle, an der die Zeichenfläche arbeitet.
  //
  // **Die erste der Wahl**, so wie MSEs `getAShape()` eine aus der Menge
  // greift (`src/gui/symbol/selection.cpp`, Zeilen 66–71, Fassung `f9d9356`
  // vom 2021-07-06). Griffe und Symmetrieanfasser gehören zu **einem** Teil;
  // sie über eine Mehrfachwahl zu legen, hieße zu raten, welches gemeint ist.
  gewaehlt: number

  waehlen: (stelle: number, hinzu: boolean) => void

  // nachWegfall rückt die Wahl nach, wenn ein Teil gelöscht wurde.
  nachWegfall: (stelle: number) => void

  gruppieren: () => void
  aufloesen: () => void
  verschieben: (von: number, nach: number) => void

  // kannGruppieren und kannAufloesen sagen, ob der Weg heute offensteht.
  //
  // **MSE graut seine Menüeinträge aus** (`src/gui/symbol/select_editor.cpp`,
  // Zeilen 155–164, Fassung `f9d9356` vom 2021-07-06); dieselbe Auskunft, an
  // derselben Stelle entschieden. Die Go-Seite prüft es **noch einmal** — sie
  // muss, denn eine Bindung hat keine Menüleiste, die sich ausgrauen ließe.
  kannGruppieren: boolean
  kannAufloesen: boolean
}

export function useTeilliste({
  teile,
  eintragen,
  klagen,
  schutz,
}: TeillisteEigenschaften): Teillistenwege {
  const [wahl, setzeWahl] = useState<Wahl>([0])

  const waehlen = useCallback((stelle: number, hinzu: boolean) => {
    setzeWahl((bisher) => gewaehltNach(bisher, stelle, hinzu))
  }, [])

  const nachWegfall = useCallback((stelle: number) => {
    setzeWahl((bisher) => nachEinemWegfall(bisher, stelle))
  }, [])

  // aendern ruft eine der drei Bindungen und trägt das Ergebnis ein.
  //
  // **Ein Fehler räumt die Zeichnung nicht** (§7, Regel 3) — dieselbe Zusage
  // wie bei der Symmetrie, und die Go-Seite hält sie mit: sie gibt die Liste
  // von vorher zurück.
  const aendern = useCallback(
    (ruf: () => Promise<SymbolAntwort>, danach: Wahl, marke: string) => {
      void schutz.starte(async (nochAktuell) => {
        try {
          const antwort = await ruf()
          const neue = auspacken(antwort, () => antwort.teile)

          // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
          if (!nochAktuell()) {
            return
          }

          eintragen(neue, { marke })
          setzeWahl(danach)
        } catch (ursache: unknown) {
          klagen(ursache)
        }
      })
    },
    [eintragen, klagen, schutz],
  )

  // **Nach dem Gruppieren steht die Gruppe da, wo das erste gewählte Teil
  // stand** — und sie ist danach gewählt. Wer gruppiert, will als Nächstes
  // etwas mit der Gruppe tun; eine Wahl, die auf die alten Stellen zeigte,
  // zeigte nach dem Zusammenrücken auf fremde Teile.
  const gruppieren = useCallback(() => {
    aendern(
      () => api().TeileGruppieren(teile, [...wahl]),
      [wahl[0]],
      'gruppieren',
    )
  }, [aendern, teile, wahl])

  // **Nach dem Auflösen ist das erste Kind gewählt.** Es steht an der Stelle,
  // an der die Gruppe stand.
  const aufloesen = useCallback(() => {
    aendern(
      () => api().GruppeAufloesen(teile, [...wahl]),
      [wahl[0]],
      'aufloesen',
    )
  }, [aendern, teile, wahl])

  // **Nach dem Verschieben ist das bewegte Teil gewählt**, an seiner neuen
  // Stelle — sonst spränge die Wahl auf das, was nachgerückt ist.
  const verschieben = useCallback(
    (von: number, nach: number) => {
      aendern(
        () => api().TeilVerschieben(teile, von, nach),
        [nach],
        `verschieben:${String(von)}:${String(nach)}`,
      )
    },
    [aendern, teile],
  )

  return {
    wahl,
    gewaehlt: wahl[0],
    waehlen,
    nachWegfall,
    gruppieren,
    aufloesen,
    verschieben,
    kannGruppieren: wahl.length >= 2,
    // **Eine Gruppe in der Wahl genügt** — so hält es MSE:
    // `if (p->isSymbolGroup() && !p->isSymbolSymmetry()) { ev.Enable(true);
    // return; }` (`src/gui/symbol/select_editor.cpp`, Zeilen 157–164, Fassung
    // `f9d9356` vom 2021-07-06). Was gewählt ist und keine Gruppe, bleibt beim
    // Auflösen stehen.
    //
    // **Nicht `every`**: dann schaltete Gruppe plus Form den Knopf ab, und
    // der Dienst, der genau dieses Verhalten trägt, wäre produktiv **nicht
    // erreichbar**.
    kannAufloesen: wahl.some((wo) => teile[wo]?.art === 'Gruppe'),
  }
}
