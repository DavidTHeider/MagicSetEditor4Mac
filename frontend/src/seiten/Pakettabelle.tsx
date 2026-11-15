// Die Tabelle im Zentrum des Paketarbeitsplatzes — herausgelöst aus
// `Paketliste.tsx`, als sie mit Navigator und Inspector (DOM-01, #1696) über
// die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort, **was**
// die Seite zeigt und tut, hier, **wie** die Zeilen stehen.
//
// **Und das Kontextmenü der gewählten Zeile liegt um sie** — AP-5
// (#2362),
// Befund B-01. Ein `ul` darf nicht in einer Tabellenzeile stehen; deshalb
// ein Rahmen um die Tabelle, und eine andere Zeile wird erst gewählt, wie in
// der Kartenliste (`useAnstoss`).
import { useState, type ComponentProps } from 'react'

import type { PaketDTO } from '../api'
import { Kontextmenue } from '../bausteine/Kontextmenue'
import { useAnstoss } from '../haken/useAnstoss'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { paketeintraege, paketkennung } from './paketeintraege'
import { Paketzeilen } from './Paketzeilen'

type Eigenschaften = Omit<
  ComponentProps<typeof Paketzeilen>,
  'markenOffen' | 'setzeMarkenOffen' | 'anstossen'
>

// inhaltsspalte ist eine Spalte, die nur so breit ist wie ihr Inhalt.
const inhaltsspalte = 'w-px whitespace-nowrap py-1 pr-4'

// artspalte steht erst ab 42rem Breite der Tabelle da — darunter steht die Art
// unter dem Namen. Bei 1.024 × 700 maß die Tabelle 464 Punkte, und
// „mse-export-template“ nahm davon 170.
const artspalte = 'hidden @2xl:table-cell ' + inhaltsspalte

export function Pakettabelle(zeilen: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.pakete

  const [markenOffen, setzeMarkenOffen] = useState<string | null>(null)

  const pakete = zeilen.zeilen.flatMap((zeile) =>
    zeile.paket === undefined ? [] : [zeile.paket],
  )
  const gewaehlt = pakete.find(
    (paket) => paketkennung(paket) === zeilen.gewaehlt,
  )
  const { anstoss, verbraucht, anstossen } = useAnstoss<PaketDTO>(
    zeilen.gewaehlt,
    zeilen.waehlen,
  )

  return (
    <Kontextmenue
      bezeichnung={
        gewaehlt === undefined
          ? ''
          : einsetzen(eigene.handlungenAn, { name: gewaehlt.name })
      }
      eintraege={
        gewaehlt === undefined
          ? []
          : paketeintraege(gewaehlt, texte, {
              markenOffen,
              setzeMarkenOffen,
              planen: zeilen.planen,
              laeuft: zeilen.laeuft,
            })
      }
      objekt={
        gewaehlt === undefined ? undefined : `paket/${paketkennung(gewaehlt)}`
      }
      anstoss={anstoss}
      anstossVerbraucht={verbraucht}
      nurAngestossen
    >
      {/* **`role="treegrid"` und nicht die Vorgaberolle** (#1290): eine
          Tabelle ist ein Dokument, ein Gitter ist bedienbar — eine
          Tabstelle, und die Pfeiltasten bewegen darin. **Ein Baumgitter seit
          UXR-17** (#1925), weil die Gruppenzeile selbst auf- und zuklappt.
          Warum das hier zählt, steht in `Paketzeilen.tsx`. */}
      {/* **Ein Behälter für die schmale Form** — die Breite der Tabelle,
          nicht die des Fensters, entscheidet, ob die Art eine Spalte hat. */}
      <div className="@container">
        <table role="treegrid" className="w-full border-collapse text-left">
          <caption className="sr-only">{eigene.katalogUeberschrift}</caption>

          {/* **Spalten nach Inhalt** — AP-12 (#2369), B-28. Hier standen
          feste Prozente (34/20/18/20/8), gemessen für ein breites Zentrum;
          zwischen Navigator und Inspector lief die Fassung dann in den Chip
          daneben („2025-06-18eingebunden“). Jetzt nehmen Art, Fassung, Stand
          und Menü, was ihr Inhalt braucht (`w-px` und kein Umbruch), und der
          Name den Rest (`w-full` — ohne das verteilte WebKit den Platz nach
          eigenem Maß, und der Name brach in drei Zeilen, während die Art Luft
          hatte; am Fenster gesehen) — er ist der Grund, warum jemand hersieht, und trägt
          den Einzug der Gruppenstufen. Die Spalten folgen dem Artboard
          `Pakete`: Paket, Art, Fassung, Stand, Menü. */}

          {/* **Der Kopf bleibt stehen.** Wer 508 Zeilen rollt, hat nach zwei
          Umdrehungen vergessen, welche Spalte die Fassung ist. */}
          <thead className="sticky top-0 bg-grund">
            <tr className="border-b border-kante">
              <th scope="col" className="w-full py-1 pr-4">
                {eigene.katalogName}
              </th>
              <th scope="col" className={artspalte}>
                {eigene.katalogArt}
              </th>
              <th scope="col" className={inhaltsspalte}>
                {eigene.katalogVersion}
              </th>
              <th scope="col" className={inhaltsspalte}>
                {eigene.plan.spalteStand}
              </th>
              {/* **Das Menü hat seinen Platz, bevor es erscheint** — es steht
                nur an der gewählten Zeile; ohne feste Breite schob seine
                Spalte beim Wählen alle anderen, und ein Klick landete
                zwischen zwei Zeilen (am Fenster gesehen, `dom01`). */}
              <th scope="col" className="w-12 py-1">
                <span className="sr-only">{eigene.katalogHandlung}</span>
              </th>
            </tr>
          </thead>

          <Paketzeilen
            {...zeilen}
            markenOffen={markenOffen}
            setzeMarkenOffen={setzeMarkenOffen}
            anstossen={anstossen}
          />
        </table>
      </div>
    </Kontextmenue>
  )
}
