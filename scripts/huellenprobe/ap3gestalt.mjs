// **Ein Hauptgewicht, eine Lasche, ein gewählter Stand — am nativen Fenster**
// — AP-3 (#2360),
// Befunde B-08 und B-10, Entwurf `Zustaende09`.
//
// # Was es prüft, hell und dunkel
//
// - **Gewählt ist nicht gefüllt**: das gewählte Segment der Ansicht auf
//   *Sets* (seit AP-6, #2363; vorher ein Umschalter) und »Normal« im Kartenformular tragen den Rand der Marke und den
//   Grund `betont` — **nicht** den Grund des Hauptknopfs, den »Set anlegen« in der
//   Schublade der Setanlage trägt. Und ein ungedrückter Umschalter sieht anders aus als ein
//   gedrückter. Gemessen an den berechneten Farben, nicht an Klassen.
// - **Der Fokus bleibt unterscheidbar**: mit Tab erreicht, trägt der
//   gedrückte Umschalter und die gewählte Lasche den Ring (`outline`), und der
//   Ring ist ein anderer als der Rand der Auswahl.
// - **Eine Lasche**: die Laschen des Setarbeitsplatzes und der Schublade
//   *Eigenschaften* tragen dieselbe Unterlinie — Breite, Art, Farbe der
//   gewählten, Schriftschnitt —, die gewählte in der Marke, die anderen ohne.
// - **Ein Pfeil**: das native Feld »Sprache« und die Auswahl mit Suche
//   »Zeitzone« in den Einstellungen tragen denselben berechneten Pfeil.
//
// # Was es nicht belegt
//
// **Wie viele gefüllte Knöpfe eine Sicht zeigt** — das zählen `uirevue` und
// `uirevue09zustaende` (R1 je Blick). **Ob es gut aussieht** — die Bilder
// sieht ein Mensch an. **VoiceOver** — die Ansage von »gedrückt« und
// »ausgewählt« ist eine Handprüfung.

import { pruefeKontraste } from '../huellenkontrast.mjs'

const sichtbar = '(e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })'
const knopfMit = (name, wo = 'document') => `[...${wo}.querySelectorAll('button')].find((b) => (${sichtbar})(b) && (b.getAttribute('aria-label') ?? b.innerText).trim().replace(/\\s+/g, ' ') === ${JSON.stringify(name)})`

// stil liest, was ein Mensch an einem Element sieht.
const stil = (wer) => `const e = ${wer}; if (!e) return null; const s = getComputedStyle(e)
  return { bg: s.backgroundColor, rand: s.borderTopColor, unten: s.borderBottomColor, untenBreite: s.borderBottomWidth, untenArt: s.borderBottomStyle,
    schnitt: s.fontWeight, ring: s.outlineStyle, ringFarbe: s.outlineColor, pfeil: s.backgroundImage, lage: s.backgroundPosition, groesse: s.backgroundSize,
    gedrueckt: e.getAttribute('aria-pressed'), gewaehlt: e.getAttribute('aria-selected') }`

// marke ist die Farbe der Marke, aufgelöst wie jede andere.
const marke = "const s = document.createElement('span'); s.style.color = 'var(--color-marke)'; document.body.appendChild(s); const f = getComputedStyle(s).color; s.remove(); return f"

