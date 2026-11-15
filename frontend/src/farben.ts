// Farben zwischen MSEs Form und der des Browsers.
//
// # Warum es zwei Formen gibt
//
// **MSE schreibt `rgb(r,g,b)`**, bei Durchsichtigkeit `rgba(r,g,b,a)`
// (`format_color`, `src/gfx/color.cpp`, Zeilen 76–83, Fassung f9d9356 vom
// 2021-07-06). Ein `<input type="color">` kennt nur `#rrggbb`.
//
// **Geschrieben wird MSEs Form**, nicht die des Browsers: eine `.mse-set`, die
// wir zurückschreiben, soll MSE öffnen können — und ein Feldskript vergleicht
// mit dem, was in der Vorlage steht.
//
// # Was gelesen wird
//
// MSE liest mehr, als es schreibt: `rgb()`, `rgba()`, `#rgb`, `#rgba`,
// `#rrggbb`, `#rrggbbaa`, `transparent` und die Farbnamen der wx-Datenbank
// (`parse_color`, `src/gfx/color.cpp`, Zeilen 42–73).
//
// **Die Farbnamen fehlen hier**, und das ist eine bewusste Grenze: wx bringt
// rund 150 mit, ihre Liste ist eine Eigenheit von wxWidgets, und keine der
// 105 Farbstiloptionen des Packs benutzt einen. Wer eine findet, sieht sie im
// Textfeld daneben — sie geht nicht verloren, sie wird nur nicht als Kachel
// gezeigt.
//
// # Warum ohne Fanggruppen
//
// **Ein Zugriff auf `treffer[1]` ist unter `noUncheckedIndexedAccess`
// `string | undefined`**, und der Rückfall daneben wäre ein Zweig, den kein
// Test betreten kann — der Ausdruck hat ihn ja gerade ausgeschlossen. Geprüft
// wird deshalb mit `test`, zerlegt wird mit `slice` und `map`.

// Farbe ist ein Farbwert, wie das Fenster ihn braucht — `#rrggbb`.
export type Farbe = `#${string}`

// alsBrowserfarbe macht aus MSEs Form die des Browsers.
//
// **`undefined`, wenn sich nichts lesen lässt** — nicht Schwarz: eine Farbe zu
// erfinden hieße, einen unlesbaren Wert als gelesen auszugeben.
export function alsBrowserfarbe(wert: string): Farbe | undefined {
  const roh = wert.trim()

  if (roh === 'transparent') {
    return '#000000'
  }

  if (roh.startsWith('#')) {
    return ausHex(roh.slice(1))
  }

  return ausZahlen(roh)
}

// ausHex liest `rgb`, `rgba`, `rrggbb` und `rrggbbaa` — ohne das Doppelkreuz.
//
// **Eine kurze Ziffer zählt doppelt**, wie `parse_hex` es tut: aus `#abc` wird
// `#aabbcc`. Die Durchsichtigkeit fällt weg — ein Farbwähler kennt keine.
function ausHex(ziffern: string): Farbe | undefined {
  if (!/^[0-9a-fA-F]+$/.test(ziffern)) {
    return undefined
  }

  const klein = ziffern.toLowerCase()

  if (klein.length === 3 || klein.length === 4) {
    return `#${Array.from(klein.slice(0, 3), (ziffer) => ziffer + ziffer).join('')}`
  }

  if (klein.length === 6 || klein.length === 8) {
    return `#${klein.slice(0, 6)}`
  }

  return undefined
}

// ausZahlen liest `rgb(r,g,b)` und `rgba(r,g,b,a)`.
function ausZahlen(roh: string): Farbe | undefined {
  if (!/^rgba?\(\s*\d+\s*(?:,\s*\d+\s*){2,3}\)$/.test(roh)) {
    return undefined
  }

  const zahlen = roh
    .slice(roh.indexOf('(') + 1, -1)
    .split(',')
    .map((teil) => Number(teil.trim()))

  return `#${zahlen.slice(0, 3).map(zweistellig).join('')}`
}

// alsMsefarbe macht aus `#rrggbb` MSEs `rgb(r,g,b)`.
//
// **Was keine Browserfarbe ist, bleibt, wie es war.** Sonst machte ein
// unerwarteter Wert aus einer Rückgabe eine Erfindung.
export function alsMsefarbe(farbe: string): string {
  const roh = farbe.trim()

  if (!/^#[0-9a-fA-F]{6}$/.test(roh)) {
    return farbe
  }

  const zahlen = [1, 3, 5].map((ab) =>
    Number.parseInt(roh.slice(ab, ab + 2), 16),
  )

  return `rgb(${zahlen.join(',')})`
}

// zweistellig macht aus einer Zahl von 0 bis 255 zwei Hexziffern.
//
// **Geklemmt und nicht abgewiesen**: MSE gibt `rgb(300,0,0)` an `wxColour`
// weiter, und die klemmt. Ein Befund hier machte aus einer schrulligen Vorlage
// eine unlesbare.
function zweistellig(wert: number): string {
  return Math.min(255, Math.max(0, wert)).toString(16).padStart(2, '0')
}
