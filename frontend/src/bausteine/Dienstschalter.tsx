// **Die Schalter der fremden Kartendienste** — Werkstück 2 von Phase 4.8
// (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// **Ohne Prüfung auf eine leere Liste**: seit Werkstück I von Phase 4.10
// lässt die Einstellungsseite das **Gebiet** weg, wo keine Dienste sind
// (#1325) —
// eine Lasche, hinter der nichts liegt, verspricht etwas, was nicht kommt. Die
// Zusage steht damit an einer Stelle, und der Zweig hier wäre einer, den kein
// Aufruf mehr erreicht (§3).
//
// **Ein eigener Bereich und keine vier Zeilen zwischen Sprache und
// Schriftgröße.** Was hier steht, betrifft nicht die Anwendung, sondern was
// sie **nach außen** tut — und darunter steht der Katalog, der sagt, was
// jeder Dienst kann und ob er gegen sein Original gelaufen ist.
import { Wahl } from '../seiten/Wahl'
import { schalternamen } from './dienstnamen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import type { EinstellungDTO } from '../api'

// **Die Auskunft, welcher Dienst gegen sein Original gelaufen ist, steht
// seit Werkstück A1 von Phase 5.3 im [Dienstkatalog](./Dienstkatalog.tsx)** —
// je Fähigkeit, mit Datum und Fassung, und aus der Go-Seite statt aus einer
// Liste hier. Die Liste stand bis dahin von Hand da und war einmal falsch
// (magicthegathering.io fehlte); ein Wächter hielt sie gegen die Läufe. Jetzt
// hält er den Katalog (#1440).

export function Dienstschalter({
  dienste,
  angehalten,
  setze,
  titel,
  hinweis,
}: {
  dienste: EinstellungDTO[]
  angehalten: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void

  // titel und hinweis sagen, **worüber** diese Schalter urteilen — seit
  // UX-04.AK-04 stehen dieselben Schalter in zwei Gruppen
  // (#1676).
  //
  // **Ohne sie stand „Fremde Kartendienste" über den Sprachmodellen** —
  // gefunden in der Sichtprüfung am 2026-09-21. Eine Aufschrift, die etwas
  // anderes nennt als das, was darunter steht, ist schlechter als keine.
  titel?: string
  hinweis?: string
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  // **Für jeden Dienst dieselben beiden Stellungen** — deshalb einmal hier
  // statt zweimal je Dienst im Sprachkatalog.
  const werte = { aus: eigene.dienstAus, an: eigene.dienstAn }

  return (
    <section
      aria-label={titel ?? eigene.dienste}
      className="mt-8 flex flex-col gap-3"
    >
      <h2 className="text-lg font-semibold">{titel ?? eigene.dienste}</h2>

      <Erklaerungsabsatz text={hinweis ?? eigene.diensteHinweis} />

      <div className="grid max-w-xl gap-5">
        {dienste.map((eintrag) => (
          <div key={eintrag.schluessel} className="flex flex-col gap-1">
            <Wahl
              eintrag={eintrag}
              // **Ohne Eigennamen gar keine Beschriftung** —
              // #1540.
              //
              // Die Adressen aus A18 (#1533) sind keine Eigennamen und stehen
              // im Katalog.
              //
              // **`undefined` ist hier die richtige Antwort**: [Wahl] schlägt
              // dann selbst in `namen` nach und fällt erst danach auf den
              // Schlüssel zurück. Ein Rückfall an **zwei** Stellen hieß, dass
              // der zweite nie drankam.
              beschriftung={schalternamen[eintrag.schluessel]}
              // **Ohne das stünden hier leere Auswahllisten.** `Wahl` schlüge
              // sonst `dienst.scryfall.aus` im Katalog nach, und den Eintrag
              // gibt es nicht.
              werteNamen={werte}
              angehalten={angehalten(eintrag.schluessel)}
              aendern={(wert) => {
                setze(eintrag.schluessel, wert)
              }}
            />
          </div>
        ))}
      </div>
    </section>
  )
}
