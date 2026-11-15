// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: 'Una cuenta propia',
  eigenesErklaerung:
    'Una dirección o un número de teléfono es aquí un identificador: dice quién inicia sesión. El programa no comprueba si el buzón o el número existen realmente: no envía nada y no necesita ningún servicio externo para ello.',

  kennung: 'Dirección o número de teléfono',
  kennungHinweis: 'Por ejemplo nombre@example.org o +4915112345678',

  kennwort: 'Contraseña',
  kennwortHinweis: 'Al menos doce caracteres',
  kennwortAnmelden: 'Contraseña',

  anzeigename: 'Nombre visible (opcional)',
  anzeigenameHinweis: 'Si se deja vacío, se forma a partir del identificador',

  eintragen: 'Crear cuenta',
  eingetragen: 'Creada: {name}',

  anmelden: 'Iniciar sesión',
  angemeldet: 'Sesión iniciada como {name}',

  ohneAngaben: 'Sin datos',
  ohneAngabenErklaerung:
    'Una cuenta anónima recibe un identificador aleatorio como nombre, sin contraseña y sin dirección. Y quien no quiera ninguna, no necesita ninguna: el usuario del primer inicio está ahí y la aplicación funciona.',
  anonym: 'Crear cuenta anónima',
  anonymAngelegt: 'Creada: {name}',

  kennwortNieZurueck:
    'La contraseña se guarda protegida (con hash) y no vuelve a mostrarse nunca. Quien la olvida define una nueva en lugar de leer la antigua.',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: 'Segundo factor',
  erklaerung:
    'Se exige un segundo factor al iniciar sesión y al cambiar de perfil. El código de una aplicación de autenticación sirve en todas partes; Touch ID solo funciona en este Mac.',

  totp: 'Código de una aplicación',
  biometrie: 'Biometría',

  einrichten: 'Configurar',
  entfernen: 'Quitar',

  totpEinrichten: 'Configurar un código de una aplicación',
  totpEntfernen: 'Quitar el código de una aplicación',
  biometrieEinrichten: 'Configurar Touch ID o Face ID',
  biometrieEntfernen: 'Quitar Touch ID o Face ID',

  totpEntfernenFrage: '¿Quitar el código de una aplicación?',
  biometrieEntfernenFrage: '¿Quitar Touch ID o Face ID?',
  entfernenErklaerung:
    'El inicio de sesión y el cambio de perfil ya no lo piden. Quien lo quiera de nuevo lo configura otra vez, con una aplicación y un secreto nuevo.',
  entfernenJa: 'Quitar',

  einmal: 'Esto se muestra una sola vez',
  einmalErklaerung:
    'Introduce este secreto en tu aplicación de autenticación y guarda los códigos de recuperación. Después no se pueden volver a consultar; quien los pierde, configura de nuevo.',
  codesUeberschrift: '{zahl} códigos de recuperación, cada uno válido una vez:',
  bestaetigen:
    'Introduce un código de tu aplicación. Solo entonces entra en vigor el segundo factor.',
} as const
