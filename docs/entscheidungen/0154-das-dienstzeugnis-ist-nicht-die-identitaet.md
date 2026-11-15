# 0154 — Das Zeugnis des Dienstes ist nicht die Identität des Profils

**Stand:** angenommen · **Datum:** 2026-09-30
**Schreibt fort:** [0115](0115-die-schleife-ist-eine-vorgabe.md). Der Absatz
„Kein Zertifikat einer Zertifizierungsstelle“ fällt als Verbot und bleibt als
Vorgabe.
**Berührt:** [0113](0113-kein-libp2p-im-lan.md), `docs/NETZEIGENSCHAFTEN.md`
(Abschnitt *Als Dienst*)
**Issue:** NET-03 (#1707)

## Die Frage

`mse-rest` und `mse-mcp` zeigen jenseits der Schleife ein Zeugnis vor, das
**die Peeridentität trägt**: selbst unterschrieben und Ed25519. Das hat zwei
Folgen, die mit dem Dienst nichts zu tun haben:

- **Wer sich verbindet, kann nur den Fingerabdruck vergleichen.** Eine
  Zertifizierungsstelle, der ein Haus vertraut, kann er nicht nutzen.
- **`/usr/bin/curl` von macOS scheitert am Handschlag.** Gemessen am
  2026-09-15: LibreSSL 3.3.6 kennt Ed25519 in einem Serverzeugnis nicht.

## Was gilt

1. **Zwei Identitäten, zwei Zwecke.** Die **Peeridentität** unterschreibt
   Nachrichten und nagelt die Peerleitung fest, Kennung gegen Kennung. Das
   **Dienstzeugnis** weist einen HTTPS-Dienst gegenüber einem Klienten aus,
   der ihn nach Namen und Kette prüft. Keines ersetzt das andere, und keines
   leitet sich aus dem anderen ab.
2. **Die Vorgabe bleibt**: ohne Angaben zeigt der Dienst das Zeugnis auf der
   Peeridentität, wie seit N6. Niemand braucht eine eigene PKI, damit es
   funktioniert.
3. **Eine eigene PKI ist möglich**: Zertifikat, Schlüssel, Zwischenkette und
   Vertrauensanker. Die vier sind Einstellungen des Geräts
   (`dienst.tls.*`), und `--zeugnis`, `--schluessel`, `--kette` und
   `--anker` übersteuern sie, Feld für Feld. **Eine halbe Angabe startet
   nichts** — wer einen Anker einträgt und kein Zertifikat, bekäme sonst
   still das Zeugnis der Identität, und seine Klienten wiesen es ab. Die Einstellungen sind die
   **Konfigurationsdatei**: sie liegen im Bestand dieses Rechners, das
   Fenster und `mse einstellung` schreiben sie, und die Dienste lesen sie beim
   Start.
4. **Geprüft wird vor dem Start und nicht beim ersten Klienten**, und zwar vor
   `net.Listen`:
   - der Schlüssel passt zum Zertifikat;
   - es gilt jetzt, also nicht vor `NotBefore` und nicht nach `NotAfter`;
   - es nennt wenigstens einen Namen oder eine Adresse (SAN), und bei einer
     bestimmten Bindungsadresse **diese**;
   - die Kette reicht bis zu einem Anker, zum Serverzweck;
   - die Schlüsseldatei ist für niemanden sonst lesbar (`0600`).

   Scheitert eines davon, startet der Dienst nicht und sagt, was fehlt.
5. **Rotation**: am Handschlag wird nachgesehen, ob sich die Dateien geändert
   haben, höchstens alle fünf Sekunden. Ein neues Paar, das die Prüfung
   besteht, gilt ab dem nächsten Handschlag. Eines, das sie nicht besteht,
   gilt nicht. Dann bleibt das alte, solange es gültig ist, und der Befund
   steht im Protokoll. **Ein abgelaufenes Zeugnis wird nie vorgezeigt**: der
   Handschlag scheitert, statt eines zu zeigen, das der Klient abweisen
   müsste.
6. **Angesagt wird der Fingerabdruck des vorgezeigten Zeugnisses** (SHA-256
   über das Zertifikat, in der Form von `openssl x509 -fingerprint -sha256`),
   dazu das Ende seiner Gültigkeit.
7. **Die Namen des Zeugnisses zählen als eigene Namen**: wer über den
   Namen kommt, den die Stelle beglaubigt hat, schickt ihn als `Origin`,
   und die Herkunftsprüfung lässt ihn durch.
8. **Eine eigene PKI schaltet keine Prüfung ab.** Der Zugang bleibt Pflicht
   jenseits der Schleife, `Origin` wird geprüft wie bisher, und kein Weg in
   diesem Programm setzt `InsecureSkipVerify`.

## Was es kostet

- **Vier Einstellungen und vier Flaggen mehr**, und eine Prüfung, die
  Dateien liest. Die Rechte der Schlüsseldatei sind eine Eigenschaft von
  macOS und Linux; unter Windows würde nach diesem Satz jede Datei
  abgewiesen. Gebaut wird nur für macOS (§12), und `docs/PORTIERUNG.md`
  bekommt die Zeile.
- **Wer eine PKI betreibt, betreibt auch ihre Rotation.** Das Programm liest
  neu, wenn sich die Datei ändert, bestellt aber kein neues Zertifikat.

## Verworfen

- **Das Dienstzeugnis auf ECDSA umstellen und die Peeridentität lassen**:
  das wäre ein zweites Schlüsselpaar desselben Profils, ohne dass jemand es
  bestellt hat, und es löste nur das `curl`-Problem, nicht die Kette.
- **ACME**: ein Dienst auf einem Heimrechner hat meist keinen öffentlichen
  Namen, und ein Konto bei einem Dritten wäre wieder das Problem, das
  niemand hat (0115).
- **Neu lesen per Signal**: `SIGHUP` erreicht einen Dienst, den das Fenster
  gestartet hat, nur über einen Umweg, und am Handschlag nachzusehen kostet
  einen `stat` alle fünf Sekunden.

## Gemessen

- **`internal/leitung`**: jede Prüfung mit ihrem Nein, dazu die Rotation mit
  ihren drei Zuständen und ein prüfender TLS-Klient (`zeugnispruefung_test.go`,
  `zeugnisrotation_test.go`). Die Gegenproben: ohne die Rechteprüfung, ohne
  die Adressprüfung, ohne die Ablaufsperre und ohne das Merken eines
  kaputten Standes wird je ein Lauf rot.
- **Am gebauten Dienst** (`internal/flaechen/dienstpki_test.go`), mit
  `mse-rest` und `mse-mcp` an `0.0.0.0`:
  - **`/usr/bin/curl` von macOS** (LibreSSL) mit `--cacert` und ohne `-k`
    bekommt `200`. Am Zeugnis der Identität scheitert es am Handschlag.
  - ein **Go-Klient mit dem Anker** bekommt `200`, einer mit den Ankern des
    Systems bekommt keine Antwort, und ohne Marke kommt `401`;
  - **MCP `initialize`** über HTTPS gibt die Protokollfassung zurück;
  - **über die Einstellungen** gesetzt, ohne Flaggen gestartet: `200`.
