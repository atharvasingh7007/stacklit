#!/usr/bin/env node
const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');
const ext = os.platform() === 'win32' ? '.exe' : '';
const bin = path.join(__dirname, 'stacklit' + ext);
try {
  execFileSync(bin, process.argv.slice(2), { stdio: 'inherit' });
} catch (err) {
  if (err.status !== null) process.exit(err.status);
  console.error('stacklit binary not found. Run: npm install stacklit');
  process.exit(1);
}
