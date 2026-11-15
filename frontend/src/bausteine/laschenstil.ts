// **Eine Gestalt für jede Lasche** — AP-3
// (#2360),
// Befund B-10, Entwurf `Zustaende09`.
//
// Bis dahin gab es zwei: einen Kasten mit Rand und Grund in `Laschen` und im
// Setarbeitsplatz, eine Unterlinie in der Setschublade. **Die Unterlinie ist
// die ruhigere** (Empfehlung des Audits), und sie bleibt von einem gewählten
// Segment verschieden — das trägt Rand und Grund (`Knopf`, `Segmente`); mit
// einer Lasche im Kastenstil sahen beide gleich aus.
//
// **Und sie trägt auch dort, wo die Leiste umbricht**: der Kasten war an die
// Stelle eines Reiterblatts getreten, weil ein Blatt, das am Feld klebt, in
// der zweiten Zeile nichts mehr verbindet (#1166). Die Unterlinie behauptet
// keine Verbindung zum Feld — sie steht unter der gewählten Lasche, gleich in
// welcher Zeile.
//
// **Farben gemessen** in `docs/KONTRASTE.md`: `text-leise` und `text-text` auf
// `bg-grund` und `bg-flaeche`, die Unterlinie `border-marke` gegen beide.
// Der Fokusring liegt außen, wie an jedem Knopf.
export const lasche =
  'border-b-2 border-transparent px-3 py-1.5 text-sm text-leise hover:text-text ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring ' +
  'aria-selected:border-marke aria-selected:font-semibold aria-selected:text-text'
