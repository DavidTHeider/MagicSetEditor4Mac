# WebKit-Probe — uxrf1-suchauswahl.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk1 — WebCore im eigenen Prozess |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:54:54Z |

- ✓ Z14 `motor wk1`
- ✓ Z15 `lade /`
- ✓ Z16 `frist 20000`
- ✓ Z17 `warte nav`
- ✓ Z18 `wartetext Erste Schritte`
- ✓ Z19 `taste escape`
- ✓ Z20 `taste cmd+,`
- ✓ Z21 `warte [role=tablist][aria-orientation=vertical]`
- ✓ Z22 `pruefe ([...document.querySelectorAll('[role=tab]')].find((t) => t.innerText.startsWith('Allgemein')).dataset.ziel = 'gruppe', true)`
- ✓ Z23 `klicke [data-ziel="gruppe"]`
- ✓ Z24 `warte input[data-einstellung="zeitzone"]`
- ✓ Z25 `notiz Zu: das Feld mit seiner Rolle, seinem Namen und seinem Wert — keine Liste im Baum`
- ✓ Z26 `pruefe (document.querySelector('input[data-einstellung="zeitzone"]').focus(), document.activeElement.dataset.einstellung === 'zeitzone')`
- ✓ Z27 `fokus`
  ```
  Ansage: AXComboBox «Zeitzone» · combo box · Wert «Wie das System» · fokussiert — DOM: "input#_r_q_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
      "wrapper 0x76571cb220 { object {role: StaticText, ID 365, RenderText 0x120bb8f50 #text 0x120b658f0 length=14 \"Wie das System\"} }"
  )
    AXChildrenInNavigationOrder = (
      "wrapper 0x76571cb220 { object {role: StaticText, ID 365, RenderText 0x120bb8f50 #text 0x120b658f0 length=14 \"Wie das System\"} }"
  )
    AXDOMIdentifier = _r_q_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x76571a4400 [0x1f91bef70]>{length = 28, bytes = 0x1d000000920100000000000000000000 ... 0000000000010000}
    AXExpanded = 0
    AXExpanded = 0
    AXFocusableAncestor = wrapper 0x76571c8ba0 { object {role: TabPanel, ID 58, RenderBlock 0x120b56970 DIV 0x1201bb430 id='_r_f_-feld' class='min-w-0 flex-1'} }
    AXFocused = 1
    AXHasPopup = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = false
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXNumberOfCharacters = 14
    AXOrientation = AXUnknownOrientation
    AXParent = wrapper 0x76571c9360 { object {role: Generic, ID 120, RenderBlock (relative positioned) 0x120b57c90 DIV 0x1201ba980 class='relative'} }
    AXPosition = NSPoint: {-19439, -19571}
    AXRequired = 0
    AXRequired = 0
    AXRequired = 0
    AXRole = AXComboBox
    AXRoleDescription = combo box
    AXSelected = 0
    AXSelectedTextRange = NSRange: {14, 0}
    AXSize = NSSize: {320, 36}
    AXStartTextMarker = <AXTextMarker 0x76571a4840 [0x1f91bef70]>{length = 28, bytes = 0x1d000000980000000000000000000000 ... 0000000000010000}
    AXTitle = Zeitzone
    AXTitleUIElement = wrapper 0x76571c9300 { object {role: Label, ID 117, RenderFlexibleBox 0x120aeea20 LABEL 0x120a80340 class='flex max-w-xs flex-col gap-1'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x7656d8cf00>
    AXValue = Wie das System
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 14}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x7656d8cf00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z28 `baum input[data-einstellung="zeitzone"]`
  ```
  AXComboBox «Zeitzone» · combo box · Wert «Wie das System» · fokussiert
    AXStaticText «» · text · Wert «Wie das System»
  ```
- ✓ Z29 `gleich document.querySelectorAll('[role=listbox]').length => 0`
- ✓ Z30 `notiz Ein Klick öffnet die Liste: das Feld meldet „aufgeklappt", die Liste ist eine benannte listbox`
- ✓ Z31 `klicke input[data-einstellung="zeitzone"]`
- ✓ Z32 `warte [role=listbox]`
- ✓ Z33 `fokus`
  ```
  Ansage: AXComboBox «Zeitzone» · combo box · aufgeklappt · fokussiert — DOM: "input#_r_q_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
      "wrapper 0x765793d360 { object {role: Generic, ID 809, RenderBlock 0x120bc5980 DIV 0x1208ac850} }"
  )
    AXChildrenInNavigationOrder = (
      "wrapper 0x765793d360 { object {role: Generic, ID 809, RenderBlock 0x120bc5980 DIV 0x1208ac850} }"
  )
    AXDOMIdentifier = _r_q_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x7657af4200 [0x1f91bef70]>{length = 28, bytes = 0x1d000000990100003100000000000000 ... 3100000000010000}
    AXExpanded = 1
    AXExpanded = 1
    AXFocusableAncestor = wrapper 0x76571c8ba0 { object {role: TabPanel, ID 58, RenderBlock 0x120b56970 DIV 0x1201bb430 id='_r_f_-feld' class='min-w-0 flex-1'} }
    AXFocused = 1
    AXHasPopup = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = false
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
      "wrapper 0x76579ed8a0 { object {role: ListBox, ID 405, RenderBlock 0x120bc5ba0 DIV 0x120bee3e0 id='_r_s_' class='mantine-Combobox-options max-h-64 overflow-y-auto py-1'} }"
  )
    AXNumberOfCharacters = 0
    AXOrientation = AXUnknownOrientation
    AXParent = wrapper 0x76571c9360 { object {role: Generic, ID 120, RenderBlock (relative positioned) 0x120b57c90 DIV 0x1201ba980 class='relative'} }
    AXPlaceholderValue = Wie das System
    AXPosition = NSPoint: {-19439, -19570}
    AXRequired = 0
    AXRequired = 0
    AXRequired = 0
    AXRole = AXComboBox
    AXRoleDescription = combo box
    AXSelected = 0
    AXSelectedTextRange = NSRange: {0, 0}
    AXSize = NSSize: {320, 36}
    AXStartTextMarker = <AXTextMarker 0x7657af4640 [0x1f91bef70]>{length = 28, bytes = 0x1d000000980000000000000000000000 ... 0000000000010000}
    AXTitle = Zeitzone
    AXTitleUIElement = wrapper 0x76571c9300 { object {role: Label, ID 117, RenderFlexibleBox 0x120aeea20 LABEL 0x120a80340 class='flex max-w-xs flex-col gap-1'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x7656d8cf00>
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 0}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x7656d8cf00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z34 `gleich document.querySelector('input[data-einstellung="zeitzone"]').getAttribute('aria-expanded') => true`
- ✓ Z35 `gleich document.querySelector('[role=listbox]').getAttribute('aria-label') => Zeitzone`
- ✓ Z36 `gleich document.querySelector('input[data-einstellung="zeitzone"]').getAttribute('aria-controls') === document.querySelector('[role=listbox]').id => true`
- ✓ Z37 `zeige document.querySelectorAll('[role=listbox] [role=option]').length + ' Zeilen, die erste: ' + document.querySelector('[role=listbox] [role=option]').innerText`
  ```
  "200 Zeilen, die erste: Wie das System"
  ```
- ✓ Z38 `abzug offen`
  ```
  Bild: offen.png
  ```
- ✓ Z39 `notiz Tippen grenzt ein — die Liste im Baum ist die eine Zeile, und die Live-Region sagt die Zahl`
- ✓ Z40 `tippe new york`
- ✓ Z41 `bis document.querySelectorAll('[role=listbox] [role=option]').length === 1`
- ✓ Z42 `baum [role=listbox]`
  ```
  AXList «Zeitzone» · list
    AXStaticText «America/New_York» · text · gewählt
  ```
- ✓ Z43 `zeige document.querySelector('[data-suchauswahl]').innerText`
  ```
  "Treffer: 1"
  ```
- ✓ Z44 `abzug gefiltert`
  ```
  Bild: gefiltert.png
  ```
- ✓ Z45 `notiz Enter wählt und schließt — das Feld nennt die Zone, die Liste ist aus dem Baum`
- ✓ Z46 `taste enter`
- ✓ Z47 `bis document.querySelectorAll('[role=listbox]').length === 0`
- ✓ Z48 `gleich document.querySelector('input[data-einstellung="zeitzone"]').value => America/New_York`
- ✓ Z49 `gleich document.querySelector('input[data-einstellung="zeitzone"]').getAttribute('aria-expanded') => false`
- ✓ Z50 `fokus`
  ```
  Ansage: AXComboBox «Zeitzone» · combo box · Wert «America/New_York» · fokussiert — DOM: "input#_r_q_ "
    AXBlockQuoteLevel = 0
    AXChildren = (
      "wrapper 0x76572d87e0 { object {role: StaticText, ID 812, RenderText 0x120cb9ce0 #text 0x120229b90 length=16 \"America/New_York\"} }"
  )
    AXChildrenInNavigationOrder = (
      "wrapper 0x76572d87e0 { object {role: StaticText, ID 812, RenderText 0x120cb9ce0 #text 0x120229b90 length=16 \"America/New_York\"} }"
  )
    AXDOMIdentifier = _r_q_
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x7657b535c0 [0x1f91bef70]>{length = 28, bytes = 0x1d000000920100000000000000000000 ... 0000000000010000}
    AXExpanded = 0
    AXExpanded = 0
    AXFocusableAncestor = wrapper 0x76571c8ba0 { object {role: TabPanel, ID 58, RenderBlock 0x120b56970 DIV 0x1201bb430 id='_r_f_-feld' class='min-w-0 flex-1'} }
    AXFocused = 1
    AXHasPopup = 1
    AXInsertionPointLineNumber = 0
    AXInvalid = false
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXNumberOfCharacters = 16
    AXOrientation = AXUnknownOrientation
    AXParent = wrapper 0x76571c9360 { object {role: Generic, ID 120, RenderBlock (relative positioned) 0x120b57c90 DIV 0x1201ba980 class='relative'} }
    AXPosition = NSPoint: {-19439, -19570}
    AXRequired = 0
    AXRequired = 0
    AXRequired = 0
    AXRole = AXComboBox
    AXRoleDescription = combo box
    AXSelected = 0
    AXSelectedTextRange = NSRange: {16, 0}
    AXSize = NSSize: {320, 36}
    AXStartTextMarker = <AXTextMarker 0x7657b53180 [0x1f91bef70]>{length = 28, bytes = 0x1d000000980000000000000000000000 ... 0000000000010000}
    AXTitle = Zeitzone
    AXTitleUIElement = wrapper 0x76571c9300 { object {role: Label, ID 117, RenderFlexibleBox 0x120aeea20 LABEL 0x120a80340 class='flex max-w-xs flex-col gap-1'} }
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x7656d8cf00>
    AXValue = America/New_York
    AXValueAutofillAvailable = 0
    AXVisibleCharacterRange = NSRange: {0, 16}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x7656d8cf00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z51 `baum input[data-einstellung="zeitzone"]`
  ```
  AXComboBox «Zeitzone» · combo box · Wert «America/New_York» · fokussiert
    AXStaticText «» · text · Wert «America/New_York»
  ```
- ✓ Z52 `abzug gewaehlt`
  ```
  Bild: gewaehlt.png
  ```
- ✓ Z53 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 40 Schritte.
