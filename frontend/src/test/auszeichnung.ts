// **Einen Katalogtext so suchen, wie er auf dem Bildschirm steht.**
//
// # Wofür
//
// `bausteine/Satz` löst `**fett**` und `` `code` `` in **eigene Elemente**
// auf. Danach findet weder der rohe Katalogtext noch der bereinigte einen
// Treffer: `findByText` vergleicht je Textknoten, und der Satz ist auf
// mehrere verteilt.
//
// **Was daraus folgte**: ein Lauf, der ihn roh suchte, war nur grün, solange
// die Auszeichnung **nicht** aufgelöst wurde — er sicherte also zu, dass die
// Sternchen auf dem Bildschirm stehen bleiben. Genau das ist am 2026-09-06 an
// zwölf Stellen der Nutzerseite herausgekommen.

// ohneAuszeichnung gibt den Satz ohne seine Zeichen.
//
// **Dieselben zwei Formen wie in `Satz` und keine dritte** — was dort nicht
// aufgelöst wird, bleibt auch hier stehen.
export function ohneAuszeichnung(satz: string): string {
  return satz
    .replaceAll(/\*\*([^*]+)\*\*/g, '$1')
    .replaceAll(/`([^`]+)`/g, '$1')
}

// alsSatz gibt einen Sucher für einen ausgezeichneten Katalogtext.
//
// **Über das Element und nicht über den Textknoten**: der Satz steht in
// mehreren, und nur ihr gemeinsamer Elternteil trägt ihn ganz.
//
// **Nur das innerste Element zählt.** Ohne diese Bedingung träfe der Sucher
// auch jeden Vorfahren — und `findByText` bräche mit „mehrere Treffer" ab.
export function alsSatz(
  satz: string,
): (inhalt: string, element: Element | null) => boolean {
  const gesucht = ohneAuszeichnung(satz)

  return (_inhalt, element) => {
    if (element?.textContent !== gesucht) {
      return false
    }

    return ![...element.children].some((kind) => kind.textContent === gesucht)
  }
}
