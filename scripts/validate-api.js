#!/usr/bin/env node

/**
 * OpenAPI Specification Validator
 * 
 * This script validates the OpenAPI specification file and checks for common issues.
 * 
 * Usage: node scripts/validate-api.js
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

// Colors for console output
const colors = {
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  reset: '\x1b[0m'
};

function log(color, message) {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function validateOpenAPISpec() {
  const specPath = path.join(__dirname, '..', 'api', 'openapi.yaml');
  
  try {
    // Check if file exists
    if (!fs.existsSync(specPath)) {
      log('red', '❌ OpenAPI specification file not found at: ' + specPath);
      process.exit(1);
    }

    // Read and parse YAML
    const specContent = fs.readFileSync(specPath, 'utf8');
    const spec = yaml.load(specContent);

    log('blue', '🔍 Validating OpenAPI specification...\n');

    // Basic structure validation
    const requiredFields = ['openapi', 'info', 'paths'];
    const missingFields = requiredFields.filter(field => !spec[field]);
    
    if (missingFields.length > 0) {
      log('red', `❌ Missing required fields: ${missingFields.join(', ')}`);
      process.exit(1);
    }

    // Validate OpenAPI version
    if (!spec.openapi.startsWith('3.0')) {
      log('yellow', `⚠️  OpenAPI version ${spec.openapi} detected. This validator is optimized for 3.0.x`);
    }

    // Validate info section
    if (!spec.info.title || !spec.info.version) {
      log('red', '❌ Info section must include title and version');
      process.exit(1);
    }

    // Count endpoints
    const pathCount = Object.keys(spec.paths).length;
    let operationCount = 0;
    
    Object.values(spec.paths).forEach(pathItem => {
      const methods = ['get', 'post', 'put', 'delete', 'patch', 'options', 'head'];
      methods.forEach(method => {
        if (pathItem[method]) {
          operationCount++;
        }
      });
    });

    // Validate components
    const componentCounts = {
      schemas: spec.components?.schemas ? Object.keys(spec.components.schemas).length : 0,
      responses: spec.components?.responses ? Object.keys(spec.components.responses).length : 0,
      parameters: spec.components?.parameters ? Object.keys(spec.components.parameters).length : 0,
      securitySchemes: spec.components?.securitySchemes ? Object.keys(spec.components.securitySchemes).length : 0
    };

    // Check for common issues
    const issues = [];
    
    // Check for missing descriptions
    Object.entries(spec.paths).forEach(([path, pathItem]) => {
      Object.entries(pathItem).forEach(([method, operation]) => {
        if (typeof operation === 'object' && operation.operationId === undefined) {
          issues.push(`Missing operationId for ${method.toUpperCase()} ${path}`);
        }
        if (typeof operation === 'object' && !operation.description && !operation.summary) {
          issues.push(`Missing description/summary for ${method.toUpperCase()} ${path}`);
        }
      });
    });

    // Check for unused schemas
    const usedSchemas = new Set();
    const allSchemas = spec.components?.schemas ? Object.keys(spec.components.schemas) : [];
    
    function findSchemaReferences(obj, visited = new Set()) {
      if (!obj || typeof obj !== 'object' || visited.has(obj)) return;
      visited.add(obj);
      
      if (obj.$ref && typeof obj.$ref === 'string') {
        const schemaName = obj.$ref.replace('#/components/schemas/', '');
        if (allSchemas.includes(schemaName)) {
          usedSchemas.add(schemaName);
        }
      }
      
      Object.values(obj).forEach(value => {
        if (typeof value === 'object') {
          findSchemaReferences(value, visited);
        }
      });
    }
    
    findSchemaReferences(spec);
    
    const unusedSchemas = allSchemas.filter(schema => !usedSchemas.has(schema));
    if (unusedSchemas.length > 0) {
      issues.push(`Potentially unused schemas: ${unusedSchemas.join(', ')}`);
    }

    // Display results
    log('green', '✅ OpenAPI specification is valid!\n');
    
    log('blue', '📊 Specification Statistics:');
    console.log(`   • OpenAPI Version: ${spec.openapi}`);
    console.log(`   • API Title: ${spec.info.title}`);
    console.log(`   • API Version: ${spec.info.version}`);
    console.log(`   • Paths: ${pathCount}`);
    console.log(`   • Operations: ${operationCount}`);
    console.log(`   • Schemas: ${componentCounts.schemas}`);
    console.log(`   • Responses: ${componentCounts.responses}`);
    console.log(`   • Security Schemes: ${componentCounts.securitySchemes}`);
    
    if (spec.servers && spec.servers.length > 0) {
      console.log(`   • Servers: ${spec.servers.length}`);
      spec.servers.forEach((server, index) => {
        console.log(`     ${index + 1}. ${server.url} (${server.description || 'No description'})`);
      });
    }

    if (spec.tags && spec.tags.length > 0) {
      console.log(`   • Tags: ${spec.tags.length}`);
      spec.tags.forEach(tag => {
        console.log(`     • ${tag.name}: ${tag.description || 'No description'}`);
      });
    }

    // Display issues
    if (issues.length > 0) {
      log('yellow', '\n⚠️  Issues found:');
      issues.forEach(issue => {
        console.log(`   • ${issue}`);
      });
    } else {
      log('green', '\n🎉 No issues found!');
    }

    // Endpoint summary
    log('blue', '\n📋 API Endpoints:');
    Object.entries(spec.paths).forEach(([path, pathItem]) => {
      const methods = [];
      ['get', 'post', 'put', 'delete', 'patch'].forEach(method => {
        if (pathItem[method]) {
          methods.push(method.toUpperCase());
        }
      });
      if (methods.length > 0) {
        console.log(`   ${methods.join(', ')} ${path}`);
      }
    });

    log('green', '\n✅ Validation completed successfully!');

  } catch (error) {
    if (error.name === 'YAMLException') {
      log('red', `❌ YAML parsing error: ${error.message}`);
    } else {
      log('red', `❌ Validation error: ${error.message}`);
    }
    process.exit(1);
  }
}

// Check if js-yaml is available
try {
  require.resolve('js-yaml');
} catch (e) {
  log('red', '❌ js-yaml package is required. Install it with: npm install js-yaml');
  process.exit(1);
}

// Run validation
validateOpenAPISpec();