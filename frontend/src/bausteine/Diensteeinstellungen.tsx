// **Der Bereich »Fremde Kartendienste«** — abgetrennt am 2026-09-11, weil
// `Einstellungen.tsx` mit dem Modellkatalog über die Zeilengrenze aus §2 ging
// (A6, #1445).
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **welche Bereiche** die Seite hat und wie sie geladen werden;
// hier steht, **was in diesem einen** steht: darf gefragt werden und wer
// antwortet worauf. **Unter welchen Bedingungen** — der Katalog — steht seit
// #2465 in einem Abschnitt darunter (`seiten/einstellungsbeiwerk.tsx`), und
// was es kostet und was ein Modell kann, in der Gruppe *KI*.
import { Anbieterstand } from './Anbieterstand'
import { Dienstfragen } from './Dienstfragen'
import { Dienstschalter } from './Dienstschalter'
import { useTexte } from '../sprache/kontext'
import { KILaufzeit } from './KILaufzeit'

import type { EinstellungDTO } from '../api'

interface DiensteeinstellungenEigenschaften {
  dienste: EinstellungDTO[]
  fragen: EinstellungDTO[]

  // ki sind die zwei Zeilen der lokalen Laufzeit — A7, #1446.
  ki: EinstellungDTO[]
  angehalten: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void
}

export function Diensteeinstellungen({
  dienste,
  fragen,
  angehalten,
  setze,
}: Omit<DiensteeinstellungenEigenschaften, 'ki'>) {
  return (
    <div className="flex flex-col gap-6">
      <Dienstschalter dienste={dienste} angehalten={angehalten} setze={setze} />

      <Dienstfragen
        fragen={fragen}
        dienste={dienste}
        angehalten={angehalten}
        setze={setze}
      />
    </div>
  )
}

// KIeinstellungen ist, was die KI angeht — UX-04.AK-04
// (#1676).
//
// # Warum sie aus den Diensten herausgeschnitten sind
//
// **Weil das Kriterium es verlangt**: „KI erhält einen eigenen eindeutig
// benannten Bereich." Unter *Fremde Kartendienste* stünden Kontingent,
// Laufzeit und Modelle hinter einer Aufschrift, die sie nicht nennt, und dort
// sucht sie niemand.
//
// **Die Schalter der Anbieter reisen mit** (`dienst.anthropic` und die
// zwei daneben): sie heißen `dienst.`, aber sie schalten ein Sprachmodell.
// Wo ein Schlüssel hingehört, sagt `seiten/einstellungsgruppen.ts`.
export function KIeinstellungen({
  modelle,
  ki,
  angehalten,
  setze,
}: Omit<DiensteeinstellungenEigenschaften, 'dienste' | 'fragen'> & {
  // modelle sind die Schalter der drei Sprachmodelle — sie heißen
  // `dienst.anthropic` und ihresgleichen und gehören trotzdem hierher.
  modelle: EinstellungDTO[]
}) {
  const eigene = useTexte().seiten.einstellungen

  return (
    <div className="flex flex-col gap-6">
      {/* **Über denselben Schalter wie ein Kartendienst** — er weiß, wie
          ein Anbieter heißt (`bausteine/dienstnamen.ts`). Die gewöhnliche
          Wahl schlüge den Namen im Katalog nach und fände keinen; genau das
          stand am 2026-09-21 als `dienst.anthropic` auf der Seite. */}
      <Dienstschalter
        dienste={modelle}
        angehalten={angehalten}
        setze={setze}
        titel={eigene.kiAnbieter}
        hinweis={eigene.kiAnbieterHinweis}
      />

      {/* **Die Einrichtung steht vor dem Katalog** (A7, #1446) — wer nichts
          sieht, soll zuerst lesen, wo etwas einzuschalten ist. */}
      <KILaufzeit zeilen={ki} angehalten={angehalten} setze={setze} />

      {/* **Das Kontingent** (A5 #1444): hier entscheidet jemand, ob
          gefragt werden darf — und was das kostet, gehört in dieselbe
          Entscheidung. **Die Modelle** (A6 #1445) stehen seit #2465 in einem
          Abschnitt darunter: sie werden nachgeschlagen. */}
      <Anbieterstand />
    </div>
  )
}
