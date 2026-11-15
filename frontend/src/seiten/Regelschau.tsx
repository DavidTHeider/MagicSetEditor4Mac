// **Was dieses Programm durchsetzt** — Werkstück G4 von Phase 5.5
// (#1461),
// nach dem Regelbaum aus PLAY-02.AK-02
// (#1710).
//
// # Warum das vor dem Anfangen steht
//
// **Das Issue verlangt es**: „Spieler sehen vor Start, welche Regeln
// automatisch durchgesetzt werden, und erhalten reproduzierbare
// Format-/Hausregeln ohne falsche Vollständigkeitszusage."
//
// **Und die Lücken stehen auch bei einem Ja da.** Wer „begleitet" bekommt,
// soll sehen, dass „streng" nicht zur Wahl steht — sonst erführe er es erst
// beim Versuch.
//
// # Warum jeder Bereich dasteht und nicht nur die Zusagen
//
// **Weil eine Liste, die nur die Zusagen zeigte, wie eine vollständige
// aussähe.** Der Unterschied zwischen „das setzen wir durch" und „das macht
// ihr selbst" ist die ganze Auskunft.
//
// # Was hier nicht steht
//
// **Die Hindernisse einer Absage und die Widersprüche einer Abwahl** — sie
// stehen seit dem 2026-09-24 in `Regelhinweise`, sichtbar in der Regeltafel.
// Diese Schau steht dort eingeklappt (das Artboard *Eine Partie anfangen*,
// Befund F13), und was den Anfang anhält, darf nicht mit ihr verschwinden.
//
// # Warum gruppiert und nicht dreizehn Zeilen
//
// **Weil dreizehn gleichrangige Namen keine Ordnung sind** (AK-02). Der Baum
// kommt aus dem Kern und nicht aus dieser Datei: eine zweite Gruppierung im
// Fenster wäre eine zweite Wahrheit über dieselbe Frage (§16).
import { useTexte } from '../sprache/kontext'
import {
  regelbereichwort,
  regelgradwort,
  regelgruppenwort,
} from './regelwoerter'

import type { Matrixeintrag, RegelfreigabeDTO } from '../apiregeln'

export function Regelschau({ freigabe }: { freigabe: RegelfreigabeDTO }) {
  const eigene = useTexte().seiten.spielen

  return (
    <section className="flex max-w-prose flex-col gap-3">
      <h3 className="text-base font-semibold">{eigene.regelnTitel}</h3>

      <p className="text-sm text-leise">
        {eigene.regelwerk}: {freigabe.regelwerk} · {eigene.fassung}{' '}
        {String(freigabe.fassung)} · {eigene.wirksamAb} {freigabe.wirksam}
      </p>

      <p className="text-sm">
        {eigene.hoechsteStrenge}: {freigabe.hoechste}
      </p>

      {freigabe.baum.map((gruppe) => (
        <Gruppe key={gruppe.kennung} gruppe={gruppe} freigabe={freigabe} />
      ))}

      <Uebrige freigabe={freigabe} />
    </section>
  )
}

// Gruppe zeigt einen Ast samt seinen Bereichen.
function Gruppe({
  gruppe,
  freigabe,
}: {
  gruppe: RegelfreigabeDTO['baum'][number]
  freigabe: RegelfreigabeDTO
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <section className="flex flex-col gap-2">
      <h4 className="text-sm font-medium">
        {regelgruppenwort(eigene, gruppe.kennung)}
      </h4>

      <ul className="flex flex-col gap-2">
        {gruppe.bereiche.map((bereich) => (
          <li key={bereich} className="text-sm">
            <Bereich eintrag={eintragZu(freigabe, bereich)} name={bereich} />
          </li>
        ))}
      </ul>
    </section>
  )
}

// Bereich zeigt eine Zeile — Name, Grad, Quelle und Grenzen.
function Bereich({
  eintrag,
  name,
}: {
  eintrag: Matrixeintrag | undefined
  name: string
}) {
  const eigene = useTexte().seiten.spielen

  // **Ein Bereich ohne Eintrag ist `unbekannt`** — dieselbe Vorgabe wie im
  // Kern (`Regelmatrix.Grad`), und nicht eine leere Zeile, die wie eine
  // Zusage aussähe.
  const grad = eintrag?.grad ?? 'unbekannt'
  const quelle = eintrag?.quelle ?? ''
  const grenzen = eintrag?.grenzen ?? []

  return (
    <>
      <span className="font-medium">{regelbereichwort(eigene, name)}</span>
      {': '}
      {regelgradwort(eigene, grad)}
      {quelle !== '' && <span className="text-leise"> · {quelle}</span>}

      {grenzen.length > 0 && (
        <ul className="ml-4 list-disc text-leise">
          {grenzen.map((grenze) => (
            <li key={grenze}>{grenze}</li>
          ))}
        </ul>
      )}
    </>
  )
}

// Uebrige zeigt, was in der Matrix steht und in keinem Ast.
//
// # Warum es diese Liste gibt
//
// **Weil eine Fläche, die etwas verschweigt, eine Regel verschweigt.** Der
// Baum kommt aus dem Kern; kommt von dort einmal ein Bereich, den kein Ast
// führt, stünde er ohne diese Liste nirgends — und niemand sähe es, weil
// zwölf Zeilen wie dreizehn aussehen (§10).
//
// **Im Normalfall ist sie leer**: `internal/domain` hält fest, dass jeder
// Bereich in genau einer Gruppe steht.
function Uebrige({ freigabe }: { freigabe: RegelfreigabeDTO }) {
  const imBaum = new Set(freigabe.baum.flatMap((gruppe) => gruppe.bereiche))
  const uebrig = freigabe.matrix.filter(
    (eintrag) => !imBaum.has(eintrag.bereich),
  )

  if (uebrig.length === 0) {
    return null
  }

  return (
    <ul className="flex flex-col gap-2">
      {uebrig.map((eintrag) => (
        <li key={eintrag.bereich} className="text-sm">
          <Bereich eintrag={eintrag} name={eintrag.bereich} />
        </li>
      ))}
    </ul>
  )
}

// eintragZu sucht die Aussage über einen Bereich.
function eintragZu(
  freigabe: RegelfreigabeDTO,
  bereich: string,
): Matrixeintrag | undefined {
  return freigabe.matrix.find((eintrag) => eintrag.bereich === bereich)
}
