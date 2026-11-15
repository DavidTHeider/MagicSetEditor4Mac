// **Ein Ort für „hier raus"** —
// #677.
//
// # Warum es diese Fläche gibt
//
// Aus Nutzersicht gemeldet: *„Zu den diversen Export- und Speicheroptionen
// habe ich keine konsolidierte und leicht verständliche Sicht gefunden. Es
// wird zwischen Druckbögen und Exports unterschieden. Das verwirrt."*
//
// # Seit UXR-09 ein Assistent
//
// **Bis dahin ein Wegweiser** — fünf Flächen mit sieben Überschriften; zwei
// Wege führten auf eine **andere** Lasche, einer war nur ein Satz. „Wohin?"
// beantwortete man, indem man die Lasche wechselte, und die Einstellungen
// des Druckbogens standen auf ihrer eigenen, ob man drucken wollte oder
// nicht (#1917).
//
// **Nach dem Artboard *Ausgeben***: links die Frage „Wohin?", rechts das,
// was das gewählte Ziel braucht — und nur das. Wie im Artboard steht die
// erste Frage weiter da, während die zweite beantwortet wird: wer das Ziel
// wechseln will, muss nicht zurück.
//
// **Drei Seiten, ein Assistent** (`ausgabe.ts`): `bilder` und `druckbogen`
// schlagen ihn mit ihrem Ziel auf.
import { Activity, useCallback, useState } from 'react'

import { api, auspacken, type SetDTO } from '../api'
import { OhneKarten } from '../bausteine/OhneKarten'
import { Leerzustand } from '../bausteine/Leerzustand'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { useAusgabeziel } from '../haken/useAusgabeziel'
import { useLaden } from '../haken/useLaden'
import { meldungAus } from '../meldung'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import type { Seitenschluessel } from '../navigation'
import { verlangtesZiel, type Ausgabeseite, type Ausgabeziel } from '../ausgabe'
import { Ausgabeflaeche } from './Ausgabeflaeche'
import { Ausgabewahl } from './Ausgabewahl'

interface AusgebenEigenschaften {
  setID: string

  // seite ist die Seite, als die der Assistent dasteht — `bilder` und
  // `druckbogen` wählen ihr Ziel vor.
  seite: Ausgabeseite

  wechseln: (seite: Seitenschluessel) => void
}

