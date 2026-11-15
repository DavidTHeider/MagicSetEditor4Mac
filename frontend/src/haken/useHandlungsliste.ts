// **Eine Handlungsliste je Objekt** — GESTE
// (#2358),
// Befund B-01; Artboard `Kontext09`.
//
// # Was es ist
//
// Die Seite baut die Liste **einmal** — als `Menueeintrag[]`, wie sie schon
// das sichtbare `Menue` und das `Kontextmenue` nehmen — und gibt sie diesem
// Haken. Er gibt sie unverändert zurück, für den Überlauf an der Zeile und
// den Rechtsklick samt ⇧F10, und meldet sie dem `Objekthandlungsraum`, der
// sie in die Menüleiste bringt. **Eine Liste, drei Orte, keine zweite
// fachliche Liste** — das ist die Abnahme.
//
// # Gesperrt
//
// Was die Liste im Menü grau und erklärt zeigt (`Menueliste`), steht in der
// Leiste grau: **angehalten**, weil sein Vorgang läuft oder es nichts zu tun
// gibt, und **schreibend für einen Leser** (#2288). Einen Grund kann ein
// Eintrag der Leiste nicht tragen; das Kontextmenü trägt ihn.
//
// # Die Epoche
//
// **Zählt mit, sobald sich Objekt oder Liste ändern.** Ein Klick in der Leiste
// trägt die Epoche, mit der sie gebaut wurde; gehört er zu einer anderen,
// bewirkt er nichts (§7, Regel 2) — die Karte ist inzwischen eine andere,
// oder der Eintrag ist es.
import { use, useEffect, useId, useRef, useState } from 'react'

import type { Menueeintrag } from '../bausteine/Menue'
import { ObjekthandlungsAnmeldung } from '../objekthandlungsraum'
import type { Objekthandlung } from '../objekthandlungen'
import { useNurLesend } from './schreibrecht'

export interface Handlungslistenangaben {
  // titel steht über dem Menü der Leiste — aus dem Katalog (§9).
  titel: string
  // art und kennung nennen das Objekt: `karte` und ihre Kennung.
  art: string
  kennung: string
  eintraege: readonly Menueeintrag[]
}

// gesperrtFuer sagt, ob ein Eintrag hier nichts tun darf.
function gesperrtFuer(eintrag: Menueeintrag, nurLesend: boolean): boolean {
  return eintrag.angehalten === true || (eintrag.schreibt === true && nurLesend)
}

export function useHandlungsliste({
  titel,
  art,
  kennung,
  eintraege,
}: Handlungslistenangaben): readonly Menueeintrag[] {
  const anmelden = use(ObjekthandlungsAnmeldung)
  const nurLesend = useNurLesend()
  const platz = useId()

  const nativ: Objekthandlung[] = eintraege.map((eintrag) => ({
    schluessel: eintrag.schluessel,
    name: eintrag.name,
    gesperrt: gesperrtFuer(eintrag, nurLesend),
    ...(eintrag.kuerzel === undefined ? {} : { kuerzel: eintrag.kuerzel }),
  }))
  const gestalt = JSON.stringify([titel, art, kennung, nativ])

  // **Die Epoche als abgeleiteter Zustand** — gezählt im Zeichnen, wenn sich
  // die Gestalt geändert hat, wie React es für Werte aus dem vorigen Zeichnen
  // vorsieht; ohne Effekt, der einmal zu spät zählte.
  const [stand, setzeStand] = useState({ gestalt, epoche: 1 })
  if (stand.gestalt !== gestalt) {
    setzeStand({ gestalt, epoche: stand.epoche + 1 })
  }
  const epoche = stand.gestalt === gestalt ? stand.epoche : stand.epoche + 1

  // **Die Einträge zur Zeit des Klicks**, nicht die des Anmeldens: ihre
  // Rückrufe ändern sich mit jedem Zeichnen, ihre Gestalt nicht.
  const heute = useRef({ eintraege, nurLesend })
  useEffect(() => {
    heute.current = { eintraege, nurLesend }
  })

  useEffect(() => {
    // **Aus der Gestalt gelesen**: sie ist, woran die Anmeldung hängt, und
    // nur bei neuer Gestalt oder Epoche wird neu angemeldet.
    const [neuerTitel, neueArt, neueKennung, neueEintraege] = JSON.parse(
      gestalt,
    ) as [string, string, string, Objekthandlung[]]
    anmelden?.(platz, {
      eintraege: heute.current.eintraege,
      meldung: {
        titel: neuerTitel,
        art: neueArt,
        kennung: neueKennung,
        epoche,
        eintraege: neueEintraege,
      },
      ausloesen: (schluessel) => {
        const { eintraege: jetzt, nurLesend: leser } = heute.current
        const eintrag = jetzt.find(
          (kandidat) => kandidat.schluessel === schluessel,
        )
        if (
          eintrag === undefined ||
          !('aufWahl' in eintrag) ||
          gesperrtFuer(eintrag, leser)
        ) {
          return
        }
        eintrag.aufWahl()
      },
    })
  }, [anmelden, platz, gestalt, epoche])

  useEffect(
    () => () => {
      anmelden?.(platz, undefined)
    },
    [anmelden, platz],
  )

  return eintraege
}
