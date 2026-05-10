const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');

function mustExist(relativePath) {
  const target = path.join(root, relativePath);
  if (!fs.existsSync(target)) {
    throw new Error(`missing required path: ${relativePath}`);
  }
}

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8');
}

JSON.parse(read('package.json'));

for (const dir of ['skills', 'skills/maestro-snap', 'packages', 'brew', '.github/workflows']) {
  mustExist(dir);
}

const skill = read('skills/maestro-snap/SKILL.md');
if (!/^---\n/.test(skill) || !/^name:\s*maestro-snap$/m.test(skill) || !/^description:\s*.+$/m.test(skill)) {
  throw new Error('skills/maestro-snap/SKILL.md is missing required frontmatter');
}
