// **Die Gruppen stehen als Liste, nicht als Leiste** — UX-04.AK-01
// (#1676).
//
// # Der Anlass
//
// **Das Artboard `Einstellungen` zeichnet eine Spalte von 232 Punkten**: das
// Suchfeld oben, darunter die neun Gruppen untereinander mit Name und
// Untertitel; rechts die Fläche. Gebaut waren waagerechte [Laschen], und bei
// 960 Punkten Inhaltsbreite brachen sie **auf zwei Reihen** um — gesehen in
// der Sichtprüfung am 2026-09-21. Sie sind nicht zu eng geraten, sie sind
// die falsche Form.
//
// # Warum kein [Arbeitsplatz](./Arbeitsplatz.tsx)
//
// **Weil [Vermerk 0076](../../../docs/entscheidungen/0076-die-ordnung-der-arbeitsplaetze.md)
// die Einstellungsseite ausdrücklich ausnimmt**: „Start, Hilfe,
// **Einstellungen**, Pakete, Turniere und Nutzer haben keinen Navigator".
// Diese Spalte ist auch keiner: ein Navigator führt **Objekte** — die Karten
// eines Sets, die Dateien eines Pakets —, und was er zeigt, hängt davon ab,
// was offen ist. Hier stehen neun feste Orte derselben Seite. Sie klappt
// nicht ein, sie merkt sich nichts, und ⌥⌘[ fasst sie nicht an; einzuklappen
// gäbe es auch nichts, denn sie ist der einzige Weg zur nächsten Gruppe.
//
// # Was sie mit den Laschen teilt
//
// **Welches Gebiet offen ist und wie die Tastatur es wechselt** — das steht
// in [useGebietswahl](../haken/useGebietswahl.ts). Verschieden ist nur die
// Gestalt: dort eine Leiste und Pfeil links/rechts, hier eine Liste und Pfeil
// hoch/runter.
//
// # Was sie nicht tut
//
//   - **Sie trägt kein Symbol je Gruppe.** Der Entwurf setzt eines davor;
//     dass ein Zeichen in jeder Schriftgröße und in beiden
//     Erscheinungsbildern trägt, ist eine eigene Frage und gehört zu UX-10.
//   - **Sie merkt sich nichts.** Wer die Seite verlässt, kommt bei der
//     ersten Gruppe wieder an — dieselbe Zusage wie bei den Laschen.
//   - **Kein Lauf hält die Umbruchschwelle.** jsdom rechnet keine
//     Behälterabfrage aus; dass die Liste unter 672 Punkten über die Fläche
//     rückt, ist **am laufenden Fenster gemessen** (2026-09-21: bei 800
//     Punkten Fenster 455 und 455 übereinander, bei 1.024 daneben 232 und
//     431) und steht sonst nirgends.
import { useEffect } from 'react'

import { useGebietswahl } from '../haken/useGebietswahl'

import type { ReactNode } from 'react'

import type { Gebiet } from '../haken/useGebietswahl'

// Gruppengebiet ist ein Gebiet samt dem Satz, der sagt, was darin steht.
export interface Gruppengebiet extends Gebiet {
  // unter steht unter dem Namen — im Entwurf „Anbieter, Modelle, Kosten"
  // unter „KI".
  //
  // **Nicht im Feld** über den Einstellungen: dort sähe ihn nur, wer die
  // Gruppe schon aufgeschlagen hat — und die Frage „bin ich hier richtig?"
  // stellt sich davor.
  //
  // **Wahlweise** (AP-11, #2368): die Hilfe führt ihre Ziele ohne Untertitel
  // — ein Inhaltsverzeichnis nennt, was darin steht, und der Name tut das.
  unter?: string | undefined

  // marke steht rechts in der Zeile — seit UX-04.AK-03 die Trefferzahl
  // (#1676).
  //
  // **Sonst wäre ein Treffer in einer anderen Gruppe unsichtbar.** Vorher tat
  // das der Knopf „Im Gebiet „X" zeigen" unter jedem Treffer; mit der
  // Trefferliste fiel er weg.
  marke?: ReactNode
}

