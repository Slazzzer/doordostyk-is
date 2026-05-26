let robotoBase64 = null

async function loadRobotoBase64() {
  if (robotoBase64) return robotoBase64
  const res = await fetch('/fonts/Roboto-Regular.ttf')
  if (!res.ok) throw new Error('Не удалось загрузить шрифт Roboto для PDF')
  const buf = await res.arrayBuffer()
  const bytes = new Uint8Array(buf)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  robotoBase64 = btoa(binary)
  return robotoBase64
}

/** Регистрирует Roboto в jsPDF для кириллицы. */
export async function applyRobotoFont(doc) {
  const b64 = await loadRobotoBase64()
  if (!doc.getFontList().Roboto) {
    doc.addFileToVFS('Roboto-Regular.ttf', b64)
    doc.addFont('Roboto-Regular.ttf', 'Roboto', 'normal')
  }
  doc.setFont('Roboto', 'normal')
}
