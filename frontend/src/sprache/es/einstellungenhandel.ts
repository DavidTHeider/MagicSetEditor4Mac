// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'Permitir comerciar',
  'richtlinie.handel.markt':
    'Permitir comerciar a través de un proveedor externo',
  'richtlinie.handel.menge': 'Cantidad máxima por línea',
  'richtlinie.handel.wert':
    'Valor máximo de una oferta en la unidad más pequeña (en EUR: céntimos)',
  'richtlinie.handel.frist': 'Plazo máximo de reserva en minutos',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    'Desactiva a la vez compartir, préstamo de mazos, intercambios y el mercado. **Un límite, no una preferencia**: lo que se desactiva aquí no se puede volver a activar en ningún perfil.',
  'richtlinie.handel.markt':
    'Por defecto está **desactivado**. El mercado recurre a un proveedor externo y a dinero; todo lo demás queda entre dos personas.',
  'richtlinie.handel.menge':
    'Cuántos ejemplares como máximo puede indicar una sola línea de una oferta. **Vacío significa: sin límite.**',
  'richtlinie.handel.wert':
    'Cuánto puede valer como máximo una oferta en total. Solo se aplica si la valoración **forma parte del contrato**; una simple vista del mercado no la limita. **Vacío significa: sin límite.**',
  'richtlinie.handel.frist':
    'Cuánto tiempo puede una promesa mantener reservados ejemplares. **Vacío significa: sin límite.**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'Desactivado',
  'richtlinie.handel.an': 'Activado',
  'richtlinie.handel.markt.aus': 'Desactivado',
  'richtlinie.handel.markt.an': 'Activado',
}
