const assert = require('node:assert/strict');

const {
  bumpVersion,
  commitMessage,
  isStaleHead,
  shouldSkipCiForBump,
} = require('./bump-suse-prerelease.cjs');

assert.equal(bumpVersion('2.10.3-pre.9'), '2.10.3-pre.10');
assert.equal(bumpVersion('2.10.3'), '2.10.4-pre.1');
assert.throws(() => bumpVersion('2.10-pre.3'), /Unsupported chart version/);

assert.equal(
  shouldSkipCiForBump({ publishedHead: 'ABCDEF', targetHead: 'abcdef' }),
  true,
);
assert.equal(
  shouldSkipCiForBump({ publishedHead: 'abcdef', targetHead: '123456' }),
  false,
);
assert.equal(
  shouldSkipCiForBump({ publishedHead: '', targetHead: '123456' }),
  true,
);

assert.equal(
  commitMessage('stable/suse-observability', '2.10.3-pre.10', true),
  "Updating 'stable/suse-observability' helm chart version to 2.10.3-pre.10 [skip ci]",
);
assert.equal(
  commitMessage('stable/suse-observability', '2.10.3-pre.10', false),
  "Updating 'stable/suse-observability' helm chart version to 2.10.3-pre.10",
);

assert.equal(
  isStaleHead({
    errors: [
      {
        type: 'STALE_DATA',
        message: 'Expected branch to point to "abc" but it did not.',
      },
    ],
  }),
  true,
);
assert.equal(
  isStaleHead({
    errors: [
      {
        type: 'VALIDATION',
        message: 'Something else failed',
      },
    ],
  }),
  false,
);

console.log('bump-suse-prerelease tests passed');
