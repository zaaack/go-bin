import sharp from 'sharp'
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const favicon = readFileSync(join(__dirname, '..', '..', 'internal', 'web', 'assets', 'static', 'favicon.png'))
const androidRes = join(__dirname, '..', 'android', 'app', 'src', 'main', 'res')

// Full launcher icons
const densities = [
  { dir: 'mipmap-mdpi', size: 48 },
  { dir: 'mipmap-hdpi', size: 72 },
  { dir: 'mipmap-xhdpi', size: 96 },
  { dir: 'mipmap-xxhdpi', size: 144 },
  { dir: 'mipmap-xxxhdpi', size: 192 },
]

for (const { dir, size } of densities) {
  const out = join(androidRes, dir)
  if (!existsSync(out)) mkdirSync(out, { recursive: true })

  const buf = await sharp(favicon).resize(size, size).png().toBuffer()
  writeFileSync(join(out, 'ic_launcher.png'), buf)
  writeFileSync(join(out, 'ic_launcher_round.png'), buf)
  console.log(`wrote ${dir}/ic_launcher.png (${size}x${size})`)

  // Adaptive foreground: 108dp base, icon centered with safe zone padding
  const fgSize = Math.round(size * 108 / 48)
  const innerSize = Math.round(fgSize * 0.66)
  const inner = await sharp(favicon).resize(innerSize, innerSize).toBuffer()
  const fg = await sharp({
    create: { width: fgSize, height: fgSize, channels: 4, background: { r: 0, g: 0, b: 0, alpha: 0 } },
  }).composite([{ input: inner, gravity: 'center' }]).png().toBuffer()
  writeFileSync(join(out, 'ic_launcher_foreground.png'), fg)
  console.log(`wrote ${dir}/ic_launcher_foreground.png (${fgSize}x${fgSize})`)
}
