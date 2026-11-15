// **Listen ordnen sich in der Sprache des Fensters** — F01
// (#2243),
// am nativen Fenster.
//
// Die Vektoren in `sprache/vergleich.test.ts` laufen unter Node mit dessen
// ICU. **Was WKWebView ordnet, sagt nur WKWebView** — es bringt seine eigene
// ICU mit, die des Systems. Hier stehen vier Sets, zwei kyrillisch, zwei
// lateinisch, nach Namen geordnet:
//
//   - auf **Russisch** zuerst Kyrillisch (Банан, Яблоко, Apple, Banane),
//   - auf **Deutsch** zuerst Lateinisch (Apple, Banane, Банан, Яблоко).
//
// Dieselben Daten, dasselbe System — nur die Sprache des Fensters wechselt.
const namen = ['Яблоко', 'Apple', 'Банан', 'Banane']

const erwartet = {
  ru: ['Банан', 'Яблоко', 'Apple', 'Banane'],
  de: ['Apple', 'Banane', 'Банан', 'Яблоко'],
}

// reihe liest die Namen in der gezeigten Reihenfolge.
const reihe = `return [...document.querySelectorAll('[data-eintrag]')].filter((e) => e.checkVisibility()).map((e) => ${JSON.stringify(namen)}.find((n) => e.innerText.includes(n))).filter(Boolean)`

// nachName stellt die Ordnung der Liste auf »nach Name« — über das native
// Auswahlfeld, dessen Wert React liest.
const nachName = "const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && [...e.options].some((o) => o.value === 'name')); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, 'name'); s.dispatchEvent(new Event('change', { bubbles: true })); return s.value === 'name'"

export default async function f01(p) {
  const ort = p.ort({ zuhause: true })

  for (const name of namen) {
    p.mse(ort, 'set-neu', name, 'probe', 'probe-schlicht')
  }

  for (const sprache of ['ru', 'de']) {
    p.notiz(`Die Sets nach Namen, das Fenster auf ${sprache}`)
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    await p.starte({ ort })

    const lang = await p.warte('das Dokument nennt seine Sprache', () => p.seite('return document.documentElement.lang || null'), 20000)
    p.pruefe(`das Fenster spricht ${sprache}`, lang === sprache, lang)

    await p.taste('2', 'befehl')
    await p.warte('die Sets stehen da', async () => (await p.seite(reihe)).length === namen.length, 20000)
    p.pruefe('die Ordnung steht auf »nach Name«', await p.seite(nachName))

    // **Ohne Befund gewartet** — ob die Ordnung stimmt, sagt die Prüfung
    // darunter einmal; `p.warte` vermerkte einen Fehlschlag zweimal.
    let gezeigt = await p.seite(reihe)
    for (const ende = Date.now() + 10000; gezeigt.join('|') !== erwartet[sprache].join('|') && Date.now() < ende;) {
      await new Promise((weiter) => { setTimeout(weiter, 200) })
      gezeigt = await p.seite(reihe)
    }

    p.zeige(`gezeigt: ${gezeigt.join(', ')}`)
    p.pruefe(`auf ${sprache} ordnet WKWebView wie die Vektoren`, gezeigt.join('|') === erwartet[sprache].join('|'), erwartet[sprache].join(', '))
    await p.bild(`f01-sets-${sprache}`)
    p.pruefe('⌘Q beendet', await p.beende())
  }
}
