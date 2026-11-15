// **Die WebKit-Probe** — Abnahmen, die sonst nur ein Mensch am nativen
// Fenster machen konnte: echte Tastatur und Maus, WebGL, WebKits eigene
// Vorschlagslisten, helles und dunkles Erscheinungsbild und das, was ein
// Vorleser ansagt.
//
//     webkitprobe <drehbuch> [--ausgabe <verzeichnis>] [--basis <adresse>]
//                 [--motor wk1|wk2] [--breite 1440] [--hoehe 900]
//                 [--dunkel] [--sichtbar] [--stand <commit>]
//
// Der Aufruf steht in `scripts/webkitprobe.sh`; was die zwei Motoren können
// und was nicht, in `Motoren.swift`; die Sprache des Drehbuchs in
// `Drehbuch.swift`.
//
// **Keine Freigabe nötig, und keine berührt**: das Programm schickt Ereignisse
// nur an sein eigenes Fenster, wird nie aktiv und liegt ohne `--sichtbar`
// außerhalb des Bildschirms.
import AppKit

struct Argumente {
  var drehbuch = ""
  var ausgabe = "webkitprobe-ausgabe"
  var basis = "http://localhost:34115/"
  var motor: String?
  var breite = 1440.0
  var hoehe = 900.0
  var dunkel = false
  var sichtbar = false
  var stand = ProcessInfo.processInfo.environment["MSE_STAND"] ?? "unbekannt"
}

func argumenteLesen(_ liste: [String]) -> Argumente {
  var argumente = Argumente()
  var rest = liste.dropFirst().makeIterator()

  while let wort = rest.next() {
    switch wort {
    case "--ausgabe": argumente.ausgabe = rest.next() ?? argumente.ausgabe
    case "--basis": argumente.basis = rest.next() ?? argumente.basis
    case "--motor": argumente.motor = rest.next()
    case "--breite": argumente.breite = Double(rest.next() ?? "") ?? argumente.breite
    case "--hoehe": argumente.hoehe = Double(rest.next() ?? "") ?? argumente.hoehe
    case "--dunkel": argumente.dunkel = true
    case "--sichtbar": argumente.sichtbar = true
    case "--stand": argumente.stand = rest.next() ?? argumente.stand
    default: argumente.drehbuch = wort
    }
  }

  return argumente
}

@MainActor func starten() async -> Int32 {
  NSApp.mainMenu = bearbeitenmenue()
  let argumente = argumenteLesen(CommandLine.arguments)
  guard !argumente.drehbuch.isEmpty, let text = try? String(contentsOfFile: argumente.drehbuch, encoding: .utf8) else {
    FileHandle.standardError.write(Data("Aufruf: webkitprobe <drehbuch> [--ausgabe <verzeichnis>] …\n".utf8))
    return 2
  }

  let schritte: [(zeile: Int, schritt: Schritt, roh: String)]
  do {
    schritte = try drehbuchLesen(text)
  } catch {
    FileHandle.standardError.write(Data("\(error)\n".utf8))
    return 2
  }

  // **Der Motor steht im Drehbuch**, und der Aufruf darf ihn überstimmen.
  var motorname = "wk2"
  if case .motor(let genannt)? = schritte.first?.schritt { motorname = genannt }
  if let genannt = argumente.motor { motorname = genannt }

  let groesse = NSSize(width: argumente.breite, height: argumente.hoehe)
  let motor: Motor = motorname == "wk1" ? Wk1(groesse: groesse) : Wk2(groesse: groesse)

  // **Außerhalb des Bildschirms, solange niemand zusehen will** — ein
  // randloses Fenster darf dort liegen; ein betiteltes schöbe AppKit zurück.
  let ursprung = argumente.sichtbar ? NSPoint(x: 60, y: 60) : NSPoint(x: -20000, y: -20000)
  let fenster = Probefenster(
    contentRect: NSRect(origin: ursprung, size: groesse),
    styleMask: argumente.sichtbar ? [.titled] : [.borderless], backing: .buffered, defer: false)
  fenster.isReleasedWhenClosed = false
  fenster.contentView = motor.ansicht
  if argumente.dunkel {
    fenster.appearance = NSAppearance(named: .darkAqua)
    motor.ansicht.appearance = fenster.appearance
  }
  fenster.orderFrontRegardless()
  fenster.makeFirstResponder(motor.ansicht)

  try? FileManager.default.createDirectory(atPath: argumente.ausgabe, withIntermediateDirectories: true)
  let bestanden = await Lauf(argumente: argumente, motor: motor, fenster: fenster).fahren(schritte)
  return bestanden ? 0 : 1
}

// bearbeitenmenue gibt der Probe die Einträge, die das Fenster auch hat —
// dieselben Rollen wie `internal/app/menue.go` (#255): ⌘A, ⌘C, ⌘V, ⌘X, ⌘Z
// und ⇧⌘Z gehen in einer Anwendung über die Menüleiste und nicht an die
// Seite. Ohne sie hängte ⌘A in einem Feld an, statt zu markieren.
@MainActor func bearbeitenmenue() -> NSMenu {
  let leiste = NSMenu()
  let bearbeiten = NSMenu(title: "Bearbeiten")
  let eintraege: [(String, Selector, String, NSEvent.ModifierFlags)] = [
    ("Widerrufen", Selector(("undo:")), "z", [.command]),
    ("Wiederholen", Selector(("redo:")), "z", [.command, .shift]),
    ("Ausschneiden", #selector(NSText.cut(_:)), "x", [.command]),
    ("Kopieren", #selector(NSText.copy(_:)), "c", [.command]),
    ("Einsetzen", #selector(NSText.paste(_:)), "v", [.command]),
    ("Alles auswählen", #selector(NSText.selectAll(_:)), "a", [.command]),
  ]
  for (titel, aktion, taste, umschalter) in eintraege {
    let eintrag = NSMenuItem(title: titel, action: aktion, keyEquivalent: taste)
    eintrag.keyEquivalentModifierMask = umschalter
    bearbeiten.addItem(eintrag)
  }
  let halter = NSMenuItem(title: "Bearbeiten", action: nil, keyEquivalent: "")
  halter.submenu = bearbeiten
  leiste.addItem(halter)
  return leiste
}

let anwendung = NSApplication.shared
anwendung.setActivationPolicy(.accessory)
DispatchQueue.main.async {
  Task { @MainActor in
    exit(await starten())
  }
}
anwendung.run()
