// **Zwei Motoren, und jeder kann etwas, das der andere nicht kann.**
//
// `wk2` ist ein WKWebView **so eingerichtet wie in Wails 3 beta.23**
// (`pkg/application/webview_window_darwin.go`, Zeilen 139–194):
// `suppressesIncrementalRendering`, „wails.io" im User-Agent, dazu, was das
// Fenster dieser Anwendung setzt (`tabFocusesLinks`) — dieselbe Seite,
// derselbe Web-Prozess, dieselbe GPU. Hier stehen
// WebGL, WebKits eigene Vorschlagsliste und das Bild.
//
// `wk1` ist WebKits älterer Weg, `WebView`: WebCore läuft **im eigenen
// Prozess**. Dieselbe Zugänglichkeitsberechnung wie im Web-Prozess von
// `wk2` — Rolle, Name, Zustand —, aber ohne die Brücke über Prozesse, die
// nur ein vertrauenswürdiger Hilfsdienst wie VoiceOver überqueren darf.
// Hier steht, **was ein Vorleser sagt**.
//
// **Was keiner von beiden ist: das Wails-Fenster selbst.** Die Seite kommt
// über den Servermodus (`scripts/bauen.sh browser`) und nicht über Wails'
// IPC; die native Menüleiste und der Fensterrahmen gehören nicht dazu.
import AppKit
import WebKit

@MainActor protocol Motor: AnyObject {
  var name: String { get }
  var ansicht: NSView { get }
  func lade(_ adresse: URL) async throws
  func auswerten(_ skript: String) async throws -> String
  func abzug() async throws -> Data
}

struct Motorfehler: Error, CustomStringConvertible {
  let description: String
}

// umhuellt macht aus einem Ausdruck ein Skript, das immer Text zurückgibt —
// JSON oder einen Fehler als `{"__fehler": …}`. Beide Motoren geben nur so
// verlässlich dasselbe zurück.
func umhuellt(_ ausdruck: String) -> String {
  """
  (() => { try { const w = (() => (\(ausdruck)))(); \
  return JSON.stringify(w === undefined ? null : w) } \
  catch (e) { return JSON.stringify({ __fehler: String(e) }) } })()
  """
}

func pngAus(_ rep: NSBitmapImageRep) throws -> Data {
  guard let png = rep.representation(using: .png, properties: [:]) else {
    throw Motorfehler(description: "das Bild ließ sich nicht als PNG schreiben")
  }
  return png
}

@MainActor final class Wk2: NSObject, Motor, WKNavigationDelegate {
  let name = "wk2"
  let web: Probeansicht
  var ansicht: NSView { web }
  private var geladen: CheckedContinuation<Void, Error>?

  init(groesse: NSSize) {
    let einrichtung = WKWebViewConfiguration()
    einrichtung.suppressesIncrementalRendering = true
    einrichtung.applicationNameForUserAgent = "wails.io"

    // **Und was das Fenster dazu setzt** (`internal/fenster/ereignisse.go`):
    // Tab erreicht Knöpfe und Verweise. `internal/werkzeuge/webkitprobe_test.go`
    // hält beide Seiten gleich.
    einrichtung.preferences.tabFocusesLinks = true

    // **Die Fehler der Seite werden gesammelt**, vom ersten Skript an — ein
    // roter Schritt nennt sie im Bericht. Ohne das stand beim ersten Anlauf
    // der Tischmessung eine leere Seite da, und niemand wusste, warum.
    einrichtung.userContentController.addUserScript(WKUserScript(
      source: """
        window.__probeFehler = [];
        addEventListener('error', (e) => window.__probeFehler.push(String(e.message) + ' @ ' + e.filename + ':' + e.lineno));
        addEventListener('unhandledrejection', (e) => window.__probeFehler.push('unbehandelt: ' + String(e.reason && (e.reason.stack || e.reason))));
        """,
      injectionTime: .atDocumentStart, forMainFrameOnly: true))
    web = Probeansicht(frame: NSRect(origin: .zero, size: groesse), configuration: einrichtung)
    super.init()
    web.navigationDelegate = self

    // **Die Verdeckung abschalten** — WebKits eigener Testweg. Ein Fenster
    // außerhalb des Bildschirms gälte sonst als verdeckt, und WebKit hielte
    // Zeitgeber und `requestAnimationFrame` an: der Tisch zöge kein Bild.
    let wahl = NSSelectorFromString("_setWindowOcclusionDetectionEnabled:")
    if web.responds(to: wahl) {
      typealias Setzer = @convention(c) (AnyObject, Selector, Bool) -> Void
      unsafeBitCast(web.method(for: wahl), to: Setzer.self)(web, wahl, false)
    }
  }

  func lade(_ adresse: URL) async throws {
    try await withCheckedThrowingContinuation { (fortsetzung: CheckedContinuation<Void, Error>) in
      geladen = fortsetzung
      web.load(URLRequest(url: adresse))
    }
  }

