# WebKit-Probe — ux10-vorleser.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk1 — WebCore im eigenen Prozess |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:52:08Z |

- ✓ Z10 `motor wk1`
- ✓ Z11 `lade /`
- ✓ Z12 `frist 20000`
- ✓ Z13 `warte nav`
- ✓ Z14 `wartetext Erste Schritte`
- ✓ Z15 `taste escape`
- ✓ Z19 `taste cmd+k`
- ✓ Z20 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z21 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((x) => x.getClientRects().length > 0 && x.innerText.trim() === 'Sammlung').dataset.ziel = 'seite', true)`
- ✓ Z22 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z23 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z24 `bis [...document.querySelectorAll('main [role=grid][aria-label="Exemplare"] [role=gridcell][title]')].some((z) => z.checkVisibility())`
- ✓ Z25 `pruefe ([...document.querySelectorAll('main [role=grid][aria-label="Exemplare"] [role=gridcell][title]')].find((z) => z.checkVisibility()).dataset.ziel = 'zeile', true)`
- ✓ Z26 `klicke [data-ziel="zeile"]`
- ✓ Z27 `bis [...document.querySelectorAll('aside[aria-label="Exemplar"] button[data-aktion="loeschen"]')].some((b) => b.checkVisibility())`
- ✓ Z28 `notiz Der Löschknopf eines Exemplars — sein Zeichen steht im Baum nicht`
- ✓ Z29 `pruefe (document.querySelector('aside[aria-label="Exemplar"] button[data-aktion="loeschen"]').focus(), document.activeElement.dataset.aktion === 'loeschen')`
- ✓ Z30 `fokus`
  ```
  Ansage: AXButton «Löschen» · button · fokussiert — DOM: "button "
    AXBlockQuoteLevel = 0
    AXChildren = (
  )
    AXChildrenInNavigationOrder = (
  )
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x7668b76b00 [0x1f91bef70]>{length = 28, bytes = 0x2e000000ee0100000800000000000000 ... 0800000000010000}
    AXFocusableAncestor = wrapper 0x76692d6ce0 { object {role: LandmarkMain, ID 67, RenderBlock 0x1221c71f0 MAIN 0x1221b91e0 id='inhalt' class='flex-1 px-8 pb-8 pt-4 focus:outline-none'} }
    AXFocused = 1
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXParent = wrapper 0x76692d4ca0 { object {role: LandmarkComplementary, ID 81, RenderFlexibleBox 0x122b104a0 ASIDE 0x1221bb3a0 id='_r_o_-inspector' class='flex min-h-0 w-80 shrink-0 flex-col overflow-y-auto'} }
    AXPosition = NSPoint: {-18777, -19507}
    AXRole = AXButton
    AXRoleDescription = button
    AXSelected = 0
    AXSize = NSSize: {107, 44}
    AXStartTextMarker = <AXTextMarker 0x7668b76cc0 [0x1f91bef70]>{length = 28, bytes = 0x2e000000780000000000000000000000 ... 0000000000010000}
    AXTitle = Löschen
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x7668d88f00>
    AXVisibleCharacterRange = NSRange: {0, 11}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x7668d88f00>
    _AXPrimaryScreenHeight = 956
  ```
- ✓ Z31 `baum aside[aria-label="Exemplar"] button[data-aktion="loeschen"]`
  ```
  AXButton «Löschen» · button · fokussiert
  ```
- ✓ Z32 `notiz Und der Knopf zum Bearbeiten daneben — ✎ steht davor, gesagt wird es nicht`
- ✓ Z33 `pruefe (document.querySelector('aside[aria-label="Exemplar"] button[data-aktion="bearbeiten"]').focus(), document.activeElement.dataset.aktion === 'bearbeiten')`
- ✓ Z34 `fokus`
  ```
  Ansage: AXButton «Bearbeiten» · button · fokussiert — DOM: "button "
    AXBlockQuoteLevel = 0
    AXChildren = (
  )
    AXChildrenInNavigationOrder = (
  )
    AXElementBusy = 0
    AXEnabled = 1
    AXEndTextMarker = <AXTextMarker 0x7668b76d00 [0x1f91bef70]>{length = 28, bytes = 0x2e000000ee0100000800000000000000 ... 0800000000010000}
    AXFocusableAncestor = wrapper 0x76692d6ce0 { object {role: LandmarkMain, ID 67, RenderBlock 0x1221c71f0 MAIN 0x1221b91e0 id='inhalt' class='flex-1 px-8 pb-8 pt-4 focus:outline-none'} }
    AXFocused = 1
    AXInvalid = false
    AXLanguage = de
    AXLinkedUIElements = (
  )
    AXParent = wrapper 0x76692d4ca0 { object {role: LandmarkComplementary, ID 81, RenderFlexibleBox 0x122b104a0 ASIDE 0x1221bb3a0 id='_r_o_-inspector' class='flex min-h-0 w-80 shrink-0 flex-col overflow-y-auto'} }
    AXPosition = NSPoint: {-18912, -19507}
    AXRole = AXButton
    AXRoleDescription = button
    AXSelected = 0
    AXSize = NSSize: {128, 44}
    AXStartTextMarker = <AXTextMarker 0x7668b76d40 [0x1f91bef70]>{length = 28, bytes = 0x2e000000780000000000000000000000 ... 0000000000010000}
    AXTitle = Bearbeiten
    AXTopLevelUIElement = <webkitprobe.Probefenster: 0x7668d88f00>
    AXVisibleCharacterRange = NSRange: {0, 14}
    AXVisited = 0
    AXWindow = <webkitprobe.Probefenster: 0x7668d88f00>
    _AXPrimaryScreenHeight = 956
  ```

**Ergebnis: bestanden** — 22 Schritte.
