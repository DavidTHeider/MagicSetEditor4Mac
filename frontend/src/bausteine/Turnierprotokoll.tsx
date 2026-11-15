// **Das Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk 0139: was geschah, in der Reihenfolge, in der es geschah.
//
// **Die Sätze entstehen hier**, in der Sprache des Fensters: das Protokoll
// trägt einen Vorgang und sprachfreie Angaben — einen Namen, eine
// Rundennummer, einen Spielstand. Eine Paarung heißt nach ihren Namen, und
// nach ihrer Kennung, wenn es sie im Turnier nicht mehr gibt.
//
// # Das Neueste, und davor stückweise
//
// **Seit UXR-12** (#1920).
// Ein Abend mit sechzehn Teilnehmenden und vier Runden schreibt rund fünfzig
// Einträge — über 1.200 Punkte untereinander (R4 in `docs/UIREVIEW.md`),
// und gesucht wird fast immer, was eben geschah. **Die Reihenfolge bleibt**;
// gezeigt werden die letzten `vorn`, und darüber holt ein Knopf die früheren
// dazu — mit den Worten und dem Fokusweg der `Faltliste`, die von vorn
// zeigt und deshalb hier nicht passt.
import { useRef, useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Knopf } from './Knopf'
import { Ladefehler } from './Ladehinweise'
import { useLaden } from '../haken/useLaden'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import type { ProtokolleintragDTO, TurnierDTO } from '../apiturniere'

// vorn ist, wie viele der letzten Einträge zu Anfang dastehen; schritt, wie
// viele frühere ein Klick dazu holt.
const vorn = 20
const schritt = 50

export function Turnierprotokoll({ turnier }: { turnier: TurnierDTO }) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere
  const worte = texte.bausteine.listenansicht
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  const [zahl, setzeZahl] = useState(vorn)
  const stand = useRef<HTMLParagraphElement>(null)

  const protokoll = useLaden<ProtokolleintragDTO[]>(async () => {
    const antwort = await api().TurnierProtokoll(turnier.id)

    return auspacken(antwort, () => antwort.eintraege)
  }, [turnier.id, turnier.fassung])

  const namen = new Map(
    turnier.teilnehmer.map((teilnehmer) => [teilnehmer.id, teilnehmer.name]),
  )
  const paarungen = new Map(
    turnier.runden.flatMap((runde) =>
      runde.paarungen.map((paarung) => [
        paarung.id,
        `${namen.get(paarung.a) ?? paarung.a} ${eigene.gegen} ${namen.get(paarung.b) ?? eigene.bye}`,
      ]),
    ),
  )

  const saetze: Record<string, string> = {
    angelegt: eigene.protokollAngelegt,
    angemeldet: eigene.protokollAngemeldet,
    ausgestiegen: eigene.protokollAusgestiegen,
    eingestiegen: eigene.protokollEingestiegen,
    runde: eigene.protokollRunde,
    ergebnis: eigene.protokollErgebnis,
    zurueckgenommen: eigene.protokollZurueckgenommen,
    abgeschlossen: eigene.protokollAbgeschlossen,
    wiedereroeffnet: eigene.protokollWiedereroeffnet,
    eingelesen: eigene.protokollEingelesen,
  }

  const satz = (eintrag: ProtokolleintragDTO) =>
    einsetzen(saetze[eintrag.vorgang] ?? eintrag.vorgang, {
      angaben: eintrag.angaben,
      paarung: paarungen.get(eintrag.paarungID) ?? eintrag.paarungID,
    })

  if (protokoll.fehler !== undefined) {
    return (
      <Ladefehler fehler={protokoll.fehler} wiederholen={protokoll.nachladen} />
    )
  }

  if (protokoll.daten?.length === 0) {
    return <p className="text-sm text-leise">{eigene.protokollLeer}</p>
  }

  const eintraege = protokoll.daten ?? []
  const ab = Math.max(0, eintraege.length - zahl)
  const frueher = ab

  return (
    <div className="flex flex-col gap-2">
      {/* **Gedeckelt wird sichtbar** (§10, wie bei der `Faltliste`): eine
          Liste, die einfach anfängt, liest sich wie eine vollständige. */}
      {eintraege.length > vorn && (
        <div className="flex flex-wrap items-center gap-3">
          <p
            ref={stand}
            tabIndex={-1}
            role="status"
            className="text-sm text-leise"
          >
            {einsetzen(worte.zahl, {
              gezeigt: String(eintraege.length - ab),
              gesamt: String(eintraege.length),
            })}
          </p>

          {frueher > 0 && (
            <Knopf
              aufKlick={() => {
                // **Der Fokus geht nicht verloren** — verschwindet der Knopf
                // mit dem letzten Stück, geht er auf die Zahl daneben.
                if (frueher <= schritt) {
                  stand.current?.focus()
                }
                setzeZahl(zahl + schritt)
              }}
            >
              {einsetzen(worte.weitere, {
                anzahl: String(Math.min(schritt, frueher)),
              })}
            </Knopf>
          )}
        </div>
      )}

      <ol aria-label={eigene.gebietProtokoll} className="flex flex-col gap-1">
        {eintraege.slice(ab).map((eintrag, stelle) => (
          <li
            key={`${eintrag.zeit}:${String(ab + stelle)}`}
            className="flex flex-wrap gap-3 text-sm"
          >
            <span className="text-leise">
              {alsZeitpunkt(eintrag.zeit, darstellung, zeitzone)}
            </span>
            <span>{satz(eintrag)}</span>
          </li>
        ))}
      </ol>
    </div>
  )
}
