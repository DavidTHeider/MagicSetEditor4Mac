// Die Formen, die über die Wails-Naht gehen.
//
// **Getrennt von der Fassade, weil es eine andere Frage ist** (AGENTS.md §2):
// hier steht, *was* hinüberkommt, in `api.ts`, *wie* es gerufen wird. Der
// Schnitt kam, als `api.ts` über 400 Zeilen wuchs — und er folgt der Frage,
// nicht der Zeilenzahl.
//
// Jede Form entspricht einem DTO aus `internal/app`. Läuft eine der beiden
// Seiten weg, merkt es kein Test: die Frontend-Tests gehen über `test/api.ts`
// und nie durch die JSON-Serialisierung (§3). Was hier steht, gehört an der
// gebauten Anwendung nachgesehen.

export interface SetDTO {
  id: string
  name: string
  spiel: string
  stil: string
  herkunft: string
  karten: number

  // herkunftFehlt sagt, dass die Datei, aus der dieses Set kam, nicht mehr
  // dasteht (#891).
  //
  // **Das Set bleibt trotzdem offen** — es liegt im Bestand, nicht in der
  // Datei. Leer heißt nicht »fehlt«: ein Set, das in der Anwendung entstand,
  // hat gar keine Herkunft.
  herkunftFehlt: boolean
}

// KartenfeldwertDTO ist der Wert **eines** Feldes **einer** Karte (#891).
export interface KartenfeldwertDTO {
  id: string
  wert: string
}

// KartenzeileDTO ist eine Karte, wie die **Liste** sie zeigt (#891).
//
// **Ohne Feldwerte.** `Black.mse-set` hat 3.788 Karten; sie mit jedem Wert zu
// laden waren 4.761.182 Bytes JSON je Aufruf — gemessen am 2026-08-25.
export interface KartenzeileDTO {
  id: string
  nummer: number

  // name ist der Wert des Feldes `name` — leer, wenn es keinen gibt.
  name: string

  // felder ist die **Zahl** der Feldwerte, nicht die Werte.
  //
  // Die Liste zeigt „ohne Felder" an einer Karte, die keine hat: der
  // Unterschied zwischen „heißt nicht" und „ist leer".
  felder: number
}

export interface WertDTO {
  feld: string
  wert: string
}

export interface KarteDTO {
  id: string
  nummer: number
  werte: WertDTO[]
}

export interface VorbehaltDTO {
  datei: string
  zeile: number
  inhalt: string
}

// BewegungAntwort sagt, was eine Rücknahme bewegt hat.
//
// `bewegt` ist falsch, wenn es nichts zurückzunehmen gab — **kein Fehler**: ⌘Z
// am Anfang des Verlaufs ist eine gewöhnliche Lage.
//
// `karteID` nennt die betroffene Karte, damit nur **sie** nachgeladen wird
// (AGENTS.md §7); `urheber` sagt, wessen Änderung bewegt wurde — `fenster` oder
// `server` (#52).
export interface BewegungAntwort {
  bewegt: boolean
  karteID: string
  feld: string

  // umfang ist die Zahl der Karten, die der Schritt berührt hat — eins bei
  // einer Einzeländerung, viele bei einer Massenänderung.
  umfang: number

  urheber: string
  fehler: FehlerDTO | null
}

export interface VorschauDTO {
  bild: string
  breite: number
  hoehe: number

  // dpi ist die Auflösung der Vorlage — mit ihr wird aus Punkten eine
  // Druckgröße (U5, #1416); 0 heißt: die Vorlage nennt keine.
  dpi: number
  befunde: string[]

  // flaechen sagen, welches Feld wo im Bild liegt (#246).
  flaechen: FeldflaecheDTO[]

  // stil ist der Wert `stylesheet`, mit dem die Vorschau gezeichnet ist —
  // getrimmt, leer heißt „der des Sets" (Zweitblick über die Folgen vor
  // 0.8.0, Befund 8). Unter ihm legt das Formular die Flächen ab.
  stil: string

  // gezeichnet ist das Paket, mit dem der Plan gebaut wurde — leer beim
  // Behelfsplan (#2314). **Nicht dasselbe wie `stil`**: eine Karte, die den
  // Stil des Sets erbt, schreibt `''`, und welches Paket das ist, ändert ein
  // fremder Setstilwechsel, ohne dass die Karte sich ändert.
  gezeichnet: string

