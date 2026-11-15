// **Die lokale Sitzung am nativen Fenster** — ID-02
// (#1701):
//
// - AK-01: Kennwort setzen, sperren, und **nach ⌘Q und Neustart gesperrt** —
//   der Zustand steht im Bestand, nicht im Arbeitsspeicher,
// - AK-01: gesperrt steht die Anwendung gar nicht da; dass auch keine
//   Bindung den Bestand eines Profils erreicht, belegt
//   `TestGesperrtErreichtNichtsDenBestand` (`internal/app/sitzung_test.go`),
// - AK-01: falsches Kennwort bleibt gesperrt, richtiges entsperrt,
// - AK-01: abmelden, dann als derselbe Mensch mit Kennwort anmelden,
// - AK-02: „Lokal entsperren (Touch ID)" heißt nirgends Passkey.
//
// Die Übergänge selbst belegen `internal/domain/sitzung_test.go`,
// `internal/service/sitzung*_test.go` und `internal/app/sitzung_test.go`.

const kennwort = 'probe-kennwort-lang'

const feld = (beschriftung) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === ${JSON.stringify(beschriftung)})`

async function eintippen(p, beschriftung, text) {
  await p.klickeAuf(feld(beschriftung), `das Feld »${beschriftung}«`)
  await p.tippeInsFeld(text)
}

// ohneEinfuehrung schließt „Erste Schritte", falls die frisch aufgebaute
// Anwendung sie wieder zeigt — wie `bereit()` beim Start.
async function ohneEinfuehrung(p) {
  await new Promise((fertig) => setTimeout(fertig, 800))
  if (await p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility() && d.innerText.includes('Erste Schritte'))")) {
    await p.taste('Escape')
  }
}

const ueberschrift = (p) => p.seite("return document.querySelector('h1')?.innerText.trim() ?? ''")

export default async function id02(p) {
  const ort = p.ort()
  await p.starte({ ort })

  p.notiz('AK-01 — ein Kennwort setzen')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /Nutzer und Profile/.test(b.getAttribute('aria-label') ?? b.innerText))", '»Nutzer und Profile«')
  await p.klickeAuf("[...document.querySelectorAll('[role=tab]')].find((t) => t.checkVisibility() && t.innerText.trim() === 'Konto')", 'die Lasche »Konto«')
  await p.wartetext('Kennwort ändern', 20000)
  // **Seit UXR-20 (#1928) eingeklappt** unter Sperren und Abmelden.
  await p.klickeText('Kennwort ändern')
  await eintippen(p, 'Neues Kennwort', kennwort)
  await p.klickeText('Kennwort setzen')
  await p.wartetext('Das Kennwort ist gesetzt.', 10000)
  p.pruefe('das Kennwort ist gesetzt', true)
  await p.seite("[...document.querySelectorAll('section')].find((s) => s.getAttribute('aria-label') === 'Sitzung')?.scrollIntoView({ block: 'center' })")
  await p.bild('sitzung-auf-der-nutzerseite')

  p.notiz('AK-01 — sperren')
  await p.klickeText('Sperren')
  await p.warte('die Sperranzeige steht da', async () => (await ueberschrift(p)) === 'Gesperrt', 15000)
  p.pruefe('statt der Anwendung steht die Sperranzeige', !(await p.seite("return !!document.querySelector('nav')")))
  p.pruefe('Touch ID heißt nirgends Passkey', !(await p.seite("return /passkey/i.test(document.body.innerText) && !/kein Passkey/i.test(document.body.innerText)")))
  await p.bild('gesperrt')

  p.notiz('AK-01 — ⌘Q und Neustart: gesperrt bleibt gesperrt')
  p.pruefe('⌘Q beendet', await p.beende())
  p.pruefe('an der Kommandozeile steht die Sitzung gesperrt', p.mse(ort, 'sitzung').startsWith('gesperrt'), p.mse(ort, 'sitzung').trim())
  await p.starte({ ort })
  await p.warte('nach dem Neustart steht die Sperranzeige', async () => (await ueberschrift(p)) === 'Gesperrt', 20000)
  p.pruefe('nach dem Neustart gesperrt', true)

  p.notiz('AK-01 — ein falsches Kennwort bleibt gesperrt, das richtige entsperrt')
  await eintippen(p, 'Kennwort', 'falsches-kennwort')
  await p.klickeText('Entsperren')
  await p.wartetext('Dieses Kennwort stimmt nicht.', 10000)
  p.pruefe('das falsche Kennwort steht am Feld', (await ueberschrift(p)) === 'Gesperrt')
  await p.bild('falsches-kennwort')
  // **Das Feld leeren, dann nativ tippen** — geleert über den Wertsetzer,
  // den React beobachtet; ⌘A erreicht das Feld über die Probe nicht.
  await p.seite(`const e = ${feld('Kennwort')}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`)
  await p.klickeAuf(feld('Kennwort'), 'das Feld »Kennwort«')
  await p.tippe(kennwort)
  await p.warte('im Feld steht das richtige Kennwort', () => p.seite(`return ${feld('Kennwort')}?.value === ${JSON.stringify(kennwort)}`))
  await p.klickeText('Entsperren')
  await p.warte('die Anwendung steht wieder da', () => p.seite("return !!document.querySelector('nav')"), 20000)
  p.pruefe('entsperrt steht die Anwendung da', true)
  await ohneEinfuehrung(p)
  await p.bild('entsperrt')

  p.notiz('AK-01 — abmelden und wieder anmelden')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /Nutzer und Profile/.test(b.getAttribute('aria-label') ?? b.innerText))", '»Nutzer und Profile«')
  await p.klickeAuf("[...document.querySelectorAll('[role=tab]')].find((t) => t.checkVisibility() && t.innerText.trim() === 'Konto')", 'die Lasche »Konto«')
  await p.klickeText('Abmelden')
  await p.warte('die Anmeldung steht da', async () => (await ueberschrift(p)) === 'Abgemeldet', 15000)
  await p.bild('abgemeldet')
  await eintippen(p, 'Kennwort', kennwort)
  await p.klickeText('Anmelden')
  await p.warte('die Anwendung steht wieder da', () => p.seite("return !!document.querySelector('nav')"), 20000)
  p.pruefe('angemeldet steht die Anwendung da', p.mse(ort, 'sitzung').startsWith('angemeldet'))

  p.pruefe('⌘Q beendet', await p.beende())
}
