# Beleg: I01 — trägt ein eigener Herausgeber die Bindung über eine Fassung?

**Stand:** Commit `c6e379ba6` · **Datum:** 2026-10-11 06:55 UTC · **Befehl:** `node scripts/bindunggast.mjs mse080-152` · **Rückgabe:** 0

**Maschine:** Lume-VM `mse080-152` — macOS 15.2 (24C101), LibreSSL 3.3.6. **Nur ein künstlicher Bund** mit Probekennwort, kurz vorn in der Suchliste des Gastes; zwei selbst ausgestellte Herausgeber (A, B), im selben Bund und mit ihm gelöscht. **Auf dem Wirt wurde kein Herausgeberzeugnis eingerichtet.** Aufgeräumt: ja.

**Ergebnis:** jede der 12 Erwartungen trifft.

| Fall | Leser | erwartet | gemessen | trifft |
| --- | --- | --- | --- | --- |
| E1 heute | security | liest | liest | ja |
| E1 heute | leser-a1 | liest nicht | müsste fragen | ja |
| E2 gebunden | security | liest nicht | bekommt nichts (rc=0) | ja |
| E2 gebunden | leser-a1 | liest | liest | ja |
| E3 neue Fassung, Herausgeber A | leser-a2 | liest | liest | ja |
| E3 anderer Herausgeber, gleicher Bezeichner | leser-b1 | liest nicht | müsste fragen | ja |
| E3 ad hoc, gleicher Bezeichner | leser-adhoc | liest nicht | müsste fragen | ja |
| E5 fremder Code in der gebundenen Fassung | fremd in leser-a1 | liest | liest | ja |
| E5 gehärtete Fassung von A | leser-a3 | liest | liest | ja |
| E5 fremder Code in der gehärteten Fassung | fremd in leser-a3 | liest nicht | läuft nicht | ja |
| E4 gesperrt | security | liest nicht | bekommt nichts (rc=0) | ja |
| E4 gesperrt | leser-a1 | liest nicht | müsste fragen | ja |

**Die Anforderungen der Fassungen** (`codesign -d -r-`, dazu die Signaturflags — `runtime` heißt gehärtet):

- `leser-a1` (cdhash `f59d488dfdbf…`, flags `0x0(none)`): `identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"`
- `leser-a2` (cdhash `3d0375483891…`, flags `0x0(none)`): `identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"`
- `leser-a3` (cdhash `2067f7aec13a…`, flags `0x10000(runtime)`): `identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"`
- `leser-adhoc` (cdhash `cae43d9fe120…`, flags `0x2(adhoc)`): `cdhash H"cae43d9fe120079c1f799bc65ad26662f7759cb2"`
- `leser-b1` (cdhash `c2a0aac2b578…`, flags `0x0(none)`): `identifier "de.mse4mac.i01.leser" and certificate leaf = H"658e9f8689d433da0d677e7956b6422a4bc9ad73"`

**Gatekeeper** (`spctl --assess --type execute`), getrennt gemessen:

- `leser-a1`: /Users/lume/i01/leser-a1: rejected origin=MSE4Mac I01 Herausgeber A
- `leser-a2`: /Users/lume/i01/leser-a2: rejected origin=MSE4Mac I01 Herausgeber A
- `leser-a3`: /Users/lume/i01/leser-a3: rejected origin=MSE4Mac I01 Herausgeber A
- `leser-adhoc`: /Users/lume/i01/leser-adhoc: rejected
- `leser-b1`: /Users/lume/i01/leser-b1: rejected origin=MSE4Mac I01 Herausgeber B

**Die ACL des gebundenen Eintrags** (`security dump-keychain -a`, gekürzt): ` "type"<uint32>=<NULL> access: 3 entries entry 0: authorizations (1): encrypt don't-require-password description: dienst2 applications: <null> entry 1: authorizations (6): decrypt derive export_clear export_wrapped mac sign don't-require-password description: dienst2 applications (1): 0: /Users/lume/i01/leser-a1 (OK) requirement: identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf`

