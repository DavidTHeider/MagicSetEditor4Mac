// **Das Messprogramm der Quick-Look-Probe** — G06
// (#2253).
//
// `qlmanage -t` blockiert über `lume ssh`: ohne Anmeldesitzung erreicht es
// den Vorschaudienst nicht (gemessen am 2026-10-11, 22 Minuten ohne
// Antwort, schon ohne Erweiterung). Dieses Programm wird deshalb mit `open`
// in der Sitzung gestartet und fragt `QLThumbnailGenerator` — dieselbe
// Stelle, die der Finder fragt. Es schreibt die **Art** der Antwort (ein
// Symbol des Systems oder ein Vorschaubild), Größe und die Farbe der Mitte
// als JSON und endet.
//
// API nachgelesen im SDK von Xcode 27 (`QLThumbnailGenerator.h`,
// `QLThumbnailGenerationRequest.h`, `QLThumbnailRepresentation.h`).
import CoreGraphics
import Foundation
import QuickLookThumbnailing

let umgebung = ProcessInfo.processInfo.environment
guard let datei = umgebung["MSE_G06_DATEI"], let ziel = umgebung["MSE_G06_AUS"] else { exit(2) }

// farbe liest den Punkt in der Mitte und nennt ihn, wenn er einer der beiden
// Probefarben ist.
func farbe(_ bild: CGImage) -> String {
  var punkt = [UInt8](repeating: 0, count: 4)
  let raum = CGColorSpaceCreateDeviceRGB()
  guard let kontext = CGContext(data: &punkt, width: 1, height: 1, bitsPerComponent: 8, bytesPerRow: 4,
    space: raum, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue) else { return "?" }
  kontext.draw(bild, in: CGRect(x: -CGFloat(bild.width / 2), y: -CGFloat(bild.height / 2),
    width: CGFloat(bild.width), height: CGFloat(bild.height)))
  let (r, g, b) = (punkt[0], punkt[1], punkt[2])
  if r > 200 && b > 200 && g < 60 { return "magenta" }
  if g > 200 && b > 200 && r < 60 { return "cyan" }
  return "rgb(\(r),\(g),\(b))"
}

let anfrage = QLThumbnailGenerator.Request(fileAt: URL(fileURLWithPath: datei),
  size: CGSize(width: 256, height: 256), scale: 1, representationTypes: .all)
let fertig = DispatchSemaphore(value: 0)
var ergebnis: [String: Any] = ["fehler": "keine Antwort in 60 s"]
QLThumbnailGenerator.shared.generateBestRepresentation(for: anfrage) { bild, fehler in
  if let bild {
    let arten = ["symbol", "niedrig", "vorschaubild"]
    let art = bild.type.rawValue < arten.count ? arten[bild.type.rawValue] : "\(bild.type.rawValue)"
    ergebnis = ["art": art, "breite": bild.cgImage.width, "hoehe": bild.cgImage.height, "farbe": farbe(bild.cgImage)]
  } else {
    ergebnis = ["fehler": fehler.map { "\($0)" } ?? "?"]
  }
  fertig.signal()
}
_ = fertig.wait(timeout: .now() + 60)
let json = (try? JSONSerialization.data(withJSONObject: ergebnis, options: [.sortedKeys])) ?? Data()
FileManager.default.createFile(atPath: ziel, contents: json)
