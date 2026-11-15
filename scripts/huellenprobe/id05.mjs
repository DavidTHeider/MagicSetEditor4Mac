// **Wogegen ein Anmeldeanbieter geprüft ist, am nativen Fenster** — ID-05
// (#1704):
//
// - AK-01: auf der Nutzerseite unter *Konto* (seit UXR-20 im Kasten
//   *Anmeldung über einen Anbieter*) steht bei jedem der vier
//   Anbieter *nicht eingerichtet · nur gegen eine Nachbildung geprüft*, und
//   nach dem Eintragen einer Kennung für Microsoft dort *eingerichtet*.
//
// **Microsoft und nicht GitHub**: GitHub verlangt seit dem Nachlauf über
// ID-05 (#1890) auch ein Geheimnis, eine Kennung allein richtet es nicht ein;
// Microsoft verbietet eines. Das Drehbuch lief danach nicht mehr und blieb
// stehen, bis UXR-20 es am 2026-10-02 wieder fuhr.
//
// Die Regel belegen `internal/domain/anbieterpruefstand_test.go` und
// `internal/app/anbieterstaende_test.go`.

const zeilen = "return [...document.querySelectorAll('section')].find((s) => s.getAttribute('aria-label') === 'Anmeldung über einen Anbieter')?.innerText ?? ''"

export default async function id05(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'anbieter.microsoft.kennung', 'probe-kennung')
  await p.starte({ ort })

  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /Nutzer und Profile/.test(b.getAttribute('aria-label') ?? b.innerText))", '»Nutzer und Profile«')
  await p.klickeAuf("[...document.querySelectorAll('[role=tab]')].find((t) => t.checkVisibility() && t.innerText.trim() === 'Konto')", 'die Lasche »Konto«')
  // **Seit UXR-20 (#1928) eingeklappt** hinter Sitzung und zweitem Faktor.
  await p.klickeText('Anmeldung über einen Anbieter')
  await p.wartetext('nur gegen eine Nachbildung geprüft', 20000)

  const text = await p.seite(zeilen)
  const stand = text.split('\n').filter((z) => z.includes('geprüft'))
  p.pruefe('jeder der vier Anbieter nennt seinen Prüfstand', stand.length === 4, stand.join(' | '))
  p.pruefe('Microsoft ist eingerichtet, die anderen nicht', stand.filter((z) => z.startsWith('eingerichtet')).length === 1 && stand.filter((z) => z.startsWith('nicht eingerichtet')).length === 3, stand.join(' | '))

  await p.seite("[...document.querySelectorAll('section')].find((s) => s.getAttribute('aria-label') === 'Anmeldung über einen Anbieter')?.scrollIntoView({ block: 'center' })")
  await p.bild('anbieterstand')
  p.pruefe('⌘Q beendet', await p.beende())
}
