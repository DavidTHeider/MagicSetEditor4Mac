// **Die Zonenliste einer Partie** — Werkstück G1 von Phase 5.5
// (#1458),
// herausgelöst mit G3
// (#1460).
//
// # Warum sie eine eigene Datei ist
//
// **Weil sie eine eigene Frage beantwortet** (§2): *was liegt wo* — während
// `Partietisch.tsx` beantwortet, *was eine Partie gerade ist*. Der Anlass war
// die Zeilengrenze: mit dem gezeichneten Tisch daneben ging die eine Datei
// auf 404 Zeilen. Der Schnitt folgt aber der Frage und nicht der Zahl; dass
// beides zusammenfiel, ist der übliche Fall.
//
// **Und sie ist seit G3 die zweite Darstellung, nicht die einzige.** Der
// gezeichnete Tisch zeigt dieselben Karten mit denselben Knöpfen — sichtbar
// ist immer nur eine von beiden.
import { useGeste } from '../haken/useGeste'
import { useTexte } from '../sprache/kontext'
import { Zonenkarten } from './Zonenkarten'
import { zonenDerPartie } from '../spiel/partiezonenliste'
import { zonenwort } from './partiewoerter'

import type { Partie, Spielobjekt } from '../apipartie'
import type { Wirkungsart } from '../tisch/tischwirkungen'
import type { Objekthandlung } from '../apipartie'
import type { Klick } from '../spiel/kartengriff'
import type { Texturlauf } from '../haken/useSpieltexturen'
import { Abschnittstitel } from '../bausteine/Abschnittstitel'

// Partiezonen zeigt jede Zone mit dem, was in ihr liegt.
export function Partiezonen({
  partie,
  texturen,
  gewaehlt,
  waehle,
  wirkungen,
  erlaubt,
  handlungen,
  greife,
  tastenFuer,
  taste,
  befehlen,
}: {
  partie: Partie
  texturen: Texturlauf
  gewaehlt: string | undefined
  waehle: (kennung: string) => void

  // wirkungen sagt je Kennung, was gerade geschehen ist — PLAY-04.AK-02
  // (#1712).
  wirkungen: Map<string, Wirkungsart>

  // erlaubt und befehlen sind der dritte Weg an eine Karte — PLAY-04.AK-01
  // (#1712).
  // **Dieselbe Quelle wie die Handlungsleiste**: was der Kern erlaubt.
  erlaubt: readonly string[]
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void

  // handlungen und greife sind Auskunft und Klick je Karte (P4, #2372).
  handlungen: readonly Objekthandlung[]
  greife: (kennung: string, klick: Klick) => void
  tastenFuer: (kennung: string) => string
  taste: (kennung: string, taste: string) => boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const zonen = zonenDerPartie(partie)

  // **Der Tastenweg aus `Tisch09`** (GESTE, #2358): Leertaste greift, Pfeile
  // wählen die Zone, Eingabe legt ab, Escape bricht ab. Gestellt wird
  // **derselbe Befehl** wie beim Ziehen mit der Maus (`tischbefehle.ts`), und
  // ob er gilt, entscheidet der Kern (Vermerk 0126). **Der Griff gilt nur in
  // der Fassung, aus der er kam**: zieht ein Gegenüber derweil, ist er weg.
  const geste = useGeste({
    stand: { profil: `${partie.id}/${partie.fuer}`, epoche: partie.fassung },
    // **Dieselbe gedeutete Geste wie mit der Maus** (P5, #2373): der
    // Ziel-Sitz steht in der Absicht, gesetzt mit dem Ziel ([zielsitz]).
    ausfuehren: (absicht) => {
      befehlen('objekt.ziehen', {
        objekt: absicht.objekt,
        zone: absicht.ziel,
        sitz: absicht.zielSitz,
      })

      return Promise.resolve()
    },
  })
  const zieleFuer = (objekt: Spielobjekt) =>
    erlaubt.includes('objekt.ziehen')
      ? zonen
          .filter((zone) => zone !== objekt.zone)
          .map((zone) => ({
            kennung: zone,
            name: zonenwort(eigene, zone),
            sitz: zielsitz(objekt, zone),
          }))
      : undefined

  return (
    <div className="flex flex-col gap-3">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.zonen}
      </Abschnittstitel>

      <ul className="flex flex-col gap-2">
        {zonen.map((zone) => (
          <li key={zone} className="flex flex-wrap items-baseline gap-2">
            <span className="w-32 shrink-0 text-sm font-medium">
              {zonenwort(eigene, zone)}
            </span>

            <Zonenkarten
              objekte={partie.objekte.filter((eines) => eines.zone === zone)}
              gewaehlt={gewaehlt}
              waehle={waehle}
              texturen={texturen}
              spiel={partie.spiel}
              sitze={partie.sitze}
              wirkungen={wirkungen}
              handlungen={handlungen}
              greife={greife}
              tastenFuer={tastenFuer}
              taste={taste}
              befehlen={befehlen}
              geste={geste}
              zieleFuer={zieleFuer}
            />
          </li>
        ))}
      </ul>

      {/* **Eine Live-Region und kein zweiter Status** — der Tisch hat
          schon eine Statuszeile; die Ansage der Geste spricht, ohne sich als
          zweite auszugeben. */}
      <p data-gestenansage aria-live="polite" className="sr-only">
        {geste.ansage}
      </p>
    </div>
  )
}

// zielsitz sagt, wessen Zone eine Geste ohne Ort am Tisch meint — das
// Schlachtfeld des Beherrschers (CR 110.2), Hand, Bibliothek und Friedhof des
// Eigners (CR 400.3, beide Fassung 2026-09-25). Den Stapel teilen alle; dort
// nennt die Geste den Beherrscher wie die Zugdeutung im Kern.
function zielsitz(objekt: Spielobjekt, zone: string): string {
  return zone === 'tisch' || zone === 'stapel'
    ? objekt.beherrscher
    : objekt.eigner
}
