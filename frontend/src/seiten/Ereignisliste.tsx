// **Was zuletzt geschah** — Werkstück G1 von Phase 5.5
// (#1458),
// lesbar seit #1836.
//
// # Und warum der private Teil hier das Ergebnis einer Suche ist
//
// **Wer seine Bibliothek durchsieht, bekommt ihren Inhalt als
// empfängergebundene Ergänzung** — sie steht hier und in keiner anderen
// Sicht.
import { useEffect, useId, useRef } from 'react'

import { gesehene } from '../spiel/blickkarten'
import { handlungswort } from './partiewoerter'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Partie, Spielereignis } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'
import { Abschnittstitel } from '../bausteine/Abschnittstitel'

export function Ereignisliste({ lauf }: { lauf: Partielauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const titelID = useId()

  return (
    <div className="flex flex-col gap-1">
      <Abschnittstitel id={titelID} className="text-base font-semibold">
        {eigene.ereignisse}
      </Abschnittstitel>

      {/* **Eine Lücke wird genannt** (Z3): sonst sähe eine Liste, der vorn
          Züge fehlen, vollständig aus. */}
      {lauf.ereignisluecke && (
        <p className="text-sm text-leise">{eigene.ereignisseLuecke}</p>
      )}

      {lauf.ereignisse.length === 0 && (
        <p className="text-sm text-leise">{eigene.ereignisseLeer}</p>
      )}

      {/* **Ein Protokoll, das immer dasteht** (#1992): `role="log"` sagt an,
          was dazukommt — gemessen am 2026-10-01 in der Abnahme-VM. Eine
          Live-Region, die erst mit ihrem ersten Eintrag entsteht, meldet ihn
          nicht; deshalb steht das Protokoll auch leer. **Die Rolle steht am
          Rahmen und nicht an der Liste** (P4, #2372, axe in WebKit): eine
          `ul` mit fremder Rolle ist keine Liste mehr — und eine leere Liste
          sagte „null Einträge“, wo der Satz darüber es schon sagt. */}
      <div role="log" aria-labelledby={titelID}>
        {lauf.ereignisse.length > 0 && (
          <ul className="flex flex-col gap-2 text-sm">
            {lauf.ereignisse.map((ereignis) => (
              <li key={ereignis.folge} className="flex flex-col gap-1">
                {/* **Mit dem Wort der Handlung** und nicht mit ihrer Kennung:
                „Hineinsehen" liest sich, `zone.blick` nicht. Eine Art ohne
                Wort bleibt, wie sie heißt (§8). */}
                <span>{ereignissatz(ereignis, lauf.partie, texte)}</span>

                <Gesehenes
                  privat={ereignis.privat}
                  titel={eigene.ereignisPrivat}
                />
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

// ereignissatz sagt, wer was womit getan hat —
// #1992.
//
// **Ohne Bild verständlich**: gemessen am 2026-10-01 in der Abnahme-VM las
// VoiceOver nur „Tap". Sitz und Karte stehen öffentlich im Ereignis
// (`objekt`, `sitz`), ihre Namen in der Sicht dieses Sitzes — **und nur
// dort**: eine Karte, deren Name diese Sicht nicht kennt, bleibt ohne Namen.
function ereignissatz(
  ereignis: Spielereignis,
  partie: Partie | undefined,
  texte: ReturnType<typeof useTexte>,
): string {
  const handlung = handlungswort(
    texte.seiten.handlungen.handlungsname,
    ereignis.art,
  )

  if (partie === undefined) {
    return handlung
  }

  const sitz = partie.sitze.find(
    (einer) => einer.kennung === ereignis.oeffentlich.sitz,
  )?.name

  if (sitz === undefined) {
    return handlung
  }

  const eines = partie.objekte.find(
    (kandidat) => kandidat.kennung === ereignis.oeffentlich.objekt,
  )

  // **Der Name aus den Eigenschaften, wie in der Zonenliste** — gemessen am
  // 2026-10-01 in der Abnahme-VM: `name` am Objekt trägt nur ein Spielstein.
  // Ohne Namen keine Kennung als Ersatz: „p0-1" sagt niemandem etwas.
  const objekt =
    eines === undefined || eines.geheim
      ? ''
      : (eines.eigenschaften?.name ?? '') || eines.name

  return objekt === ''
    ? einsetzen(texte.seiten.spielen.ereignisSatzOhneKarte, { sitz, handlung })
    : einsetzen(texte.seiten.spielen.ereignisSatz, { sitz, handlung, objekt })
}

// Gesehenes zeigt die Karten eines Blicks — der Reihe nach, von oben.
//
// **Leer heißt: gar nicht da.** Die meisten Ereignisse tragen keine private
// Ergänzung, und eine Überschrift ohne Liste darunter wäre eine Zusage, die
// niemand einlöst (§10).
//
// **Eine Ergänzung ohne Karten steht, wie sie kam.** Heute legt nur der
// Blick eine; legt der Kern morgen eine zweite, soll sie nicht still
// verschwinden, bis jemand ihr eine Gestalt gibt.
function Gesehenes({
  privat,
  titel,
}: {
  privat: Record<string, unknown>
  titel: string
}) {
  const karten = gesehene(privat)
  const titelID = useId()
  const liste = useRef<HTMLOListElement>(null)

  // **Das Ergebnis kommt in den Blick, wenn es erscheint** — und nur dann:
  // der Effekt hängt an nichts, und ein neues Ergebnis ist ein neuer
  // Eintrag (der `key` ist die Folgenummer). Die Liste steht am Ende der
  // Spalte, unter allen Handlungen; wer „Hineinsehen…" stellte, sah bis zum
  // 2026-09-24 nur, dass eine Mischpflicht dazukam.
  //
  // **`nearest` und nicht `start`**: es rollt so wenig wie möglich, und das
  // Fenster steht dann meist still — der Tisch soll nicht aus dem Blick.
  //
  // **Und Platz für drei Ziffern** (`pl-10`): eine Bibliothek mit hundert
  // Karten zählt bis „100.", und die Zahlen stehen vor dem Inhalt. Mit
  // `pl-5` schnitt die Spalte ab der zehnten Karte die erste Ziffer ab —
  // gesehen in WebKit am 2026-09-24.
  useEffect(() => {
    liste.current?.scrollIntoView({ block: 'nearest' })
  }, [])

  if (karten.length === 0) {
    return Object.keys(privat).length === 0 ? undefined : (
      <span className="font-mono text-xs text-leise">
        {titel}: {JSON.stringify(privat)}
      </span>
    )
  }

  return (
    <div className="flex flex-col gap-1 text-leise">
      <span id={titelID} className="text-xs">
        {titel}
      </span>

      <ol ref={liste} aria-labelledby={titelID} className="list-decimal pl-10">
        {karten.map((karte) => (
          <li key={karte.objekt}>{karte.name}</li>
        ))}
      </ol>
    </div>
  )
}
