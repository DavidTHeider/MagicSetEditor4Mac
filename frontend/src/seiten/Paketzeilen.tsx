// Die Zeilen der Paketliste — Gruppen und Pakete (#244).
//
// **Was sie tun, steht hier; wie sie aussehen, in den Klassen** (§2).
//
// # Warum eine Tabelle und kein `role="tree"`
//
// Weil die Zeilen Spalten haben: Art, Fassung, Stand. Ein ARIA-Baum
// verlangte eine eigene Tippsuche und Kürzel zum Auf- und Zuklappen — viel
// Fläche, auf der sich etwas verstecken kann.
//
// # Und warum sie **ein Gitter** ist
//
// Eine Tabelle mit Umschaltknöpfen hat von selbst eine Bedienung: `Tab` läuft
// die Knöpfe ab. **Das stimmt, und es hat einen Preis.** Gemessen mit
// `scripts/serienmass.mjs` bei 1280 × 800: **413 Tabstellen** auf dieser Seite,
// bei 331 installierten Paketen. „Tab läuft die Knöpfe ab" ist bei fünf Zeilen
// eine Bedienung und bei 331 eine Sperre.
//
// Die Tabelle bleibt eine Tabelle — sie bekommt nur den Tastaturvertrag, den
// eine Liste auf jeder Plattform hat: **eine** Tabstelle, ↑ und ↓ zwischen den
// Zeilen, → und ← in die Zeile hinein
// (#1290).
//
// **Seit UXR-17 ein Baumgitter** (`treegrid`, #1925): die Gruppenzeile
// schaltet selbst, und `aria-expanded` an einer Zeile erlaubt ARIA nur dort —
// axe weist es im Gitter ab. Was ein Baumgitter dazu gibt, ist die Stufe
// (`aria-level`); die Tasten bleiben die des Gitters.
//
// **Die Gruppenzeile ist ein `th` über die volle Breite.** Damit steht in der
// Vorlesereihenfolge „Gruppe magic, ausgeklappt" vor den Paketen darunter,
// und nicht eine Zelle ohne Bezug.
//
// **Die Gruppenzeile liegt auf `bg-flaeche` / `bg-slate-900`, nicht auf
// `bg-betont`.** Der erste Anlauf nahm die 100er — und sie ist im hellen
// Erscheinungsbild genau der Anwendungsgrund, auf dem die Paketseite steht
// (`Rahmen.tsx`, Zeile 162). Eine Hervorhebung, die die Farbe des Untergrunds
// hat, ist keine. Weiß auf dem Grund ist dieselbe Sprache wie die
// Seitenleiste: ein Blatt darauf.
//
// **Der Name der Gruppenzeile steht als `aria-label` da** — seit UXR-17
// (#1925) an der Zeile selbst, die schaltet —, nicht als versteckter
// Textknoten daneben. Der erste Anlauf setzte ein `sr-only`-Span vor die
// Aufschrift; zusammengesetzt ergab das je nach Leerraum „Gruppe magic",
// „Gruppe  magic" oder „Gruppemagic" — ein Name, den niemand vorhersagen
// kann, ist kein Name.
import { Fragment, type KeyboardEvent, type MouseEvent } from 'react'

import type { PaketDTO } from '../api'
import { Menue } from '../bausteine/Menue'
import { einsetzen } from '../sprache/einsetzen'
import { einzug } from './paketeinzug'
import { useTexte } from '../sprache/kontext'
import type { Baumzeile } from './paketbaum'
import { useRovierendeZeile } from '../haken/useRovierendeZeile'
import { gitterzeilentaste } from './gitterzeilentaste'
import { Paketmarken, Paketmarkenzeile } from './Paketmarken'
import { Gruppenzeile } from './Paketgruppenzeile'
import type { Markenfilterstand } from '../haken/useMarkenfilter'
import { paketeintraege, paketkennung } from './paketeintraege'

interface Eigenschaften {
  zeilen: readonly Baumzeile[]

  // eingeklappt sind die Schlüssel der **zugeklappten** Gruppen.
  //
  // Die Ausnahmen und nicht die Regel: MSE2 klappt beim Öffnen alles auf
  // (`packages_window.cpp`, Zeile 222, Fassung f9d9356 vom 2021-07-06), und
  // eine Menge, die im Normalfall leer ist, sagt genau das.
  eingeklappt: ReadonlySet<string>

  // suchend sagt, dass gerade gefiltert wird — dann steht ohnehin alles da.
  //
  // **Und dann ist die Gruppenzeile kein Knopf.** Der erste Anlauf ließ ihn
  // stehen: der Pfeil zeigte „zugeklappt", die Kinder standen darunter, und
  // ein Klick änderte sichtbar nichts. Ein Bedienelement, das etwas
  // verspricht, was nicht kommt, ist schlechter als keines — und eine
  // Fokusstation weniger ist beim Durchtabben durch 332 Gruppen keine
  // Kleinigkeit (§6).
  suchend: boolean

