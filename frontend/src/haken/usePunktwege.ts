// **Was mit den Punkten einer Form geschieht** — verschieben, setzen,
// einfügen, löschen, und die Bézierhandgriffe ziehen.
//
// # Warum ein Haken
//
// Dieselbe Begründung wie bei `useSymmetrie` und `useTeilliste` daneben:
// `Symbol.tsx` steht an der Zeilengrenze aus §2, und der Schnitt folgt der
// Frage — *was* mit einem Punkt geschieht, steht hier; *wie* er aussieht, auf
// der Zeichenfläche.

import { useCallback, useRef } from 'react'

import type { Handgriff } from './useZeichenverlauf'
import type { Zeichenschutz } from './useZeichenschutz'
import type {
  SymbolAntwort,
  SymbolpunktweiseDTO,
  SymbolteilDTO,
} from '../antworten'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { gesetzt, verschoben } from '../seiten/symbolformen'
import { useLauf } from './useLauf'

interface PunktwegeEigenschaften {
  teile: SymbolteilDTO[]
  eintragen: (teile: SymbolteilDTO[], handgriff: Handgriff) => void
  klagen: (ursache: unknown) => void

  // schutz hält §7 Regeln 2 und 4 über der Zeichnung (Befund 2 des
  // Zweitblicks über Phase 4.11).
  //
  // **`griffAuf` behält zusätzlich seinen eigenen `useLauf`.** Ein Zug
  // schickt eine Anfrage je Bewegung; die dürfen einander überholen, und der
  // Wächter hier ließe nur die erste durch. Dort geht es um die Reihenfolge
  // **innerhalb** eines Zuges, hier um zwei Handlungen nebeneinander.
  schutz: Zeichenschutz
}

export interface Punktwege {
  // zugBeginnt sagt an, dass ein neuer Zug anfängt — MSEs `resetActions()`
  // (`src/gui/symbol/point_editor.cpp` Zeile 358, Fassung `f9d9356` vom
  // 2021-07-06).
  //
  // **Ohne diesen Ruf wären zwei Züge an demselben Punkt ein Schritt.** Die
  // Marke eines Zuges trägt den Punkt, damit die vielen Zeigerbewegungen eines
  // Zuges zusammenfallen; was sie **trennt**, ist diese Zählung. In MSE
  // erledigt das die Lebensdauer der Handlung, hier eine Zahl.
  zugBeginnt: () => void

  verschieben: (
    teilstelle: number,
    punktstelle: number,
    dx: number,
    dy: number,
  ) => void
  setzenAuf: (
    teilstelle: number,
    punktstelle: number,
    x: number,
    y: number,
  ) => void
  griffAuf: (
    teil: number,
    punkt: number,
    davor: boolean,
    x: number,
    y: number,
  ) => void
  punktEinfuegen: (teil: number, punkt: number) => Promise<void>
  punktLoeschen: (teil: number, punkt: number) => Promise<void>
}

