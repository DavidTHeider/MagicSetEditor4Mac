// **Was der Import nicht verstanden hat** — die Zeilen, die unverändert
// mitreisen.
//
// Eigene Datei, weil `Karten.tsx` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, was mit einer **Karte** geschieht, hier,
// was am **Set** nicht gedeutet werden konnte — und das hängt an nichts als
// seiner Kennung.
import { useCallback } from 'react'

import { api, auspacken } from '../api'
import type { VorbehaltDTO } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Flaeche } from '../bausteine/Flaeche'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface VorbehalteEigenschaften {
  setID: string
}

export function Vorbehalte({ setID }: VorbehalteEigenschaften) {
  const texte = useTexte()

  const holen = useCallback(async () => {
    const antwort = await api().Vorbehalte(setID)

    return auspacken(antwort, () => antwort.vorbehalte)
  }, [setID])

  const vorbehalte = useLaden<VorbehaltDTO[]>(holen, [setID])

  // **Ohne Vorbehalte gibt es keinen Kasten** — Werkstück I von Phase 4.10
  // (#1325).
  //
  // Seit die fünf Formatangaben gedeutet werden, ist die Liste bei jedem
  // kanonischen Set **leer**: am 2026-09-03 über die Bindung nachgezählt, null
  // von zehn haben noch einen Vorbehalt. Eine ganze Fläche für den Satz
  // „nichts" ist Platz für eine Nachricht, die keine ist.
  //
  // **Der Satz bleibt trotzdem** — als Zeile im Kasten darüber. Wer wissen
  // will, ob der Import alles verstanden hat, bekommt eine Antwort; er bekommt
  // nur keinen Rahmen darum.
  if (vorbehalte.daten?.length === 0) {
    return (
      <p className="text-sm text-leise">{texte.seiten.sets.vorbehalte.keine}</p>
    )
  }

  return (
    <Flaeche ueberschrift={texte.seiten.sets.vorbehalte.ueberschrift}>
      <Erklaerungsabsatz
        text={texte.seiten.sets.vorbehalte.erklaerung}
        klasse="mb-2 max-w-prose text-leise"
      />

      <Ladefeld
        laedt={vorbehalte.laedt}
        fehler={vorbehalte.fehler}
        leer={false}
      >
        <></>
      </Ladefeld>

      {/* **Die Zahl steht offen, die Zeilen liegen dahinter** — Werkstück H
          von Phase 4.10
          (#1325).

          Der Block stand offen am Ende der Kartenseite und zeigte bei
          `Vanguard` fünf Zeilen Rohdaten. Dass etwas nicht gedeutet wurde, ist
          die Nachricht; **welche** Zeile es war, ist die Antwort auf eine
          Frage, die man erst stellt.

          **Ohne Vorbehalte steht kein Schalter**, sondern der Satz darüber:
          ein Bedienelement, hinter dem nichts liegt, verspricht etwas, was
          nicht kommt. */}
      {vorbehalte.daten && vorbehalte.daten.length > 0 && (
        <Aufklappkasten
          titel={einsetzen(texte.seiten.sets.vorbehalte.mitZahl, {
            anzahl: String(vorbehalte.daten.length),
          })}
        >
          <ul className="space-y-1 font-mono text-sm">
            {vorbehalte.daten.map((vorbehalt) => (
              <li key={`${vorbehalt.datei}:${String(vorbehalt.zeile)}`}>
                <span className="text-leise">
                  {vorbehalt.datei}:{vorbehalt.zeile}
                </span>{' '}
                {vorbehalt.inhalt}
              </li>
            ))}
          </ul>
        </Aufklappkasten>
      )}
    </Flaeche>
  )
}
