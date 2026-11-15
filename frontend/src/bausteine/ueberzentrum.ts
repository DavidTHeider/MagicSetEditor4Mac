// **Was über dem Zentrum liegt, und wem Escape gehört** — angestrichen vom
// Zweitblick über Phase 6.3 (M1): der Horcher des Arbeitsplatzes schloss
// die Spalte über dem Zentrum auf **jedes** Escape — in einem Feld des
// Zentrums, das Escape selbst brauchte, bei einer ruhenden Überlagerung,
// und nahm dabei den Fokus aus dem Feld. Dieselben drei Prüfungen wie an der
// Schublade (`Schublade.tsx`).

// ueberZentrum ist die Marke an der Spalte, die über dem Zentrum liegt.
export const ueberZentrum = '[data-ueber="zentrum"]'

// imUeberZentrum sagt, ob der Fokus in der Spalte über dem Zentrum steht.
export function imUeberZentrum(aktiv: Element | null): boolean {
  return aktiv?.closest(ueberZentrum) != null
}

// escapeSchliesst sagt, ob ein Escape die Spalte über dem Zentrum schließt:
// **nur, wenn sie sichtbar ist**, niemand es schon behandelt hat, keine
// Eingabe im Werden ist, und der Fokus in ihr oder nirgends steht.
export function escapeSchliesst(
  ereignis: KeyboardEvent,
  sichtbar: boolean,
): boolean {
  const aktiv = document.activeElement

  return (
    ereignis.key === 'Escape' &&
    sichtbar &&
    !ereignis.defaultPrevented &&
    !ereignis.isComposing &&
    (aktiv === null || aktiv === document.body || imUeberZentrum(aktiv))
  )
}

// verdecktVon sagt, ob die Spalte über dem Zentrum ein Element verdeckt —
// N12 aus dem Nachlauf über Phase 6.3 (#1664): **seine Mitte liegt unter
// ihr**. Ein Element ohne Fläche verdeckt sie nicht; es ist gar nicht zu
// sehen, und das ist nicht ihre Schuld.
export function verdecktVon(ziel: Element, spalte: Element): boolean {
  const flaeche = ziel.getBoundingClientRect()
  const ueber = spalte.getBoundingClientRect()

  if (flaeche.width === 0 || flaeche.height === 0) {
    return false
  }

  const mitteX = flaeche.left + flaeche.width / 2
  const mitteY = flaeche.top + flaeche.height / 2

  return (
    mitteX > ueber.left &&
    mitteX < ueber.right &&
    mitteY > ueber.top &&
    mitteY < ueber.bottom
  )
}
