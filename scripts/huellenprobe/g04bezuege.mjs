// **Neuere Paketstände sehen und sicher aktualisieren** — G04
// (#2251),
// am nativen Fenster und **gegen GitHub selbst**.
//
// Bezogen wird `MagicSetEditorPacks/Magic-Mini-Packs` — rund 11 MB, die
// kleinste Sammlung des Kontos — an einem **älteren** Commit (`97bb53a`).
// **Der trägt nur seine README**, kein Paket; die bringt erst `c3d9b16`
// (`Deckbox/`, `Scrolls/`). Der erste Lauf am 2026-10-10 war daran rot: ein
// Baum ohne Paket stand nirgends und ließ sich nicht aktualisieren.
// Dann im Fenster unter *Pakete › Bezogene Stände*:
//
//   1. ungefragt: der bezogene Stand, „kein Paket", „nicht in der Liste
//      geprüfter Quellen", „Noch nicht gefragt" — und nichts ging ins Netz,
//   2. „Quellen fragen": die Spitze ist ein Commit (`c3d9b16`) samt Datum,
//      einen Commit voraus,
//   3. „Aktualisierung ansehen": die Vorschau aus GitHubs Vergleich,
//   4. „Auf c3d9b16 aktualisieren": der Baum steht am neuen Commit, der
//      bisherige liegt aufgehoben,
//   5. „Zum aufgehobenen Stand zurück": wieder `97bb53a`.
//
// **Gelesen wird über die Bindungen**, nicht vom Bildschirm allein: was der
// Katalog sagt, ist, was gilt.
//
// **Was er nicht belegt**: eine Vorschau, die eine Bestätigung verlangt — die
// Quelle hier ist voraus, verliert kein Paket, das ein Set nennt, und ihre
// Bäume sind ganz; das fahren die Läufe in
// `frontend/src/seiten/bezugsstaende.test.tsx`.

const adresse = 'https://github.com/MagicSetEditorPacks/Magic-Mini-Packs'
const alt = '97bb53a0e3f8f093cab19f18044c45792625e616'
const neu = 'c3d9b16ecd539a939e448a9fb8c6e5df2a619395'
const binde = (aufruf) => `return await window.go.app.Bindungen.${aufruf}`

async function bezug(p) {
  const antwort = await p.seite(binde('PaketBezuege(false)'))

  return antwort.bezuege.find((b) => b.name === 'Magic-Mini-Packs')
}

