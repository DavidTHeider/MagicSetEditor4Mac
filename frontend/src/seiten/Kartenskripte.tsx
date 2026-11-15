// **Karten aus einem Skript, im Fenster** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was MSE2 hier tut
//
// Es zeigt die Skripte als **Aufklappmenü an „Karte hinzufügen"**
// (`src/gui/set/cards_panel.cpp`, Zeilen 160–176, Fassung f9d9356 vom
// 2021-07-06). Hier stehen sie als Kasten, den das Menü „Mehr" neben dem
// Anlegen öffnet (UXR-05, #1913) — nahe an MSE2, aber kein Menü in einem
// Menü: §6 sagt „keine verschachtelten Dialoge". **Titel und Schließen
// trägt der Kasten in `Kartenwerkzeuge.tsx`**, hier steht sein Inhalt.
//
// # Warum sie nicht immer dastehen
//
// **Welche es gibt, entscheidet das Spielpaket.** Ein Set ohne Spielpaket hat
// keine, und das Menü nennt den Eintrag dann gar nicht — ein leerer Kasten
// wäre eine Zeile, die erklärt werden müsste (§6: leere Zustände erklären
// den nächsten Schritt, statt „Keine Daten" zu zeigen).
import { useTexte } from '../sprache/kontext'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import type { KartenskriptDTO } from '../antwortenkartenskripte'

interface SkriptEigenschaften {
  skripte: readonly KartenskriptDTO[]

  // laeuft hält die Knöpfe an, solange einer arbeitet (§7, Regel 4).
  laeuft: boolean

  // ausfuehren führt das Skript mit diesem Namen.
  ausfuehren: (name: string) => void

  // bericht ist, was der letzte Lauf gemeldet hat — leer, wenn keiner lief.
  bericht: string | undefined

  // befunde ist, was dabei auffiel.
  befunde: readonly string[]
}

export function Kartenskripte({
  skripte,
  laeuft,
  ausfuehren,
  bericht,
  befunde,
}: SkriptEigenschaften) {
  const eigene = useTexte().seiten.karten

  return (
    // **Zu, bis jemand ihn braucht** — seit Werkstück C von Phase 4.10.
    // Dreizehn Knöpfe kosteten 230 Punkte über der Kartenliste; benutzt wird
    // das Werkzeug einmal je Set, die Liste bei jedem Blick.
    //
    // **Die `h2` steht nicht mehr hier**: der Kasten trägt denselben Namen
    // als seine Bezeichnung, und zwei Namen für dieselbe Sache sind einer zu
    // viel.
    <div className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.kartenskripteErklaerung}
        klasse="max-w-3xl text-xs text-leise"
      />

      <div className="flex flex-wrap gap-2">
        {skripte.map((skript) => (
          <Knopf
            key={skript.name}
            gewicht="neben"
            bezeichnung={
              skript.beschreibung === ''
                ? skript.name
                : `${skript.name} — ${skript.beschreibung}`
            }
            angehalten={laeuft}
            aufKlick={() => {
              ausfuehren(skript.name)
            }}
          >
            {skript.name}
          </Knopf>
        ))}
      </div>

      {bericht !== undefined && <p className="text-sm text-text">{bericht}</p>}

      {befunde.map((befund) => (
        <p key={befund} className="text-sm text-hinweis">
          {befund}
        </p>
      ))}
    </div>
  )
}

// berichttext sagt, wie viele Karten angelegt wurden.
//
// **Eine Zahl und kein „fertig".** Wer zwölf Grundlande erwartet und zwei
// bekommt, soll es sehen.
export function berichttext(
  anzahl: number,
  muster: { eine: string; mehrere: string },
): string {
  if (anzahl === 1) {
    return muster.eine
  }

  return einsetzen(muster.mehrere, { anzahl: String(anzahl) })
}