  // vorgaben sind die **gerechneten** Feldwerte — `default:`, `initial:` oder
  // die erste `choice:`, für Felder, die die Karte nicht selbst führt
  // (#563).
  //
  // **Sie hängen an der Vorschau und nicht an `Kartenfelder`**, weil sie zur
  // Karte gehören: ein `default:` ist ein Skript über `card`, und dieselbe
  // Vorlage gibt für zwei Karten zwei Werte.
  vorgaben: VorgabeDTO[]

  // nennungen sind die Sätze, die neben diesem Bild stehen müssen
  // (#1025).
  //
  // **Sie hängen an der Vorschau**, weil die Vorschau die Oberfläche ist, die
  // den Bildausschnitt zeigt — auch dann noch, wenn der Zuschnittkasten
  // längst zu ist.
  //
  // **Nie `null`.** Go schickt eine leere Liste; der Umschlagwächter
  // (`internal/app/umschlag_test.go`) hält das fest.
  nennungen: NennungDTO[]
}

// NennungDTO ist, was neben einem übernommenen Wert stehen muss.
//
// **Der Satz kommt fertig aus Go** (`service.Nennungen`): dort steht, woran
// eine Kartenseite bei Scryfall erkannt wird, und dort steht das Copyright als
// Konstante.
export interface NennungDTO {
  adresse: string
  felder: string[]
  kuenstler: string
  text: string
}

// VorgabeDTO ist ein gerechneter Feldwert.
export interface VorgabeDTO {
  feld: string
  wert: string
}

// TrefferDTO ist eine gefundene Stelle.
//
// `kartenname` ist der Name der Karte, **nicht** der getroffene Wert: eine
// Trefferliste, die nur die Fundstelle zeigt, lässt einen raten, wozu sie
// gehört.
export interface TrefferDTO {
  karteID: string
  setID: string
  nummer: number
  kartenname: string
  feld: string
  wert: string
}

// FeldflaecheDTO sagt, **wo** ein Feld in der Vorschau liegt — in Bildpunkten.
//
// Damit lässt sich in der Karte anfassen, was man sieht
// (#246).
export interface FeldflaecheDTO {
  feld: string
  links: number
  oben: number
  breite: number
  hoehe: number
}

// FeldbeschreibungDTO sagt, **welche Form** ein Kartenfeld hat.
//
// Sie kommt aus dem Spielpaket und entscheidet, welches Steuerelement der
// Editor zeigt (#560).
export interface FeldbeschreibungDTO {
  name: string

  // aufschrift ist `caption:` in der Sprache der Oberfläche — **die
  // Beschriftung, nicht der Bezeichner**. `name` bleibt die Kennung: daran
  // hängt, welches Feld geschrieben wird
  // (#560).
  aufschrift: string

  // art ist `type:` aus dem Paket — **unübersetzt** (§9), und das Frontend
  // schaltet darauf.
  art: string

  mehrzeilig: boolean

  // gespeichert ist `save value:` — ob der Wert in die Datei geht oder aus
  // anderen entsteht (#1215, Befund 1).
  //
  // **Nicht aus `teilfelder` zu erraten**: ein berechnetes Feld, dessen
  // Skript keine Glieder hergibt, hat ebenfalls keine.
  gespeichert: boolean
  wahlen: string[]
  beschreibung: string

  // vorgabe ist der **Anfangswert** — was eine Karte ohne eigenen Wert in
  // diesem Feld trägt. Ohne ihn zeigte ein `select` seine erste Option, und
  // `rarity` beginnt mit `basic land`, hat aber `initial: common`
  // (#562).
  vorgabe: string

  // editierbar ist `editable:` — ob jemand dieses Feld bearbeiten darf.
  //
  // **Nicht dasselbe wie „wird gespeichert"**: MSE lässt ein nicht
  // editierbares Feld nicht einmal auswählen. 24 der 249 Kartenfelder von
  // `magic.mse-game` tragen es — `full name`, `cmc`, `card number` und weitere
  // (#708).
  editierbar: boolean

  // teilfelder sind die Felder, aus denen dieses Feld **entsteht** — leer bei
  // allem, was für sich steht.
  //
  // `text` trägt bei `magic.mse-game` `save value: false` und setzt sich aus
  // Regeltext und Flavortext zusammen. Der Editor bietet an seiner Stelle
  // **diese** an; ein Kasten „Text" nähme eine Eingabe entgegen, die nirgends
  // ankommt (#696).
  teilfelder: string[]

