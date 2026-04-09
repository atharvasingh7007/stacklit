#!/usr/bin/env node

const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const ext = os.platform() === 'win32' ? '.exe' : '';
const binPath = path.join(__dirname, 'stacklit' + ext);

if (!fs.existsSync(binPath)) {
  console.error('stacklit binary not found. Try reinstalling:');
  console.error('  npm install -g stacklit');
  console.error('Or install directly:');
  console.error('  go install github.com/glincker/stacklit/cmd/stacklit@latest');
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
} catch (err) {
  process.exit(err.status || 1);
}