export function usePunktwege({
  teile,
  eintragen,
  klagen,
  schutz,
}: PunktwegeEigenschaften): Punktwege {
  const lauf = useLauf()

  // zug zählt die Züge. Er ist ein Verweis und kein Zustand: er ändert nichts
  // am Bild, und ein Zustand ließe die Fläche bei jedem Aufsetzen neu zeichnen.
  const zug = useRef(0)

  const zugBeginnt = useCallback(() => {
    zug.current += 1
  }, [])

  // verschieben bewegt einen Punkt des gewählten Teils.
  const verschieben = useCallback(
    (teilstelle: number, punktstelle: number, dx: number, dy: number) => {
      eintragen(
        teile.map((teil, wo) =>
          wo === teilstelle ? verschoben(teil, punktstelle, dx, dy) : teil,
        ),
        // **Nicht fortschreibbar** — `point_editor.cpp` Zeilen 355–380 legt je
        // Tastendruck eine neue Handlung an, und `ControlPointMoveAction`
        // kennt kein `merge`. Wer viermal die Pfeiltaste drückt, bekommt vier
        // Schritte, und genau das tut MSE auch.
        { marke: `punkt:${String(teilstelle)}:${String(punktstelle)}` },
      )
    },
    [teile, eintragen],
  )

  // setzenAuf stellt einen Punkt **auf** eine Stelle, statt ihn zu verschieben.
  //
  // **Der Zeiger führt, nicht eine Differenz.** `verschieben` bleibt für die
  // Tastatur: dort ist ein Schritt genau das, was gemeint ist. Am Zeiger ist
  // die Stelle gemeint, und wer sie aus einem gemerkten Anfangswert
  // ausrechnet, rechnet ab der zweiten Bewegung falsch (#890).
  const setzenAuf = useCallback(
    (teilstelle: number, punktstelle: number, x: number, y: number) => {
      eintragen(
        teile.map((teil, wo) =>
          wo === teilstelle ? gesetzt(teil, punktstelle, x, y) : teil,
        ),
        // **Fortschreibbar** — `ExtendableAction`: die vielen Bewegungen eines
        // Zuges sind **ein** Schritt. Die Zugnummer trennt zwei Züge an
        // demselben Punkt.
        { fortschreibbar: true, marke: `zug:${String(zug.current)}` },
      )
    },
    [teile, eintragen],
  )

  // punktaendern ruft eine der beiden Punktbindungen und trägt das Ergebnis
  // ein.
  //
  // **Ein Fehler räumt die Zeichnung nicht.** Wer den letzten Punkt einer Form
  // löschen will, bekommt eine Meldung — und behält, was er hat (§7, Regel 3).
  const punktaendern = useCallback(
    (
      ruf: (
        teile: SymbolteilDTO[],
        weise: SymbolpunktweiseDTO,
      ) => Promise<SymbolAntwort>,
      teil: number,
      punkt: number,
    ) =>
      schutz.starte(async (nochAktuell) => {
        try {
          const antwort = await ruf(teile, { teil, punkt })
          const neue = auspacken(antwort, () => antwort.teile)

          // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
          if (!nochAktuell()) {
            return
          }

          eintragen(neue, {
            marke: `punktzahl:${String(teil)}:${String(punkt)}`,
          })
        } catch (ursache: unknown) {
          klagen(ursache)
        }
      }),
    [teile, eintragen, klagen, schutz],
  )

  // griffAuf zieht einen Handgriff über die Go-Seite auf eine Stelle.
  //
  // **Dort, weil die Sperre dort steht**: `direction` behält die Länge des
  // anderen Griffs, `size` spiegelt ihn. Eine zweite Rechnung hier liefe mit
  // jener auseinander.
  //
  // **Und die Antwort wird gegen den Lauf gehalten** (§7, Regel 2). Ein Zug
  // schickt eine Anfrage je Bewegung; sie kommen über dieselbe Naht zurück,
  // aber nicht zwingend in der Reihenfolge, in der sie hinausgingen. Eine
  // überholte Antwort trüge den Griff dorthin zurück, wo er vor zwei
  // Bewegungen stand — und der Zeiger stünde woanders.
  const griffAuf = useCallback(
    (teil: number, punkt: number, davor: boolean, x: number, y: number) => {
      const nummer = lauf.naechster()

      // **Und der gemeinsame Stand dazu** (Nachlauf über Phase 4.11).
      //
      // Der eigene Lauf ordnet die Antworten **innerhalb** eines Zuges; er
      // weiß nichts davon, dass inzwischen ein anderes Symbol offen ist oder
      // eine andere Umformung gelaufen ist. Eine verspätete Griffantwort galt
      // in ihrem privaten Lauf weiter als aktuell und überschrieb den
      // neueren Gesamtstand.
      //
      // **Nicht unter `starte`**: ein Zug schickt eine Anfrage je Bewegung,
      // und dort käme nur die erste durch — der Griff bliebe kleben.
      const nochAktuell = schutz.beobachte()

      void (async () => {
        try {
          const antwort = await api().GriffZiehen(teile, {
            teil,
            punkt,
            davor,
            x,
            y,
          })

          const neue = auspacken(antwort, () => antwort.teile)
          if (!lauf.istAktuell(nummer) || !nochAktuell()) {
            return
          }

          eintragen(neue, {
            fortschreibbar: true,
            marke: `griff:${String(zug.current)}`,
          })
        } catch (ursache: unknown) {
          if (lauf.istAktuell(nummer) && nochAktuell()) {
            klagen(ursache)
          }
        }
      })()
    },
    [teile, lauf, schutz, eintragen, klagen],
  )

  // **Sie geben ihren Lauf zurück**, seit Befund 2 des Zweitblicks über Phase
  // 4.11: ein Lauf, der auf das Ende warten will, hätte sonst nichts, worauf
  // er wartet. Die Oberfläche ruft sie weiter ohne `await`.
  const punktEinfuegen = useCallback(
    (teil: number, punkt: number) =>
      punktaendern(api().PunktEinfuegen, teil, punkt),
    [punktaendern],
  )

  const punktLoeschen = useCallback(
    (teil: number, punkt: number) =>
      punktaendern(api().PunktLoeschen, teil, punkt),
    [punktaendern],
  )

  return {
    zugBeginnt,
    verschieben,
    setzenAuf,
    griffAuf,
    punktEinfuegen,
    punktLoeschen,
  }
}
