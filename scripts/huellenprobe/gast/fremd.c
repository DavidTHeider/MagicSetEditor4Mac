// **Fremder Code im gebundenen Programm** — E5 der Bindungsprobe, I01
// (#2259),
// für die Entscheidung über #2295.
//
// Eine Bibliothek, die über `DYLD_INSERT_LIBRARIES` in einen Leser geladen
// wird und **vor dessen `main`** dasselbe fragt wie er: liest der Eintrag
// still, wenn fremder Code im Namen des gebundenen Programms fragt? Sie
// nennt wie der Leser nur Länge oder Status, nie den Wert, und tut nichts,
// solange `MSE_I01_FREMD` keinen Dienst nennt.
//
// API nachgelesen im SDK von Xcode 27 (`SecItem.h`, `CFDictionary.h`).
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdio.h>
#include <stdlib.h>

__attribute__((constructor)) static void fremd(void) {
  const char *dienst = getenv("MSE_I01_FREMD");
  if (dienst == NULL) return;
  CFStringRef name = CFStringCreateWithCString(NULL, dienst, kCFStringEncodingUTF8);
  const void *schluessel[] = {kSecClass, kSecAttrService, kSecReturnData, kSecMatchLimit};
  const void *werte[] = {kSecClassGenericPassword, name, kCFBooleanTrue, kSecMatchLimitOne};
  CFDictionaryRef anfrage = CFDictionaryCreate(NULL, schluessel, werte, 4,
      &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
  CFTypeRef ergebnis = NULL;
  OSStatus status = SecItemCopyMatching(anfrage, &ergebnis);
  if (status == errSecSuccess && ergebnis != NULL) {
    printf("FREMD GELESEN %ld Bytes\n", (long)CFDataGetLength((CFDataRef)ergebnis));
  } else {
    printf("FREMD STATUS %d\n", (int)status);
  }
  fflush(stdout);
  if (ergebnis != NULL) CFRelease(ergebnis);
  CFRelease(anfrage);
  CFRelease(name);
}