## Das Protokoll

```text
$ security create-keychain -p *** '/Users/lume/i01/bund.keychain-db'

rc=0
$ security unlock-keychain -p *** '/Users/lume/i01/bund.keychain-db'

rc=0
$ security set-keychain-settings '/Users/lume/i01/bund.keychain-db'

rc=0
$ security list-keychains -d user
    "/Users/lume/i01/bund.keychain-db"
    "/Users/lume/Library/Keychains/login.keychain-db"
rc=0
$ security list-keychains -d user -s '/Users/lume/i01/bund.keychain-db' '/Users/lume/Library/Keychains/login.keychain-db'

rc=0
$ openssl req -x509 -newkey rsa:2048 -nodes -days 2 -keyout '/Users/lume/i01/A.key' -out '/Users/lume/i01/A.crt' -subj '/CN=MSE4Mac I01 Herausgeber A' -addext 'keyUsage=critical,digitalSignature' -addext 'extendedKeyUsage=critical,codeSigning'
Generating a 2048 bit RSA private key
.............................+++++
..........+++++
writing new private key to '/Users/lume/i01/A.key'
-----
rc=0
$ openssl pkcs12 -export -inkey '/Users/lume/i01/A.key' -in '/Users/lume/i01/A.crt' -out '/Users/lume/i01/A.p12' -passout pass:***

rc=0
$ security import '/Users/lume/i01/A.p12' -k '/Users/lume/i01/bund.keychain-db' -P *** -T /usr/bin/codesign
1 identity imported.
rc=0
$ openssl req -x509 -newkey rsa:2048 -nodes -days 2 -keyout '/Users/lume/i01/B.key' -out '/Users/lume/i01/B.crt' -subj '/CN=MSE4Mac I01 Herausgeber B' -addext 'keyUsage=critical,digitalSignature' -addext 'extendedKeyUsage=critical,codeSigning'
Generating a 2048 bit RSA private key
........................................................................+++++
.....................................+++++
writing new private key to '/Users/lume/i01/B.key'
-----
rc=0
$ openssl pkcs12 -export -inkey '/Users/lume/i01/B.key' -in '/Users/lume/i01/B.crt' -out '/Users/lume/i01/B.p12' -passout pass:***

rc=0
$ security import '/Users/lume/i01/B.p12' -k '/Users/lume/i01/bund.keychain-db' -P *** -T /usr/bin/codesign
1 identity imported.
rc=0
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/fremd.dylib' '/Users/lume/i01/fremd.dylib'

rc=0
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/leser-v1' '/Users/lume/i01/leser-a1'

rc=0
$ codesign --force --sign 'MSE4Mac I01 Herausgeber A' --keychain '/Users/lume/i01/bund.keychain-db'  -i de.mse4mac.i01.leser '/Users/lume/i01/leser-a1'
/Users/lume/i01/leser-a1: replacing existing signature
rc=0
$ codesign -dv '/Users/lume/i01/leser-a1'
Executable=/Users/lume/i01/leser-a1
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Signature size=1647
Signed Time=Oct 10, 2026 at 11:55:45â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ codesign -d -r- '/Users/lume/i01/leser-a1'
Executable=/Users/lume/i01/leser-a1
designated => identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"
rc=0
$ codesign -dvvv '/Users/lume/i01/leser-a1'
Executable=/Users/lume/i01/leser-a1
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Hash type=sha256 size=32
CandidateCDHash sha256=f59d488dfdbfb695f20ef6b7d8a29ee79d1091c7
CandidateCDHashFull sha256=f59d488dfdbfb695f20ef6b7d8a29ee79d1091c762c803014d2cade21dbb557b
Hash choices=sha256
CMSDigest=f59d488dfdbfb695f20ef6b7d8a29ee79d1091c762c803014d2cade21dbb557b
CMSDigestType=2
CDHash=f59d488dfdbfb695f20ef6b7d8a29ee79d1091c7
Signature size=1647
Authority=MSE4Mac I01 Herausgeber A
Signed Time=Oct 10, 2026 at 11:55:45â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ spctl --assess --type execute -vv '/Users/lume/i01/leser-a1'
/Users/lume/i01/leser-a1: rejected
origin=MSE4Mac I01 Herausgeber A
rc=3
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/leser-v2' '/Users/lume/i01/leser-a2'

rc=0
$ codesign --force --sign 'MSE4Mac I01 Herausgeber A' --keychain '/Users/lume/i01/bund.keychain-db'  -i de.mse4mac.i01.leser '/Users/lume/i01/leser-a2'
/Users/lume/i01/leser-a2: replacing existing signature
rc=0
$ codesign -dv '/Users/lume/i01/leser-a2'
Executable=/Users/lume/i01/leser-a2
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Signature size=1647
Signed Time=Oct 10, 2026 at 11:55:45â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ codesign -d -r- '/Users/lume/i01/leser-a2'
Executable=/Users/lume/i01/leser-a2
designated => identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"
rc=0
$ codesign -dvvv '/Users/lume/i01/leser-a2'
Executable=/Users/lume/i01/leser-a2
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Hash type=sha256 size=32
CandidateCDHash sha256=3d03754838911369255cbad7390b4cc62ad83a84
CandidateCDHashFull sha256=3d03754838911369255cbad7390b4cc62ad83a849f12da59e9f3aabfafee4e95
Hash choices=sha256
CMSDigest=3d03754838911369255cbad7390b4cc62ad83a849f12da59e9f3aabfafee4e95
CMSDigestType=2
CDHash=3d03754838911369255cbad7390b4cc62ad83a84
Signature size=1647
Authority=MSE4Mac I01 Herausgeber A
Signed Time=Oct 10, 2026 at 11:55:45â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ spctl --assess --type execute -vv '/Users/lume/i01/leser-a2'
/Users/lume/i01/leser-a2: rejected
origin=MSE4Mac I01 Herausgeber A
rc=3
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/leser-v1' '/Users/lume/i01/leser-a3'

rc=0
$ codesign --force --sign 'MSE4Mac I01 Herausgeber A' --keychain '/Users/lume/i01/bund.keychain-db' --options runtime -i de.mse4mac.i01.leser '/Users/lume/i01/leser-a3'
/Users/lume/i01/leser-a3: replacing existing signature
rc=0
$ codesign -dv '/Users/lume/i01/leser-a3'
Executable=/Users/lume/i01/leser-a3
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20500 size=629 flags=0x10000(runtime) hashes=14+2 location=embedded
Signature size=1647
Signed Time=Oct 10, 2026 at 11:55:46â¯PM
Info.plist=not bound
TeamIdentifier=not set
Runtime Version=27.0.0
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ codesign -d -r- '/Users/lume/i01/leser-a3'
Executable=/Users/lume/i01/leser-a3
designated => identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"
rc=0
$ codesign -dvvv '/Users/lume/i01/leser-a3'
Executable=/Users/lume/i01/leser-a3
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20500 size=629 flags=0x10000(runtime) hashes=14+2 location=embedded
Hash type=sha256 size=32
CandidateCDHash sha256=2067f7aec13af00ec21585e4383970800a826946
CandidateCDHashFull sha256=2067f7aec13af00ec21585e4383970800a8269466dfc3e1575cf0b356ded8c0e
Hash choices=sha256
CMSDigest=2067f7aec13af00ec21585e4383970800a8269466dfc3e1575cf0b356ded8c0e
CMSDigestType=2
CDHash=2067f7aec13af00ec21585e4383970800a826946
Signature size=1647
Authority=MSE4Mac I01 Herausgeber A
Signed Time=Oct 10, 2026 at 11:55:46â¯PM
Info.plist=not bound
TeamIdentifier=not set
Runtime Version=27.0.0
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ spctl --assess --type execute -vv '/Users/lume/i01/leser-a3'
/Users/lume/i01/leser-a3: rejected
origin=MSE4Mac I01 Herausgeber A
rc=3
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/leser-v1' '/Users/lume/i01/leser-adhoc'

rc=0
$ codesign --force --sign '-' --keychain '/Users/lume/i01/bund.keychain-db'  -i de.mse4mac.i01.leser '/Users/lume/i01/leser-adhoc'
/Users/lume/i01/leser-adhoc: replacing existing signature
rc=0
$ codesign -dv '/Users/lume/i01/leser-adhoc'
Executable=/Users/lume/i01/leser-adhoc
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x2(adhoc) hashes=14+2 location=embedded
Signature=adhoc
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=0 size=12
rc=0
$ codesign -d -r- '/Users/lume/i01/leser-adhoc'
Executable=/Users/lume/i01/leser-adhoc
# designated => cdhash H"cae43d9fe120079c1f799bc65ad26662f7759cb2"
rc=0
$ codesign -dvvv '/Users/lume/i01/leser-adhoc'
Executable=/Users/lume/i01/leser-adhoc
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x2(adhoc) hashes=14+2 location=embedded
Hash type=sha256 size=32
CandidateCDHash sha256=cae43d9fe120079c1f799bc65ad26662f7759cb2
CandidateCDHashFull sha256=cae43d9fe120079c1f799bc65ad26662f7759cb231708a257cc33ed5cf2a9831
Hash choices=sha256
CMSDigest=cae43d9fe120079c1f799bc65ad26662f7759cb231708a257cc33ed5cf2a9831
CMSDigestType=2
CDHash=cae43d9fe120079c1f799bc65ad26662f7759cb2
Signature=adhoc
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=0 size=12
rc=0
$ spctl --assess --type execute -vv '/Users/lume/i01/leser-adhoc'
/Users/lume/i01/leser-adhoc: rejected
rc=3
$ cp '/Volumes/My Shared Files/i01-gast-L6mFIV/i01/leser-v1' '/Users/lume/i01/leser-b1'

rc=0
$ codesign --force --sign 'MSE4Mac I01 Herausgeber B' --keychain '/Users/lume/i01/bund.keychain-db'  -i de.mse4mac.i01.leser '/Users/lume/i01/leser-b1'
/Users/lume/i01/leser-b1: replacing existing signature
rc=0
$ codesign -dv '/Users/lume/i01/leser-b1'
Executable=/Users/lume/i01/leser-b1
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Signature size=1647
Signed Time=Oct 10, 2026 at 11:55:47â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ codesign -d -r- '/Users/lume/i01/leser-b1'
Executable=/Users/lume/i01/leser-b1
designated => identifier "de.mse4mac.i01.leser" and certificate leaf = H"658e9f8689d433da0d677e7956b6422a4bc9ad73"
rc=0
$ codesign -dvvv '/Users/lume/i01/leser-b1'
Executable=/Users/lume/i01/leser-b1
Identifier=de.mse4mac.i01.leser
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=621 flags=0x0(none) hashes=14+2 location=embedded
Hash type=sha256 size=32
CandidateCDHash sha256=c2a0aac2b57898f93100c5f6a777bd7dbb2f6f56
CandidateCDHashFull sha256=c2a0aac2b57898f93100c5f6a777bd7dbb2f6f56c398a4352873fe77399ec961
Hash choices=sha256
CMSDigest=c2a0aac2b57898f93100c5f6a777bd7dbb2f6f56c398a4352873fe77399ec961
CMSDigestType=2
CDHash=c2a0aac2b57898f93100c5f6a777bd7dbb2f6f56
Signature size=1647
Authority=MSE4Mac I01 Herausgeber B
Signed Time=Oct 10, 2026 at 11:55:47â¯PM
Info.plist=not bound
TeamIdentifier=not set
Sealed Resources=none
Internal requirements count=1 size=96
rc=0
$ spctl --assess --type execute -vv '/Users/lume/i01/leser-b1'
/Users/lume/i01/leser-b1: rejected
origin=MSE4Mac I01 Herausgeber B
rc=3
$ security add-generic-password -a i01 -s dienst1 -w *** '/Users/lume/i01/bund.keychain-db'

rc=0
$ sh -c 'w=$(security find-generic-password -s dienst1 -w *** 2>/dev/null); echo "rc=$? laenge=${#w}"'
rc=0 laenge=11
rc=0
$ '/Users/lume/i01/leser-a1' dienst1
v1 STATUS -25308 User interaction is not allowed.
rc=0
$ security add-generic-password -a i01 -s dienst2 -w *** -T '/Users/lume/i01/leser-a1' '/Users/lume/i01/bund.keychain-db'

rc=0
$ security dump-keychain -a '/Users/lume/i01/bund.keychain-db'
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: 0x00000010 
attributes:
    0x00000000 <uint32>=0x00000010 
    0x00000001 <blob>=0x303930393445443100  "09094ED1\000"
    0x00000002 <blob>=<NULL>
    0x00000003 <uint32>=0x00000001 
    0x00000004 <uint32>=0x00000000 
    0x00000005 <uint32>=0x00000000 
    0x00000006 <blob>=0xC752E7AB48A9BF0DFA7A572B9B3193B9A9CDCFA6  "\307R\347\253H\251\277\015\372zW+\2331\223\271\251\315\317\246"
    0x00000007 <blob>=<NULL>
    0x00000008 <blob>=0x7B38373139316361322D306663392D313164342D383439612D3030303530326235323132327D00  "{87191ca2-0fc9-11d4-849a-000502b52122}\000"
    0x00000009 <uint32>=0x0000002A  "\000\000\000*"
    0x0000000A <uint32>=0x00000800 
    0x0000000B <uint32>=0x00000800 
    0x0000000C <blob>=0x0000000000000000 
    0x0000000D <blob>=0x0000000000000000 
    0x0000000E <uint32>=0x00000001 
    0x0000000F <uint32>=0x00000001 
    0x00000010 <uint32>=0x00000001 
    0x00000011 <uint32>=0x00000000 
    0x00000012 <uint32>=0x00000001 
    0x00000013 <uint32>=0x00000001 
    0x00000014 <uint32>=0x00000001 
    0x00000015 <uint32>=0x00000001 
    0x00000016 <uint32>=0x00000001 
    0x00000017 <uint32>=0x00000001 
    0x00000018 <uint32>=0x00000001 
    0x00000019 <uint32>=0x00000001 
    0x0000001A <uint32>=0x00000001 
access: 3 entries
    entry 0:
        authorizations (6): decrypt derive export_clear export_wrapped mac sign
        don't-require-password
        description: A.p12
        applications (1):
            0: /usr/bin/codesign (OK)
                requirement: identifier "com.apple.security.codesign" and anchor apple
    entry 1:
        authorizations (1): encrypt
        don't-require-password
        description: A.p12
        applications: <null>
    entry 2:
        authorizations (1): change_acl
        don't-require-password
        description: A.p12
        applications (0):
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: 0x00000010 
attributes:
    0x00000000 <uint32>=0x00000010 
    0x00000001 <blob>=0x344341353038354200  "4CA5085B\000"
    0x00000002 <blob>=<NULL>
    0x00000003 <uint32>=0x00000001 
    0x00000004 <uint32>=0x00000000 
    0x00000005 <uint32>=0x00000000 
    0x00000006 <blob>=0x4330E656537691682A8C03FC9D3DAD2588ECB999  "C0\346VSv\221h*\214\003\374\235=\255%\210\354\271\231"
    0x00000007 <blob>=<NULL>
    0x00000008 <blob>=0x7B38373139316361322D306663392D313164342D383439612D3030303530326235323132327D00  "{87191ca2-0fc9-11d4-849a-000502b52122}\000"
    0x00000009 <uint32>=0x0000002A  "\000\000\000*"
    0x0000000A <uint32>=0x00000800 
    0x0000000B <uint32>=0x00000800 
    0x0000000C <blob>=0x0000000000000000 
    0x0000000D <blob>=0x0000000000000000 
    0x0000000E <uint32>=0x00000001 
    0x0000000F <uint32>=0x00000001 
    0x00000010 <uint32>=0x00000001 
    0x00000011 <uint32>=0x00000000 
    0x00000012 <uint32>=0x00000001 
    0x00000013 <uint32>=0x00000001 
    0x00000014 <uint32>=0x00000001 
    0x00000015 <uint32>=0x00000001 
    0x00000016 <uint32>=0x00000001 
    0x00000017 <uint32>=0x00000001 
    0x00000018 <uint32>=0x00000001 
    0x00000019 <uint32>=0x00000001 
    0x0000001A <uint32>=0x00000001 
access: 3 entries
    entry 0:
        authorizations (6): decrypt derive export_clear export_wrapped mac sign
        don't-require-password
        description: B.p12
        applications (1):
            0: /usr/bin/codesign (OK)
                requirement: identifier "com.apple.security.codesign" and anchor apple
    entry 1:
        authorizations (1): encrypt
        don't-require-password
        description: B.p12
        applications: <null>
    entry 2:
        authorizations (1): change_acl
        don't-require-password
        description: B.p12
        applications (0):
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: "genp"
attributes:
    0x00000007 <blob>="dienst1"
    0x00000008 <blob>=<NULL>
    "acct"<blob>="i01"
    "cdat"<timedate>=0x32303236313031313036353534375A00  "20261011065547Z\000"
    "crtr"<uint32>=<NULL>
    "cusi"<sint32>=<NULL>
    "desc"<blob>=<NULL>
    "gena"<blob>=<NULL>
    "icmt"<blob>=<NULL>
    "invi"<sint32>=<NULL>
    "mdat"<timedate>=0x32303236313031313036353534375A00  "20261011065547Z\000"
    "nega"<sint32>=<NULL>
    "prot"<blob>=<NULL>
    "scrp"<sint32>=<NULL>
    "svce"<blob>="dienst1"
    "type"<uint32>=<NULL>
access: 3 entries
    entry 0:
        authorizations (1): encrypt
        don't-require-password
        description: dienst1
        applications: <null>
    entry 1:
        authorizations (6): decrypt derive export_clear export_wrapped mac sign
        don't-require-password
        description: dienst1
        applications (1):
            0: /usr/bin/security (OK)
                requirement: identifier "com.apple.security" and anchor apple
    entry 2:
        authorizations (1): change_acl
        don't-require-password
        description: dienst1
        applications (0):
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: "genp"
attributes:
    0x00000007 <blob>="dienst2"
    0x00000008 <blob>=<NULL>
    "acct"<blob>="i01"
    "cdat"<timedate>=0x32303236313031313036353534385A00  "20261011065548Z\000"
    "crtr"<uint32>=<NULL>
    "cusi"<sint32>=<NULL>
    "desc"<blob>=<NULL>
    "gena"<blob>=<NULL>
    "icmt"<blob>=<NULL>
    "invi"<sint32>=<NULL>
    "mdat"<timedate>=0x32303236313031313036353534385A00  "20261011065548Z\000"
    "nega"<sint32>=<NULL>
    "prot"<blob>=<NULL>
    "scrp"<sint32>=<NULL>
    "svce"<blob>="dienst2"
    "type"<uint32>=<NULL>
access: 3 entries
    entry 0:
        authorizations (1): encrypt
        don't-require-password
        description: dienst2
        applications: <null>
    entry 1:
        authorizations (6): decrypt derive export_clear export_wrapped mac sign
        don't-require-password
        description: dienst2
        applications (1):
            0: /Users/lume/i01/leser-a1 (OK)
                requirement: identifier "de.mse4mac.i01.leser" and certificate leaf = H"6d7dbacaf1fa6de6d7f5f246f2e0e6469e6a7b67"
    entry 2:
        authorizations (1): change_acl
        don't-require-password
        description: dienst2
        applications (0):
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: 0x80001000 
attributes:
    "alis"<blob>="MSE4Mac I01 Herausgeber A"
    "cenc"<uint32>=0x00000003 
    "ctyp"<uint32>=0x00000001 
    "hpky"<blob>=0xC752E7AB48A9BF0DFA7A572B9B3193B9A9CDCFA6  "\307R\347\253H\251\277\015\372zW+\2331\223\271\251\315\317\246"
    "issu"<blob>=0x30243122302006035504030C194D5345344D6163204930312048657261757367656265722041  "0$1"0 \006\003U\004\003\014\031MSE4Mac I01 Herausgeber A"
    "labl"<blob>="MSE4Mac I01 Herausgeber A"
    "skid"<blob>=<NULL>
    "snbr"<blob>=0x00FAC133E54B2BF8C9  "\000\372\3013\345K+\370\311"
    "subj"<blob>=0x30243122302006035504030C194D5345344D6163204930312048657261757367656265722041  "0$1"0 \006\003U\004\003\014\031MSE4Mac I01 Herausgeber A"
no access control for this item
keychain: "/Users/lume/i01/bund.keychain-db"
version: 256
class: 0x80001000 
attributes:
    "alis"<blob>="MSE4Mac I01 Herausgeber B"
    "cenc"<uint32>=0x00000003 
    "ctyp"<uint32>=0x00000001 
    "hpky"<blob>=0x4330E656537691682A8C03FC9D3DAD2588ECB999  "C0\346VSv\221h*\214\003\374\235=\255%\210\354\271\231"
    "issu"<blob>=0x30243122302006035504030C194D5345344D6163204930312048657261757367656265722042  "0$1"0 \006\003U\004\003\014\031MSE4Mac I01 Herausgeber B"
    "labl"<blob>="MSE4Mac I01 Herausgeber B"
    "skid"<blob>=<NULL>
    "snbr"<blob>=0x00ABB6716B18CA36D6  "\000\253\266qk\030\3126\326"
    "subj"<blob>=0x30243122302006035504030C194D5345344D6163204930312048657261757367656265722042  "0$1"0 \006\003U\004\003\014\031MSE4Mac I01 Herausgeber B"
no access control for this item
rc=0
$ sh -c 'w=$(security find-generic-password -s dienst2 -w *** 2>/dev/null); echo "rc=$? laenge=${#w}"'
rc=36 laenge=0
rc=0
$ '/Users/lume/i01/leser-a1' dienst2
v1 GELESEN 11 Bytes
rc=0
$ '/Users/lume/i01/leser-a2' dienst2
v2 GELESEN 11 Bytes
rc=0
$ '/Users/lume/i01/leser-b1' dienst2
v1 STATUS -25308 User interaction is not allowed.
rc=0
$ '/Users/lume/i01/leser-adhoc' dienst2
v1 STATUS -25308 User interaction is not allowed.
rc=0
$ MSE_I01_FREMD=dienst2 DYLD_INSERT_LIBRARIES='/Users/lume/i01/fremd.dylib' '/Users/lume/i01/leser-a1' keiner
FREMD GELESEN 11 Bytes
v1 STATUS -25300 The specified item could not be found in the keychain.
rc=0
$ '/Users/lume/i01/leser-a3' dienst2
v1 GELESEN 11 Bytes
rc=0
$ MSE_I01_FREMD=dienst2 DYLD_INSERT_LIBRARIES='/Users/lume/i01/fremd.dylib' '/Users/lume/i01/leser-a3' keiner
v1 STATUS -25300 The specified item could not be found in the keychain.
rc=0
$ security lock-keychain '/Users/lume/i01/bund.keychain-db'

rc=0
$ sh -c 'w=$(security find-generic-password -s dienst2 -w *** 2>/dev/null); echo "rc=$? laenge=${#w}"'
rc=36 laenge=0
rc=0
$ '/Users/lume/i01/leser-a1' dienst2
v1 STATUS -25308 User interaction is not allowed.
rc=0
$ security list-keychains -d user -s '/Users/lume/Library/Keychains/login.keychain-db'

rc=0
$ security delete-keychain '/Users/lume/i01/bund.keychain-db'

rc=0
```
