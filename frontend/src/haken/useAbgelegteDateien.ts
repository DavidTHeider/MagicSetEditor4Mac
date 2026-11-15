// **Eine gezogene Datei** —
// #246 und
// #891.
//
// # Zwei Hälften, zwei Seiten
//
// Der Browser sieht bei einem Ablegen nur einen **Namen**, keinen Pfad — das
// ist eine Sicherheitszusage jeder Webansicht. Wails hebt sie für die eigene
// Anwendung auf und meldet die **absoluten** Pfade über `runtime.OnFileDrop`
// (`EnableFileDrop` in den Optionen).
//
// **Über welchem Feld losgelassen wurde, weiß dagegen nur das Fenster.** Go
// bekommt Koordinaten; sie in ein Feld umzurechnen hieße, den Aufbau der Seite
// ein zweites Mal zu kennen — und beim ersten geänderten Rand läge er daneben.
//
// # Ein Set ist kein Bild
//
// Bis zum 2026-08-25 ging **jeder** Pfad ins Feld darunter: wer eine
// `.mse-set` auf eine Karte zog, bekam sie als Bild eingesetzt. Die Endung
// entscheidet jetzt zuerst.
import { useEffect } from 'react'

import { useAblage } from '../ablage'
import { laufzeit } from '../fenster'

// setendung ist die Endung, an der ein Setpaket zu erkennen ist.
//
// **Die Endung und nicht der Inhalt.** Ein Setpaket kann ein Verzeichnis oder
// ein Archiv sein; hineinzusehen hieße, Go zu fragen, und die Antwort käme zu
// spät für die Entscheidung, ob es ein Bild ist.
const setendung = '.mse-set'

// useAbgelegteDateien meldet jede gezogene Datei an — **einmal fürs Fenster**.
//
// `aufSet` bekommt den Pfad eines Setpakets; alles andere geht an das Feld,
// über dem losgelassen wurde, falls dort eines ist.
export function useAbgelegteDateien(aufSet: (pfad: string) => void): void {
  const ablage = useAblage()

  useEffect(() => {
    const laufende = laufzeit()
    if (!laufende?.OnFileDrop) {
      return
    }

    laufende.OnFileDrop((x, y, pfade) => {
      // **Das gemerkte Feld gilt genau einmal** — `nimmFeld` gibt es und
      // vergisst es. Wer danach neben die Karte zieht, setzte sonst ein Bild
      // in das Feld von vorhin.
      //
      // **Und ohne gemerktes Feld zählt die Stelle** (ENG-02): unter Wails 3
      // nimmt eine eigene Fläche über der Webansicht die Dateien aus dem
      // Finder entgegen, und die Webansicht sieht kein `dragover` mehr. Das
      // Feld ergibt sich dann aus der Stelle, die Go mit dem Ablegen meldet.
      const feld = ablage.nimmFeld() ?? bildfeldAn(x, y)

      // **Sets und Bilder werden getrennt** — angestrichen vom Zweitblick
      // über Phase Y als Befund 6.
      //
      // Bis zum 2026-08-25 zählte allein `pfade[0]`. Die Begründung galt für
      // Bilder — ein Feld trägt eines —, aber derselbe Weg behandelt auch
      // Setpakete: wer zwei Sets zusammen zog, öffnete eines; wer ein Bild
      // **und** ein Set zog, verlor je nach Reihenfolge das eine oder das
      // andere. Und zwar **wortlos**: es sah aus, als sei der ganze Zug
      // angekommen.
      const sets = pfade.filter((pfad) =>
        pfad.toLowerCase().endsWith(setendung),
      )
      const bilder = pfade.filter(
        (pfad) => !pfad.toLowerCase().endsWith(setendung),
      )

      // **Jedes Set wird geöffnet**, und das letzte gewinnt die Anzeige —
      // dieselbe Ordnung wie beim Doppelklick im Finder
      // (`internal/app/dateioeffnen.go`).
      for (const pfad of sets) {
        aufSet(pfad)
      }

      // **Und das erste Bild geht ins Feld darunter.** Hier bleibt es bei
      // einem, denn ein Feld trägt ein Bild — wer fünf zieht, bekäme sonst
      // fünfmal dasselbe Feld überschrieben.
      const erstesBild = bilder[0]

      if (erstesBild !== undefined && feld !== undefined) {
        ablage.bildweg()?.(feld, erstesBild)
      }
    }, false)

    return () => {
      laufzeit()?.OnFileDropOff?.()
    }
  }, [ablage, aufSet])
}

// bildfeldAn gibt das Bildfeld an einer Stelle des Fensters — `undefined`,
// wenn dort keines liegt.
//
// **Über dieselbe Marke wie die Vorschau** (`data-bildfeld` in
// `Kartenvorschau.tsx`) und nicht über eine zweite Rechnung: wo ein Feld auf
// der Karte liegt, weiß nur die Seite, die es zeichnet.
//
// **`elementFromPoint` fehlt unter jsdom**; die Suche kommt deshalb von
// außen, und im Betrieb ist es die des Dokuments.
export function bildfeldAn(
  x: number,
  y: number,
  suchen: (x: number, y: number) => Element | null = (x, y) =>
    typeof document.elementFromPoint === 'function'
      ? document.elementFromPoint(x, y)
      : null,
): string | undefined {
  return (
    suchen(x, y)?.closest('[data-bildfeld]')?.getAttribute('data-bildfeld') ??
    undefined
  )
}
