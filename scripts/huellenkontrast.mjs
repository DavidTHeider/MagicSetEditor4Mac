// **Kontraste am gerenderten Element** — ABN-01.AK-03
// (#2175).
//
// Ein Drehbuch nennt Elemente — als Ausdruck in der Seite — und bekommt je
// Element das Verhältnis zwischen dem, was vorn steht, und dem Grund, auf dem
// es steht. **Dieselbe Rechnung wie `scripts/kontrastlauf.mjs`**: Farben über
// einen Canvas (Tailwind 4 gibt `oklch()`), der Grund geschichtet vom
// Element nach oben, Weiß als letzter Grund, WCAG 2.2 für das Verhältnis.
//
// Drei Arten, weil ein Rand und eine Trennlinie nicht dort liegen, wo Text
// liegt:
//
// - `text`: die Schriftfarbe gegen den Grund des Elements selbst,
// - `rand`: eine Randfarbe — die obere, oder die genannte Seite — gegen den
//   Grund **außerhalb**, den des Elternteils, denn ein Rand trennt das
//   Element von seiner Umgebung,
// - `platzhalter`: die Farbe von `::placeholder` gegen den Grund des Felds
//   (Nachlauf über Phase A, A03). **WebKit gibt für das Pseudo-Element die
//   Farbe des Felds zurück** (gemessen unter macOS 15.2); dann gilt die
//   letzte greifende `::placeholder`-Regel der Blätter, über ein Probeelement
//   aufgelöst. Greift keine, kommt der Platzhalter mit `fehlt` zurück,
// - `symbol`: die Strich- oder Füllfarbe eines SVG gegen den Grund — ein
//   Bedienelement ohne Text, 3 : 1 (WCAG 1.4.11),
// - `flaeche`: die Hintergrundfarbe gegen den Grund des Elternteils — eine
//   Trennlinie aus `bg-kante`.
//
// **Was es nicht misst**: einen Schatten, einen Verlauf oder ein Bild als
// Grund — und es sagt nicht, ob das Element überhaupt zu sehen ist. Ein
// Element, das nicht gefunden wird, kommt mit `fehlt` zurück und ist für das
// Drehbuch ein roter Befund, nicht ein übersprungener. **Ebenso ein Rand ohne
// Breite**: sonst trüge er die Schriftfarbe (`currentColor`), und ein falsch
// gezieltes Element hielte scheinbar 13 : 1.

export const rechnung = `
  const malflaeche = document.createElement('canvas')
  malflaeche.width = 1
  malflaeche.height = 1
  const stift = malflaeche.getContext('2d', { willReadFrequently: true })
  const alsRGBA = (farbe) => {
    stift.clearRect(0, 0, 1, 1)
    stift.fillStyle = farbe
    stift.fillRect(0, 0, 1, 1)
    const [r, g, b, a] = stift.getImageData(0, 0, 1, 1).data
    return [r, g, b, a / 255]
  }
  const ueber = ([r, g, b, a], [gr, gg, gb]) => [r * a + gr * (1 - a), g * a + gg * (1 - a), b * a + gb * (1 - a)]
  const kanal = (wert) => { const anteil = wert / 255; return anteil <= 0.03928 ? anteil / 12.92 : Math.pow((anteil + 0.055) / 1.055, 2.4) }
  const helligkeit = ([r, g, b]) => 0.2126 * kanal(r) + 0.7152 * kanal(g) + 0.0722 * kanal(b)
  const verhaeltnis = (vorne, hinten) => { const a = helligkeit(vorne); const b = helligkeit(hinten); return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05) }
  const grundVon = (knoten) => {
    const schichten = []
    for (let auf = knoten; auf !== null; auf = auf.parentElement) {
      const farbe = alsRGBA(getComputedStyle(auf).backgroundColor)
      if (farbe[3] === 0) continue
      schichten.push(farbe)
      if (farbe[3] === 1) break
    }
    let grund = [255, 255, 255]
    for (const schicht of schichten.reverse()) grund = ueber(schicht, grund)
    return grund
  }
  const platzhalterfarbe = (element) => {
    const direkt = getComputedStyle(element, '::placeholder').color
    if (direkt && direkt !== getComputedStyle(element).color) return direkt
    let farbe = ''
    const durch = (regeln) => {
      for (const regel of regeln) {
        if (regel.cssRules) durch(regel.cssRules)
        if (!regel.selectorText || !regel.style?.color) continue
        for (const teil of regel.selectorText.split(',').map((t) => t.trim())) {
          if (!teil.endsWith('::placeholder')) continue
          try { if (element.matches(teil.slice(0, -13) || '*')) farbe = regel.style.color } catch { /* ein Wähler, den die Seite nicht kennt */ }
        }
      }
    }
    for (const blatt of document.styleSheets) { try { durch(blatt.cssRules) } catch { /* fremdes Blatt */ } }
    if (!farbe) return ''
    const probe = document.createElement('span')
    probe.style.color = farbe
    element.parentElement.appendChild(probe)
    const aufgeloest = getComputedStyle(probe).color
    probe.remove()
    return aufgeloest
  }
  const messe = (name, element, art, grenze, seite = 'Top') => {
    if (!element) return { name, art, grenze, fehlt: true }
    const stil = getComputedStyle(element)
    if (art === 'rand' && parseFloat(stil['border' + seite + 'Width']) === 0) return { name, art, grenze, fehlt: true }
    const platzhalter = art === 'platzhalter' ? platzhalterfarbe(element) : ''
    if (art === 'platzhalter' && !platzhalter) return { name, art, grenze, fehlt: true }
    const hinten = art === 'text' || art === 'platzhalter' ? grundVon(element) : grundVon(element.parentElement)
    const symbolfarbe = art === 'symbol' ? [stil.stroke, stil.fill].find((f) => f && f !== 'none') ?? '' : ''
    if (art === 'symbol' && !symbolfarbe) return { name, art, grenze, fehlt: true }
    const vorn = alsRGBA(art === 'text' ? stil.color : art === 'platzhalter' ? platzhalter : art === 'symbol' ? symbolfarbe : art === 'rand' ? stil['border' + seite + 'Color'] : stil.backgroundColor)
    return { name, art, grenze, verhaeltnis: Math.round(verhaeltnis(ueber(vorn, hinten), hinten) * 100) / 100 }
  }
`