// perTab drückt Tab, bis das Element den Fokus hat — der Weg eines Menschen,
// und der einzige, auf dem WebKit `:focus-visible` setzt. **Vom Anfang der
// Seite und mit mindestens einem Tab**: der erste Lauf fand den Umschalter
// schon fokussiert — von einem Klick, und ohne Ring. **Gelesen wird im selben
// Aufruf, in dem der Fokus ankommt**, und noch einmal danach: der dritte Lauf
// las den Ring erst nach einer Pause und fand ihn nicht. **Und nach jedem
// Tab eine Pause, bevor gelesen wird**: die Taste kommt als Ereignis in die
// Warteschlange der Anwendung, und der vierte Lauf las schneller, als sie
// ankam — ein Tab lag noch an, und der Fokus stand 250 ms später beim
// nächsten Knopf.
//
// **Und das Fenster vorn**: ohne Fokus des Fensters setzt WebKit weder
// `:focus` noch `:focus-visible`.
async function perTab(p, wer, name) {
  await p.aktiviere()
  await p.seite('document.activeElement?.blur(); return true')
  for (let schritt = 0; schritt < 80; schritt++) {
    await p.taste('Tab')
    await new Promise((los) => setTimeout(los, 150))
    const da = await p.seite(`const e = ${wer}; if (!e || document.activeElement !== e) return null; const s = getComputedStyle(e)
      return { fenster: document.hasFocus(), fokus: e.matches(':focus'), sichtbar: e.matches(':focus-visible'), ring: s.outlineStyle, breite: s.outlineWidth, abstand: s.outlineOffset, farbe: s.outlineColor, rand: s.borderTopColor, randBreite: s.borderTopWidth }`)
    if (da !== null) {
      await new Promise((los) => setTimeout(los, 250))
      const danach = await p.seite(`const a = document.activeElement; return { gleich: a === (${wer}), wer: a ? a.tagName.toLowerCase() + ' ' + (a.getAttribute('aria-label') ?? a.innerText ?? '').trim().slice(0, 40) : 'keiner', fenster: document.hasFocus() }`)
      p.zeige(`${name}: nach ${String(schritt + 1)} Tabs ${JSON.stringify(da)}; 250 ms später ${JSON.stringify(danach)}`)
      return da
    }
  }
  p.zeige(`${name}: mit 80 Tabs nicht erreicht`)
  return null
}

// segmentMit gibt das sichtbare Feld eines Segments — **seit AP-6 (#2363)
// ist die Ansicht der Sets ein Segment** und kein Paar Umschalter; gemessen
// wird dieselbe Frage am Feld, das man sieht.
const segmentMit = (name) => `[...document.querySelectorAll('[role=radiogroup] label')].find((l) => (${sichtbar})(l) && l.innerText.trim() === ${JSON.stringify(name)})?.querySelector('span')`
const segmentStil = (wer) => `const e = ${wer}; if (!e) return null; const s = getComputedStyle(e)
  return { bg: s.backgroundColor, rand: s.borderTopColor, gedrueckt: String(e.closest('label').querySelector('input').checked) }`

// perTabSegment drückt Tab, bis das Optionsfeld des Segments den Fokus hat,
// und liest den Ring am Beschriftungsfeld — dort hängt er (`focus-within`,
// `bausteine/Segmente.tsx`), denn das Optionsfeld selbst ist unsichtbar.
async function perTabSegment(p, wer, name) {
  await p.aktiviere()
  await p.seite('document.activeElement?.blur(); return true')
  for (let schritt = 0; schritt < 80; schritt++) {
    await p.taste('Tab')
    await new Promise((los) => setTimeout(los, 150))
    const da = await p.seite(`const e = ${wer}; const l = e?.closest('label'); if (!l || document.activeElement !== l.querySelector('input')) return null; const s = getComputedStyle(l)
      return { fenster: document.hasFocus(), fokus: true, ring: s.outlineStyle, breite: s.outlineWidth, abstand: s.outlineOffset, farbe: s.outlineColor, randBreite: getComputedStyle(e).borderTopWidth }`)
    if (da !== null) {
      p.zeige(`${name}: nach ${String(schritt + 1)} Tabs ${JSON.stringify(da)}`)
      return da
    }
  }
  p.zeige(`${name}: mit 80 Tabs nicht erreicht`)
  return null
}

// ringAmSegment ist `ring` ohne `:focus-visible` — das Segment zeigt seinen
// Ring an `focus-within`, gemessen in `uxrf2`.
const ringAmSegment = (f) => f !== null && f.ring !== 'none' && parseFloat(f.breite) >= 2 && parseFloat(f.abstand) >= 1 && parseFloat(f.breite) > parseFloat(f.randBreite)

// ring sagt, ob ein Fokusring dasteht, der sich vom Rand der Auswahl
// unterscheidet — **an Breite und Abstand, nicht an der Farbe**: in
// Pergament dunkel ist der Ring so gelb wie die Marke, und damit wie der Rand
// eines gewählten Segments.
const ring = (f) => f !== null && f.sichtbar === true && f.ring !== 'none' && parseFloat(f.breite) >= 2 && parseFloat(f.abstand) >= 1 && parseFloat(f.breite) > parseFloat(f.randBreite)