export function Ausgeben({ setID, seite, wechseln }: AusgebenEigenschaften) {
  const texte = useTexte()
  const { ausgeben } = texte.seiten
  const vorgang = useVorgang()

  // **Der Satz wird hier geholt und nicht durchgereicht.** Zwei der Ziele
  // brauchen ihn ganz — Name, Spiel, Herkunft; die Seite bekommt von außen
  // nur seine Kennung.
  const holen = useCallback(async () => {
    const antwort = await api().Sets()

    return auspacken(antwort, () => antwort.sets)
  }, [])

  const { daten, laedt, fehler } = useLaden<SetDTO[]>(holen, [setID])
  const satz = daten?.find((eintrag) => eintrag.id === setID)

  // **Ein Wunsch der Seite wählt vor — einmal.** ⌥⌘0 oder „Ablage ›
  // Drucken" verlangen den Druckbogen; wer danach im Assistenten etwas
  // anderes wählt, hat entschieden, und ein Wunsch, der nur stehen
  // geblieben ist, überstimmt das nicht. Verglichen wird beim Zeichnen mit
  // dem vorigen Wunsch — die Form, die React für „Zustand an eine geänderte
  // Eigenschaft anpassen" vorsieht.
  //
  // **Und ohne Wunsch gilt, was dieses Set zuletzt zeigte** (AP-10, #2367,
  // B-20) — über einen Neustart, aus `ausgabe.ziele`. Gemerkt wird jede Wahl
  // und jeder Wunsch; der erste Wunsch auch, deshalb beginnt `letzterWunsch`
  // leer.
  const verlangt = verlangtesZiel(seite)
  const gedaechtnis = useAusgabeziel(setID)
  const [gewaehlt, setzeZiel] = useState(verlangt)
  const [letzterWunsch, setzeLetzterWunsch] = useState<Ausgabeziel>()
  const ziel = gewaehlt ?? gedaechtnis.ziel

  // **Was einmal gewählt war, bleibt gezeichnet** — verborgen, bis man
  // zurückkommt. Wer am Druckbogen „Letter" einstellt, kurz bei den Bildern
  // nachsieht und zurückkehrt, findet „Letter" (S11.NAT-03).
  const [besucht, setzeBesucht] = useState<readonly Ausgabeziel[]>(
    verlangt === undefined ? [] : [verlangt],
  )

  const gehe = (neu: Ausgabeziel) => {
    setzeZiel(neu)
    setzeBesucht((bisher) => (bisher.includes(neu) ? bisher : [...bisher, neu]))
    gedaechtnis.merke(neu)
  }

  // **Das gemerkte Ziel steht gezeichnet da, ohne besucht zu sein.**
  const gezeichnet =
    ziel === undefined || besucht.includes(ziel) ? besucht : [...besucht, ziel]

  if (verlangt !== letzterWunsch) {
    setzeLetzterWunsch(verlangt)

    if (verlangt !== undefined) {
      gehe(verlangt)
    }
  }

  // **Eine Wahl auf `bilder` oder `druckbogen` führt auf `ausgeben`** —
  // sonst stünde über der Exportvorlage „Druckbogen" im Weg über dem Kopf.
  // Auf `ausgeben` selbst wechselt sie keine Seite: jeder Klick stünde sonst
  // in der Seitengeschichte.
  const waehlen = (neu: Ausgabeziel) => {
    gehe(neu)

    if (seite !== 'ausgeben') {
      wechseln('ausgeben')
    }
  }

  return (
    <>
      {/* **Ebene zwei, weil Ausgeben im Setarbeitsplatz steht**
          (#812). */}
      <Seitenkopf
        titel={ausgeben.titel}
        erklaerung={ausgeben.erklaerung}
        ebene={2}
      />

      {/* **Nebeneinander, solange der Platz reicht** — gefragt wird der
          Behälter und nicht das Fenster, wie bei den Schubladen (UX-08):
          darunter stehen die beiden Schritte übereinander, und nichts rollt
          waagerecht (R5). */}
      <div className="@container">
        <div className="flex flex-col gap-4 @2xl:flex-row @2xl:items-start">
          <div className="@2xl:w-72 @2xl:shrink-0">
            <Ausgabewahl gewaehlt={ziel} waehlen={waehlen} />
            {gedaechtnis.fehlschlag !== undefined && (
              <p role="alert" className="mt-2 text-sm text-fehler">
                {`${ausgeben.nichtGemerkt} ${meldungAus(gedaechtnis.fehlschlag, texte)}`}
              </p>
            )}
          </div>

          <div className="min-w-0 flex-1">
            {/* **Ein leeres Set sagt es vorher** (G02, #2249): Bilder und
                Druckbogen zeigten sonst „0 Karten" an einer Auswahl, die
                sich nicht bedienen ließ — und keinen Weg zu den Karten. */}
            <OhneKarten
              karten={satz?.karten}
              erklaerung={ausgeben.ohneKarten}
              wechseln={wechseln}
            />

            {ziel === undefined && (
              <Leerzustand
                titel={ausgeben.keinZiel}
                erklaerung={ausgeben.keinZielErklaerung}
              />
            )}

            {/* **Ein Schlüssel je Set** (#749, §7 Regel 1): ein anderes Set
                ist eine andere Sache — Vorlagen, Optionen und Meldungen von
                Set A stünden sonst unter Set B. */}
            {gezeichnet.map((eines) => (
              <Activity
                key={`${setID}/${eines}`}
                mode={eines === ziel ? 'visible' : 'hidden'}
              >
                <Ausgabeflaeche
                  ziel={eines}
                  setID={setID}
                  satz={satz}
                  laedt={laedt}
                  fehler={fehler}
                  vorgang={vorgang}
                />
              </Activity>
            ))}
          </div>
        </div>
      </div>
    </>
  )
}
