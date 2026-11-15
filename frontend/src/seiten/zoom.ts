// **Die Größe der Vorschau** — Werkstück U5 von Phase 5.1
// (#1416).
//
// MSE2 hält je Stil ein `card_zoom` als beliebigen Faktor
// (`StyleSheetSettings`, `src/data/settings.hpp`; `CardViewer::getRotation`,
// `src/gui/control/card_viewer.cpp:106–111`, Fassung f9d9356) — ohne Stufen,
// gesetzt im Einstellungsdialog des Stils. Hier sind es **Stufen und zwei
// Antworten auf eine Frage** (adapted): *Anpassen* passt die Karte in die
// Spalte — im Arbeitsplatz in Breite und Höhe (UXR-06) —, die
// Prozentstufen sind Punkte der Vorlage mal Faktor, *Druckgröße* ist die
// Karte so groß, wie sie aus dem Drucker käme — aus `card dpi` des Stils.
//
// # Warum gerastert wird, was größer als 100 % ist
//
// Ein Bild in Vorlagenpunkten hochzuziehen macht es weich. Ab 150 % rastert
// der Go-Teil deshalb mit der Vergrößerung neu (`Vorschau(karteID, 1.5)`);
// darunter wird das Bild von 100 % verkleinert, und das kostet nichts.
import type { VorschauDTO } from '../api'

export type Zoomwahl = 'anpassen' | 'hundert' | 'druck' | 50 | 150 | 200

// stufen sind die Prozentstufen in der Leiste — zwischen *Anpassen* und
// *Druckgröße*.
export const stufen: readonly (50 | 'hundert' | 150 | 200)[] = [
  50,
  'hundert',
  150,
  200,
]

// vergroesserungFuer sagt, mit welchem Faktor der Go-Teil rastern soll.
export function vergroesserungFuer(zoom: Zoomwahl): number {
  if (zoom === 150) {
    return 1.5
  }

  if (zoom === 200) {
    return 2
  }

  return 1
}

// cssBreite gibt die Breite des Bildes als CSS-Maß — oder nichts, wenn das
// Bild seine Spalte füllen soll.
//
// `gezeigt.breite` ist die **gerasterte** Breite, also Vorlagenpunkte mal
// Vergrößerung; die Vorlagenpunkte sind der Bezug jeder Stufe.
export function cssBreite(
  zoom: Zoomwahl,
  gezeigt: Pick<VorschauDTO, 'breite' | 'dpi'>,
): string | undefined {
  const punkte = gezeigt.breite / vergroesserungFuer(zoom)

  switch (zoom) {
    case 'anpassen':
      return undefined
    case 'hundert':
      return `${String(punkte)}px`
    case 'druck':
      return gezeigt.dpi > 0 ? `${String(punkte / gezeigt.dpi)}in` : undefined
    default:
      return `${String((punkte * zoom) / 100)}px`
  }
}

// passendeBreite gibt der Karte bei *Anpassen* im Arbeitsplatz ihre Breite:
// die der Spalte, **aber nie mehr, als die Höhe zulässt** — UXR-06
// (#1914).
// `hoehe` ist, was unter dem Rahmen bis zum Rand der Spalte bleibt; die
// Breite folgt aus dem Seitenverhältnis des gerasterten Bildes.
export function passendeBreite(
  gezeigt: Pick<VorschauDTO, 'breite' | 'hoehe'>,
  hoehe: number,
): string {
  const breite = Math.floor((hoehe * gezeigt.breite) / gezeigt.hoehe)

  return `min(100%, ${String(breite)}px)`
}

// beschnittProzent gibt, wie viel Prozent der Kante ein Schnitt von 3 mm
// frisst — die übliche Toleranz einer Druckerei. **Neu**, nicht aus MSE2:
// dort gibt es keinen Beschnitt, Karten werden auf Maß gedruckt; die Linie
// hier ist eine Hilfe für den, der zu Hause schneidet.
export function beschnittProzent(
  gezeigt: Pick<VorschauDTO, 'breite' | 'hoehe' | 'dpi'>,
): { x: number; y: number } | undefined {
  if (gezeigt.dpi <= 0) {
    return undefined
  }

  const punkteJeMillimeter = gezeigt.dpi / 25.4
  const schnitt = 3 * punkteJeMillimeter * (gezeigt.breite / gezeigt.breite)

  return {
    x: (schnitt / gezeigt.breite) * 100,
    y: (schnitt / gezeigt.hoehe) * 100,
  }
}