// kontraste misst jedes genannte Element: `[name, ausdruck, art, grenze]`,
// der Ausdruck in der Seite ausgewertet — bei einem Rand als fünftes die
// Seite (`'Bottom'` für eine Unterlinie), sonst die obere.
export function kontraste(p, liste) {
  const rufe = liste.map(([name, ausdruck, art, grenze, seite]) => `messe(${JSON.stringify(name)}, ${ausdruck}, ${JSON.stringify(art)}, ${String(grenze)}${seite === undefined ? '' : `, ${JSON.stringify(seite)}`})`)

  return p.seite(`${rechnung}\n  return [${rufe.join(', ')}]`)
}

// pruefeKontraste misst und prüft — ein fehlendes Element ist rot.
export async function pruefeKontraste(p, wo, liste) {
  for (const w of await kontraste(p, liste)) {
    p.pruefe(`Kontrast ${wo}: ${w.name} (${w.art}) hält ${String(w.grenze).replace('.', ',')} : 1`, !w.fehlt && w.verhaeltnis >= w.grenze, w.fehlt ? 'nicht gefunden' : String(w.verhaeltnis).replace('.', ','))
  }
}

// texteMessen misst jeden sichtbaren Text unter `wurzel` (Wähler in der
// Seite) und gibt die Messungen zurück — leer heißt: nichts gefunden.
//
// **Die Grenzen sind einstellbar** (H02, #2256): WCAG verlangt 4,5 : 1 und
// für große Schrift 3 : 1; die Einstellung „Kontrast hoch" sagt 7 : 1 für
// jeden Text zu. Aus `a03.mjs` hierher gezogen, damit beide dieselbe Messung
// fahren.
export const texteMessen = (p, wurzel, { klein = 4.5, gross = 3 } = {}) => p.seite(`${rechnung}
  const wurzeln = [...document.querySelectorAll(${JSON.stringify(wurzel)})].filter((w) => w.checkVisibility())
  const mitText = (e) => e.checkVisibility() && [...e.childNodes].some((k) => k.nodeType === 3 && k.textContent.trim() !== '')
  const elemente = wurzeln.flatMap((w) => [w, ...w.querySelectorAll('*')]).filter(mitText)
  return elemente.map((e) => {
    const stil = getComputedStyle(e)
    const groesse = parseFloat(stil.fontSize)
    const istGross = groesse >= 24 || (groesse >= 18.66 && Number(stil.fontWeight) >= 700)
    return messe(e.tagName.toLowerCase() + ' »' + e.textContent.trim().slice(0, 32) + '«', e, 'text', istGross ? ${String(gross)} : ${String(klein)})
  })`)
