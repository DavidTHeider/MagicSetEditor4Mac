// **Die Symbolbibliothek** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Eine Bibliothek ist leer, solange niemand ein Symbol hineinlegt, und die
// einzigen Wege hinein ohne Dateidialog laufen über ein Set, das **schon ein
// Symbol trägt** (`BibliothekssymbolAufnehmen`, `internal/service/symbolbibliothek.go`).
// Das Saatspiel `probe` führt kein Setfeld `symbol`; der Dienst liest aber den
// gespeicherten Wert (`symbolwert`) und nicht das Spielpaket. Deshalb legt die
// Gruppe zuerst eine Zeichnung neben das Set (`SymbolSetzen`) und trägt sie als
// Setsymbol ein (`SetwertSetzen`) — so macht es auch das Fenster, das die zwei
// Schritte getrennt tut (`internal/app/symbol.go`, „Sie wird nicht gleich als
// Setsymbol eingetragen").
//
// **Am Ende steht der Setwert wieder leer, und beide Symbole sind aus der
// Bibliothek.** Die abgelegte Datei neben dem Set bleibt liegen; das Fenster
// räumt sie ebenso nicht (Beiwerk des Sets).
//
// Braucht `k.set` aus der ersten Gruppe.

// punkt macht einen Kontrollpunkt ohne Griffe, mit freier Sperre.
const punkt = (x, y) => ({ x, y, griffDavorX: 0, griffDavorY: 0, griffDanachX: 0, griffDanachY: 0, kurveDanach: false, sperre: 'free' })

// form macht ein Teil der Zeichnung — eine geschlossene Form, vereinigt.
const form = (teilname, punkte) => ({
  art: 'Form', name: teilname, kombination: 'merge', punkte,
  spiegelart: '', kopien: 0, mitteX: 0, mitteY: 0, griffX: 0, griffY: 0, teile: [],
})

const quadrat = [form('Quadrat', [punkt(0.2, 0.2), punkt(0.8, 0.2), punkt(0.8, 0.8), punkt(0.2, 0.8)])]
const dreieck = [form('Dreieck', [punkt(0.5, 0.1), punkt(0.9, 0.9), punkt(0.1, 0.9)])]

export const aSymbole = {
  name: 'Symbolbibliothek (A)',
  rufe: [
    // Vorbedingung: ein Set, das ein Symbol trägt.
    { name: 'SymbolSetzen', argumente: (k) => [k.set.id, quadrat, ''], nimm: 'a.name', nach: (w, k) => { k.aSymboldatei = w } },
    { name: 'SetwertSetzen', argumente: (k) => [k.set.id, 'symbol', k.aSymboldatei, []] },
    {
      name: 'BibliothekssymbolAufnehmen', argumente: (k) => [k.set.id, 'Vertragssymbol'],
      nimm: 'a.id', nach: (w, k) => { k.aSymbol = w },
    },
    {
      name: 'BibliothekssymbolKopieren', argumente: (k) => [k.aSymbol, 'Vertragssymbol Kopie'],
      nimm: 'a.id', nach: (w, k) => { k.aSymbolkopie = w },
    },
    { name: 'BibliothekssymbolUmbenennen', argumente: (k) => [k.aSymbolkopie, 'Vertragssymbol Zwei'] },
    // **Die Kopie trägt das Set danach** — dasselbe Quadrat, unter neuem Dateinamen.
    { name: 'BibliothekssymbolEinsetzen', argumente: (k) => [k.aSymbolkopie, k.set.id] },
    // **Eine neue Fassung des Originals**: das Set trägt noch die vorige und
    // wird in `veraltet` genannt; `Nachziehen` zeichnet es neu, `Zuruecknehmen`
    // holt das Quadrat zurück. Ein Sichern mit derselben Zeichnung änderte
    // nichts und ließe die beiden letzten ins Leere laufen.
    { name: 'BibliothekssymbolSichern', argumente: (k) => [k.aSymbol, dreieck] },
    { name: 'BibliothekssymbolNachziehen', argumente: (k) => [k.aSymbol] },
    { name: 'BibliothekssymbolZuruecknehmen', argumente: (k) => [k.aSymbol] },
    // Ans Ende, was wegnimmt. **Mit Bestätigung**: ein Set kann das Symbol noch tragen.
    { name: 'SetwertSetzen', argumente: (k) => [k.set.id, 'symbol', '', []] },
    { name: 'BibliothekssymbolLoeschen', argumente: (k) => [k.aSymbolkopie, true] },
    { name: 'BibliothekssymbolLoeschen', argumente: (k) => [k.aSymbol, true] },
  ],
}
