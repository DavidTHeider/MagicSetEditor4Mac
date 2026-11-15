// Der Name einer Karte.
//
// **In einer eigenen Datei**, seit die Kartenliste aus `Karten.tsx`
// herausgelöst ist (#288):
// beide brauchen ihn, und ein Import der einen Seite in die andere nur wegen
// einer Hilfsfunktion verbände zwei Dinge, die nichts miteinander zu tun haben.
import { type KarteDTO, type KartenzeileDTO } from '../api'
import { ohneAuszeichnung } from './ohneauszeichnung'

// kartenname holt den Kartennamen aus den Feldwerten.
//
// **Über den Feldnamen `name`**, wie ihn jedes gesichtete Spielpaket führt —
// und mit einem Rückfall, weil eine Karte ohne Namen in einem fremden Set
// vorkommt und trotzdem anklickbar bleiben muss.
//
// **Und ohne seine Auszeichnung**
// (#952): ein
// Name ist ein Feldwert wie jeder andere und kann `<i>` tragen. In einer Liste
// steht er als Text, und dort ist eine spitze Klammer kein Hinweis, sondern
// ein Befund. **Der Wert selbst bleibt roh** — hier wird nur angezeigt.
export function kartenname(karte: KarteDTO, rueckfall: string): string {
  const treffer = karte.werte.find((wert) => wert.feld === 'name')

  return zeilenname(treffer?.wert ?? '', rueckfall)
}

// zeilenname ist dasselbe für eine **Listenzeile** (#891).
//
// Sie trägt den Namen schon einzeln — die Liste holt keine Feldwerte mehr.
export function zeilenname(
  karte: KartenzeileDTO | string,
  rueckfall: string,
): string {
  const roh = typeof karte === 'string' ? karte : karte.name

  return roh.trim() !== '' ? ohneAuszeichnung(roh) : rueckfall
}