  func webView(_ w: WKWebView, didFinish n: WKNavigation!) {
    geladen?.resume(); geladen = nil
  }

  func webView(_ w: WKWebView, didFail n: WKNavigation!, withError fehler: Error) {
    geladen?.resume(throwing: fehler); geladen = nil
  }

  func webView(_ w: WKWebView, didFailProvisionalNavigation n: WKNavigation!, withError fehler: Error) {
    geladen?.resume(throwing: fehler); geladen = nil
  }

  func auswerten(_ skript: String) async throws -> String {
    let ergebnis = try await web.evaluateJavaScript(umhuellt(skript))
    return (ergebnis as? String) ?? "null"
  }

  func abzug() async throws -> Data {
    let bild = try await web.takeSnapshot(configuration: nil)
    guard let tiff = bild.tiffRepresentation, let rep = NSBitmapImageRep(data: tiff) else {
      throw Motorfehler(description: "WebKit gab kein Bild")
    }
    return try pngAus(rep)
  }
}

@MainActor final class Wk1: NSObject, Motor, @preconcurrency WebFrameLoadDelegate {
  let name = "wk1"
  let web: WebView
  var ansicht: NSView { web }
  private var geladen: CheckedContinuation<Void, Error>?

  init(groesse: NSSize) {
    web = WebView(frame: NSRect(origin: .zero, size: groesse))
    web.applicationNameForUserAgent = "wails.io"
    super.init()
    web.frameLoadDelegate = self
    ersterKlickAngenommen("WebHTMLView")
  }

  func lade(_ adresse: URL) async throws {
    try await withCheckedThrowingContinuation { (fortsetzung: CheckedContinuation<Void, Error>) in
      geladen = fortsetzung
      web.mainFrame.load(URLRequest(url: adresse))
    }
  }

  func webView(_ sender: WebView!, didFinishLoadFor frame: WebFrame!) {
    guard frame == web.mainFrame else { return }

    // **Die Dokumentansicht nimmt die Tasten**, nicht die äußere `WebView`:
    // gewöhnliche Tasten reicht sie weiter, ⌘-Kürzel nicht — sie gingen
    // sonst am Dokument vorbei, und ⌘3 wechselte keine Seite.
    web.window?.makeFirstResponder(web.mainFrame.frameView.documentView)
    geladen?.resume(); geladen = nil
  }

  func webView(_ sender: WebView!, didFailLoadWithError fehler: Error!, for frame: WebFrame!) {
    if frame == web.mainFrame { geladen?.resume(throwing: fehler); geladen = nil }
  }

  func webView(_ sender: WebView!, didFailProvisionalLoadWithError fehler: Error!, for frame: WebFrame!) {
    if frame == web.mainFrame { geladen?.resume(throwing: fehler); geladen = nil }
  }

  func auswerten(_ skript: String) async throws -> String {
    web.stringByEvaluatingJavaScript(from: umhuellt(skript)) ?? "null"
  }

  func abzug() async throws -> Data {
    guard let rep = web.bitmapImageRepForCachingDisplay(in: web.bounds) else {
      throw Motorfehler(description: "keine Bildfläche")
    }
    web.cacheDisplay(in: web.bounds, to: rep)
    return try pngAus(rep)
  }
}

// Probefenster gibt sich als Schlüsselfenster aus.
//
// **Damit die Probe dem Menschen am Rechner nichts wegnimmt**: das Programm
// wird nie aktiv, fängt also keinen Tastendruck ab, der für ein anderes
// Fenster gedacht war. WebKit fragt aber, ob sein Fenster das
// Schlüsselfenster ist — sonst gilt die Seite als unscharf, `:focus-visible`
// greift nicht, und `document.hasFocus()` sagt nein.
final class Probefenster: NSWindow {
  override var isKeyWindow: Bool { true }
  override var canBecomeKey: Bool { true }
  override var canBecomeMain: Bool { true }
}

// **Den ersten Klick annehmen.** Ein Programm, das nie aktiv wird, bekommt
// jeden Klick als „ersten" — AppKit nimmt ihn zum Aktivieren und reicht ihn
// nur weiter, wenn die getroffene Ansicht `acceptsFirstMouse` bejaht.
// WKWebView verneint es; ohne diese Zeilen kam kein einziger Mausklick an.
final class Probeansicht: WKWebView {
  override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }
}

// ersterKlickAngenommen tut dasselbe für eine Klasse, die sich nicht
// ableiten lässt, weil WebKit sie selbst anlegt — `WebHTMLView` in `wk1`.
func ersterKlickAngenommen(_ klassenname: String) {
  guard let klasse = NSClassFromString(klassenname),
        let methode = class_getInstanceMethod(klasse, #selector(NSView.acceptsFirstMouse(for:))) else { return }
  let immer: @convention(block) (AnyObject, NSEvent?) -> Bool = { _, _ in true }
  method_setImplementation(methode, imp_implementationWithBlock(immer))
}
