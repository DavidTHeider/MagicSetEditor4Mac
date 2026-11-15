// **Die Zwischenspeicher, am nativen Fenster** — OPS-08
// (#1725):
//
// - AK-01: unter *Einstellungen › Diagnose* steht je Speicher eine Zeile mit
//   Ort, Trefferquote, Einträgen, Größe und Grenze — alle acht,
// - AK-03: die Grenze der Platte ist eingestellt und steht in der Zeile,
// - der Knopf „Alle leeren" steht da und sagt vorher, was er tut.
//
// **Gedrückt wird er nicht**: der Speicher auf der Platte liegt unter
// `~/Library/Caches` dieses Rechners, und eine Probe wirft dort nichts weg.
// Dass er leert, belegen `internal/app/zwischenspeicher_intern_test.go` und
// die Läufe unter jsdom.

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const tafel = "[...document.querySelectorAll('table')].find((t) => t.querySelector('caption')?.innerText.trim() === 'Zwischenspeicher')"

export default async function ops08(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'zwischenspeicher.grenze', '64')
  await p.starte({ ort })

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.klickeText('Diagnose')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Zwischenspeicher und Suchindex')
  await p.warte('die Tafel der Zwischenspeicher steht da', () => p.seite(`return !!${tafel}`), 20000)

  p.notiz('AK-01 — acht Speicher mit Ort, Quote, Einträgen, Größe und Grenze')
  const zeilen = await p.seite(`return [...${tafel}.querySelectorAll('tbody tr')].map((z) => z.innerText.replaceAll('\\t', ' | '))`)
  p.pruefe('acht Zeilen', zeilen.length === 8, String(zeilen.length))
  for (const name of ['Antworten von Scryfall', 'Kartenbilder von Scryfall', 'Sammeldatei von Scryfall', 'Abgetastete Ebenen', 'Bilder der Vorlagenpakete', 'Zerteilte Skripte', 'Kartenvorschauen', 'Kartenbilder am Tisch']) {
    p.pruefe(`»${name}« steht da`, zeilen.some((z) => z.startsWith(name)))
  }

  p.notiz('AK-03 — die eingestellte Grenze steht in der Zeile')
  const antworten = zeilen.find((z) => z.startsWith('Antworten von Scryfall')) ?? ''
  p.pruefe('Grenze 64,0 MB', antworten.includes('64,0 MB'), antworten)
  p.pruefe('Ort Platte', antworten.includes('Platte'), antworten)
  const ebenen = zeilen.find((z) => z.startsWith('Abgetastete Ebenen')) ?? ''
  p.pruefe('die Ebenen liegen im Arbeitsspeicher, Grenze 64,0 MB', ebenen.includes('Arbeitsspeicher') && ebenen.includes('64,0 MB'), ebenen)

  p.pruefe('der Knopf steht da und die Folge davor', await p.seite("return [...document.querySelectorAll('button')].some((b) => b.innerText.trim() === 'Alle leeren') && document.body.innerText.includes('Leeren wirft alles weg.')"))
  await p.seite(`${tafel}.scrollIntoView({ block: 'center' })`)
  await p.bild('zwischenspeicher')

  p.pruefe('⌘Q beendet', await p.beende())
}
