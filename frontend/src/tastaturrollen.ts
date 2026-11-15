// **Pfeiltasten rollen einen Bereich** — N3 des Nachlaufs über Phase 6.13
// (#1906).
//
// **Gemessen am 2026-10-04 an der Hüllenprobe `uirevue`**: WebKit rollt einen
// fokussierten, waagerecht rollbaren Bereich mit Namen und Tabstelle **nicht**
// mit den Pfeiltasten — der Bereich war erreichbar, aber sein rechter Teil
// blieb der Maus (WCAG 2.1.1). Dieser Weg rollt ihn, und zwar **jeden**
// Bereich mit `role="region"` und Tabstelle und nicht einen bestimmten: ein
// Haken je Tabelle wäre die Zusage, die der nächste vergisst.
//
// **Was er nicht tut**: senkrecht rollen (das kann WebKit) und eingreifen,
// wenn eine Sondertaste mitgedrückt ist oder nichts zu rollen ist.

// schritt ist, wie weit ein Druck rollt — etwa eine Spalte.
const schritt = 40

function tastaturrollen(ereignis: KeyboardEvent): void {
  if (ereignis.key !== 'ArrowLeft' && ereignis.key !== 'ArrowRight') {
    return
  }

  if (
    ereignis.altKey ||
    ereignis.metaKey ||
    ereignis.ctrlKey ||
    ereignis.shiftKey
  ) {
    return
  }

  const ziel = ereignis.target

  if (
    !(ziel instanceof HTMLElement) ||
    ziel.getAttribute('role') !== 'region' ||
    ziel.tabIndex < 0 ||
    ziel.scrollWidth <= ziel.clientWidth
  ) {
    return
  }

  ziel.scrollBy({ left: ereignis.key === 'ArrowRight' ? schritt : -schritt })
  ereignis.preventDefault()
}

// installiereTastaturrollen hängt den Weg an ein Dokument und gibt zurück,
// wie er wieder abgeht.
export function installiereTastaturrollen(dokument: Document): () => void {
  dokument.addEventListener('keydown', tastaturrollen)

  return () => {
    dokument.removeEventListener('keydown', tastaturrollen)
  }
}
