function svgToDataUri(svg) {
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
}

const icons = {
  default: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="18" y="10" width="60" height="76" rx="6" fill="#26a69a"/>
      <rect x="26" y="18" width="44" height="24" fill="#80cbc4"/>
      <rect x="26" y="48" width="44" height="30" fill="#4db6ac"/>
      <circle cx="66" cy="48" r="3" fill="#ffd54f"/>
    </svg>`
  ),
  entryDoor: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="26" y="6" width="44" height="84" rx="5" fill="#6d4c41"/>
      <rect x="30" y="12" width="36" height="72" fill="#8d6e63"/>
      <rect x="34" y="18" width="28" height="20" fill="#a1887f"/>
      <rect x="34" y="44" width="28" height="34" fill="#795548"/>
      <circle cx="59" cy="49" r="3" fill="#ffca28"/>
      <rect x="22" y="6" width="4" height="84" fill="#cfd8dc"/>
    </svg>`
  ),
  interiorDoor: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="30" y="8" width="36" height="80" rx="5" fill="#607d8b"/>
      <rect x="34" y="14" width="28" height="68" fill="#78909c"/>
      <rect x="36" y="18" width="24" height="18" fill="#b0bec5"/>
      <rect x="36" y="42" width="24" height="36" fill="#90a4ae"/>
      <circle cx="57" cy="49" r="2.8" fill="#ffd54f"/>
      <rect x="24" y="8" width="4" height="80" fill="#dce6eb"/>
    </svg>`
  ),
  handle: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="20" y="30" width="56" height="10" rx="5" fill="#b0bec5"/>
      <rect x="58" y="30" width="16" height="30" rx="4" fill="#78909c"/>
      <rect x="18" y="52" width="60" height="8" rx="4" fill="#90a4ae"/>
    </svg>`
  ),
  lock: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="24" y="38" width="48" height="36" rx="6" fill="#546e7a"/>
      <path d="M34 38v-8a14 14 0 0 1 28 0v8" fill="none" stroke="#90a4ae" stroke-width="6"/>
      <circle cx="48" cy="54" r="4" fill="#ffd54f"/>
      <rect x="46" y="58" width="4" height="10" fill="#ffd54f"/>
    </svg>`
  ),
  foam: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="36" y="18" width="24" height="52" rx="6" fill="#ffca28"/>
      <rect x="42" y="8" width="12" height="12" rx="2" fill="#90a4ae"/>
      <circle cx="30" cy="54" r="8" fill="#ffe082"/>
      <circle cx="66" cy="50" r="10" fill="#fff3c4"/>
      <circle cx="52" cy="40" r="6" fill="#fff8e1"/>
    </svg>`
  ),
  hinge: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="32" y="18" width="12" height="60" rx="3" fill="#78909c"/>
      <rect x="52" y="18" width="12" height="60" rx="3" fill="#607d8b"/>
      <circle cx="48" cy="30" r="3" fill="#cfd8dc"/>
      <circle cx="48" cy="48" r="3" fill="#cfd8dc"/>
      <circle cx="48" cy="66" r="3" fill="#cfd8dc"/>
    </svg>`
  ),
  trim: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="22" y="22" width="52" height="10" fill="#8d6e63"/>
      <rect x="22" y="42" width="52" height="10" fill="#a1887f"/>
      <rect x="22" y="62" width="52" height="10" fill="#bcaaa4"/>
    </svg>`
  ),
  wreath: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <circle cx="48" cy="48" r="28" fill="none" stroke="#2e7d32" stroke-width="8"/>
      <circle cx="48" cy="48" r="18" fill="none" stroke="#43a047" stroke-width="6"/>
      <circle cx="30" cy="42" r="5" fill="#c62828"/>
      <circle cx="66" cy="44" r="5" fill="#c62828"/>
      <circle cx="48" cy="68" r="5" fill="#c62828"/>
      <rect x="44" y="12" width="8" height="14" rx="2" fill="#8d6e63"/>
    </svg>`
  ),
  nameplate: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="18" y="36" width="60" height="28" rx="6" fill="#b0bec5"/>
      <rect x="22" y="40" width="52" height="20" rx="4" fill="#eceff1"/>
      <text x="48" y="54" text-anchor="middle" font-size="10" fill="#546e7a" font-family="sans-serif">HOME</text>
    </svg>`
  ),
  knocker: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="42" y="16" width="12" height="24" rx="3" fill="#8d6e63"/>
      <circle cx="48" cy="58" r="18" fill="#ffb300"/>
      <circle cx="48" cy="58" r="12" fill="#ffc107"/>
    </svg>`
  ),
  sticker: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="24" y="20" width="48" height="56" rx="4" fill="#f48fb1"/>
      <circle cx="36" cy="40" r="6" fill="#f06292"/>
      <circle cx="56" cy="36" r="5" fill="#ec407a"/>
      <path d="M30 58 Q48 72 66 54" fill="none" stroke="#ad1457" stroke-width="3"/>
    </svg>`
  ),
  numberPlate: svgToDataUri(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect x="20" y="38" width="56" height="24" rx="4" fill="#37474f"/>
      <text x="48" y="54" text-anchor="middle" font-size="14" fill="#eceff1" font-family="sans-serif">12</text>
    </svg>`
  )
}

export function getProductIcon(product) {
  const name = String(product?.product_name || '').toLowerCase()
  const catId = Number(product?.category_id || 0)

  if (name.includes('ручк')) return icons.handle
  if (name.includes('замок')) return icons.lock
  if (name.includes('пен')) return icons.foam
  if (name.includes('петл')) return icons.hinge
  if (name.includes('налич')) return icons.trim
  if (name.includes('венок')) return icons.wreath
  if (name.includes('таблич')) return icons.nameplate
  if (name.includes('молоток')) return icons.knocker
  if (name.includes('номер') || name.includes('квартир')) return icons.numberPlate
  if (name.includes('наклей')) return icons.sticker

  if (catId === 1) return icons.entryDoor
  if (catId === 2) return icons.interiorDoor
  if (catId === 3) return icons.handle
  if (catId === 4) return icons.trim
  if (catId === 5) return icons.wreath

  return icons.default
}
