// **Welche Formen gerade erscheinen** — aus der Tabelle in
// `befehlsformen.ts`, nach Gruppe, Erlaubnis und dem Zustand der gewählten
// Karte (#1623).
//
// **Eine eigene Datei seit #1839**: die Tabelle stand mit den vier Befehlen
// des Beginns über der Zeilengrenze aus §2, und der Schnitt folgt der Frage —
// dort steht, **was** ein Befehl braucht, hier, **wann** er erscheint.
import { befehlsformen } from './befehlsformen'

import type { Befehlsform, Gruppe } from './befehlsformen'

// formenDerGruppe gibt die Formen einer Gruppe, die gerade erlaubt sind.
//
// **Aus `erlaubt` und nicht aus der Tabelle**: der Kern rechnet je Sitz aus,
// was gestellt werden darf — samt Priorität (CR 117.1). Ein Knopf, den der
// Kern abwiese, wäre eine Einladung, die er danach zurücknimmt.
export function formenDerGruppe(
  gruppe: Gruppe,
  erlaubt: readonly string[],
  objekt?: { getappt: boolean; haengtAn: string; haufen: string },
): Befehlsform[] {
  return befehlsformen.filter(
    (form) =>
      form.gruppe === gruppe &&
      erlaubt.includes(form.art) &&
      passtZumZustand(form, objekt),
  )
}

// passtZumZustand sagt, ob ein Befehl an diesem Objekt erscheint.
//
// **Ohne Objekt erscheint alles** — die Gruppen außerhalb von `karte`
// meinen keines, und dort wäre jede Bedingung eine über nichts.
function passtZumZustand(
  form: Befehlsform,
  objekt?: { getappt: boolean; haengtAn: string; haufen: string },
): boolean {
  if (form.nurBei === undefined || objekt === undefined) {
    return true
  }

  switch (form.nurBei) {
    case 'getappt':
      return objekt.getappt

    case 'enttappt':
      return !objekt.getappt

    case 'haengt':
      return objekt.haengtAn !== ''
  }

  return objekt.haufen !== ''
}
