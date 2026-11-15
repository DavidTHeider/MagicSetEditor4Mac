export const fehler = {
  unbekannt:
    'Se ha producido un error inesperado. Los detalles están en el registro: en «Ajustes › Diagnóstico» puedes ver su final y guardarlo.',
  unbekanntMitVorgang:
    'Se ha producido un error inesperado. Los detalles están en el registro: en «Ajustes › Diagnóstico» puedes ver su final y guardarlo. Busca allí {vorgang}: todas las líneas de esta operación llevan ese identificador.',
  nichtGefunden:
    'No se ha encontrado la entrada buscada. Quizá la haya eliminado otra persona; vuelve a cargar la lista.',
  ohnePeeridentitaet:
    'Este perfil aún no tiene identidad. Créala en el recuadro «Con qué se identifica este perfil»; sin ella, nadie puede identificarse.',
  peerNichtErreichbar:
    'Ahora mismo no se puede contactar con tu interlocutor. ¿Está escuchando y es correcta su dirección? Si no, vuelve a intentarlo más tarde.',
  peerOhneAdresse:
    'No hay ninguna dirección registrada para este interlocutor. Introdúcela en su contacto, en «Dirección y nota»; sin ella, este programa solo lo encuentra en la misma red.',
  ohneNetzerlaubnis:
    'Nada salió a la red local. Si macOS acaba de pedir permiso, concédelo y vuelve a intentarlo. Desde aquí no se puede saber el motivo con seguridad: o bien falta el permiso — Ajustes del Sistema → Privacidad y seguridad → Red local —, o bien la red no deja pasar el multicast, por ejemplo detrás de una VPN.',
  ohneNetzroute:
    'Ahora mismo no hay ninguna ruta hacia la red local. ¿Hay una interfaz de red conectada? En Ajustes → Conexiones puedes elegir una en concreto.',
  schluesselbundAblage:
    'La clave de este perfil no se pudo guardar en el llavero. ¿Está desbloqueado el llavero de inicio de sesión? Inténtalo de nuevo después.',
  einladungZuNeu:
    'Esta invitación procede de una versión más reciente de este programa. Pide a la otra parte una invitación de su versión anterior, o consigue tú la versión más reciente.',
  fernhorchenLaeuft:
    'Este perfil ya es accesible a través de otras redes. Detenlo primero si quieres una invitación nueva: la anterior dejará de ser válida.',
  anmeldungAbgelaufen:
    'El inicio de sesión no volvió en diez minutos. Vuelve a empezarlo cuando estés listo en el navegador.',
  umschlagFassung:
    'Este sobre procede de otra versión de este programa. Pide a la otra parte uno de su versión anterior, o consigue tú la versión más reciente.',
  umschlagUnterschrift:
    'La firma de este sobre no es válida. Se ha modificado por el camino o procede de alguien distinto de quien dice ser: no lo aceptes y consúltalo con la otra parte.',
  umschlagInhalt:
    'Una parte de este sobre no coincide con su manifiesto. Pide a la otra parte que lo envíe de nuevo: por el camino algo se ha dañado o modificado.',
  fremderNutzer:
    'Esto solo puede hacerlo ese mismo usuario. Cambia primero a ese usuario; la aplicación te pedirá el segundo factor.',
  nichtFreigegeben:
    'Tu contraparte no comparte nada contigo para intercambiar. Pide que te comparta una vista, o usa un archivo para intercambiar.',
  zeileNichtVerbuchbar:
    'La línea {zeile} no podría registrarse en el lado que la recibe: falta el estado o el idioma, o algún dato no existe. Quítala y vuelve a añadirla con estado, idioma y acabado.',
  zusageUnpassend:
    'Este ejemplar no cumple ninguna de las líneas que das: el estado, el idioma o el acabado no coinciden. Elige uno que coincida o añádelo a tu colección.',
  erlaubnis:
    'No tienes el rol necesario para esto. Quien administra este perfil puede dártelo.',
  unbekannteHerkunftsart:
    'El programa no conoce este origen. Elige «tomado de un perfil» o «importado».',
  // The command centre (#1423).
  befehlBestaetigen:
    'Este comando cambia algo que no se puede deshacer fácilmente: lee la vista previa y confirma.',
  befehlArgumente:
    'Los argumentos no son un objeto JSON. Escribe un objeto entre llaves, con un campo por parámetro.',
  herkunftOhneProfil:
    'Un origen de un perfil lo nombra; una importación no nombra ninguno: indica uno o déjalo en blanco.',
  nurLesend:
    'En este perfil solo puedes leer. Pertenece a otra persona y tu rol en él es «lector». Cambia a tu propio perfil para trabajar, o pide a quien lo administra que te dé el rol «colaborador».',
  bestandUnbrauchbar:
    'Este archivo está dañado o no es una biblioteca de MagicSetEditor4Mac, y no se puede abrir. Restaura una copia de seguridad; la última está donde la guardaste.',
  bestandNeuereFassung:
    'Esta biblioteca procede de una versión más reciente de MagicSetEditor4Mac. El archivo está bien: no lo sustituyas. Consigue la versión más reciente o restaura una copia de seguridad anterior.',
  bestandAeltereFassung:
    'Esta biblioteca procede de una versión anterior a 0.7.0, y esta versión ya no puede actualizarla. El archivo está bien: no lo sustituyas. Ábrela una vez con MagicSetEditor4Mac 0.7.0; después, esta versión también la abrirá.',
  bestandNurLesbar:
    'La biblioteca se puede leer pero no escribir: al archivo o a su carpeta les falta el permiso de escritura. El archivo está bien, no lo sustituyas. Dale permiso de escritura en el Finder, en «Obtener información», y vuelve a intentarlo.',
  bestandEA:
    'El disco informó de un error al leer la biblioteca. No se sabe si el archivo está intacto, así que no lo sustituyas todavía. Revisa el disco, por ejemplo con la Utilidad de Discos, y vuelve a intentarlo.',
  kaumPlatz:
    'La biblioteca no se pudo abrir ni escribir, y apenas queda espacio libre en el disco. Probablemente el archivo está bien: no lo sustituyas. Libera espacio y vuelve a intentarlo; si no basta, revisa el disco, por ejemplo con la Utilidad de Discos, y el permiso de escritura.',
  startUngeklaert:
    'No se pudo abrir la biblioteca y no está claro por qué; el motivo aparece debajo. Vuelve a intentarlo. Restaura una copia de seguridad solo cuando estés seguro de que este archivo ya no se puede abrir.',
  sicherungVerworfen:
    'La copia de seguridad no superó su comprobación y se ha descartado: no se ha guardado nada. Vuelve a intentarlo o guárdala en otro lugar.',
  sicherungUnvermerkt:
    'La copia de seguridad está escrita y comprobada; solo no se pudo registrar su hora, así que la antigüedad de la última copia sigue mostrando la anterior. Conserva el archivo y vuelve a guardar cuando la biblioteca esté libre.',
  schluesselFehlt:
    'Esta biblioteca está cifrada y su clave no está en el llavero de este Mac. El archivo está bien: no lo sustituyas. La clave se creó al activar el cifrado y no puede volver a calcularse; sin ella, los datos son inaccesibles.',
  belegt:
    'La biblioteca está ocupada en este momento; probablemente otra ventana o un comando esté escribiendo. Inténtalo de nuevo en un momento.',
  bestandWirdGetauscht:
    'La biblioteca se está sustituyendo en este momento: al restaurar una copia de seguridad, cambiar el cifrado o empezar de cero. Espera un momento y vuelve a intentarlo.',
  anwendungEndet:
    'MagicSetEditor4Mac se está cerrando; este paso ya no se ha realizado. Vuelve a abrir la aplicación para seguir trabajando.',
  vorgangAbgebrochen:
    'La operación se canceló antes de terminar. Vuelve a iniciarla si todavía la necesitas.',
  konflikt: 'La entrada se ha modificado entretanto. Vuelve a cargar.',
  konfliktgrund:
    'No ha funcionado: {grund}. Comprueba el estado actual antes de volver a intentarlo.',
  doppelt:
    'Ya existe una entrada con estos datos. Cambia alguno de ellos o edita la entrada existente.',
  doppelterName: 'Ya existe este nombre aquí; elige otro.',
  doppeltesExemplar:
    'La colección ya contiene esta impresión, con el mismo estado, idioma y acabado. Aumenta allí la cantidad o cambia alguno de los datos.',
  gesperrt:
    'Esta aplicación no puede acceder a esta ubicación: está bloqueada. Elige otra o concede el acceso, en los permisos del archivo o en Ajustes del Sistema › Privacidad y seguridad › Archivos y carpetas, y vuelve a intentarlo.',
  voll: 'No queda espacio en el disco. Libera espacio y vuelve a intentarlo.',
  netzName:
    'No se pudo resolver el nombre «{wirt}». Comprueba la conexión a internet y, si hay uno configurado, el proxy en «Ajustes › Conexiones». Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  netzNichtErreichbar:
    'En «{wirt}» nadie acepta conexiones. ¿Está en marcha el servicio y es correcta su dirección? Si no, vuelve a intentarlo más tarde. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  netzKeineAntwort:
    '«{wirt}» no respondió en {sekunden} segundos. Vuelve a intentarlo más tarde. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  netzAntwortStockt:
    'La respuesta de «{wirt}» se detuvo durante {sekunden} segundos. Comprueba la conexión y vuelve a intentarlo. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  netzAbgerissen:
    'La conexión con «{wirt}» se cortó antes de que llegara la respuesta completa. Vuelve a intentarlo; si sigue ocurriendo, el problema está en el servicio o en un proxy intermedio. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  netzProxyAnmeldung:
    'El proxy «{wirt}» exige iniciar sesión. Indica sus credenciales mediante HTTPS_PROXY en el entorno: en «Ajustes › Conexiones» el proxy no admite contraseña a propósito. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  gedrosselt:
    'La otra parte informa de demasiadas solicitudes. Espera un momento y vuelve a intentarlo. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  gedrosseltBis:
    '«{wirt}» no acepta solicitudes en este momento. Vuelve a intentarlo a partir de las {uhrzeit}; hasta entonces este programa no volverá a preguntar. Lo que informó la conexión figura en «Ajustes › Diagnóstico» bajo {vorgang}.',
  kontingent:
    'Demasiadas llamadas en poco tiempo. Espera {sekunden} s y vuelve a intentarlo.',
  prozessVeraltet:
    'Este proceso procede de una versión anterior a la de la biblioteca y se está cerrando: no se ha escrito nada. Un servicio instalado se reinicia solo, con el programa para el que está instalado; reinicia cualquier otro proceso con la nueva versión y vuelve a intentarlo.',
  bestandVomDienstGehalten:
    'El servicio para scripts mantiene abierta la biblioteca. Revócalo en Ajustes › Diagnóstico › «Configurar la línea de comandos y el servidor de herramientas» y vuelve a restaurar.',
  falschesSpiel:
    'El mazo pertenece a un juego distinto del de este torneo. Elige un mazo del mismo juego.',
  turnierformatSpiel:
    'El formato pertenece a un juego distinto del de este torneo. Elige un formato del mismo juego, o ninguno.',
  paketGeaendert:
    'Este paquete difiere de cómo se instaló. Confirma la sustitución: el cambio se conserva como versión anterior.',
  symbolZuVieleTeile:
    'Este símbolo tiene {anzahl} partes y se permiten {grenze}. Une partes o elige un dibujo más sencillo.',
  symbolZuVielePunkte:
    'Este símbolo tiene {anzahl} puntos y se permiten {grenze}. Simplifica las formas o elige un dibujo más sencillo.',
  symbolZuTief:
    'Los grupos de este símbolo están anidados en {anzahl} niveles y se permiten {grenze}. Elige un dibujo más plano o desagrupa los grupos interiores.',
  symbolZuGross:
    'Este archivo es mayor que {grenze} y, por tanto, no es un archivo de símbolo. Elige un archivo .mse-symbol.',
  symbolGespeichertZuGross:
    'Al guardarlo, este símbolo superaría {grenze}, más de lo que cabe en un archivo de símbolo. Simplifica las formas o usa menos puntos.',
  bildZuGross:
    'Esta imagen es demasiado grande para descomprimirla. Elige una más pequeña o redúcela antes.',
  setOhneSymbol:
    'Este set aún no tiene símbolo. Dibuja uno y guárdalo; entonces se podrá añadir.',
  // **Die eigenen Übersetzungen** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '«{kennung}» no es un identificador de idioma según BCP 47. Usa uno como pt-BR o zh-Hant.',
  sprachkennungEingebaut:
    '«{kennung}» es un idioma integrado. Elige un identificador propio como {kennung}-x-meine.',
  platzhalterAbweichend:
    'Esta traducción usa otros marcadores de posición que la frase alemana. Escribe exactamente estos: {erwartet}.',
  auszeichnungUnsicher:
    'Esta traducción contiene algo que no es texto, o un formato que no se cierra. Elimina las etiquetas y cierra ** y `.',
  sprachpaketNeuer:
    'Esta traducción procede de un programa más reciente ({format}). Obtén la versión nueva.',
  sprachpaketOhneLizenz:
    'Sin licencia no se puede compartir ninguna traducción. Escribe una, por ejemplo CC0-1.0.',
  bereichAbgeschaltet:
    'Esta área está desactivada ({schalter}). Vuelve a activarla en Ajustes › Áreas y funciones: no se ha perdido nada.',
  dienstAus:
    'El servicio {dienst} está desactivado. Actívalo en Ajustes › Servicios si quieres consultarlo.',
  fremdeEigeneKarten:
    'Las cartas propias del remitente no llegaron — su identificador solo vale en su biblioteca: {karten}. Si las necesitas, abre su set y añádelas aquí.',
  nichtGelegtePlaetze:
    'Estas entradas no se pudieron añadir aquí: {plaetze}. El resto del mazo se aceptó; revísalas y añádelas a mano.',
  fremdeKartenUndNichtGelegte:
    'Las cartas propias del remitente no llegaron — su identificador solo vale en su biblioteca: {karten}. Y estas entradas no se pudieron añadir aquí: {plaetze}. Revisa ambas cosas y añádelas a mano.',
  symbolVerwendet:
    'Estos sets llevan el símbolo: {sets}. Confirma la eliminación: conservan su copia.',
  symbolOhneVorige:
    'Este símbolo no tiene versión anterior. Guarda primero uno nuevo; entonces se podrá recuperar el actual.',
  turnierAbgeschlossen:
    'Este torneo está cerrado. Elige «Reabrir» si quieres cambiar algo.',
  sitzGegangen:
    'Este asiento ha abandonado la partida y ya no puede actuar. Elige un asiento que siga en juego.',
  fremderWert:
    'Otra persona ha cambiado este campo entretanto. Revisa el nuevo valor y decide si conservas tu borrador.',
  vorschlagNurGegenseite:
    'Solo la otra parte puede cambiar una oferta abierta. Espera su respuesta o cancela el intercambio y vuelve a ofrecer.',
  turnierNichtFertig:
    'Introduce primero todos los resultados: un torneo se puede cerrar cuando tiene al menos una ronda y todos los emparejamientos tienen resultado.',
  turniernameLeer: 'Un torneo necesita un nombre; escribe uno.',
  turnierspielLeer:
    'Un torneo pertenece a un juego; de lo contrario no hay formato con el que comprobar. Elige uno.',
  validierungsgrund:
    'No ha funcionado: {grund}. Corrige el dato y vuelve a intentarlo.',

  validierung:
    'La entrada no es válida. La indicación del campo explica el motivo; corrígela allí.',

  // **The rule itself is the remedy** (#1350).
  // **The password of a backup** (OPS-06, #1723).
  sicherungskennwortKurz:
    'La contraseña de una copia de seguridad tiene al menos doce caracteres. Elige una más larga: una frase de cuatro palabras es buena y fácil de recordar.',
  sicherungskennwortFehlt:
    'Esta copia de seguridad está cifrada. Introduce su contraseña para verla o restaurarla.',
  sicherungskennwortFalsch:
    'La contraseña no corresponde a esta copia de seguridad. Compruébala y vuelve a introducirla.',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} no es un JSON legible: en el byte {stelle} el archivo termina o está mal formado. Elige otro archivo.',
  mtgjsonKeinSet:
    '{datei} es JSON, pero no contiene ningún set en formato MTG JSON (difiere en el byte {stelle}). Elige un archivo con un set, o una carpeta entera.',
  sicherungImKlartext:
    'La colección está cifrada; sin contraseña, la copia de seguridad quedaría en texto claro. Define una contraseña en Ajustes › Datos y copia de seguridad o elige allí expresamente «copia de seguridad sin contraseña».',

  planVerschluesselt:
    'La colección está cifrada y el plan hace copias sin contraseña, así que el plan no la copia. Haz una copia manual con contraseña.',

  paketnameUngueltig:
    'El nombre de un paquete se compone de letras, dígitos, guiones y guiones bajos, y se convierte en el nombre del directorio. Elige uno con esos caracteres.',
  ordnerangabe:
    '«{wert}» no es una ubicación completa. Escribe una que empiece por / o por ~/, sin .. en medio.',
  ordnerFehlt:
    'No se puede acceder a la carpeta «{ordner}». Conecta la unidad o elige otra carpeta.',
  ordnerIstDatei:
    '«{ordner}» es un archivo, no una carpeta. Elige la carpeta en la que está, u otra.',
  ordnerGesperrt:
    'Esta aplicación no puede escribir en «{ordner}». Elige otra carpeta o concede acceso a esta.',
  ordnerVoll:
    'En la unidad de «{ordner}» solo quedan {frei} libres. Libera espacio o elige otra carpeta.',
  karteInDeck:
    'Esta carta se usa en {anzahl} mazo(s): {decks}. Al eliminarla también se quita de ellos; confirma si es lo que quieres.',
  pflichtfeld: 'Este campo no puede estar vacío; escribe algo.',
  nichtEditierbar:
    'Este campo no se puede editar: la plantilla solo lo muestra. Cambia el valor del que se deriva.',
  berechnetesFeld:
    'Este campo se deriva de otros y no se puede establecer en muchas cartas a la vez. Sus partes aparecen por separado; elige una de ellas.',
  keinSetMitStil:
    'Ningún set abierto usa esta hoja de estilo. Abre un set que la use o cambia la hoja de estilo del set abierto.',
  vorlageFehlt:
    'No hay ninguna plantilla instalada para esta carta. Comprueba en la página Paquetes el juego y la hoja de estilo del set; sin ellos, una carta no tiene campos ni áreas.',
  sicherungNeuereFassung:
    'Esta copia de seguridad procede de una versión más reciente de MagicSetEditor4Mac. Contiene {unbekannt} y esta versión llega hasta {stand}. Consigue la versión más reciente o elige una copia de seguridad anterior.',
  sicherungAeltereFassung:
    'Esta copia de seguridad procede de una versión anterior a 0.7.0: le faltan pasos del esquema hasta {grundlinie}, y esta versión ya no puede actualizarla. Restáurala con MagicSetEditor4Mac 0.7.0; después, esta versión abrirá la biblioteca.',
  sicherungUnbrauchbar:
    'Este archivo no es una copia de seguridad de MagicSetEditor4Mac. Elige una que haya escrito esta aplicación.',
  // The parser's own sentence is German for now — see issue #1178.
  skriptFehler:
    'No se puede evaluar la expresión: {meldung} Corrígela en el lugar indicado.',

  // **The missing second factor** — piece 7 of phase 4.12 (#1351).
  // **The local session** — ID-02 (#1701).
  kennwortFalsch:
    'Esta contraseña no es correcta. Introdúcela de nuevo o define una nueva con la hoja de recuperación.',
  kennwortLaenge:
    'Una contraseña tiene al menos {mindestens} y como máximo {hoechstens} caracteres. Elige una de esa longitud.',
  sperreOhneRueckweg:
    'Sin contraseña y sin desbloqueo local, el bloqueo no se podría levantar. Define primero una contraseña.',
  lokaleEntsperrungFehlt:
    'El desbloqueo local no está configurado para este usuario en este Mac. Introduce la contraseña en su lugar.',
  codeFalsch:
    'Este código no es correcto, o ya se ha usado. Comprueba el código de la hoja: cada uno vale una sola vez.',
  zweiterFaktor:
    'Este usuario exige un segundo factor. Introduce el código de tu aplicación de autenticación o confirma con Touch ID.',
  faktorGesperrt:
    'Demasiados códigos incorrectos. Espera un momento y vuelve a intentarlo.',
  sicherungFremdeFassung:
    'Esta copia de seguridad no corresponde a esta versión de MagicSetEditor4Mac. Indica {dort} y esta aplicación es {hier}. Restáurala con la versión que la escribió.',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: 'el carácter {0} no forma parte del lenguaje',
    einzelnesGleich: 'un único `=`; se asigna con `:=` y se compara con `==`',
    offeneZeichenkette: 'una cadena sin comillas de cierre',
    zuTiefGeschachtelt: 'la expresión está anidada a más de {0} niveles',
    keinSkript: 'no existe ningún script llamado {0}',
    skriptkreis: 'estos campos se leen entre sí en círculo: {0}',
    keinVerweis: '`+` no puede concatenar una referencia a otro script',
    keineZahl: '{0} no es un número',
    formUnbekannt: 'la forma {0} no se puede evaluar',
    nameUnbekannt: 'el nombre {0} no se conoce aquí',
    keineFunktion: 'aquí se llama a {0}, y eso solo funciona con una función',
    ohneRumpf: 'esta función no tiene cuerpo',
    nichtWahrheitswert: '`not` necesita un valor de verdad, aquí hay {0}',
    vorzeichenZahl: 'un signo necesita un número, aquí hay {0}',
    zaehlschleifeZahl:
      'el {0} de un bucle de conteo necesita un número, aquí hay {1}',
    rechnenLinks: '`{0}` calcula con números, a la izquierda hay {1}',
    rechnenRechts: '`{0}` calcula con números, a la derecha hay {1}',
    vergleichLinks: '`{0}` compara números, a la izquierda hay {1}',
    vergleichRechts: '`{0}` compara números, a la derecha hay {1}',
    keineListe: 'sort_list ordena una lista, aquí hay {0}',
    zuTief:
      'la evaluación tiene más de {0} niveles de profundidad: ¿se llama algo a sí mismo aquí?',
    schleifeSteht:
      'un bucle de conteo no avanza en {0}: el número es demasiado grande',
    funktionVerbinden:
      'una función solo se puede combinar con otra función, aquí hay {0} y {1}',
    fehltNach: 'después de {0} falta {1}',
    fehltHier: 'aquí falta {0}',
    klammerOffen: 'este {0} no se cierra: falta {1}',
    ausdruckGehtWeiter: 'aquí termina la expresión, pero el texto continúa',
    zuweisungBrauchtNamen: 'se asigna a un nombre',
    durchgereicht: '{0}',
    bedingungWahrheit:
      'una condición necesita un valor de verdad, aquí hay {0}',
    fuerJedenListe: '`for each` recorre una lista, aquí hay {0}',
    wahrheitLinks: '`{0}` necesita valores de verdad, a la izquierda hay {1}',
    wahrheitRechts: '`{0}` necesita valores de verdad, a la derecha hay {1}',
    bindenFunktion: '`@` liga argumentos a una función, aquí hay {0}',
  },
} as const
