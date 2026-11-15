// **Chinesische Zeichenformen in chinesischer Umgebung** — G18 aus dem
// Zweitblick über Phase 6.3 (#1664), Phase 6.10 (#1671).
//
// jsdom löst weder `:lang()` noch Variablen auf; ob WebKit die Reihe wählt,
// sieht man nur am Fenster. Gelesen wird die **berechnete** Schriftkette
// einer Überschrift (Anzeigeschrift) und des Rumpfs (Mantine) — welche
// Han-Schrift zuerst kommt:
//
// - Deutsch: Hiragino vor Songti und PingFang (Kana nur japanisch),
// - vereinfachtes Chinesisch: Songti SC und PingFang SC zuerst,
// - Japanisch: Hiragino zuerst.
//
// Dazu ein **Element mit eigener Sprache** in einer deutschen Oberfläche —
// eingehängt von der Probe, weil die Anwendung heute keines trägt.
//
// **Was es nicht tut**: die Glyphen ansehen. Ob die gewählte Schrift das
// Zeichen auch zeichnet, zeigt das Bild; geprüft wird die Reihenfolge.

// ersteHan gibt die erste Han-Schrift einer berechneten Kette.
const ersteHan = (feld) => `const kette = ${feld}; const treffer = kette.match(/Hiragino[^,]*|Songti [ST]C|PingFang [ST]C/); return { kette, erste: treffer ? treffer[0].replace(/["']/g, '') : '' }`

const ueberschrift = "getComputedStyle(document.querySelector('main h1, main h2, h1, h2')).fontFamily"
const rumpf = 'getComputedStyle(document.body).fontFamily'

const faelle = [
  { sprache: 'de', anzeige: /^Hiragino/, text: /^Hiragino/ },
  { sprache: 'zh', anzeige: /^Songti SC$/, text: /^PingFang SC$/ },
  { sprache: 'ja', anzeige: /^Hiragino/, text: /^Hiragino/ },
]

export default async function g18(p) {
  const ort = p.ort()

  for (const fall of faelle) {
    p.notiz(`Oberfläche ${fall.sprache}`)
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', fall.sprache)
    await p.starte({ ort })
    await p.warte('eine Überschrift steht', () => p.seite("return !!document.querySelector('h1, h2')"), 20000)
    p.zeige(`lang = ${await p.seite('return document.documentElement.lang')}`)
    const anzeige = await p.seite(ersteHan(ueberschrift))
    const text = await p.seite(ersteHan(rumpf))
    p.zeige(`Überschrift: ${anzeige.kette}`)
    p.zeige(`Rumpf: ${text.kette}`)
    p.pruefe(`Überschrift, erste Han-Schrift ${anzeige.erste} (${fall.sprache})`, fall.anzeige.test(anzeige.erste), anzeige.erste)
    p.pruefe(`Rumpf, erste Han-Schrift ${text.erste} (${fall.sprache})`, fall.text.test(text.erste), text.erste)

    if (fall.sprache === 'de') {
      // **Ein Element mit eigener Sprache** — die Regel gilt jedem Element,
      // nicht nur der Wurzel.
      const eigen = await p.seite(`const h = document.createElement('h2'); h.lang = 'zh-Hant'; h.textContent = '測試'; h.id = 'g18-probe'; document.querySelector('main').prepend(h); ${ersteHan("getComputedStyle(h).fontFamily")}`)
      p.zeige(`h2 mit lang=zh-Hant: ${eigen.kette}`)
      p.pruefe('ein Element mit lang=zh-Hant wählt Songti TC', eigen.erste === 'Songti TC', eigen.erste)
      await p.seite("document.getElementById('g18-probe')?.remove(); return true")
    }

    await p.bild(`g18-${fall.sprache}`)
    p.pruefe(`⌘Q beendet die Anwendung (${fall.sprache})`, await p.beende())
  }
}
