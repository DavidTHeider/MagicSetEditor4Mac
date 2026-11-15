// **Echte Ereignisse, im eigenen Fenster** — Maus und Tastatur so, wie AppKit
// sie an WebKit reicht.
//
// # Warum nicht über `CGEvent`
//
// Ein `CGEvent` geht an das ganze System und braucht dafür die Freigabe
// „Bedienungshilfen" — und er landete in jedem Fenster, das gerade vorn ist,
// auch in dem eines Menschen am Rechner. Ein `NSEvent`, das dieses Programm
// seinem **eigenen** Fenster schickt, braucht keine Freigabe und erreicht
// niemanden sonst. Für WebKit ist es ein Ereignis der Plattform und kein
// Skriptereignis: `isTrusted` ist wahr, und ein Klick verhält sich wie einer.
import AppKit

// tastencodes sind die virtuellen Codes einer ANSI-Tastatur
// (`HIToolbox/Events.h`, `kVK_ANSI_*`). WebKit leitet daraus `code` ab;
// `key` kommt aus den Zeichen.
let tastencodes: [Character: UInt16] = [
  "a": 0, "s": 1, "d": 2, "f": 3, "h": 4, "g": 5, "z": 6, "x": 7, "c": 8, "v": 9,
  "b": 11, "q": 12, "w": 13, "e": 14, "r": 15, "y": 16, "t": 17, "1": 18, "2": 19,
  "3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "9": 25, "7": 26, "-": 27, "8": 28,
  "0": 29, "]": 30, "o": 31, "u": 32, "[": 33, "i": 34, "p": 35, "l": 37, "j": 38,
  "'": 39, "k": 40, ";": 41, "\\": 42, ",": 43, "/": 44, "n": 45, "m": 46, ".": 47,
  " ": 49, "`": 50,
]

// sondertasten sind die benannten Tasten: Code und das Zeichen, das AppKit
// für sie liefert (`NSEvent.SpecialKey`).
let sondertasten: [String: (code: UInt16, zeichen: String)] = [
  "tab": (48, "\t"),
  "enter": (36, "\r"),
  "return": (36, "\r"),
  "escape": (53, "\u{1b}"),
  "space": (49, " "),
  "backspace": (51, "\u{7f}"),
  "delete": (117, String(UnicodeScalar(0xF728)!)),
  "up": (126, String(UnicodeScalar(0xF700)!)),
  "down": (125, String(UnicodeScalar(0xF701)!)),
  "left": (123, String(UnicodeScalar(0xF702)!)),
  "right": (124, String(UnicodeScalar(0xF703)!)),
  "home": (115, String(UnicodeScalar(0xF729)!)),
  "end": (119, String(UnicodeScalar(0xF72B)!)),
]

struct Tastenfehler: Error, CustomStringConvertible {
  let description: String
}

// tasteLesen zerlegt „cmd+shift+k" in Umschalter, Code und Zeichen.
func tasteLesen(_ kombination: String) throws -> (NSEvent.ModifierFlags, UInt16, String) {
  var umschalter: NSEvent.ModifierFlags = []
  let teile = kombination.lowercased().split(separator: "+").map(String.init)
  guard let taste = teile.last, !taste.isEmpty else {
    throw Tastenfehler(description: "keine Taste in »\(kombination)«")
  }

  for teil in teile.dropLast() {
    switch teil {
    case "cmd", "command": umschalter.insert(.command)
    case "shift": umschalter.insert(.shift)
    case "alt", "option": umschalter.insert(.option)
    case "ctrl", "control": umschalter.insert(.control)
    default: throw Tastenfehler(description: "unbekannter Umschalter »\(teil)«")
    }
  }

  if let sonder = sondertasten[taste] {
    return (umschalter, sonder.code, sonder.zeichen)
  }

  guard taste.count == 1, let zeichen = taste.first else {
    throw Tastenfehler(description: "unbekannte Taste »\(taste)«")
  }

  return (umschalter, tastencodes[zeichen] ?? 0, taste)
}