export function Gruppenliste({
  bezeichnung,
  gebiete,
  oeffne,
  darueber,
  aufWechsel,
}: {
  // bezeichnung benennt die Liste für Bedienhilfen.
  bezeichnung: string

  gebiete: readonly Gruppengebiet[]

  // oeffne schlägt ein Gebiet von außen auf — daran hängt der Sprung aus der
  // Suche. Die Nummer hinter dem `#` macht denselben Wunsch zweimal wirksam.
  oeffne?: string | undefined

  // darueber steht in derselben Spalte über der Liste. Im Entwurf ist das
  // das Suchfeld.
  darueber?: ReactNode

  // aufWechsel sagt dem Aufrufer, welches Gebiet offen ist — UX-04.AK-03
  // (#1676).
  //
  // **Weil eine Zeile hier angeklickt wird und die Seite es wissen muss**:
  // sie entscheidet, ob sie bei einer neuen Suchfrage umschaltet, und das
  // hängt daran, ob die **offene** Gruppe einen Treffer hat. Ohne diese
  // Meldung könnte sie nur raten oder die Liste ganz führen — und das hieße,
  // den Zustand aus [useGebietswahl] zu verdoppeln.
  aufWechsel?: ((schluessel: string) => void) | undefined
}) {
  const { offen, kennung, leiste, waehle, aufTaste } = useGebietswahl(
    gebiete,
    oeffne,
    'ArrowDown',
    'ArrowUp',
  )

  const offener = offen?.schluessel

  useEffect(() => {
    if (offener !== undefined) {
      aufWechsel?.(offener)
    }
    // **Nur am Gebiet**: der Rückruf kommt bei jeder Zeichnung neu, und mit
    // ihm in der Liste liefe der Effekt endlos.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [offener])

  if (offen === undefined) {
    return null
  }

  return (
    // **Das Bauteil misst sich selbst und nicht das Fenster** — dieselbe
    // Lehre wie bei `arbeitsplatzmasse.engeGrenze`: dort stand die „Inhaltsbreite"
    // 1.200 als Fensterbreite, und der Navigator klappte auf keinem Fenster
    // ein, das es gibt. Hier stand `md:`, und bei 800 Punkten Fenster blieb
    // der Fläche **207 Punkte** — gemessen am 2026-09-21, denn die
    // Seitenleiste nimmt gut 200 davon weg.
    //
    // **Die Schwelle ist die Summe der Teile**: 232 für die Spalte, 16
    // Abstand, und der Fläche soll mehr bleiben als der Spalte — `@2xl` sind
    // 672 Punkte, also 424 für die Fläche.
    //
    // **Die Spalte haftet** — UXR-19
    // (#1927).
    // Unter einer langen Gruppe lag links ein leerer Streifen, und wer zur
    // nächsten wollte, rollte erst ganz hinauf (beobachtet an `ops09b`,
    // #1961). `top-4` wie die Vorschau der Stilwahl; **und sie rollt
    // selbst**, wenn sie höher ist als das Fenster — bei 700 Punkten fehlt
    // dafür nicht viel, und ohne das stünde die letzte Gruppe unter dem
    // Rand, bis die Fläche zu Ende ist. `p-1`, damit der Fokusring des
    // Suchfelds nicht am Rand der rollenden Spalte abgeschnitten wird.
    <div className="@container">
      <div className="flex flex-col gap-4 @2xl:flex-row @2xl:items-start">
        <div className="flex flex-col gap-3 @2xl:sticky @2xl:top-4 @2xl:max-h-[calc(100dvh_-_2rem)] @2xl:w-58 @2xl:shrink-0 @2xl:overflow-y-auto @2xl:p-1">
          {darueber}

          <div
            ref={leiste}
            role="tablist"
            aria-orientation="vertical"
            aria-label={bezeichnung}
            className="flex flex-col gap-1 rounded-lg border border-kante p-1.5"
          >
            {gebiete.map((gebiet) => {
              const dieses = gebiet.schluessel === offen.schluessel

              return (
                <button
                  key={gebiet.schluessel}
                  type="button"
                  role="tab"
                  id={`${kennung}-${gebiet.schluessel}`}
                  data-lasche={gebiet.schluessel}
                  // **Der Name ist der Gruppenname, die Beschreibung der
                  // Untertitel.** Beides steht sichtbar da; wer den ganzen Text
                  // zum Namen machte, hieße „KI Anbieter, Modelle, Kosten" —
                  // und eine Vorlesehilfe sagte den Satz bei jedem Pfeil noch
                  // einmal.
                  aria-labelledby={`${kennung}-${gebiet.schluessel}-name`}
                  aria-describedby={
                    gebiet.unter === undefined
                      ? undefined
                      : `${kennung}-${gebiet.schluessel}-unter`
                  }
                  aria-selected={dieses}
                  aria-controls={`${kennung}-feld`}
                  tabIndex={dieses ? 0 : -1}
                  onClick={() => {
                    waehle(gebiet.schluessel)
                  }}
                  onKeyDown={aufTaste}
                  className="group flex flex-col gap-0.5 rounded-md border border-transparent px-2.5 py-1.5 text-left hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-selected:border-kante aria-selected:bg-betont"
                >
                  <span className="flex items-center gap-2">
                    <span
                      id={`${kennung}-${gebiet.schluessel}-name`}
                      className="font-medium group-aria-selected:font-semibold"
                    >
                      {gebiet.name}
                    </span>

                    {gebiet.marke}
                  </span>

                  {gebiet.unter !== undefined && (
                    <span
                      id={`${kennung}-${gebiet.schluessel}-unter`}
                      className="text-xs text-leise"
                    >
                      {gebiet.unter}
                    </span>
                  )}
                </button>
              )
            })}
          </div>
        </div>

        <div
          id={`${kennung}-feld`}
          role="tabpanel"
          aria-labelledby={`${kennung}-${offen.schluessel}-name`}
          tabIndex={0}
          className="min-w-0 flex-1"
        >
          {offen.inhalt}
        </div>
      </div>
    </div>
  )
}
