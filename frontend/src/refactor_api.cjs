const fs = require('fs');
const path = require('path');
const dir = '/Users/siddiqs_office/Documents/Personal Dev/GST Invoice Manager/frontend/src';

const files = fs.readdirSync(dir).filter(f => f.endsWith('.tsx') || f.endsWith('.ts'));

fs.writeFileSync(path.join(dir, 'api.ts'), "export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';\n");

files.forEach(file => {
  if (file === 'api.ts' || file === 'vite-env.d.ts') return;
  const filePath = path.join(dir, file);
  let content = fs.readFileSync(filePath, 'utf8');
  let changed = false;

  // if file already has const API_BASE, strip it
  if (content.includes("const API_BASE =")) {
    content = content.replace(/const API_BASE = .*?\n/, '');
    changed = true;
  }
  
  if (/'http:\/\/localhost:8080([^']*)'/.test(content)) {
    content = content.replace(/'http:\/\/localhost:8080([^']*)'/g, '`${API_BASE}$1`');
    changed = true;
  }
  
  if (/"http:\/\/localhost:8080([^"]*)"/.test(content)) {
    content = content.replace(/"http:\/\/localhost:8080([^"]*)"/g, '`${API_BASE}$1`');
    changed = true;
  }
  
  if (/`http:\/\/localhost:8080([^`]*)`/.test(content)) {
    content = content.replace(/`http:\/\/localhost:8080([^`]*)`/g, '`${API_BASE}$1`');
    changed = true;
  }

  if (content.includes('http://localhost:8080')) {
     content = content.replace(/http:\/\/localhost:8080/g, '${API_BASE}');
     changed = true;
  }

  if (changed) {
    if (!content.includes("import { API_BASE }")) {
      content = "import { API_BASE } from './api';\n" + content;
    }
    fs.writeFileSync(filePath, content);
    console.log(`Updated ${file}`);
  }
});
