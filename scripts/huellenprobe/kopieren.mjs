// **Karten kopieren und in ein anderes Set einsetzen** — ENG-02.AK-04,
// Kernweg 6 (#1727).
//
// **Eingesetzt wird in ein leeres Set** — über »Einfügen« im Leerzustand.
// Den gab es vorher nicht: die Handlungszeile und ⌘V hängen an der
// Kartenspalte, und die steht erst ab der ersten Karte da. Das war der Fund
// dieses Drehbuchs (`Kartenleer.tsx`).
//
// **Über die Ersatzablage des Abnahmebaus** (`internal/fenster/abnahme.go`,
// `ersatzablage`): die Zwischenablage dieses Rechners fasst die Probe nicht
// an — sie überschriebe, was ein Mensch kopiert hat. Belegt ist damit der Weg
// vom Knopf über Bindung, Anwendung und Bestand bis zur Naht und zurück,
// **nicht `NSPasteboard` selbst**; das ist Wails' `Clipboard`, eine Zeile in
// `internal/fenster/wails.go`.

// **Nur Kartenzeilen** (Kennung, Nummer, Name) — für ein leeres Set gibt die
// Kommandozeile statt keiner Zeile einen Hinweis aus.
const kartenVon = (p, ort, set) => p.mse(ort, 'karten', set).split('\n').filter((z) => /^\d+\t/.test(z))

export default async function kopieren(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  // **Mit dem Stil des Quellsets** — ein neues Set braucht einen, und
  // dieselbe Gestalt nimmt dieselben Felder auf.
  p.mse(ort, 'set-neu', 'Zielset', 'probe', 'probe-kunst')
  const sets = p.mse(ort, 'sets').trim().split('\n')
  const kennung = (name) => sets.find((z) => z.includes(name))?.split(/\s+/)[0] ?? ''
  const quelle = kennung('Werkstatt')
  const ziel = kennung('Zielset')
  p.zeige(`Quelle ${quelle}, Ziel ${ziel}; im Ziel vorher ${String(kartenVon(p, ort, ziel).length)} Karten`)
  await p.starte({ ort })

  p.notiz('Die Karten anhaken und kopieren')
  await p.taste('2', 'befehl')
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('die Kartenliste steht', async () => (await p.seite("return document.querySelectorAll('main input[type=checkbox]').length")) >= 5, 20000)
  await p.klick('Set', 'Alle Karten anhaken')
  // **Die Ablage steht im Menü „Diese Karte"** über der Karte
  // (`Kartenhandlungen.tsx`) — seit UXR-06 (#1914); vorher im Kasten
  // „Ablage und Bezug".
  await p.klickeText('Diese Karte')
  // **Der Eintrag trägt ein Zeichen vor dem Wort und das Kürzel dahinter**
  // (»⎘ Kopieren (5) ⌘C«).
  await p.klickeAuf("[...document.querySelectorAll('[role=menuitem]')].find((b) => b.checkVisibility() && b.innerText.includes('Kopieren (5)'))", '»Kopieren (5)«')
  const text = await p.warte('die Ablage trägt etwas', async () => {
    const jetzt = await p.frage({ was: 'zwischenablage' })
    return jetzt !== '' ? jetzt : false
  })
  p.pruefe('»Kopieren (5)« legt die Karten in die Ablage', text.includes('Langtext') && text.includes('Ohne Bild'), `${String(text.length)} Zeichen`)

  p.notiz('Im anderen Set einsetzen')
  const vorher = kartenVon(p, ort, ziel)
  await p.taste('2', 'befehl')
  // **Erst, wenn die Liste steht** — seit UX-11 ordnet sie sich beim Laden
  // noch um, und ein Klick auf die gemessene Stelle traf das andere Set.
  await p.warte('beide Sets stehen in der Liste', () => p.seite("return document.body.innerText.includes('Zielset') && document.body.innerText.includes('Werkstatt-Testkarten')"), 20000)
  await new Promise((los) => setTimeout(los, 1000))
  await p.klickeText('Zielset')
  await p.warte('das Zielset ist offen', () => p.seite("return (document.querySelector('nav ol')?.innerText ?? '').includes('Zielset')"), 20000)
  // **Der Leerzustand bietet »Einfügen« an** — der Knopf trägt kein Zeichen,
  // anders als die in der Handlungszeile.
  const einfuegen = "[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Einfügen')"
  await p.warte('der Leerzustand bietet »Einfügen« an', () => p.seite(`return document.body.innerText.includes('Noch keine Karte') && !!(${einfuegen})`), 20000)
  await p.klickeAuf(einfuegen, '»Einfügen« im Leerzustand')
  const nachher = await p.warte('die Karten stehen im Zielset', () => {
    const jetzt = kartenVon(p, ort, ziel)
    return jetzt.length === vorher.length + 5 ? jetzt : false
  }, 20000).catch(async (grund) => {
    p.zeige(`im Zielset: ${JSON.stringify(kartenVon(p, ort, ziel))}`)
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).filter(Boolean).slice(0, 6)"))}`)
    p.zeige(`Brotkrume: ${await p.seite("return document.querySelector('nav ol')?.innerText.replace(/\\s+/g, ' ') ?? ''")}`)
    throw grund
  })
  p.pruefe('»Einfügen« legt die Karten im anderen Set an', nachher.some((z) => z.includes('Langtext')) && nachher.some((z) => z.includes('Ohne Bild')), `${String(nachher.length - vorher.length)} Karten`)
  p.pruefe('und das Quellset bleibt, wie es war', kartenVon(p, ort, quelle).length >= 5)
  p.pruefe('⌘Q beendet', await p.beende())
}
