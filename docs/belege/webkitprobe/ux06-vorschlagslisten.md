# WebKit-Probe — ux06-vorschlagslisten.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:51Z |

- ✓ Z7 `motor wk2`
- ✓ Z8 `lade /`
- ✓ Z9 `frist 20000`
- ✓ Z10 `warte nav`
- ✓ Z11 `wartetext Erste Schritte`
- ✓ Z12 `taste escape`
- ✓ Z13 `taste cmd+3`
- ✓ Z14 `wartetext Aufnehmen …`
- ✓ Z15 `klicketext Aufnehmen …`
- ✓ Z16 `warte [role=dialog][aria-label="Ein Exemplar aufnehmen"]`
- ✓ Z17 `klicketext Weitere Angaben`
- ✓ Z18 `zeige [...document.querySelectorAll('[role=dialog] input[list]')].map(e => (e.labels && e.labels[0] ? e.labels[0].innerText.trim() : '?') + ' → ' + e.getAttribute('list') + ' mit ' + document.getElementById(e.getAttribute('list')).options.length + ' Vorschlägen')`
  ```
  ["Sprache → _r_1l_ mit 20 Vorschlägen","Währung → _r_20_ mit 307 Vorschlägen"]
  ```
- ✓ Z19 `klicketext Sprache`
- ✓ Z20 `pruefe document.activeElement.getAttribute('list') !== null`
- ✓ Z21 `tippe p`
- ✓ Z22 `fenster sprache`
  ```
  WKDataListSuggestionWindow 10×10
    Zeile 1: pt — Portugiesisch
    Zeile 2: ph — Phyrexianisch
    Zeile 3: es — Spanisch
    Zeile 4: ja — Japanisch
    Bild: sprache-1.png
  ```
- ✓ Z23 `tippe h`
- ✓ Z24 `fenster sprache-ph`
  ```
  WKDataListSuggestionWindow 10×10
    Zeile 1: ph — Phyrexianisch
    Bild: sprache-ph-1.png
  ```
- ✓ Z25 `taste down`
- ✓ Z26 `taste enter`
- ✓ Z27 `gleich document.activeElement.value => ph`
- ✓ Z28 `klicketext Währung`
- ✓ Z29 `tippe eu`
- ✓ Z30 `fenster waehrung`
  ```
  WKDataListSuggestionWindow 10×10
    Zeile 1: EUR — Euro
    Zeile 2: AON — Angolanischer Neuer Kwanza (1990–2000)
    Zeile 3: BAN — Bosnien und Herzegowina Neuer Dinar (1994–1997)
    Zeile 4: CHE — WIR-Euro
    Zeile 5: DEM — Deutsche Mark
    Zeile 6: ILS — Israelischer Neuer Schekel
    Zeile 7: MDL — Moldau-Leu
    Zeile 8: NZD — Neuseeland-Dollar
    Zeile 9: PGK — Papua-neuguineischer Kina
    Zeile 10: ROL — Rumänischer Leu (1952–2006)
    Zeile 11: RON — Rumänischer Leu
    Zeile 12: TWD — Neuer Taiwan-Dollar
    Zeile 13: XBA — Europäische Rechnungseinheit
    Zeile 14: XBB — Europäische Währungseinheit (XBB)
    Zeile 15: XBC — Europäische Rechnungseinheit (XBC)
    Zeile 16: XBD — Europäische Rechnungseinheit (XBD)
    Zeile 17: XEU — Europäische Währungseinheit (XEU)
    Zeile 18: YUM — Jugoslawischer Neuer Dinar (1994–2002)
    Zeile 19: ZRN — Zaire-Neuer Zaïre (1993–1998)
    Bild: waehrung-1.png
  ```
- ✓ Z31 `taste down`
- ✓ Z32 `taste enter`
- ✓ Z33 `gleich document.activeElement.value => EUR`
- ✓ Z34 `pruefe !document.activeElement.matches('[aria-invalid=true]')`
- ✓ Z35 `abzug formular`
  ```
  Bild: formular.png
  ```

**Ergebnis: bestanden** — 29 Schritte.