export default async function g04bezuege(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')

  p.notiz('0 — bezogen wird ein älterer Commit, über die Kommandozeile')
  p.zeige(p.mse(ort, 'paket-beziehen', adresse, alt, '', 'Magic-Mini-Packs').trim())

  await p.starte({ ort })
  await p.taste('k', 'befehl')
  await p.tippe('Pakete')
  await p.taste('Enter')
  await p.warte('die Paketseite steht da', () => p.seite("return document.querySelector('main')?.dataset.seite === 'pakete'"), 15000)

  p.notiz('1 — ungefragt')
  await p.klickeText('Bezogene Stände')
  await p.wartetext('Bezogen: 97bb53a', 15000)
  const vorher = await bezug(p)
  p.pruefe('der Katalog nennt den bezogenen Stand, kein Paket und keine Spitze', vorher?.stand === alt && vorher?.pakete === 0 && vorher?.spitze === '' && vorher?.lage === 'ungefragt', JSON.stringify(vorher))
  p.pruefe('der Baum ohne Paket sagt „kein Paket"', await p.seite("return [...document.querySelectorAll('main li')].some((e) => e.innerText.includes('Magic-Mini-Packs') && e.innerText.includes('kein Paket'))"))
  p.pruefe('ungefragt steht „Noch nicht gefragt."', await p.seite("return document.querySelector('main').innerText.includes('Noch nicht gefragt.')"))
  p.pruefe('die Quelle steht nicht in der Liste geprüfter Quellen', await p.seite("return document.querySelector('main').innerText.includes('Nicht in der Liste geprüfter Quellen')"))
  await p.bild('g04-ungefragt')

  p.notiz('2 — die Quellen fragen (GitHub)')
  await p.klickeText('Quellen fragen')
  await p.wartetext('Neuer Stand c3d9b16', 30000)
  const satz = await p.seite("return [...document.querySelectorAll('main li p')].map((e) => e.innerText).find((t) => t.startsWith('Neuer Stand')) ?? ''")
  p.pruefe('die Spitze ist ein Commit mit Datum, einen Commit voraus', satz === 'Neuer Stand c3d9b16 vom 7. Oktober 2021 — einen Commit voraus.', satz)
  await p.bild('g04-neuer')

  p.notiz('3 — die Vorschau')
  await p.klickeAuf("[...document.querySelectorAll('main button')].find((b) => b.innerText.startsWith('Aktualisierung ansehen'))", '»Aktualisierung ansehen …«')
  await p.warte('die Vorschau steht da', () => p.seite("return !!document.querySelector('[aria-label], section') && document.querySelector('main').innerText.includes('Aktualisierung von Magic-Mini-Packs')"), 30000)
  const vorschau = await p.seite("return [...document.querySelectorAll('main section')].map((e) => e.innerText).find((t) => t.includes('Aktualisierung von')) ?? ''")
  p.zeige(`Vorschau: ${vorschau.replace(/\s+/g, ' ').slice(0, 400)}`)
  p.pruefe('die Vorschau nennt Von und Zu', vorschau.includes('97bb53a → c3d9b16'), vorschau.slice(0, 120))
  p.pruefe('die Vorschau sagt, dass der bisherige Stand aufgehoben bleibt', vorschau.includes('Der bisherige Stand bleibt aufgehoben'))
  p.pruefe('die Vorschau nennt die Pakete, die kommen', vorschau.includes('magic-deckbox.mse-style'), vorschau.slice(0, 200))
  // **Vollständig, obwohl es 435 Dateien sind**: der Vergleich nennt höchstens
  // 300, und der zweite Lauf am 2026-10-10 blieb an der Bestätigung hängen,
  // die eine gekürzte Liste verlangt. Seitdem kommen die Dateien aus den
  // Bäumen beider Stände.
  p.pruefe('die Vorschau ist vollständig und verlangt keine Bestätigung', !vorschau.includes('gekürzt') && !(await p.seite("return !!document.querySelector('main section input[type=checkbox]')")))
  await p.bild('g04-vorschau')

  p.notiz('4 — aktualisieren')
  await p.klickeAuf("[...document.querySelectorAll('main section button')].find((b) => b.innerText.startsWith('Auf c3d9b16'))", '»Auf c3d9b16 aktualisieren«')
  await p.wartetext('Bezogen: c3d9b16', 60000)
  const danach = await bezug(p)
  p.pruefe('der Katalog nennt den neuen Stand, seine Pakete und den aufgehobenen', danach?.stand === neu && danach?.pakete > 0 && danach?.vorige === true && danach?.vorigerStand === alt, JSON.stringify(danach))
  p.pruefe('die Vorschau ist geschlossen', !(await p.seite("return document.querySelector('main').innerText.includes('Aktualisierung von Magic-Mini-Packs')")))
  await p.bild('g04-aktualisiert')

  p.notiz('5 — zurück')
  await p.klickeAuf("[...document.querySelectorAll('main button')].find((b) => b.innerText.startsWith('Zum aufgehobenen Stand zurück'))", '»Zum aufgehobenen Stand zurück«')
  await p.wartetext('Bezogen: 97bb53a', 30000)
  const zurueck = await bezug(p)
  p.pruefe('wieder der alte Stand ohne Paket, der neuere liegt aufgehoben', zurueck?.stand === alt && zurueck?.pakete === 0 && zurueck?.vorige === true && zurueck?.vorigerStand === neu, JSON.stringify(zurueck))
  p.pruefe('die Seite rollt nicht waagerecht', !(await p.seite('return document.scrollingElement.scrollWidth > window.innerWidth + 1')))
  await p.bild('g04-zurueck')

  p.pruefe('⌘Q beendet', await p.beende())
}
