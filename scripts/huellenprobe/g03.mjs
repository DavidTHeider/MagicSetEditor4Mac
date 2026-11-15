// **Im Kartentext prüft das System, und ersetzt wird nichts von selbst** — G03
// (#2250,
// Vermerk 0183), am nativen Fenster:
//
//   1. ein Set mit einer Karte, nativ angelegt,
//   2. das Feld »text« trägt `autocorrect="off"`, `autocapitalize="off"` und
//      prüft (`spellcheck`), und WKWebView kennt die Eigenschaft
//      `autocorrect` überhaupt,
//   3. **echte Tasten** mit zwei Tippfehlern, die macOS gewöhnlich berichtigt,
//      in das Feld der Karte — und zur **Gegenprobe** in ein eingefügtes Feld
//      ohne das Attribut. Berichtigt das System dort und hier nicht, beachtet
//      WKWebView das Attribut. Berichtigt es auch dort nicht, sagt die Messung
//      nichts über das Attribut, und das steht so im Beleg.
//
// **Die Systemeinstellung wird nicht angefasst** — gelesen wird nur, was sie
// bewirkt.

// feldNamens sucht ein Eingabefeld der Karte über seinen Feldnamen.
const feldNamens = (name) => `[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === ${JSON.stringify(name)})`

const tippfehler = 'teh wiht '

export default async function g03(p) {
  const ort = p.ort()
  await p.starte({ ort })

  p.notiz('1 — ein Set mit einer Karte')
  await p.klick('Ablage', 'Neu')
  await p.warte('»Neu« fragt nach dem Namen', () => p.seite("return document.activeElement?.labels?.[0]?.innerText.includes('Name des Sets') ?? false"))
  await p.tippeInsFeld('Schreibprobe')
  await p.klickeText('+ Set anlegen')
  await p.wartetext('Schreibprobe')
  await p.klickeText('+ Karte anlegen')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.startsWith('Ohne Namen'))", 'die neue Karte')

  p.notiz('2 — die Attribute des Feldes »text«')
  await p.warte('das Feld »text« steht da', () => p.seite(`return !!(${feldNamens('text')})`), 20000)
  const angaben = await p.seite(`const f = ${feldNamens('text')}; return { attribut: f.getAttribute('autocorrect'), gross: f.getAttribute('autocapitalize'), eigenschaft: ('autocorrect' in HTMLElement.prototype) ? f.autocorrect : 'unbekannt', prueft: f.spellcheck, art: f.tagName }`)
  p.zeige(JSON.stringify(angaben))
  p.pruefe('das Feld trägt autocorrect="off"', angaben.attribut === 'off')
  p.pruefe('und autocapitalize="off"', angaben.gross === 'off')
  p.pruefe('und es prüft weiter (spellcheck)', angaben.prueft === true)
  p.zeige(`WKWebView kennt die Eigenschaft autocorrect: ${angaben.eigenschaft === 'unbekannt' ? 'nein' : `ja (${angaben.eigenschaft})`}`)

  p.notiz('3 — echte Tasten: das Feld der Karte und ein Feld ohne das Attribut')
  // **Gelesen wird, wenn alles angekommen ist** — an der Länge und nicht am
  // Wortlaut, denn eine Berichtigung („teh" → „the") ändert ihn, die Länge
  // nicht. Der erste native Lauf (2026-10-10) las das Feld, als erst das »t«
  // verarbeitet war; im Bild stand »teh wiht« in beiden Feldern.
  await p.klickeAuf(feldNamens('text'), 'das Feld »text«')
  await p.warte('das Feld »text« hat den Fokus', () => p.seite(`return document.activeElement === (${feldNamens('text')})`), 10000)
  await p.tippe(tippfehler)
  await p.warte('alles ist im Feld »text« angekommen', () => p.seite(`return (${feldNamens('text')})?.value.length === ${String(tippfehler.length)}`), 15000)
  await p.taste('Tab')
  const imFeld = await p.seite(`return ${feldNamens('text')}?.value ?? null`)

  await p.seite("const v = document.createElement('textarea'); v.id = 'g03-vergleich'; v.setAttribute('aria-label', 'Vergleich'); v.style.cssText = 'position:fixed;top:8px;left:8px;width:300px;height:60px;z-index:99999'; document.body.append(v); return true")
  await p.klickeAuf("document.getElementById('g03-vergleich')", 'das Vergleichsfeld ohne Attribut')
  await p.warte('das Vergleichsfeld hat den Fokus', () => p.seite("return document.activeElement?.id === 'g03-vergleich'"), 10000)
  await p.tippe(tippfehler)
  await p.warte('alles ist im Vergleichsfeld angekommen', () => p.seite(`return document.getElementById('g03-vergleich')?.value.length === ${String(tippfehler.length)}`), 15000)
  await p.taste('Tab')
  const imVergleich = await p.seite("return document.getElementById('g03-vergleich')?.value ?? null")
  await p.bild('g03-tippfehler')
  await p.seite("document.getElementById('g03-vergleich')?.remove(); return true")

  p.zeige(`im Feld der Karte: ${JSON.stringify(imFeld)} · im Vergleichsfeld: ${JSON.stringify(imVergleich)}`)
  p.pruefe('im Feld der Karte steht, was getippt wurde', imFeld === tippfehler)

  if (imVergleich === tippfehler) {
    p.zeige('Das System hat auch im Vergleichsfeld nichts berichtigt — über das Attribut sagt diese Messung nichts (Systemeinstellung, Sprache oder Wörter).')
  } else {
    p.pruefe('das System berichtigt ohne das Attribut, und mit ihm nicht', imVergleich !== tippfehler && imFeld === tippfehler, imVergleich)
  }

  p.pruefe('⌘Q beendet', await p.beende())
}
