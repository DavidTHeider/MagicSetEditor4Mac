// **Das Kartenbild in den Finder ziehen** — AP-12
// (#2369),
// Befund B-24: „Hinausziehen (Kartenbild als PNG in den Finder) braucht eine
// native Ziehquelle.“
//
// # Warum ein Griff und keine Ziehquelle in Go
//
// **Wails v3.0.0-beta.27 hat keine** — `startDrag` verschiebt das Fenster
// (`pkg/application/webview_window_darwin.go` 1911), und die Schicht über der
// WebView nimmt nur abgelegte Dateien an (`webview_window_darwin_drag.m`).
// **WebKit selbst legt ein gezogenes `<img>` als PNG ab** — gemessen am
// 2026-10-17 mit `ap12ziehquelle09`, einer echten Mausgeste über den
// Fensterserver; die Datei heißt nach dem letzten Teil der Adresse. Ein
// eigener Griff, weil die Vorschau von den Feldflächen bedeckt ist: wer auf
// die Karte drückt, meint ein Feld.
//
// # Was er nicht ersetzt
//
// **Ziehen ist eine Geste für die Maus** (WCAG 2.5.7): dasselbe tut
// *Ausgeben › Bilder* mit Klicks und der Tastatur. Der Griff ist ein zweiter
// Weg, kein einziger.
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

// ziehadresse hängt den Namen der Karte hinter die Marke des Bildwegs —
// `…/vorschau/<marke>/<name>.png`. Eine andere Adresse bleibt, wie sie ist.
export function ziehadresse(bild: string, kartenname: string): string {
  // **Was ein Dateiname nicht trägt, wird ein Strich** — Schrägstrich,
  // Doppelpunkt, Rückstrich und Steuerzeichen.
  const dateiname = Array.from(kartenname, (zeichen) =>
    zeichen.charCodeAt(0) < 32 || '/:\\'.includes(zeichen) ? '-' : zeichen,
  )
    .join('')
    .trim()

  return bild.replace(
    /\/vorschau\/([^/]+)\.png$/,
    (_ganz, marke: string) =>
      `/vorschau/${marke}/${encodeURIComponent(dateiname === '' ? 'Karte' : dateiname)}.png`,
  )
}

export function Ziehgriff({
  bild,
  kartenname,
}: {
  bild: string
  kartenname: string
}) {
  const texte = useTexte()
  const worte = texte.seiten.karten.ziehen
  // **Der zweite Weg mit den Namen, die dastehen** — aus den Titeln der
  // Seiten, damit der Satz nicht auseinanderläuft, wenn sie sich ändern.
  const hinweis = einsetzen(worte.hinweis, {
    weg: `${texte.seiten.ausgeben.titel} › ${texte.seiten.bilder.titel}`,
  })

  return (
    <span
      title={hinweis}
      className="flex items-center gap-2 text-sm text-leise"
    >
      <img
        src={ziehadresse(bild, kartenname)}
        alt=""
        draggable
        className="h-8 w-auto cursor-grab rounded-sm border border-kante"
      />
      <span>{worte.beschriftung}</span>
    </span>
  )
}
