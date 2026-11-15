// **Weiterarbeiten** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Was hier steht
//
// Der jüngste Schritt als Handlung („Weiter an …"), darunter **eine Reihe**:
// was angeheftet ist, oben, dann jedes Objekt, mit dem zuletzt etwas
// geschah, nach der Zeit — mit dem, was geschah. Jedes Objekt lässt sich
// anheften oder lösen; ein Klick führt hin: ein Set wird aufgeschlagen, ein
// Deck oder Paket wechselt auf seine Seite.
//
// # Eine Reihe statt vier Listen — UXR-03
//
// #1911,
// gemessen am 2026-10-02: hier standen vier Listen mit je eigener Überschrift
// und eigenem Leerzustand, und ein Objekt stand bis zu dreimal da. Mit allem,
// was die Ablage hergibt, waren es achtzehn Zeilen und sechsunddreißig Knöpfe.
// Das Artboard *Main* zeigt eine Reihe; wie sie entsteht, steht in
// `weiterarbeitenreihe.ts`.
//
// **Vier Zeilen und der Rest auf Wunsch**, wie im Artboard — und keine Zeile
// geht verloren. **Nicht fünf wie bei der Aufmerksamkeit daneben** (UX-15,
// #1771): jede Zeile hier trägt zwei Knöpfe, und mit fünf stünden bei vollem
// Bestand 22 Bedienelemente im Bild, wo R2 aus `docs/UIREVIEW.md` 20 erlaubt.
//
// **„Lösen" trägt nicht das Hauptgewicht.** Es stand als gedrückter
// Umschalter da, und ein gedrückter `Knopf` sieht aus wie ein Hauptknopf: mit
// drei angehefteten Objekten führten vier Knöpfe einen Bereich (R1 aus
// `docs/UIREVIEW.md`). Ob etwas angeheftet ist, sagt jetzt die Zeile selbst,
// und der Knopf nennt, was er tut.
//
// # Lädt selbst, an einem Stand
//
// Wie `Zuletzt` bis zu diesem Werkstück: über `useLaden` am Bestandsstand,
// damit ein anderes Programm, das schreibt, die Liste erneuert. Anheften und
// Lösen laufen über `useVorgang` — ein Vorgang je Objekt (§7, Regel 4) — und
// laden **danach** nach (Regel 3: geschrieben ist geschrieben).
//
// # Leere Bereiche erklären den nächsten Schritt
//
// Wer noch nichts hat, sieht keine leere Liste, sondern die Handlung: ein Set
// anlegen oder öffnen, ein Paket beziehen — mit einem Knopf, der dorthin
// führt (UX-ARCHITEKTUR, Zielzustand der Übersicht).
import { useCallback, useState } from 'react'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { Ladefehler, Wirdgeladen } from '../bausteine/Ladehinweise'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'
import type {
  ArbeitsobjektDTO,
  WeiterarbeitenAntwort,
} from '../apiweiterarbeiten'
import type { Seitenschluessel } from '../navigation'
import { reihe, schluessel, type Reiheneintrag } from './weiterarbeitenreihe'

interface WeiterarbeitenEigenschaften {
  wechseln: (seite: Seitenschluessel) => void
  waehleSet: (setID: string) => void
  stand: number
}

// Gezeigt ist die Zahl der Zeilen, die ohne Klick dastehen.
const Gezeigt = 4

async function holen(): Promise<WeiterarbeitenAntwort> {
  const antwort = await api().Weiterarbeiten()

  return auspacken(antwort, () => antwort)
}

// zielSeite sagt, wohin ein Objekt führt — Sets werden aufgeschlagen, alles
// andere wechselt auf seine Seite.
function zielSeite(art: ArbeitsobjektDTO['art']): Seitenschluessel {
  return art === 'deck' ? 'decks' : 'vorlagen'
}