  // reihe ist das Grundfeld der **Feldreihe**, zu der dieses Feld gehört —
  // leer ohne Geschwister
  // (#1167).
  //
  // `name`, `name 2`, `name 3` und `name 4` tragen alle `name`. **Die Regel
  // steht in Go** (`domain.Game.Feldreihen`) und wird hier nicht nachgebaut:
  // zwei Fassungen davon liefen an dem Tag auseinander, an dem jemand eine
  // ändert.
  reihe: string

  // stelle ist die Position in der Reihe — 1 für das Grundfeld, 0 ohne Reihe.
  //
  // **Was eine Stelle bedeutet, steht nicht hier**: bei einer zweiseitigen
  // Vorlage ist Stelle 2 die Rückseite, bei `additional credit 2` die zweite
  // Zeile einer Danksagung.
  stelle: number
}

// PalettenEintragDTO ist ein Eintrag der Symbolpalette
// (#564).
// PackartDTO ist eine Packart des Spiels
// (#1072).
export interface PackartDTO {
  // name ist der Name aus dem Paket — **unübersetzt** (§9).
  name: string

  // waehlbar sagt, ob MSE sie zum Ziehen anbietet. Im echten Pack sind es vier
  // von dreißig; die übrigen sind Bausteine.
  waehlbar: boolean

  // karten ist die Zahl der Karten, aus denen sie **erreichbar** zieht.
  karten: number
}

// FehlerDTO ist die Form, in der ein Fehler aus `internal/app` ankommt.
export interface FehlerDTO {
  // **`abgebrochen` ist kein Fehler, sondern die Absicht** (§7): ein Lauf, den
  // ein neuer abgelöst hat. `useLaden` zeigt ihn nicht und lässt das vorige
  // Bild stehen (#288).
  code:
    | 'validierung'
    | 'nichtGefunden'
    | 'konflikt'
    | 'unbekannt'
    | 'abgebrochen'

    // **`erlaubnis` heißt: du bist es, du darfst nur nicht** — Werkstück 11
    // von Phase 4.12 (#1351). Die Seite bietet dafür **keinen** zweiten
    // Versuch an: eine Rolle bekommt man nicht, indem man etwas anders tippt.
    | 'erlaubnis'

    // **Die übrigen Codes aus `internal/app/fehler.go`** — `voll` kam mit
    // #1812.
    | 'belegt'
    | 'schluesselFehlt'
    | 'nurLesend'
    | 'voll'

    // **Ein abgeschalteter Bereich** — OPS-10 (#1908).
    | 'abgeschaltet'

    // **Ein fremder Dienst, der nicht antwortet** — D02 (#2233); welcher
    // Fall, sagt der Schlüssel.
    | 'netz'
  schluessel: string
  meldung: string
  feld?: string

  // **`werte` füllt die Platzhalter des Katalogsatzes**
  // (#940).
  //
  // Ohne sie stand jede Meldung mit einem konkreten Namen vor der Wahl, ihn zu
  // verlieren oder den Katalog zu umgehen. Was hier steht, sind Bezeichner —
  // Dateinamen, Feldnamen — und die werden nicht übersetzt (§9).
  werte?: Record<string, string>
}

// AuswahlDTO ist ein Ausschnitt in den Punkten der Quelle
// (#576).
export interface AuswahlDTO {
  x: number
  y: number
  breite: number
  hoehe: number
}

// ZuschnittAntwort ist, was das Zuschnittfenster zum Aufmachen braucht.
//
// **Ein leerer `pfad` heißt „abgebrochen"** — dieselbe Übereinkunft wie beim
// Öffnen eines Sets: „nichts gewählt" ist eine Aussage und kein Fehler.
export interface ZuschnittAntwort {
  pfad: string

  // quelle ist das Ursprungsbild als PNG-Data-URI, für die Anzeige
  // verkleinert. Gerechnet wird trotzdem in `quelleBreite` × `quelleHoehe`.
  quelle: string

  quelleBreite: number
  quelleHoehe: number

  zielBreite: number
  zielHoehe: number

  auswahl: AuswahlDTO

  fehler?: FehlerDTO
}

// BeispielsetDTO ist ein Set, das das Bündel mitbringt
// (#702).
export interface BeispielsetDTO {
  // name ist der Dateiname ohne Endung — **unübersetzt** (§9).
  name: string

  // pfad geht unverändert zurück an `BeispielsetOeffnen`; die Liste ist der
  // Filter, nicht das Fenster.
  pfad: string
}
