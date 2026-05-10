const rootDirectories = ['skills', 'packages', 'brew', '.github/workflows'];

const requiredDirectories = [...rootDirectories, 'skills/maestro-snap'];

const bundledSkillFiles = ['skills/maestro-snap/SKILL.md', 'skills/maestro-snap/SCAN-PATTERNS.md', 'skills/maestro-snap/OUTPUT-TEMPLATES.md'];

function hasRequiredSkillFrontmatter(skillContents) {
  return /^---\n/.test(skillContents) && /^name:\s*maestro-snap$/m.test(skillContents) && /^description:\s*.+$/m.test(skillContents);
}

module.exports = {
  bundledSkillFiles,
  hasRequiredSkillFrontmatter,
  requiredDirectories,
  rootDirectories,
};
