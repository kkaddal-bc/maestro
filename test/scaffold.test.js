const assert = require('assert');
const fs = require('fs');
const path = require('path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');

function read(file) {
  return fs.readFileSync(path.join(root, file), 'utf8');
}

function exists(file) {
  return fs.existsSync(path.join(root, file));
}

test('monorepo scaffold includes the expected top-level directories', () => {
  for (const dir of ['skills', 'packages', 'brew', '.github/workflows']) {
    assert.ok(exists(dir), `expected ${dir} to exist`);
  }
});

test('maestro-snap skill is bundled in the repo', () => {
  assert.ok(exists('skills/maestro-snap/SKILL.md'));
  assert.ok(exists('skills/maestro-snap/SCAN-PATTERNS.md'));
  assert.ok(exists('skills/maestro-snap/OUTPUT-TEMPLATES.md'));

  const skill = read('skills/maestro-snap/SKILL.md');
  assert.match(skill, /^---\n/m);
  assert.match(skill, /^name:\s*maestro-snap$/m);
  assert.match(skill, /^description:\s*.+$/m);
});

test('README documents the scaffold layout', () => {
  const readme = read('README.md');
  for (const entry of ['skills/', 'packages/', 'brew/', '.github/workflows/']) {
    assert.ok(readme.includes(entry), `expected README to mention ${entry}`);
  }
});
