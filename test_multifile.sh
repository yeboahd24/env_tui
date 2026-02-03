#!/bin/bash

# Create test env files
echo "Creating test .env files..."

# .env
cat > .env << 'EOF'
DATABASE_URL=postgres://localhost:5432/myapp
API_KEY=prod-api-key-12345
DEBUG=false
EOF

# .env.local
cat > .env.local << 'EOF'
DATABASE_URL=postgres://localhost:5432/myapp_dev
API_KEY=dev-api-key-67890
DEBUG=true
API_URL=http://localhost:3000
EOF

# .env.prod
cat > .env.prod << 'EOF'
DATABASE_URL=postgres://prod-db:5432/myapp_prod
API_KEY=prod-secret-key-abcde
DEBUG=false
API_URL=https://api.myapp.com
REDIS_URL=redis://prod-cache:6379
EOF

echo "✓ Created .env, .env.local, and .env.prod"
echo ""
echo "To test multi-file support, run:"
echo "./envtui -files '.env,.env.local,.env.prod'"
echo ""
echo "Features:"
echo "  - You'll see tabs at the top: [1:.env] [2:.env.local] [3:.env.prod]"
echo "  - Press 1, 2, or 3 to switch between files"
echo "  - The active file is highlighted in purple"
echo "  - Each file shows its own entries"
echo "  - You can add/edit/delete entries per file"
