// Der Arbeitsplatz eines Sets — die Laschen und die Schubladen darüber.
//
// **Aus `Seiteninhalt` gezogen, als die Seiten stehen blieben** (UX-03.AK-02,
// #1675):
// jede besuchte Lasche steht in einer eigenen `<Activity>`, und eine
// Schublade samt ihrer Laschen ebenso. Wer von *Stil* zu den *Karten* und
// zurück geht, findet *Stil* so, wie er es verließ.
import { Activity } from 'react'

import type { Seitenschluessel, Setschluessel } from '../navigation'
import { ausgabeplatz, istAusgabeseite } from '../ausgabe'
import { Lascheninhalt } from './Lascheninhalt'
import { Setarbeitsplatz } from '../rahmen/Setarbeitsplatz'
import { Setschublade } from '../rahmen/Setschublade'
import { schubladen, schubladeVon } from '../schubladen'
import { Statistikfilterraum } from '../statistikfilter'
import { Verdeckung } from '../verdeckung'

import type { Aufgehoben } from '../haken/useAufgehobeneSeiten'

export interface Setbereichseigenschaften {
  // seite ist die Setseite, die dasteht — oder zuletzt dastand, solange der
  // Bereich verborgen ist.
  seite: Setschluessel
  aufgehoben: Pick<Aufgehoben, 'laschen' | 'schubladen'>
  setID: string | undefined
  wechseln: (seite: Seitenschluessel) => void
  karteID: string | undefined
  oeffneKarte: (setID: string, karteID: string) => void
  symbolziel?: { karteID: string; feld: string } | undefined
  oeffneSymbol: (karteID: string, feld: string) => void
  bestandGeaendert: () => void
  verlaufsstand: number
  verlaufsfehler: string | undefined
  zuruecknehmen?: ((vorgang: string) => void) | undefined
}

// sichtbar macht aus einem Wahrheitswert die Art einer `<Activity>`.
const sichtbar = (ja: boolean) => (ja ? 'visible' : 'hidden')

export function Setbereich({
  seite,
  aufgehoben,
  setID,
  wechseln,
  verlaufsstand,
  ...fuerDieLasche
}: Setbereichseigenschaften) {
  // **Eine Seite in einer Schublade fährt über die Karten** (UX-08,
  // #1680): darunter steht die Kartenfläche, darüber die Schublade.
  const schublade = schubladeVon(seite)
  const unten: Setschluessel =
    schublade === undefined ? ausgabeplatz(seite) : 'karten'

  // **Der Platz des Ausgabeassistenten bekommt die Seite, als die er
  // dasteht** (UXR-09, #1917) — `bilder` und `druckbogen` wählen darin ihr
  // Ziel vor. Verborgen steht er als `ausgeben`: dann verlangt niemand etwas.
  const lasche = (gesetzt: string, welche: Setschluessel) => (
    <Lascheninhalt
      seite={welche === 'ausgeben' && istAusgabeseite(seite) ? seite : welche}
      wechseln={wechseln}
      setID={gesetzt}
      verlaufsstand={verlaufsstand}
      {...fuerDieLasche}
    />
  )

  return (
    // **Der Filter der Statistik gilt für dieses Set** (UXR-F6, #2013): der
    // Raum steht hier, wo `key={setID}` ihn mit dem Set abbaut — ein anderes
    // Set erbt keine Wahl (§7, Regel 1).
    <Statistikfilterraum wechseln={wechseln}>
      <Setarbeitsplatz
        setID={setID}
        aktuell={seite}
        wechseln={wechseln}
        verlaufsstand={verlaufsstand}
        oeffneKarte={fuerDieLasche.oeffneKarte}
      >
        {(gesetzt) => (
          <>
            {/* **Und sie ist gesperrt, solange die Schublade steht**: `inert`
              nimmt Fokus, Klick und Vorleseprogramm, `Verdeckung` die
              Horcher am Fenster (`verdeckung.ts`). Der Umschlag steht
              immer da, damit die Kartenseite beim Öffnen nicht neu
              entsteht; `contents` nimmt ihn aus dem Layout. */}
            <Verdeckung value={schublade !== undefined}>
              <div className="contents" inert={schublade !== undefined}>
                {aufgehoben.laschen.map((welche) => (
                  <Activity key={welche} mode={sichtbar(welche === unten)}>
                    {lasche(gesetzt, welche)}
                  </Activity>
                ))}
              </div>
            </Verdeckung>

            {schubladen.map((eine) => {
              const besucht = aufgehoben.schubladen.filter((welche) =>
                eine.laschen.includes(welche),
              )
              const [erste] = besucht
              const offen = schublade?.schluessel === eine.schluessel

              return (
                erste !== undefined && (
                  <Activity key={eine.schluessel} mode={sichtbar(offen)}>
                    <Setschublade
                      schublade={eine}
                      aktuell={offen ? seite : erste}
                      wechseln={wechseln}
                      schliessen={() => {
                        wechseln('karten')
                      }}
                    >
                      {besucht.map((welche) => (
                        <Activity
                          key={welche}
                          mode={sichtbar(welche === seite)}
                        >
                          {lasche(gesetzt, welche)}
                        </Activity>
                      ))}
                    </Setschublade>
                  </Activity>
                )
              )
            })}
          </>
        )}
      </Setarbeitsplatz>
    </Statistikfilterraum>
  )
}