  umschalten: (schluessel: string) => void

  // vorbehalte sind die Befunde je Paket — `name.art` auf ihre Zahl.
  //
  // **Eine Zahl und kein Satz.** Der Satz steht im Abschnitt darunter (#57);
  // hier soll die Zeile sagen, *dass* etwas ist, ohne den Blick von der Liste
  // zu ziehen.
  vorbehalte: ReadonlyMap<string, number>

  // gewaehlt ist das Paket im Inspector (`art/name`); waehlen setzt es,
  // planen öffnet den Plan vor einer Änderung — DOM-01 (#1696). **Das Menü
  // löscht nicht mehr selbst**: der Plan ist die Rückfrage, und er nennt,
  // was die Rückfrage vorher verschwieg — betroffene Sets und Abhängige.
  gewaehlt: string | undefined
  waehlen: (paket: PaketDTO) => void
  planen: (vorgang: 'entfernen' | 'zurueck', paket: PaketDTO) => void
  laeuft: boolean

  // marken sind die Marken je Paket (`art/name`) — UX-12 (#1684); eine
  // lokale Anmerkung des Profils, das Paket bleibt unberührt.
  marken: Markenfilterstand

  // markenOffen ist das Paket, dessen Marken gerade bearbeitet werden —
  // erreicht über das Menü der Zeile, wie jede Handlung im Gitter. Es steht
  // in `Pakettabelle`, weil der Rechtsklick dieselben Einträge baut.
  markenOffen: string | null
  setzeMarkenOffen: (kennung: string | null) => void

  // anstossen nimmt Rechtsklick und ⇧F10 an einer Zeile entgegen — AP-5
  // (#2362); `Pakettabelle` öffnet daraufhin ihr Menü.
  anstossen: (
    ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
    kennung: string,
    paket: PaketDTO,
  ) => void
}

