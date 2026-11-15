// **Die Ordnung einer Sammlung** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// # Drei Fragen in einem Kasten
//
// **Lagerorte**, **Dubletten** und die **Wunschliste** sind drei Sichten auf
// denselben Bestand. Drei Kästen nebeneinander wären dreimal derselbe Rahmen;
// drei Seiten wären drei Wechsel für eine Frage, die man im Vorbeigehen
// stellt.
//
// **„Besitze ich diese Karte?" steht nicht hier**, sondern an der Karte — dort
// wird sie gestellt. Sie kommt mit dem Deckbau aus Phase 4.6
// (#1293).
import { useState } from 'react'

import { api } from '../fassade'
import { Hinweiskasten } from './Hinweiskasten'
import { Zahlenfeld } from './Zahlenfeld'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type {
  DublettenDTO,
  ExemplarDTO,
  LagerortDTO,
} from '../antwortensammlung'

// Sicht ist, welche der drei Fragen gerade dasteht.
type Sicht = 'lagerorte' | 'dubletten' | 'wuensche'

export function Sammlungsordnung() {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const [sicht, setzeSicht] = useState<Sicht>()
  const [behalten, setzeBehalten] = useState(1)

  // befragt ist die Zahl, mit der die Dubletten geholt wurden — der Satz
  // über eine leere Antwort gehört zu ihr und nicht zum Feld (§7, Regel 2).
  const [befragt, setzeBefragt] = useState(1)
  const [fehler, setzeFehler] = useState<string>()

  const [lagerorte, setzeLagerorte] = useState<LagerortDTO[]>([])
  const [dubletten, setzeDubletten] = useState<DublettenDTO[]>([])
  const [wuensche, setzeWuensche] = useState<ExemplarDTO[]>([])

  const vorgang = useVorgang()

  // hole lädt eine Sicht — und räumt die anderen ab.
  //
  // **§7, Regel 1**: ein Wechsel der Frage räumt die Antwort auf die vorige.
  // Ohne das stünde unter „Dubletten" noch die Lagerortliste.
  const hole = (welche: Sicht) => {
    void vorgang.starte('ordnung', async () => {
      try {
        setzeLagerorte([])
        setzeDubletten([])
        setzeWuensche([])

        if (welche === 'lagerorte') {
          const antwort = await api().SammlungLagerorte()
          setzeLagerorte(auspacken(antwort, () => antwort.lagerorte))
        } else if (welche === 'dubletten') {
          const antwort = await api().SammlungDubletten(behalten)
          setzeDubletten(auspacken(antwort, () => antwort.dubletten))
          setzeBefragt(behalten)
        } else {
          const antwort = await api().SammlungWuensche()
          setzeWuensche(auspacken(antwort, () => antwort.exemplare))
        }

        setzeFehler(undefined)
        setzeSicht(welche)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const leer =
    sicht !== undefined &&
    lagerorte.length === 0 &&
    dubletten.length === 0 &&
    wuensche.length === 0

  return (
    // **Überschrift und Rahmen trägt der Kasten darum** (UXR-10): die
    // Ordnung steht eingeklappt im Navigator.
    <section aria-label={eigene.ordnungTitel} className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-3">
        <button
          type="button"
          onClick={() => {
            hole('lagerorte')
          }}
          disabled={vorgang.laeuft('ordnung')}
          className={ordnungsknopf}
        >
          {eigene.ordnungLagerorte}
        </button>

        <button
          type="button"
          onClick={() => {
            hole('dubletten')
          }}
          disabled={vorgang.laeuft('ordnung')}
          className={ordnungsknopf}
        >
          {eigene.ordnungDubletten}
        </button>

        {/* **Die behaltene Zahl steht neben den Dubletten** — sie gehört zu
            ihnen: eine Dublette ist kein Zustand, sondern ein Verhältnis
            (0057).

            **Und sie meldet sich am Feld** — Werkstück U3 von Phase 5.1
            (#1414).
            Hier stand ein `<input type="number" min={0}>`, und `min` ohne
            Formular meldet nichts (§10): eine `-2` reiste bis zu
            `service.Sammlungssichten.Dubletten` und kam als anonymer Satz
            über dem Kasten zurück. Die Regel dort ist `behalten < 0` →
            `ErrValidation` — **ohne Obergrenze**, und eine hier
            erfundene wiese eine gültige Stückzahl ab (§11). */}
        <div className="w-24">
          <Zahlenfeld
            beschriftung={eigene.ordnungBehalten}
            wert={behalten}
            kleinstens={0}
            ganzzahlig
            schritt={1}
            geaendert={setzeBehalten}
          />
        </div>

        <button
          type="button"
          onClick={() => {
            hole('wuensche')
          }}
          disabled={vorgang.laeuft('ordnung')}
          className={ordnungsknopf}
        >
          {eigene.ordnungWuensche}
        </button>
      </div>

      {fehler !== undefined && <Hinweiskasten art="fehler" meldung={fehler} />}

      {/* **Je Sicht ein eigener Satz** (G02, #2249): „Nichts zu zeigen"
          sagte nicht, was fehlt — und bei der Wunschliste nicht, wie etwas
          daraufkommt. */}
      {leer && (
        <p className="max-w-prose text-sm text-leise">
          {sicht === 'lagerorte'
            ? eigene.ordnungLeerLagerorte
            : sicht === 'dubletten'
              ? einsetzen(eigene.ordnungLeerDubletten, {
                  behalten: String(befragt),
                })
              : eigene.ordnungLeerWuensche}
        </p>
      )}

      {lagerorte.length > 0 && (
        <ul className="flex flex-col gap-1">
          {lagerorte.map((ort) => (
            <li key={ort.lagerort} className="text-sm">
              <span className="font-medium">
                {ort.lagerort === '' ? eigene.ordnungOhneOrt : ort.lagerort}
              </span>{' '}
              <span className="text-leise">
                {einsetzen(eigene.ordnungStueck, {
                  stueck: String(ort.stueck),
                  zeilen: String(ort.zeilen),
                })}
              </span>
            </li>
          ))}
        </ul>
      )}

      {dubletten.length > 0 && (
        <ul className="flex flex-col gap-1">
          {dubletten.map((dublette) => (
            <li key={dublette.exemplar.id} className="text-sm">
              <span className="font-medium">
                {dublette.exemplar.bezeichnung}
              </span>{' '}
              <span className="text-leise">
                {einsetzen(eigene.ordnungUeberzaehlig, {
                  ueberzaehlig: String(dublette.ueberzaehlig),
                })}
              </span>
            </li>
          ))}
        </ul>
      )}

      {wuensche.length > 0 && (
        <ul className="flex flex-col gap-1">
          {wuensche.map((wunsch) => (
            <li key={wunsch.id} className="text-sm">
              <span className="font-medium">{wunsch.bezeichnung}</span>{' '}
              <span className="text-leise">{String(wunsch.menge)}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

// ordnungsknopf ist die Gestalt der drei Knöpfe.
const ordnungsknopf =
  'rounded-sm border border-kante px-3 py-1 text-sm focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-50 '
