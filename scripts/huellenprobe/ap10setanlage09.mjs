// **Ein Set anlegen mit ⌘N, am nativen Fenster** — AP-10
// (#2367),
// Befund B-19, Entwurf `Wege09`: „In einer Schublade, ab dem ersten Set; ⌘N
// öffnet sie und setzt den Fokus in den Namen.“
//
// # Was es prüft
//
// **Mit einem Set**: von der Startseite ⌘N über die Menüleiste — die
// Schublade *Ein neues Set* steht, der Fokus im Namen; getippt, angelegt,
// die Schublade ist zu, das neue Set ist offen und liegt **laut `mse sets`**
// im Bestand. Auf der Setliste steht das Anlegen danach nicht mehr offen
// daneben. **Ohne Set**: das Anlegen steht offen auf der Setliste (der leere
// Einstieg), und ⌘N setzt den Fokus dort in den Namen. Hell und dunkel.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch.
const schublade = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Ein neues Set')"
const offenesAnlegen = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && !s.closest('[role=dialog]') && s.querySelector('h2, h3')?.innerText.trim() === 'Ein neues Set')"
const fokusImNamen = "return (document.activeElement?.labels?.[0]?.innerText ?? '').trim().startsWith('Name des Sets')"
const knopfIn = (wo, name) => `[...(${wo})?.querySelectorAll('button') ?? []].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const ueberschrift = (name) => `return [...document.querySelectorAll('main h1')].some((h) => h.checkVisibility() && h.innerText.trim() === ${JSON.stringify(name)})`

async function zurSeite(p, seite) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(seite)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(seite)})`, `»${seite}« in der Palette`)
}

async function mitSet(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: ⌘N von der Startseite öffnet die Schublade`)
  await p.taste('1', 'befehl')
  await p.warte('die Startseite steht', () => p.seite(ueberschrift('MagicSetEditor4Mac')), 20000)
  await p.taste('n', 'befehl')
  await p.warte('die Schublade „Ein neues Set“ steht', () => p.seite(`return !!${schublade}`), 15000)
  await p.warte('der Fokus steht im Namen', () => p.seite(fokusImNamen), 10000)
  await p.bild(`schublade-${bild}`)

  await p.tippe('Probeneu')
  await p.klickeAuf(`[...(${schublade})?.querySelectorAll('button') ?? []].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Set anlegen'))`, '»Set anlegen«')
  await p.warte('das neue Set ist offen', () => p.seite("return [...document.querySelectorAll('main h1')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Probeneu')"), 20000)
  p.pruefe('die Schublade ist zu', !(await p.seite(`return !!${schublade}`)))
  const sets = JSON.parse(p.mse(ort, 'sets', '--json')).sets.map((satz) => satz.name)
  p.pruefe('das Set liegt laut `mse sets` im Bestand', sets.includes('Probeneu'), sets.join(', '))

  await zurSeite(p, 'Sets')
  await p.warte('die Setliste steht', () => p.seite(ueberschrift('Sets')), 15000)
  p.pruefe('mit Sets steht das Anlegen nicht offen daneben', !(await p.seite(`return !!${offenesAnlegen}`)))
  p.pruefe('⌘Q beendet', await p.beende())
}

async function ohneSet(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: ohne Set steht das Anlegen offen`)
  await zurSeite(p, 'Sets')
  await p.warte('das Anlegen steht offen auf der Setliste', () => p.seite(`return !!${offenesAnlegen}`), 15000)
  p.pruefe('keine Schublade', !(await p.seite(`return !!${schublade}`)))
  await p.taste('n', 'befehl')
  await p.warte('⌘N setzt den Fokus in den Namen', () => p.seite(fokusImNamen), 10000)
  p.pruefe('auch dann keine Schublade', !(await p.seite(`return !!${schublade}`)))
  await p.bild(`leer-${bild}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap10setanlage09(p) {
  for (const bild of ['hell', 'dunkel']) {
    await mitSet(p, bild)
    await ohneSet(p, bild)
  }
}