@MainActor func tasteSenden(
  an fenster: NSWindow, umschalter: NSEvent.ModifierFlags, code: UInt16, zeichen: String,
  alsAequivalent: Bool = true
) {
  let zeit = ProcessInfo.processInfo.systemUptime

  func ereignis(_ art: NSEvent.EventType) -> NSEvent {
    NSEvent.keyEvent(
      with: art, location: .zero, modifierFlags: umschalter, timestamp: zeit,
      windowNumber: fenster.windowNumber, context: nil, characters: zeichen,
      charactersIgnoringModifiers: zeichen.lowercased(), isARepeat: false, keyCode: code)!
  }

  // **Ein Kürzel geht erst als Tastenäquivalent durch die Ansichten** — wie
  // AppKit es mit ⌘-Tasten tut. Nimmt es keine an, bekommt die Seite es als
  // gewöhnlichen Tastendruck. **WebKit 1 nimmt es an und gibt es dem
  // Dokument trotzdem nicht** — dort geht es gleich als Tastendruck hinein.
  let runter = ereignis(.keyDown)
  if alsAequivalent, umschalter.contains(.command) {
    // **Ein Eintrag der Menüleiste zuerst** — seine Aktion geht an den ersten
    // Antwortgeber des Probefensters. Über `performKeyEquivalent` des Menüs
    // ginge sie an das Schlüsselfenster der Anwendung, und ein Programm, das
    // nie aktiv wird, hat keins.
    if let eintrag = menueeintrag(zu: zeichen.lowercased(), umschalter: umschalter),
       let aktion = eintrag.action,
       NSApp.sendAction(aktion, to: fenster.firstResponder, from: eintrag) {
      return
    }

    if fenster.performKeyEquivalent(with: runter) {
      return
    }
  }

  fenster.sendEvent(runter)
  fenster.sendEvent(ereignis(.keyUp))
}

@MainActor func textSenden(an fenster: NSWindow, _ text: String) {
  for zeichen in text {
    let klein = Character(zeichen.lowercased())
    let umschalter: NSEvent.ModifierFlags = zeichen.isUppercase ? [.shift] : []
    tasteSenden(an: fenster, umschalter: umschalter, code: tastencodes[klein] ?? 0,
                zeichen: String(zeichen))
  }
}

// klickSenden klickt an einen Punkt der Seite — in CSS-Pixeln, wie
// `getBoundingClientRect` sie nennt.
@MainActor func klickSenden(an fenster: NSWindow, ansicht: NSView, punkt: NSPoint) {
  let lokal = ansicht.isFlipped ? punkt : NSPoint(x: punkt.x, y: ansicht.bounds.height - punkt.y)
  let imFenster = ansicht.convert(lokal, to: nil)
  let zeit = ProcessInfo.processInfo.systemUptime

  for art in [NSEvent.EventType.leftMouseDown, .leftMouseUp] {
    let ereignis = NSEvent.mouseEvent(
      with: art, location: imFenster, modifierFlags: [], timestamp: zeit,
      windowNumber: fenster.windowNumber, context: nil, eventNumber: 0, clickCount: 1,
      pressure: art == .leftMouseDown ? 1 : 0)!
    fenster.sendEvent(ereignis)
  }
}

// menueeintrag sucht in der Menüleiste den Eintrag zu einem Tastenäquivalent.
@MainActor func menueeintrag(zu taste: String, umschalter: NSEvent.ModifierFlags) -> NSMenuItem? {
  let relevant: NSEvent.ModifierFlags = [.command, .shift, .option, .control]
  for oben in NSApp.mainMenu?.items ?? [] {
    for eintrag in oben.submenu?.items ?? []
      where eintrag.keyEquivalent == taste
        && eintrag.keyEquivalentModifierMask.intersection(relevant) == umschalter.intersection(relevant) {
      return eintrag
    }
  }
  return nil
}
