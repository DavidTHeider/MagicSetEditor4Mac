// Die Formen der **Wörter** — Schlüsselwörter, Wortlisten und die
// Symbolpalette.
//
// **Eine eigene Datei**, weil `formen.ts` mit ihnen über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort stehen die Formen der Karte und
// ihrer Felder, hier die von vier Flächen, die alle dasselbe tun — sie bieten
// aus einer **Liste** etwas zum Einsetzen an, und ihre Einträge tragen
// Kinder.
//
// Es gilt dasselbe wie dort: jede Form entspricht einem DTO aus
// `internal/app`, und läuft eine der beiden Seiten weg, merkt es kein Test
// (§3).

// SchluesselwortDTO ist ein Schlüsselwort eines Sets oder seines Spiels
// (#1062).
export interface SchluesselwortDTO {
  // id ist die Kennung — **leer bei einem des Spiels**: das gehört keinem Set
  // und lässt sich nicht ändern.
  id: string

  // Die fünf Felder heißen wie in der Datei, **unübersetzt** (§9).
  wort: string
  muster: string
  erinnerung: string
  regeln: string
  modus: string

  // fest sagt, dass es aus dem Spiel kommt.
  fest: boolean

  // gueltig sagt, dass sein Muster greift. Ein ungültiges schaltet MSE ab —
  // ein Muster, das auf den leeren Text passt, ließe es hängen.
  gueltig: boolean
}

// WortmenueeintragDTO ist eine Zeile in einem Auswahlmenü
// (#1074).
export interface WortmenueeintragDTO {
  // name ist das Wort — **unübersetzt** (§9): es steht so im Paket und wird so
  // in die Karte geschrieben.
  name: string

  // istPraefix macht aus der Zeile ein Kästchen statt eines Punktes: sie lässt
  // sich mit anderen verbinden, statt sie zu ersetzen.
  istPraefix: boolean

  // zeileDarunter zieht einen Strich unter die Zeile.
  zeileDarunter: boolean

  // aktiv sagt, dass der Wert des Feldes diesen Eintrag enthält.
  aktiv: boolean

  // kinder sind das Untermenü einer Gruppe — leer bei einer gewöhnlichen
  // Zeile. **Keine eigene Marke für »ist eine Gruppe«**: eine Gruppe ist genau
  // eine Zeile mit Kindern, und zwei Angaben darüber liefen auseinander.
  kinder: WortmenueeintragDTO[] | null
}

// WortlistenbereichDTO ist eine Stelle im Feld mit ihrem Menü.
export interface WortlistenbereichDTO {
  // liste ist der Name der Wortliste — für die Beschriftung, wenn der Text
  // leer ist.
  liste: string

  // text ist, was an dieser Stelle steht.
  text: string

  // kaestchen sagt, dass das Menü eine Spalte für Kästchen braucht.
  kaestchen: boolean

  // gewaehlt ist die Zeile, auf der die Auswahl steht — **-1**, wenn keine
  // passt.
  gewaehlt: number

  eintraege: WortmenueeintragDTO[]
}

export interface PalettenEintragDTO {
  // art ist `code`, `custom`, `line` oder `submenu` — **unübersetzt** (§9):
  // die Wörter stehen so im Paket.
  art: string

  // code ist, was eingefügt wird. Leer bei Trenner und Untermenü.
  code: string

  beschriftung: string

  // rueckfrage ist der Satz über dem Eingabefeld eines `custom`-Eintrags.
  rueckfrage: string

  // bild ist ein PNG als Data-URI — leer, wo es keins gibt.
  bild: string

  kinder: PalettenEintragDTO[]
}
