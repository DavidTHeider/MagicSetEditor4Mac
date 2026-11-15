// **Ein Druck über seinen Namen** — AP-10
// (#2367),
// Befund B-20, Entwurf `Wege09`: „Nach Name, Set und Nummer — die
// Druckkennung steht nicht mehr im Hauptweg.“
//
// # Dieselbe Deutung wie beim Einlesen einer Liste
//
// **Gesucht wird über `ListeDeuten`** — eine Zeile `1 <Eingabe>`, also
// „Lightning Bolt“ oder „Lightning Bolt (M11) 149“, wie in jeder Deckliste.
// Die Deutung fragt Scryfall **auf Anweisung** (der Knopf, ↩) und mit
// Zwischenspeicher (§16); geschrieben wird dabei nichts.
//
// # Gewählt, nicht übernommen
//
// **Ein einziger Treffer steht gewählt da**, mehrere zur Wahl — aufgenommen
// wird erst mit „Aufnehmen“ (§16, Regel 6). Gibt es mehr Drucke, als die
// Antwort trägt, sagt der Satz darunter, wie die Wahl kürzer wird.
//
// # Nebenläufig
//
// **Eine Antwort gehört zu der Suche, aus der sie stammt** (§7, Regel 2):
// wer weitertippt und neu sucht, bekommt nicht die alte Liste unter die
// neue Frage. **Ein Fehler steht am Feld** (§6).
//
// **Eine neue Suche hebt die Wahl auf, sobald sie beginnt** — und sagt es dem
// Formular mit `waehle(undefined)`. Bis zum Zweitblick über Phase 6 von 0.9.0
// (Befund 1) räumte sie nur ihre sichtbare Auswahl: wer A wählte und B ohne
// Treffer suchte, sah nichts gewählt und nahm trotzdem A auf. **Weitertippen
// verwirft eine laufende Suche**, denn ihre Antwort gehört zu einem Text, der
// nicht mehr im Feld steht.
import { useEffect, useState } from 'react'

import { api } from '../fassade'
import type { Listentreffer } from '../antwortenlisteneinfuhr'
import { useLauf } from '../haken/useLauf'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { eingabeklasse } from './Beschriftetesfeld'
import { Feld } from './Feld'
import { Knopf } from './Knopf'

export function Drucksuche({
  waehle,
  laeuft,
}: {
  // waehle bekommt den gewählten Druck — Kennung, Name, Set und Nummer —,
  // und `undefined`, sobald keiner mehr gilt.
  waehle: (treffer: Listentreffer | undefined) => void
  // laeuft hält die Suche an, solange das Formular aufnimmt (§7, Regel 4).
  laeuft: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung.drucksuche
  const lauf = useLauf()

  const [eingabe, setzeEingabe] = useState('')
  const [sucht, setzeSucht] = useState(false)
  const [ergebnis, setzeErgebnis] = useState<{
    treffer: readonly Listentreffer[]
    weitere: boolean
  }>()
  const [befund, setzeBefund] = useState<string>()
  const [gewaehlt, setzeGewaehlt] = useState<string>()

  const nimm = (treffer: Listentreffer) => {
    setzeGewaehlt(treffer.printing)
    waehle(treffer)
  }

  // **Eine Suche gehört zu ihrer Runde** — Nachlauf zum Zweitblick, R1:
  // nach einem geglückten Aufnehmen entsteht die Suche neu, und die Antwort
  // eines alten Laufs rief sonst noch `waehle` des Formulars und setzte ihren
  // Druck still in den leeren Entwurf.
  useEffect(
    () => () => {
      lauf.naechster()
    },
    [lauf],
  )

  const suchen = async () => {
    const frage = eingabe.trim()

    // **Während das Formular aufnimmt, sucht auch ↩ nicht** (R1, §7 Regel
    // 4) — der Knopf war angehalten, der Weg über das Feld nicht.
    if (frage === '' || laeuft) {
      return
    }

    const nummer = lauf.naechster()
    setzeSucht(true)
    setzeErgebnis(undefined)
    setzeBefund(undefined)
    setzeGewaehlt(undefined)
    waehle(undefined)

    let ausgang:
      | { treffer: readonly Listentreffer[]; weitere: boolean }
      | { fehler: string }
    try {
      const antwort = await api().ListeDeuten(`1 ${frage}`)
      const [zeile] = auspacken(antwort, () => antwort.deutung)?.zeilen ?? []
      ausgang = {
        treffer: zeile?.treffer ?? [],
        weitere: zeile?.weitere === true,
      }
    } catch (ursache) {
      ausgang = { fehler: meldungAus(ursache, texte) }
    }

    // **Eine Antwort gehört zu ihrer Suche** (§7, Regel 2) — eine spätere
    // Suche hat die Wahl schon neu gestellt.
    if (!lauf.istAktuell(nummer)) {
      return
    }

    setzeSucht(false)

    if ('fehler' in ausgang) {
      setzeErgebnis(undefined)
      setzeBefund(ausgang.fehler)

      return
    }

    setzeErgebnis(ausgang)
    setzeBefund(ausgang.treffer.length === 0 ? eigene.keinTreffer : undefined)

    const [einziger, zweiter] = ausgang.treffer
    if (einziger !== undefined && zweiter === undefined) {
      nimm(einziger)
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-end gap-2">
        <Feld
          beschriftung={eigene.karte}
          hinweis={eigene.hinweis}
          {...(befund === undefined ? {} : { befund })}
        >
          {(verbindung) => (
            <input
              {...verbindung}
              // **Der Fokus beim Öffnen der Schublade** (AP-10, Kernweg 6a):
              // wer ein Exemplar aufnimmt, tippt zuerst den Namen.
              data-autofocus
              type="search"
              value={eingabe}
              size={36}
              onChange={(ereignis) => {
                setzeEingabe(ereignis.target.value)
                lauf.naechster()
                setzeSucht(false)
              }}
              onKeyDown={(ereignis) => {
                if (ereignis.key === 'Enter') {
                  ereignis.preventDefault()
                  void suchen()
                }
              }}
              className={eingabeklasse}
            />
          )}
        </Feld>

        <Knopf
          angehalten={sucht || laeuft}
          aufKlick={() => {
            void suchen()
          }}
        >
          {eigene.suchen}
        </Knopf>
      </div>

      {ergebnis !== undefined && ergebnis.treffer.length > 0 && (
        <fieldset className="flex flex-col gap-1">
          <legend className="text-sm font-medium">{eigene.drucke}</legend>
          {ergebnis.treffer.map((treffer) => (
            <label key={treffer.printing} className="flex items-center gap-2">
              <input
                type="radio"
                name="drucksuche"
                checked={gewaehlt === treffer.printing}
                onChange={() => {
                  nimm(treffer)
                }}
              />
              <span>
                {einsetzen(eigene.treffer, {
                  name: treffer.name,
                  set: treffer.setname,
                  code: treffer.setcode.toUpperCase(),
                  nummer: treffer.nummer,
                })}
              </span>
            </label>
          ))}
          {ergebnis.weitere && (
            <p className="text-sm text-leise">{eigene.weitere}</p>
          )}
        </fieldset>
      )}
    </div>
  )
}
