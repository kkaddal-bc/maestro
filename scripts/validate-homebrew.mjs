import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const formulaPath = path.join(root, 'brew', 'maestro.rb');
const formula = fs.readFileSync(formulaPath, 'utf8');

const expected = [
  'class Maestro < Formula',
  'homepage "https://github.com/kkaddal-bc/maestro"',
  'version "0.1.0"',
  'on_macos do',
  'on_arm do',
  'url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-arm64.tar.gz"',
  'sha256 "938957e5ac72f194be3bbc79d864246d51fc77354f509588ec0467204151a166"',
  'on_intel do',
  'url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-amd64.tar.gz"',
  'sha256 "943e4001be2ea33ddafde94922569425e43caf933ffa98d74489926d765a01ea"',
  'bin.install "maestro"',
  'system "#{bin}/maestro", "--help"',
];

for (const needle of expected) {
  if (!formula.includes(needle)) {
    throw new Error(`brew/maestro.rb is missing ${JSON.stringify(needle)}`);
  }
}

if (process.argv.includes('--typecheck')) {
  process.exit(0);
}

if (!formula.endsWith('\n')) {
  throw new Error('brew/maestro.rb should end with a newline');
}

const armUrlCount = (formula.match(/maestro-darwin-arm64\.tar\.gz/g) || []).length;
const amdUrlCount = (formula.match(/maestro-darwin-amd64\.tar\.gz/g) || []).length;
if (armUrlCount !== 1 || amdUrlCount !== 1) {
  throw new Error('brew/maestro.rb should reference each platform artifact exactly once');
}

console.log('homebrew formula validation passed');
