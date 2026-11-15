// **Eine Quick-Look-Erweiterung zur Probe** — G06
// (#2253).
//
// Sie zeichnet **keine** Karte, sondern beantwortet zwei Fragen der
// Machbarkeit, und die Antwort ist am Bild abzulesen:
//
//  - **Lädt macOS eine ad hoc signierte Erweiterung?** Dann hat das Bild das
//    Seitenverhältnis einer Karte (63 : 88) statt des Symbols.
//  - **Erreicht sie aus der Sandbox die Vorlagenpakete?** Eine echte
//    Kartenvorschau bräuchte die Stile in `~/Library/Application Support/
//    MagicSetEditor4Mac/pakete`. Magenta heißt lesbar, Cyan nicht.
//
// API nachgelesen im SDK von Xcode 27 (`QLThumbnailProvider.h`,
// `QLThumbnailReply.h`, `QLThumbnailRequest.h`).
import CoreGraphics
import Foundation
import QuickLookThumbnailing

class Vorschaubild: QLThumbnailProvider {
  override func provideThumbnail(
    for request: QLFileThumbnailRequest, _ handler: @escaping (QLThumbnailReply?, Error?) -> Void
  ) {
    let hoehe = request.maximumSize.height
    let groesse = CGSize(width: (hoehe * 63 / 88).rounded(), height: hoehe)

    // **Der echte Ort, nicht der Behälter**: in der Sandbox zeigt
    // `homeDirectoryForCurrentUser` in den Behälter der Erweiterung.
    let nutzer = getpwuid(getuid()).flatMap { String(validatingCString: $0.pointee.pw_dir) } ?? NSHomeDirectory()
    let pakete = nutzer + "/Library/Application Support/MagicSetEditor4Mac/pakete"
    let lesbar = (try? FileManager.default.contentsOfDirectory(atPath: pakete)) != nil

    handler(QLThumbnailReply(contextSize: groesse, drawing: { kontext in
      if lesbar {
        kontext.setFillColor(red: 1, green: 0, blue: 1, alpha: 1)
      } else {
        kontext.setFillColor(red: 0, green: 1, blue: 1, alpha: 1)
      }
      kontext.fill(CGRect(origin: .zero, size: groesse))
      return true
    }), nil)
  }
}
