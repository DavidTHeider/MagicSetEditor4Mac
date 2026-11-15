// **Der Fokus kehrt zurück — wenn es den Ort noch gibt** — Werkstück U2 von
// Phase 5.1 (#1413).
//
// Drei Bausteine geben den Fokus nach dem Schließen dorthin, wo er vorher
// war: die `Entscheidung`, der `Dialog`, das `Kontextmenue`. Was sie sich
// merken, ist `document.activeElement` — und das kann `null` sein, ein
// Element ohne `focus()` oder eines, das inzwischen aus dem Baum genommen
// wurde, weil die Zeile gelöscht ist, in der es stand.
//
// **Eine Stelle, drei Verbraucher.** Dreimal dieselbe Prüfung liefe
// auseinander, und die Fälle, die es nur im Betrieb gibt, ließen sich in
// keinem der drei belegen. Hier stehen sie einzeln.

// zuletztGedrueckt ist das Bedienelement, auf das zuletzt ein Zeiger drückte
// — UX-08.AK-03 (#1680).
//
// **WebKit gibt einem Knopf beim Klick keinen Fokus** — die Gewohnheit von
// macOS, damit ein Klick auf einen Knopf dem Textfeld den Fokus nicht nimmt.
// **Der Fokus bleibt dabei, wo er war**: wer sich beim Öffnen
// `document.activeElement` merkte, bekam das Element von vorhin — in der
// WebKit-Probe `ux10-aktionen` am 2026-09-25 „Navigator ausblenden" statt
// „Deck löschen" —, oder `body`. Die Schubladen hatten denselben Fehler schon
// einmal, damals behoben am einzelnen Knopf und nicht an der Klasse.
let zuletztGedrueckt: Element | null = null

// **Gehorcht wird ab dem Laden dieses Moduls und in der Capture-Phase**: der
// Druck, der eine Rückfrage öffnet, kommt vor ihr, und wer erst mit ihr zu
// horchen begänne, hörte ihn nie.
document.addEventListener(
  'pointerdown',
  (ereignis) => {
    zuletztGedrueckt =
      ereignis.target instanceof Element
        ? ereignis.target.closest(
            'button, a[href], input, select, textarea, summary, [role="button"], [role="menuitem"], [role="tab"], [tabindex]',
          )
        : null
  },
  true,
)

// **Und eine Taste hebt den Druck auf** — wer danach mit der Tastatur zu
// einem Knopf geht und ⏎ drückt, meint den Knopf mit dem Fokus und nicht den,
// auf den er vor einer Weile geklickt hat.
document.addEventListener(
  'keydown',
  () => {
    zuletztGedrueckt = null
  },
  true,
)

// **Und ebenso ein Fokus, der danach woanders hingeht** — ein Knopf, der
// beim Klick selbst den Fokus nimmt (die Schubladen), oder eine Seite, die
// ihn in ein Feld setzt. **Nicht aber einer, der im gedrückten Element
// bleibt oder in einen Behälter um es herum geht**: WebKit gibt den Fokus
// beim Klick dem nächsten fokussierbaren Vorfahren, in der Anwendung dem
// Inhaltsbereich — gemessen mit der zweiten Probe am 2026-09-25.
document.addEventListener(
  'focusin',
  () => {
    const aktiv = document.activeElement
    const verwandt =
      zuletztGedrueckt?.contains(aktiv) === true ||
      aktiv?.contains(zuletztGedrueckt) === true

    if (!verwandt) {
      zuletztGedrueckt = null
    }
  },
  true,
)

// **Und nach einem nativen Dialog** — UX-08.AK-03
// (#1680).
//
// Ein Dateidialog gehört dem Betriebssystem: solange er offen ist, hat das
// Fenster den Fokus abgegeben, und wenn er schließt, bekommt es ihn zurück —
// als `focus` am Fenster. **Dann kehrt er zum gedrückten Knopf zurück** —
// denn der hatte ihn in WebKit nie: der Klick ließ ihn auf `body` oder gab
// ihn dem Behälter darum. **Ob er woanders hinging, weiß schon `focusin`
// oben**: wer danach in ein Feld klickte, hat den Druck aufgehoben. Eine
// Prüfung auf `body` hier hielt den Fokus im Behälter fest — gefunden mit
// der Gegenprobe am 2026-09-25.
window.addEventListener('focus', () => {
  fokusZurueck(zuletztGedrueckt)
})

// merkeRueckkehr sagt, wohin der Fokus beim Schließen zurück soll: auf das
// Bedienelement, das zuletzt gedrückt wurde, solange es das noch gibt und
// seitdem weder eine Taste fiel noch der Fokus woanders hinging — sonst
// dorthin, wo der Fokus steht.
export function merkeRueckkehr(): Element | null {
  return zuletztGedrueckt?.isConnected === true
    ? zuletztGedrueckt
    : document.activeElement
}

// fokusZurueck setzt den Fokus auf das gemerkte Element — und lässt ihn, wo
// er ist, wenn es das Element nicht mehr gibt: ein `focus()` an einem
// losgelösten Knoten tut nichts, und der Browser setzt den Fokus dann auf den
// Anfang der Seite. Das ist der Preis, und er ist kleiner als ein Fokus im
// Nirgendwo.
export function fokusZurueck(ziel: Element | null): boolean {
  if (!(ziel instanceof HTMLElement) || !ziel.isConnected) {
    return false
  }

  // **Ohne `preventScroll`, mit Absicht** — G20 aus dem Zweitblick über
  // Phase 6.3 (#1664) fragte danach. Wer zurückkehrt, soll sehen, wo er
  // steht (WCAG 2.4.11): ein Fokus, der außerhalb des Bildes landet, ist
  // dem Zeiger wie dem Vorleser verloren. Ein Sprung der Seite ist der Preis
  // — nur dann, wenn der Knopf inzwischen außer Sicht gerollt ist.
  ziel.focus()

  return true
}

// relativZu rechnet eine Stelle auf der Seite in eine Stelle im Kasten um —
// für ein Menü, das am Zeiger stehen soll und `absolute` im Kasten liegt.
//
// **Ohne Kasten bleibt die Stelle, wie sie ist**: der Fall tritt nur auf,
// wenn ein Ereignis vor dem ersten Zeichnen käme, und dann ist die Seite
// die einzige Bezugsgröße, die es gibt.
export function relativZu(
  kasten: Element | null,
  links: number,
  oben: number,
): { left: number; top: number } {
  const rand = kasten?.getBoundingClientRect()

  return {
    left: links - (rand?.left ?? 0),
    top: oben - (rand?.top ?? 0),
  }
}

// nachDemTab schließt ein Menü erst, wenn die Tabulatortaste den Fokus
// weitergegeben hat.
//
// **Nicht im `keydown`.** Wer das Menü dort abbaut, nimmt dem Browser das
// Element, von dem aus er das nächste Ziel sucht: der Fokus fällt ins
// Dokument, und der Tab beginnt wieder vorn — beim Knopf, der das Menü
// geöffnet hat. Gefunden mit user-event 14.6.5 (#1784), das das Ziel seitdem
// wie ein Browser erst nach dem `keydown` bestimmt.
export function nachDemTab(schliessen: () => void): void {
  window.setTimeout(schliessen, 0)
}