export default async function ap3gestalt(p) {
  for (const bild of ['hell', 'dunkel']) {
    const ort = p.ort()
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    await p.starte({ ort })
    await p.groesse(1440, 900)
    const farbeMarke = await p.seite(marke)

    p.notiz(`${bild}: Sets — das gewählte Segment gegen »Set anlegen«`)
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    const kacheln = await p.seite(`return (${segmentMit('Kacheln')})?.closest('label').querySelector('input').checked === true`)
    // **Der gewählte ist der, der gilt** — welcher es ist, hängt am Bestand.
    const gedrueckt = kacheln ? 'Kacheln' : 'Liste'
    const ungedrueckt = kacheln ? 'Liste' : 'Kacheln'
    const an = await p.seite(segmentStil(segmentMit(gedrueckt)))
    const aus = await p.seite(segmentStil(segmentMit(ungedrueckt)))
    // **Der Hauptknopf steht in der Schublade der Setanlage** — seit AP-10
    // (#2367) geht sie ab dem ersten Set dort auf (⌘N), und ihr Hauptknopf
    // heißt »Set anlegen«; vorher stand »Weiter« auf der Seite, und dieser
    // Lauf war seitdem rot. Gemessen wird dieselbe Frage: gefüllt mit der
    // Marke gegen das gewählte Segment.
    const setanlage = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Ein neues Set')"
    const anlegen = `[...((${setanlage})?.querySelectorAll('button') ?? [])].find((b) => (${sichtbar})(b) && b.innerText.trim().endsWith('Set anlegen'))`
    await p.taste('n', 'befehl')
    await p.warte('die Setanlage ist offen', () => p.seite(`return !!${anlegen}`), 15000)
    const weiter = await p.seite(stil(anlegen))
    await p.taste('Escape')
    await p.warte('die Setanlage ist zu', async () => !(await p.seite(`return !!${setanlage}`)), 10000)
    p.zeige(`${bild}: »${gedrueckt}« ${JSON.stringify(an)}`)
    p.zeige(`${bild}: »${ungedrueckt}« ${JSON.stringify(aus)}`)
    p.zeige(`${bild}: »Set anlegen« ${JSON.stringify(weiter)}; Marke ${farbeMarke}`)
    p.pruefe(`${bild}: »Set anlegen« ist gefüllt mit der Marke`, weiter?.bg === farbeMarke, weiter?.bg)
    p.pruefe(`${bild}: das gewählte Segment »${gedrueckt}« ist nicht gefüllt wie »Set anlegen«`, an?.gedrueckt === 'true' && an.bg !== weiter?.bg, `${an?.bg} gegen ${weiter?.bg}`)
    p.pruefe(`${bild}: der gedrückte Umschalter trägt den Rand der Marke`, an?.rand === farbeMarke, an?.rand)
    p.pruefe(`${bild}: gedrückt und ungedrückt sehen verschieden aus`, aus?.gedrueckt === 'false' && (aus.bg !== an?.bg || aus.rand !== an?.rand), `${aus?.bg}/${aus?.rand} gegen ${an?.bg}/${an?.rand}`)
    const fokus = await perTabSegment(p, segmentMit(gedrueckt), `»${gedrueckt}«`)
    p.pruefe(`${bild}: mit Tab erreicht trägt »${gedrueckt}« einen Fokusring außerhalb seines Rands`, ringAmSegment(fokus), JSON.stringify(fokus))
    // **Der Rand trennt das Segment von seiner Umgebung** (WCAG 1.4.11):
    // gemessen gegen den Grund außerhalb.
    await pruefeKontraste(p, bild, [[`Rand von »${gedrueckt}« gewählt`, segmentMit(gedrueckt), 'rand', 3]])
    await p.bild(`sets-umschalter-${bild}`)

    p.notiz(`${bild}: das Set — Laschen des Arbeitsplatzes, »Normal« im Kartenformular`)
    await p.klickeText('Werkstatt-Testkarten')
    await p.warte('das Set ist offen', () => p.seite("return !!document.querySelector('[role=tablist] [role=tab][aria-selected=true]')"), 20000)
    const laschen = "[...document.querySelectorAll('main [role=tablist]:not([aria-orientation=vertical]) [role=tab]')]"
    const arbeitsplatz = await p.seite(`return ${laschen}.filter(${sichtbar}).map((e) => { const s = getComputedStyle(e); return { name: e.innerText.trim(), gewaehlt: e.getAttribute('aria-selected'), unten: s.borderBottomColor, breite: s.borderBottomWidth, art: s.borderBottomStyle, schnitt: s.fontWeight, bg: s.backgroundColor } })`)
    const gewaehlteLasche = arbeitsplatz.find((l) => l.gewaehlt === 'true')
    const andere = arbeitsplatz.filter((l) => l.gewaehlt !== 'true')
    p.zeige(`${bild}: gewählte Lasche ${JSON.stringify(gewaehlteLasche)}; ${String(andere.length)} weitere, erste ${JSON.stringify(andere[0])}`)
    p.pruefe(`${bild}: die gewählte Lasche trägt die Unterlinie der Marke`, gewaehlteLasche?.unten === farbeMarke && gewaehlteLasche.breite === '2px' && gewaehlteLasche.art === 'solid', JSON.stringify(gewaehlteLasche))
    p.pruefe(`${bild}: keine andere Lasche trägt sie`, andere.length > 0 && andere.every((l) => l.unten !== farbeMarke), andere.map((l) => `${l.name}: ${l.unten}`).join('; '))
    p.pruefe(`${bild}: keine Lasche ist gefüllt`, arbeitsplatz.every((l) => l.bg !== farbeMarke), arbeitsplatz.map((l) => l.bg).join(', '))
    await p.warte('die Karte steht mit Bild', () => p.seite("return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"), 30000)
    await p.warte('»Normal« steht', () => p.seite(`return !!${knopfMit('Normal', "document.querySelector('main')")}`), 15000).catch(() => false)
    const normal = await p.seite(stil(knopfMit('Normal', "document.querySelector('main')")))
    p.zeige(`${bild}: »Normal« ${JSON.stringify(normal)}`)
    p.pruefe(`${bild}: »Normal« ist gedrückt und nicht gefüllt`, normal?.gedrueckt === 'true' && normal.bg !== farbeMarke && normal.rand === farbeMarke, JSON.stringify(normal))
    const gewaehlt = `${laschen}.find((e) => e.getAttribute('aria-selected') === 'true')`
    const laschenfokus = await perTab(p, gewaehlt, 'die gewählte Lasche')
    p.pruefe(`${bild}: mit Tab erreicht trägt die gewählte Lasche den Fokusring`, laschenfokus !== null && laschenfokus.sichtbar === true && laschenfokus.ring !== 'none', JSON.stringify(laschenfokus))
    await pruefeKontraste(p, bild, [
      ['Unterlinie der gewählten Lasche', gewaehlt, 'rand', 3, 'Bottom'],
      ['Schrift der gewählten Lasche', gewaehlt, 'text', 4.5],
      ['Schrift einer anderen Lasche', `${laschen}.find((e) => e.getAttribute('aria-selected') === 'false' && e.checkVisibility())`, 'text', 4.5],
    ])
    await p.bild(`set-laschen-${bild}`)

    p.notiz(`${bild}: Schublade »Eigenschaften« — dieselbe Lasche`)
    await p.klickeAuf(knopfMit('Eigenschaften'), '»Eigenschaften«')
    const schublade = "[...document.querySelectorAll('[role=dialog] [role=tab]')]"
    await p.warte('die Schublade steht', () => p.seite(`return ${schublade}.some(${sichtbar})`), 15000)
    const inSchublade = await p.seite(`return ${schublade}.filter(${sichtbar}).map((e) => { const s = getComputedStyle(e); return { gewaehlt: e.getAttribute('aria-selected'), unten: s.borderBottomColor, breite: s.borderBottomWidth, art: s.borderBottomStyle, schnitt: s.fontWeight } })`)
    const dort = inSchublade.find((l) => l.gewaehlt === 'true')
    p.zeige(`${bild}: gewählte Lasche der Schublade ${JSON.stringify(dort)}`)
    p.pruefe(`${bild}: Schublade und Arbeitsplatz zeichnen die gewählte Lasche gleich`, dort !== undefined && gewaehlteLasche !== undefined && dort.unten === gewaehlteLasche.unten && dort.breite === gewaehlteLasche.breite && dort.art === gewaehlteLasche.art && dort.schnitt === gewaehlteLasche.schnitt, `${JSON.stringify(dort)} gegen ${JSON.stringify(gewaehlteLasche)}`)
    await p.bild(`schublade-laschen-${bild}`)
    await p.taste('Escape')

    p.notiz(`${bild}: Einstellungen — »Sprache« nativ gegen »Zeitzone« mit Suche`)
    await p.taste(',', 'befehl')
    const sprache = "document.querySelector('select[data-einstellung=sprache]')"
    const zeitzone = "(() => { const l = [...document.querySelectorAll('main label')].find((x) => x.innerText.trim().startsWith('Zeitzone')); return l ? (document.getElementById(l.htmlFor) ?? l.querySelector('input')) : [...document.querySelectorAll('main [role=combobox]')].find((x) => x.checkVisibility()) })()"
    await p.warte('»Sprache« steht', () => p.seite(`return !!${sprache}?.checkVisibility()`), 15000)
    await p.warte('»Zeitzone« steht', () => p.seite(`return !!${zeitzone}?.checkVisibility()`), 15000).catch(() => false)
    const nativ = await p.seite(stil(sprache))
    const mitSuche = await p.seite(stil(zeitzone))
    // **Das Bild gekürzt**: die Faser des Pergaments steht als ganze SVG darin.
    const kurz = (w) => (w ?? '').replace(/url\("data:.*?"\)/g, 'url(Faser)')
    p.zeige(`${bild}: »Sprache« ${kurz(nativ?.pfeil)} · ${nativ?.lage} · ${nativ?.groesse}`)
    p.zeige(`${bild}: »Zeitzone« ${kurz(mitSuche?.pfeil)} · ${mitSuche?.lage} · ${mitSuche?.groesse}`)
    // **Ein Pfeil ist ein Verlauf**: der erste Lauf verglich nur das Bild —
    // und beide trugen dieselbe Faser, also war er grün, ohne einen Pfeil.
    const pfeil = (w) => (w?.pfeil ?? '').includes('linear-gradient(45deg')
    p.pruefe(`${bild}: beide Auswahlen tragen denselben Pfeil`, pfeil(nativ) && pfeil(mitSuche) && nativ.pfeil === mitSuche.pfeil && nativ.lage === mitSuche.lage && nativ.groesse === mitSuche.groesse, `${kurz(nativ?.pfeil)} gegen ${kurz(mitSuche?.pfeil)}`)
    p.pruefe(`${bild}: die Faser des Pergaments bleibt unter dem Pfeil`, [nativ, mitSuche].every((w) => (w?.pfeil ?? '').includes('Pergaments')), `${kurz(nativ?.pfeil)} · ${kurz(mitSuche?.pfeil)}`)
    // **Die Farbe des Pfeils gegen den Grund des Felds** (WCAG 1.4.11, 3 : 1):
    // ein Verlauf lässt sich nicht als Farbe lesen. Gemessen wird deshalb eine
    // Probe mit der Farbe des Pfeils (`--color-leise`) in einem Kasten mit den
    // Klassen des Felds — derselbe Grund, dieselbe Faser darunter.
    await p.seite(`for (const [k, f] of [['nativ', ${sprache}], ['suche', ${zeitzone}]]) {
      document.getElementById('ap3pfeil-' + k)?.remove()
      const d = document.createElement('div'); d.id = 'ap3pfeil-' + k; d.className = f.className
      d.style.cssText = 'position:fixed;left:10px;top:10px;width:120px;height:30px'
      const s = document.createElement('span'); s.style.color = 'var(--color-leise)'; s.textContent = '▾'
      d.appendChild(s); f.parentElement.appendChild(d)
    } return true`)
    await pruefeKontraste(p, bild, [
      ['Pfeil im nativen Feld', "document.querySelector('#ap3pfeil-nativ span')", 'text', 3],
      ['Pfeil in der Auswahl mit Suche', "document.querySelector('#ap3pfeil-suche span')", 'text', 3],
    ])
    await p.seite("document.getElementById('ap3pfeil-nativ')?.remove(); document.getElementById('ap3pfeil-suche')?.remove(); return true")
    const zeichen = await p.seite("return [...document.querySelectorAll('main *')].some((e) => e.children.length === 0 && e.textContent === '▾' && e.checkVisibility())")
    p.pruefe(`${bild}: kein eigenes Zeichen »▾« neben einer Auswahl`, !zeichen)
    await p.bild(`einstellungen-auswahl-${bild}`)
    p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())
  }
}