export function Weiterarbeiten({
  wechseln,
  waehleSet,
  stand,
}: WeiterarbeitenEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const eigene = texte.seiten.start.weiterarbeiten
  const uebersicht = useLaden(holen, [], false, stand)
  const vorgang = useVorgang()
  const [alle, setzeAlle] = useState(false)

  const hin = useCallback(
    (objekt: ArbeitsobjektDTO) => {
      if (objekt.art === 'set') {
        waehleSet(objekt.kennung)
      } else {
        wechseln(zielSeite(objekt.art))
      }
    },
    [waehleSet, wechseln],
  )

  const umheften = (objekt: ArbeitsobjektDTO) => {
    void vorgang.starte(schluessel(objekt), async () => {
      const antwort = objekt.angeheftet
        ? await api().Loesen(objekt.art, objekt.kennung)
        : await api().Anheften(objekt.art, objekt.kennung)
      auspacken(antwort, () => undefined)
      uebersicht.nachladen()
    })
  }

  // **Laden und Scheitern sagen sich, mit Rückweg** — G11 (#2277). Der
  // Vorsatz bleibt: auf der Startseite stehen zwei Abschnitte, und er sagt,
  // dass es ums Laden ging.
  if (uebersicht.fehler !== undefined) {
    return (
      <Ladefehler
        fehler={einsetzen(texte.bausteine.zustaende.fehlerMitGrund, {
          grund: uebersicht.fehler,
        })}
        wiederholen={uebersicht.nachladen}
      />
    )
  }

  const daten = uebersicht.daten
  if (daten === undefined) {
    return uebersicht.laedt ? <Wirdgeladen /> : null
  }

  const zuletzt = daten.zuletzt
  const eintraege = reihe(daten)
  const leer = zuletzt === null && eintraege.length === 0

  return (
    <section aria-label={eigene.titel} className="flex flex-col gap-4">
      <h2 className="text-lg font-semibold">{eigene.titel}</h2>

      <Flaeche>
        {leer ? (
          // **Mit der Marke eines Leerzustands** (G02, #2249) — eigene
          // Gestalt, weil er in einer Fläche steht, aber dieselbe Auskunft.
          <div data-leerzustand="">
            <p className="text-sm text-leise">{eigene.leer}</p>
            <div className="mt-3 flex flex-wrap gap-2">
              <Knopf
                gewicht="haupt"
                aufKlick={() => {
                  wechseln('sets')
                }}
              >
                {eigene.handlungSets}
              </Knopf>
              <Knopf
                gewicht="neben"
                aufKlick={() => {
                  wechseln('pakete')
                }}
              >
                {eigene.handlungPakete}
              </Knopf>
            </div>
          </div>
        ) : (
          <>
            {zuletzt !== null && (
              <div className="mb-4 flex flex-col items-start gap-2">
                <p className="text-sm text-leise">
                  {einsetzen(eigene.zuletztWann, {
                    vorgang: eigene.vorgaenge[zuletzt.vorgang],
                    zeit: alsZeitpunkt(
                      zuletzt.zeitpunkt,
                      darstellung,
                      zeitzone,
                    ),
                  })}
                </p>
                <Knopf
                  gewicht="haupt"
                  aufKlick={() => {
                    hin(zuletzt.objekt)
                  }}
                >
                  {einsetzen(eigene.weiterAn, {
                    art: eigene.arten[zuletzt.objekt.art],
                    name: zuletzt.objekt.name,
                  })}
                </Knopf>
              </div>
            )}

            <ul className="flex flex-col gap-1">
              {(alle ? eintraege : eintraege.slice(0, Gezeigt)).map(
                (eintrag) => (
                  <Zeile
                    key={schluessel(eintrag.objekt)}
                    eintrag={eintrag}
                    hin={hin}
                    umheften={umheften}
                    laeuft={vorgang.laeuft}
                  />
                ),
              )}
            </ul>

            {eintraege.length > Gezeigt && (
              <button
                type="button"
                aria-expanded={alle}
                className="mt-2 self-start rounded-sm px-2 py-1 text-sm underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                onClick={() => {
                  setzeAlle((bisher) => !bisher)
                }}
              >
                {alle
                  ? texte.seiten.start.aufmerksamkeit.weniger
                  : einsetzen(texte.seiten.start.aufmerksamkeit.alle, {
                      anzahl: String(eintraege.length),
                    })}
              </button>
            )}

            {/* **Wie man etwas oben hält, sagt ein Satz** — solange nichts
                oben steht. Vorher hatte jede der vier Listen ihren eigenen
                Leerzustand. */}
            {daten.angeheftet.length === 0 && (
              <p className="mt-3 text-sm text-leise">{eigene.angeheftetLeer}</p>
            )}
          </>
        )}
      </Flaeche>
    </section>
  )
}

function Zeile({
  eintrag: { objekt, schritt },
  hin,
  umheften,
  laeuft,
}: {
  eintrag: Reiheneintrag
  hin: (objekt: ArbeitsobjektDTO) => void
  umheften: (objekt: ArbeitsobjektDTO) => void
  laeuft: (kennung: string) => boolean
}) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const eigene = texte.seiten.start.weiterarbeiten

  return (
    <li className="flex items-baseline justify-between gap-2">
      <button
        type="button"
        className="min-w-0 flex-1 rounded-sm px-2 py-1 text-left text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        onClick={() => {
          hin(objekt)
        }}
      >
        <span className="font-medium">{objekt.name}</span>{' '}
        <span className="text-leise">
          {eigene.arten[objekt.art]}
          {' · '}
          {/* **Was zuletzt geschah, oder dass es oben steht** — die Zeile
              sagt es, denn eine Überschrift darüber gibt es nicht mehr. */}
          {schritt === undefined
            ? eigene.angeheftet
            : `${eigene.vorgaenge[schritt.vorgang]} · ${alsZeitpunkt(schritt.zeitpunkt, darstellung, zeitzone)}`}
        </span>
      </button>
      <Knopf
        gewicht="leise"
        klein
        bezeichnung={einsetzen(
          objekt.angeheftet ? eigene.loesenVon : eigene.anheftenVon,
          { name: objekt.name },
        )}
        angehalten={laeuft(schluessel(objekt))}
        aufKlick={() => {
          umheften(objekt)
        }}
      >
        {objekt.angeheftet ? eigene.loesen : eigene.anheften}
      </Knopf>
    </li>
  )
}
