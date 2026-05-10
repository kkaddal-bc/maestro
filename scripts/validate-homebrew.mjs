import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const repository = 'kkaddal-bc/maestro';
const version = '0.1.0';
const formulaPath = path.join(root, 'brew', 'maestro.rb');
const formula = fs.readFileSync(formulaPath, 'utf8');

const expectedSnippets = [
  'class Maestro < Formula',
  `homepage "https://github.com/${repository}"`,
  `version "${version}"`,
  'on_macos do',
  'on_arm do',
  `url "https://github.com/${repository}/releases/download/v${version}/maestro-darwin-arm64.tar.gz"`,
  'sha256 "938957e5ac72f194be3bbc79d864246d51fc77354f509588ec0467204151a166"',
  'on_intel do',
  `url "https://github.com/${repository}/releases/download/v${version}/maestro-darwin-amd64.tar.gz"`,
  'sha256 "943e4001be2ea33ddafde94922569425e43caf933ffa98d74489926d765a01ea"',
  'bin.install "maestro"',
  'system "#{bin}/maestro", "--help"',
];

function countOccurrences(text, needle) {
  return text.split(needle).length - 1;
}

for (const needle of expectedSnippets) {
  if (!formula.includes(needle)) {
    throw new Error(`brew/maestro.rb is missing ${JSON.stringify(needle)}`);
  }
}

const isTypecheck = process.argv.includes('--typecheck');
if (isTypecheck) {
  process.exit(0);
}

if (!formula.endsWith('\n')) {
  throw new Error('brew/maestro.rb should end with a newline');
}

const armUrlCount = countOccurrences(formula, 'maestro-darwin-arm64.tar.gz');
const amdUrlCount = countOccurrences(formula, 'maestro-darwin-amd64.tar.gz');
if (armUrlCount !== 1 || amdUrlCount !== 1) {
  throw new Error('brew/maestro.rb should reference each platform artifact exactly once');
}

console.log('homebrew formula validation passed');
