// **Betrieb und Vertrauen, am nativen Fenster** — OPS-07
// (#1724):
//
// - AK-01/03: in der Hilfe steht ein Bereich mit Fassung, Datenstand,
//   Signatur — ad-hoc, **nicht notarisiert** —, Aktualisierung und Orten,
// - AK-02: der Prüfstand trennt Geprüftes von Vorhandenem und nennt je Zeile
//   einen Beleg,
// - NAT-01: der Datenschutz nennt Wege, die es gibt.

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

export default async function ops07(p) {
  await p.starte()

  // **Über das Menü**, wie in `eintraege.mjs`: der Eintrag der Leiste trägt
  // ein Zeichen vor dem Wort.
  await p.klick('Hilfe', 'MagicSetEditor4Mac-Hilfe')
  await p.warte('die Hilfe steht da', async () => (await hier(p)) === 'Hilfe', 20000)
  await p.klickeText('Betrieb und Vertrauen')
  await p.wartetext('Was geprüft ist', 20000)

  p.notiz('AK-01/03 — Fassung, Daten, Signatur, Aktualisierung, Orte')
  const text = await p.seite('return document.body.innerText')
  p.pruefe('die Fassung mit ihrem Bau', /\d+\.\d+\.\d+, gebaut aus [0-9a-f]{7}/.test(text))
  p.pruefe('der Datenstand', /\d+ Migrationen verbucht, zuletzt \d{4}_/.test(text))
  p.pruefe('ad-hoc und nicht notarisiert', text.includes('Ad-hoc signiert, nicht notarisiert'))
  p.pruefe('die Aktualisierung', text.includes('Aus — diese Anwendung fragt nicht nach neuen Fassungen.'))
  p.pruefe('die Orte als Tilde', /Zwischenspeicher\s+~\/Library\/Caches\/MagicSetEditor4Mac/.test(text))

  p.notiz('AK-02 — Geprüftes und Vorhandenes, je mit Beleg')
  const zeilen = await p.seite("return [...document.querySelectorAll('table')].find((t) => t.querySelector('caption')?.innerText.trim() === 'Was geprüft ist').innerText")
  p.pruefe('Scryfall geprüft, mit Beleg', /Scryfall\s+geprüft\s+docs\/HAERTUNG\.md/.test(zeilen), zeilen.split('\n').find((z) => z.startsWith('Fremde Dienste\tScryfall')) ?? '')
  p.pruefe('Anthropic nur vorhanden', /Anthropic\s+vorhanden/.test(zeilen))
  p.pruefe('VoiceOver wartet auf einen Menschen', /VoiceOver\s+wartet auf einen Menschen\s+Issue #1871/.test(zeilen))
  await p.seite("[...document.querySelectorAll('h3')].find((h) => h.innerText.trim() === 'Was geprüft ist').scrollIntoView({ block: 'start' })")
  await p.bild('pruefstand')

  p.notiz('NAT-01 — der Datenschutz nennt Wege, die es gibt')
  await p.klickeText('Was diese Anwendung über dich weiß')
  await p.wartetext('Einstellungen › Dienste › Fremde Kartendienste', 20000)
  p.pruefe('kein Verweis auf eine erfundene Gruppe', !(await p.seite("return document.body.innerText.includes('Einstellungen → ')")))
  await p.bild('datenschutz')

  p.pruefe('⌘Q beendet', await p.beende())
}
