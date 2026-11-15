// **Das Band nach einem Abbruch** — F04
// (#2343),
// Befund A-04.
//
// Endete die letzte Sitzung nicht sauber — ihre Startmarke lag noch —, sagt
// es das Fenster beim nächsten Start: wann sie begann, ob Go eine
// Absturzspur schrieb, und der Weg in die Diagnose, wo das Diagnosepaket die
// Spur mitnimmt. **Nichts verlässt den Rechner von selbst.**
//
// # Im Rahmen und nicht nur auf der Seite „Betrieb"
//
// Das Artboard `Betrieb09` zeichnet den Hinweis in den Einstellungen. Dort
// sähe ihn nur, wer hingeht — und die Zusage ist ein Hinweis **beim
// Folgestart**. Deshalb steht er neben dem Fassungsband, aus demselben Grund
// wie dieses: er trägt eine Handlung. Die Abweichung steht in
// `PROTOTYP-0.9.0.md`.
//
// # `role="status"`
//
// Ein Abbruch von gestern ist keine Störung, die einen Vorleser mitten im
// Satz unterbricht.
import { useState } from 'react'

import { useAnschluss } from '../anschlussraum'
import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'

import type { AbbruchDTO } from '../apiwiederanlauf'

const bandrahmen =
  'mt-2 flex flex-wrap items-center gap-3 rounded-sm border border-kante px-3 py-2 text-sm'

// neuester ist der Abbruch, der zuletzt begann. **Nicht der letzte in der
// Liste**: Go liest die Marken in der Reihenfolge ihrer Dateinamen, und die
// tragen die Prozessnummer.
function neuester(abbrueche: readonly AbbruchDTO[]): AbbruchDTO | undefined {
  return abbrueche.reduce<AbbruchDTO | undefined>(
    (bisher, abbruch) =>
      bisher === undefined || beginn(abbruch) > beginn(bisher)
        ? abbruch
        : bisher,
    undefined,
  )
}

// beginn ist der Beginn in Millisekunden — negativ für die Nullzeit von Go,
// die eine unlesbare Marke trägt.
function beginn(abbruch: AbbruchDTO): number {
  return Date.parse(abbruch.gestartet)
}

export function Wiederanlaufband() {
  const texte = useTexte()
  const eigene = texte.allgemein.wiederanlauf
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const anschluss = useAnschluss()
  const [ausgeblendet, setzeAusgeblendet] = useState(false)

  // **Ein Fehler zeigt nichts**: der Hinweis ist eine Auskunft über gestern,
  // und eine gescheiterte Frage danach ist keine Meldung wert.
  const geholt = useLaden(async () => {
    const antwort = await api().Wiederanlauf()

    return antwort.fehler == null ? antwort.abbrueche : []
  }, [])

  const abbrueche = geholt.daten ?? []
  const letzter = neuester(abbrueche)

  return (
    <div
      role="status"
      aria-label={eigene.bereich}
      aria-busy={geholt.laedt}
      className="px-6"
    >
      {letzter === undefined || ausgeblendet ? null : (
        <div className={bandrahmen}>
          <span>
            {beginn(letzter) > 0
              ? einsetzen(eigene.satz, {
                  zeit: alsZeitpunkt(letzter.gestartet, darstellung, zeitzone),
                })
              : eigene.satzOhneZeit}
          </span>
          <span>
            {abbrueche.some((abbruch) => abbruch.mitSpur)
              ? eigene.mitSpur
              : eigene.ohneSpur}
          </span>
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              anschluss.uebergib({
                art: 'einstellungsgruppe',
                name: texte.bausteine.diagnoseverweis.oeffnen,
                gruppe: 'diagnose',
              })
            }}
          >
            {texte.bausteine.diagnoseverweis.oeffnen}
          </Knopf>
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              setzeAusgeblendet(true)
            }}
          >
            {eigene.ausblenden}
          </Knopf>
        </div>
      )}
    </div>
  )
}
