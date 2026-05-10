const assert = require('assert');
const fs = require('fs');
const path = require('path');
const test = require('node:test');
const { bundledSkillFiles, hasRequiredSkillFrontmatter, rootDirectories } = require('../scripts/scaffold-structure.cjs');

const root = path.resolve(__dirname, '..');

function read(file) {
  return fs.readFileSync(path.join(root, file), 'utf8');
}

function exists(file) {
  return fs.existsSync(path.join(root, file));
}

test('monorepo scaffold includes the expected top-level directories', () => {
  for (const dir of rootDirectories) {
    assert.ok(exists(dir), `expected ${dir} to exist`);
  }
});

test('maestro-snap skill is bundled in the repo', () => {
  for (const file of bundledSkillFiles) {
    assert.ok(exists(file), `expected ${file} to exist`);
  }

  const skill = read('skills/maestro-snap/SKILL.md');
  assert.ok(hasRequiredSkillFrontmatter(skill), 'expected skill frontmatter to be valid');
});

test('README documents the scaffold layout', () => {
  const readme = read('README.md');
  for (const entry of ['skills/', 'packages/', 'brew/', '.github/workflows/']) {
    assert.ok(readme.includes(entry), `expected README to mention ${entry}`);
  }
});