export function Paketzeilen({
  zeilen,
  eingeklappt,
  suchend,
  umschalten,
  vorbehalte,
  gewaehlt,
  waehlen,
  planen,
  laeuft,
  marken,
  markenOffen,
  setzeMarkenOffen,
  anstossen,
}: Eigenschaften) {
  const texte = useTexte()

  // **Eine Tabstelle statt einer je Zeile** (#1290) — siehe oben.
  const rovierend = useRovierendeZeile(zeilen.length)

  // zeilentaste ist der Tastenvertrag, den jede Zeile bekommt.
  const zeilentaste =
    (stelle: number) => (ereignis: KeyboardEvent<HTMLElement>) => {
      gitterzeilentaste(ereignis, stelle, zeilen.length, rovierend.rueckenAuf)
    }

  return (
    <tbody>
      {zeilen.map((zeile, stelle) => {
        if (zeile.paket === undefined) {
          return (
            <Gruppenzeile
              key={zeile.schluessel}
              zeile={zeile}
              offen={suchend || !eingeklappt.has(zeile.schluessel)}
              schaltbar={!suchend}
              tabstelle={rovierend.annehmen(stelle)}
              aufTaste={zeilentaste(stelle)}
              umschalten={() => {
                umschalten(zeile.schluessel)
              }}
            />
          )
        }

        const paket = zeile.paket
        const kennung = paketkennung(paket)
        const anzahl = vorbehalte.get(`${paket.name}.${paket.art}`) ?? 0

        // **Das Menü steht an der aktiven Zeile, nicht an jeder** — UXR-17
        // (#1925). Ein Knopf je Paket machte 671 Bedienelemente auf der
        // Seite, 22 im Bild (Hüllenprobe `uirevue`, R2 und R6 in
        // `docs/UIREVIEW.md`). Aktiv ist die gewählte Zeile — die Maus kommt
        // über die Wahl — und die mit der Tabstelle: die Tastatur erreicht
        // den Knopf wie vorher mit → aus der Zeile.
        const aktiv = gewaehlt === kennung || rovierend.annehmen(stelle)

        return (
          <Fragment key={zeile.schluessel}>
            <tr
              tabIndex={rovierend.annehmen(stelle) ? 0 : -1}
              aria-label={paket.name}
              aria-level={zeile.stufe + 1}
              aria-selected={gewaehlt === kennung}
              onClick={() => {
                waehlen(paket)
              }}
              onContextMenu={(ereignis) => {
                anstossen(ereignis, kennung, paket)
              }}
              onKeyDown={(ereignis) => {
                anstossen(ereignis, kennung, paket)
                if (ereignis.defaultPrevented) {
                  return
                }

                // **Enter wählt** — wie ein Klick; die Pfeile bleiben beim
                // Gitter (#1290).
                if (
                  ereignis.key === 'Enter' &&
                  ereignis.target === ereignis.currentTarget
                ) {
                  ereignis.preventDefault()
                  waehlen(paket)

                  return
                }

                zeilentaste(stelle)(ereignis)
              }}
              onFocus={() => {
                rovierend.rueckenAuf(stelle)
              }}
              className="border-b border-kante focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {/* **Der Name bricht notfalls im Wort** — seit die Spalten dem
                  Inhalt folgen (AP-12, #2369), nimmt er den Rest, und ein
                  Paketname ohne Leerzeichen sprengte sonst die Tabelle. */}
              <td
                className="py-1 pr-4 [overflow-wrap:anywhere]"
                style={einzug(zeile.stufe)}
              >
                <span className="font-medium">{paket.name}</span>
                {/* **In der schmalen Tabelle steht die Art hier** (AP-12,
                    #2369) — ihre Spalte erscheint erst ab 42rem. */}
                <span className="block text-sm text-leise @2xl:hidden">
                  {paket.art}
                </span>

                <Paketmarken marken={marken.je[kennung]} />

                {/* **Ohne den Paketnamen davor.** Der erste Anlauf schrieb
                  „magic-m15: 1 Vorbehalt(e)" **neben** „magic-m15" — in der
                  Sichtprüfung stand der Name zweimal in derselben Zelle. Ein
                  Vorleseprogramm liest die Zeile ohnehin von vorn, und dort
                  steht er schon. */}
                {anzahl > 0 && (
                  <span className="ml-2 text-sm text-fehler">
                    {anzahl === 1
                      ? texte.seiten.pakete.katalogEinVorbehalt
                      : einsetzen(texte.seiten.pakete.katalogVorbehalte, {
                          anzahl: String(anzahl),
                        })}
                  </span>
                )}
              </td>

              <td className="hidden whitespace-nowrap py-1 pr-4 text-leise @2xl:table-cell">
                {paket.art}
              </td>

              <td className="whitespace-nowrap py-1 pr-4 text-leise">
                {paket.version === ''
                  ? texte.seiten.pakete.katalogOhneVersion
                  : paket.version}
              </td>

              {/* **Der Stand wie im Artboard** — ein Chip: eingebunden, oder
                lokal geändert. **Mit der Herkunft im `title`**: sie hat keine
                eigene Spalte mehr — im Zentrum zwischen Navigator und
                Inspector lief sie in den Chip hinein (Sichtprüfung am
                2026-09-27); gefiltert wird sie im Navigator, ganz steht sie
                im Inspector. */}
              <td
                title={
                  texte.seiten.pakete.plan.herkunftsarten[paket.herkunftsart]
                }
                className="py-1 pr-4"
              >
                <span
                  className={
                    paket.geaendert
                      ? 'whitespace-nowrap rounded-full border border-kante px-2 text-sm text-fehler'
                      : 'whitespace-nowrap rounded-full border border-kante px-2 text-sm text-leise'
                  }
                >
                  {paket.geaendert
                    ? texte.seiten.pakete.plan.standGeaendert
                    : texte.seiten.pakete.plan.standEingebunden}
                </span>
              </td>

              {/* **Das Menü liegt absolut in seiner Zelle** — AP-12 (#2369):
                  es steht nur an der gewählten Zeile, und seit die Spalten dem
                  Inhalt folgen, schob es beim Fokus die Breiten; die Namen
                  brachen um, und ein Klick landete zwischen zwei Zeilen. */}
              <td className="relative py-1">
                {/* **Die 3rem stehen im Fluss**, der Knopf darüber — ohne das
                    schrumpfte die Zelle neben dem Namen auf ihren Rand. */}
                <div aria-hidden="true" className="h-0 w-12" />
                <div className="absolute inset-y-0 right-0 flex items-center">
                  {/* **Ein Menü statt eines offenen Knopfs** — Werkstück U2 von
                  Phase 5.1 (#1413). Es liegt außerhalb der Tabulatorreihe,
                  denn die Zeile ist die Tabstelle; erreicht wird es mit →
                  aus der Zeile, wie jede Zelle. Löschen und Zurück öffnen
                  den **Plan** im Inspector (DOM-01, #1696); er ist die
                  Rückfrage. */}
                  {aktiv && (
                    <Menue
                      bezeichnung={einsetzen(texte.seiten.pakete.handlungenAn, {
                        name: paket.name,
                      })}
                      gewicht="leise"
                      ohneTabstelle
                      eintraege={paketeintraege(paket, texte, {
                        markenOffen,
                        setzeMarkenOffen,
                        planen,
                        laeuft,
                      })}
                    >
                      {texte.seiten.pakete.handlungenKurz}
                    </Menue>
                  )}
                </div>
              </td>
            </tr>

            {markenOffen === kennung && (
              <Paketmarkenzeile kennung={kennung} marken={marken} />
            )}
          </Fragment>
        )
      })}
    </tbody>
  )
}
