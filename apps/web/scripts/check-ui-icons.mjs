import { readdir, readFile } from "node:fs/promises";
import { join, relative } from "node:path";

// Catch legacy decorative glyphs so future agents stay on one icon family.
// Normal prose punctuation, dates, © copyright notices and typography remain allowed.
const root = join(process.cwd(), "app");
const pictogram = /[📚📝🤝📖✦♙▦◉↗←›]|\p{Extended_Pictographic}/gu;
let violations = 0;
let filesChecked = 0;
async function check(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) { await check(path); continue; }
    if (!entry.name.endsWith(".tsx")) continue;
    filesChecked++;
    const lines = (await readFile(path, "utf8")).split("\n");
    lines.forEach((line, n) => {
      const found = [...line.replaceAll("©", "").matchAll(pictogram)].map(m => m[0]);
      if (/<svg(?:\\s|>)/.test(line)) found.push("raw SVG");
      if (found.length) {
        console.error(`${relative(process.cwd(),path)}:${n+1}: replace legacy glyph(s) ${found.join(" ")} with Lucide icons`);
        violations++;
      }
    });
  }
}
await check(root);
await check(join(process.cwd(), "components"));
if (violations) {
  console.error(`OpenEdu icon audit failed: ${violations} offending lines.`);
  process.exitCode = 1;
} else {
  console.log(`OpenEdu icon audit passed: ${filesChecked} TSX files, no legacy glyphs or raw inline SVG icons.`);
}