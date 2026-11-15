// **Acht Sprachen und eine eigene, am nativen Fenster** — LOC-01 und LOC-02
// (#1687,
// #1688):
//
// - LOC-01.AK-02/04: gewählt ist Japanisch — die Menüleiste, die Seite und
//   `lang` sprechen es, die Glyphen sind japanisch (Abzug), und nach einem
//   Neustart gilt die Wahl weiter,
// - LOC-02.AK-01/03: eine eigene Übersetzung über Englisch, angelegt an der
//   Kommandozeile — das Fenster zeigt ihren Satz, was fehlt zeigt die Basis,
//   die Menüleiste spricht Englisch, und unter *Einstellungen → Allgemein*
//   steht sie mit ihren Zahlen und lässt sich bearbeiten.

const oben = (leiste) => leiste.filter((p) => !p.trenner && !p.verborgen).map((p) => p.titel)
const ueberschrift = "return document.querySelector('main h1')?.innerText.trim() ?? ''"

export default async function loc01(p) {
  const ort = p.ort()

  p.notiz('1 — Japanisch, gewählt vor dem Start')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'ja')
  await p.starte({ ort })
  await p.warte('das Dokument spricht Japanisch', async () => (await p.seite('return document.documentElement.lang')) === 'ja', 20000)

  const leiste = oben(await p.menue())
  p.zeige(`oben: ${leiste.join(' · ')}`)
  p.pruefe('die Menüleiste spricht Japanisch', leiste.includes('ヘルプ') && leiste.includes('編集'), leiste.join(' · '))
  await p.bild('japanisch')
  await p.leistenbild('japanisch-leiste')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('2 — nach dem Neustart gilt die Wahl weiter')
  await p.starte({ ort })
  await p.warte('wieder Japanisch', async () => (await p.seite('return document.documentElement.lang')) === 'ja', 20000)
  p.pruefe('die Leiste ist wieder japanisch', oben(await p.menue()).includes('ヘルプ'))
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('3 — eine eigene Übersetzung über Englisch')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'sprachpaket', 'anlegen', 'pt-BR', 'Português do Brasil', 'en')
  p.mse(ort, 'sprachpaket', 'setze', 'pt-BR', '--lizenz=CC0-1.0', 'seiten.start.ueberschrift=Olá, MagicSetEditor4Mac')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache.eigene', 'pt-BR')
  await p.starte({ ort })
  await p.warte('das Dokument nennt die eigene', async () => (await p.seite('return document.documentElement.lang')) === 'pt-BR', 20000)
  p.pruefe('ihr Satz steht da', (await p.seite(ueberschrift)) === 'Olá, MagicSetEditor4Mac', await p.seite(ueberschrift))
  const englisch = oben(await p.menue())
  p.pruefe('die Menüleiste spricht die Basis', englisch.includes('Help') && englisch.includes('Edit'), englisch.join(' · '))
  p.pruefe('was fehlt, zeigt die Basis', await p.seite("return document.body.innerText.includes('Settings')"))
  await p.bild('eigene-uebersetzung')

  p.notiz('4 — der Bereich in den Einstellungen')
  await p.taste(',', 'befehl')
  await p.wartetext('Your own translations', 20000)
  // **Seit #2014 (2026-10-02) ein eingeklappter Kasten** (`Aufklappkasten`): bis
  // zum 2026-10-05 suchte dieses Drehbuch seinen Inhalt, ohne ihn
  // aufzuklappen, und fand nur den Knopf.
  const kasten = "[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Your own translations')"
  await p.klickeAuf(kasten, 'der Kasten »Your own translations«')
  await p.warte('der Kasten ist offen', () => p.seite(`return (${kasten})?.getAttribute('aria-expanded') === 'true'`), 10000)
  await p.seite("[...document.querySelectorAll('section')].find((s) => s.getAttribute('aria-label') === 'Your own translations')?.scrollIntoView({ block: 'start' })")
  p.pruefe('sie steht mit ihren Zahlen da', await p.warte('die Übersetzung steht da', () => p.seite("return document.body.innerText.includes('Português do Brasil') && /1 of \\d+ translated/.test(document.body.innerText)"), 10000))
  await p.bild('einstellungen-eigene')

  // **Über `data-aktion` und nicht über den Text** — der Knopf trägt sein
  // Zeichen vor dem Wort.
  await p.klickeAuf("document.querySelector('section[aria-label=\"Your own translations\"] [data-aktion=bearbeiten]')", '»Edit«')
  await p.wartetext('Back to the list', 20000)
  p.pruefe('der Editor zeigt Deutsch, Basis und ein Feld je Satz', await p.seite("return document.body.innerText.includes('German') && document.body.innerText.includes('Base') && [...document.querySelectorAll('textarea')].some((t) => (t.getAttribute('aria-label') ?? '').startsWith('Translation of'))"))
  await p.bild('editor')

  p.pruefe('⌘Q beendet', await p.beende())
}
