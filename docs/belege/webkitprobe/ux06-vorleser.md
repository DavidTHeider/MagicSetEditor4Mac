# WebKit-Probe — ux06-vorleser.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk1 — WebCore im eigenen Prozess |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:49Z |

- ✓ Z7 `motor wk1`
- ✓ Z8 `lade /`
- ✓ Z9 `frist 20000`
- ✓ Z10 `warte nav`
- ✓ Z11 `wartetext Erste Schritte`
- ✓ Z12 `fokus`
  ```
  Ansage: AXButton «Zu den Sets» · button · fokussiert — DOM: "button "
    AXBlockQuoteLevel = 0
    AXChildren = (
  )
    AXChildrenInNavigationOrder = (
  )
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x77b8ba5c80 [0x1f91bef70]>{length = 28, bytes = 0x18000000e50000000a00000000000000 ... 0a00000000010000}
    AXFocused = 1
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXParent = wrapper 0x77b91c9280 { object {role: ListItem, ID 206, RenderListItem 0x124a7cfa0 LI 0x1241be350 class='rounded-md border border-kante p-3'} }
    AXPosition = NSPoint: {-19625, -19653}
    AXRole = AXButton
    AXRoleDescription = button
    AXSelected = 0
    AXSize = NSSize: {113, 42}
    AXStartTextMarker = <AXTextMarker 0x77b8ba6d80 [0x1f91bef70]>{length = 28, bytes = 0x180000002e0000000000000000000000 ... 0000000000010000}
    AXTitle = Zu den Sets
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x77b8d98f00>
    AXVisibleCharacterRange = NSRange: {0, 11}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x77b8d98f00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z13 `taste escape`
- ✓ Z14 `fokus`
  ```
  Ansage: (nichts fokussiert) — DOM: "body "
  ```
- ✓ Z15 `taste cmd+3`
- ✓ Z16 `wartetext Aufnehmen …`
- ✓ Z17 `klicketext Aufnehmen …`
- ✓ Z18 `warte [role=dialog][aria-label="Ein Exemplar aufnehmen"]`
- ✓ Z19 `klicketext Sprache`
- ✓ Z20 `fokus`
  ```
  Ansage: AXTextField «Sprache» · text field · fokussiert — DOM: "input#_r_1m_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
      "wrapper 0x77b9b27900 { object {role: Generic, ID 681, RenderBlock 0x124c90660 DIV 0x1248d8c40} }"
  )
    AXChildrenInNavigationOrder = (
      "wrapper 0x77b9b27900 { object {role: Generic, ID 681, RenderBlock 0x124c90660 DIV 0x1248d8c40} }"
  )
    AXDOMIdentifier = _r_1m_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x77b8400840 [0x1f91bef70]>{length = 28, bytes = 0x18000000aa0200001d00000000000000 ... 1d00000000010000}
    AXFocused = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXNumberOfCharacters = 0
    AXPlaceholderValue = z. B. en
    AXPosition = NSPoint: {-18606, -19392}
    AXRequired = 0
    AXRequired = 0
    AXRole = AXTextField
    AXRoleDescription = text field
    AXSelected = 0
    AXSelectedTextRange = NSRange: {0, 0}
    AXSize = NSSize: {64, 31}
    AXStartTextMarker = <AXTextMarker 0x77b8400940 [0x1f91bef70]>{length = 28, bytes = 0x180000002e0000000000000000000000 ... 0000000000010000}
    AXTitle = Sprache
    AXTitleUIElement = wrapper 0x77b9b25820 { object {role: Label, ID 436, RenderBlock 0x124c90440 LABEL 0x124aa4580 class='text-sm font-medium'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x77b8d98f00>
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 0}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x77b8d98f00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z21 `klicketext Weitere Angaben`
- ✓ Z22 `klicketext Währung`
- ✓ Z23 `fokus`
  ```
  Ansage: AXTextField «Währung» · text field · fokussiert — DOM: "input#_r_21_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
      "wrapper 0x77b9953580 { object {role: Generic, ID 744, RenderBlock 0x124c92200 DIV 0x1248d90c0} }"
  )
    AXChildrenInNavigationOrder = (
      "wrapper 0x77b9953580 { object {role: Generic, ID 744, RenderBlock 0x124c92200 DIV 0x1248d90c0} }"
  )
    AXDOMIdentifier = _r_21_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x77b89d7c80 [0x1f91bef70]>{length = 28, bytes = 0x18000000aa0200001d00000000000000 ... 1d00000000010000}
    AXFocused = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXNumberOfCharacters = 0
    AXPlaceholderValue = z. B. EUR
    AXPosition = NSPoint: {-18810, -19706}
    AXRequired = 0
    AXRequired = 0
    AXRole = AXTextField
    AXRoleDescription = text field
    AXSelected = 0
    AXSelectedTextRange = NSRange: {0, 0}
    AXSize = NSSize: {64, 31}
    AXStartTextMarker = <AXTextMarker 0x77b89d7dc0 [0x1f91bef70]>{length = 28, bytes = 0x180000002e0000000000000000000000 ... 0000000000010000}
    AXTitle = Währung
    AXTitleUIElement = wrapper 0x77b9b27aa0 { object {role: Label, ID 710, RenderBlock 0x124c91fe0 LABEL 0x124aa4a90 class='text-sm font-medium'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x77b8d98f00>
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 0}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x77b8d98f00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z24 `tippe €`
- ✓ Z25 `taste tab`
- ✓ Z26 `klicketext Währung`
- ✓ Z27 `fokus`
  ```
  Ansage: AXTextField «Währung» · text field · Wert «€» · ungültig · fokussiert — DOM: "input#_r_21_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
  )
    AXChildrenInNavigationOrder = (
  )
    AXDOMIdentifier = _r_21_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x77b8400980 [0x1f91bef70]>{length = 28, bytes = 0x18000000aa0200001d00000000000000 ... 1d00000000010000}
    AXFocused = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = true
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXNumberOfCharacters = 1
    AXPlaceholderValue = z. B. EUR
    AXPosition = NSPoint: {-18942, -19755}
    AXRequired = 0
    AXRequired = 0
    AXRole = AXTextField
    AXRoleDescription = text field
    AXSelected = 0
    AXSelectedTextRange = NSRange: {1, 0}
    AXSize = NSSize: {401, 31}
    AXStartTextMarker = <AXTextMarker 0x77b84007c0 [0x1f91bef70]>{length = 28, bytes = 0x180000002e0000000000000000000000 ... 0000000000010000}
    AXTitle = Währung
    AXTitleUIElement = wrapper 0x77b9b27aa0 { object {role: Label, ID 710, RenderBlock 0x124c91fe0 LABEL 0x124aa4a90 class='text-sm font-medium'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x77b8d98f00>
    AXValue = €
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 1}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x77b8d98f00>
    _AXPrimaryScreenHeight = 956
  ```

**Ergebnis: bestanden** — 21 Schritte.
