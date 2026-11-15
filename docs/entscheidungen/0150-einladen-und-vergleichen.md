# 0150 — Einladen und vergleichen: eine unterschriebene Einladung und ein Code je Paar

**Stand:** entschieden am 2026-09-30 · NET-01.AK-01 bis AK-04
(#1705)

## Die Frage

NET-01 verlangt, dass Verbinden als „Person wählen oder einladen, Code
vergleichen, Zweck freigeben" funktioniert. Bis hierher tauschte man eine
**Kennung** von Hand — 52 Zeichen ohne Namen, ohne Zweck, ohne Ablauf — und
bestätigte, indem jeder den **Fingerabdruck** des anderen abtippte. AK-02
verlangt, dass Code oder QR-Daten die erwartete Gegenstelle, Zweck,
Protokollfassung und Ablauf binden; AK-03 vier unterscheidbare Zustände;
AK-04 keinen halb bestätigten Kontakt.

## Was MSE2 tut

**Es verbindet keine Rechner** (gesucht nach `socket`, `pairing`, `invit`,
`fingerprint` in `/Users/david/mse2-quelle/src`, Fassung f9d9356 vom
2021-07-06: einzig `wxSocketBase::Initialize()` in `main.cpp`). Marke
**new**.

## Was gilt

1. **Die Einladung trägt den ganzen Schlüssel und ist von ihm
   unterschrieben.** `mse4mac-kontakt:1:<Daten>.<Unterschrift>`, die Daten
   JSON mit Kennung, Name, Zweck, Adresse, Ablauf, Nonce und — als
   Behauptung — der Kennung vor der letzten Rotation. Unterschrieben wird
   über `mse4mac-kontakt/1\n` und die Bytes, **die hinausgehen**; geprüft wird
   über die Bytes, **die ankamen**. Damit bindet sie, was AK-02 nennt.
2. **Sie ist kein kurzer Code**, und deshalb braucht sie kein PAKE: zu raten
   gibt es nichts. **Wiederholungsschutz** ist die Nonce, verbraucht in
   derselben Transaktion wie der Kontakt (Wanderung 0094). **Ablauf** ein
   Tag, angenommen wird höchstens sieben.
3. **Beide tauschen**, wie das Artboard es zeichnet — erst wer die Einladung
   des anderen angenommen hat, kennt beide Schlüssel. Ein Einlösen beim
   Einladenden (eine Einladung statt zwei) war ein **eigener Schnitt** — er
   kam mit [0156](0156-eine-einladung-genuegt.md), **ohne** das Drahtformat
   des Handschlags zu ändern: über eine Paarungssitzung. Der Tausch zweier
   Einladungen bleibt als Rückfall.
4. **Der Vergleichscode gehört dem Paar.** Beide sehen dieselben dreißig
   Ziffern; ein einziger Vergleich genügt. Das Verfahren ist das der
   **Sicherheitsnummern von Signal** — nachgelesen in `signalapp/libsignal`,
   `rust/protocol/src/fingerprint.rs`, Fassung 4ea5b76 vom 2025-09-03, und
   **nicht übernommen** (AGPL-3.0-only verträgt sich mit GPL-2.0-or-later
   nicht): je Seite 5.200 Runden SHA-512 über Fassung und Schlüssel, je fünf
   Bytes fünf Ziffern, die kleinere Hälfte zuerst. **Abweichend**: drei
   Gruppen je Seite statt sechs.
5. **Was der Code kostet**: wer sich zwischen zwei Menschen stellt, muss für
   jede Seite einen Schlüssel finden, dessen Hälfte die eines bestimmten
   Gegenübers trifft — rund 2^50 Schlüssel je Hälfte, jeder mit 5.200
   Hashes. Das hält einen Menschen mit Rechnern auf, keinen Staat mit
   Rechenzentrum; es steht hier, damit es niemand behauptet. **Die
   getrennten Hälften sind Signals Gestalt und kein Gewinn an Stärke**: ein
   Code über beide Schlüssel zugleich läge bei einem Geburtstag über dreißig
   Ziffern ebenfalls bei rund 2^50. Zuerst stand im Kommentar das Gegenteil;
   nachgerechnet beim Schreiben dieses Vermerks.
6. **Annehmen ist ein Schritt mit drei Ausgängen**: ohne Vergleich ein
   unbestätigter Kontakt, mit passendem Code ein bestätigter, mit falschem
   **keiner** — auch keine verbrauchte Nonce. Ein Zweck geht nur mit
   Vergleich und nur, wenn jemand ihn anhakt; vorgewählt ist er nicht.
7. **Fünf Lagen, drei Stände.** Gespeichert bleiben `unbestaetigt`,
   `bestaetigt` und `gesperrt`; die Lage leitet `unbekannt` (keine Zeile),
   `widerrufen` (gesperrt) und **`schluesselGeaendert`** ab — ein
   unbestätigter Kontakt, dessen Einladung eine hier bekannte Kennung als
   Vorgängerin nennt. **Kein Erbe**: Rechte und Stand des Vorgängers gehen
   nicht über.
8. **Keine der drei Handlungen über MCP.** Annehmen ist eine
   Vertrauensentscheidung ([0044](0044-was-ein-werkzeug-schreiben-darf.md)),
   und eine Einladung sagt, wie dieser Rechner zu erreichen ist.
9. **Der Code steht im Rumpf und nie im Pfad** — auch die lesende Prüfung ist
   ein `POST`, damit er in keinem Protokoll landet.

## Verworfen

- **Vier Wörter**, wie das Artboard sie zeigt: es gibt acht Sprachen im
  Fenster und keine geprüfte Wortliste in einer davon, und selbst eine Liste
  von 2.048 Wörtern gäbe vier Wörtern nur 44 Bit — rund 2^22 Versuche je
  Seite, in Minuten gerechnet. Ziffern sind in jeder Sprache dieselben;
  Signal hat sie aus demselben Grund. Die Abweichung steht in
  `PROTOTYP-0.6.0.md` (Entwicklungsarchiv).
- **Ein sechsstelliger Code nach Bluetooth-Art**: sicher nur mit einem
  Festlegen-und-Aufdecken im Handschlag, und den gibt es nicht. Ohne es
  wären 10^6 Versuche in Sekunden gerechnet.
- **Den Fingerabdruck ersetzen**: er bleibt, hinter „Technische Angaben" und
  als Weg, wo kein Code da ist — etwa ohne eigene Identität.

## Was offen bleibt

- **Einlösen beim Einladenden** (ein Code statt zweier), mit Ratengrenze —
  der zweite Schnitt von NET-01.
- **Ein QR-Bild** der Einladung — NET-04 teilt denselben Inhalt.
- **Die Abnahme mit zwei Rechnern am Fenster** (NAT-01): die Flächenläufe
  fahren zwei Identitäten mit eigenem Bestand, aber kein zweites Fenster.
