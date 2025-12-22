Write-Host "=== Final GraphQL Fix ===" -ForegroundColor Cyan

# Step 1: Delete auto-generated schema.resolvers.go
Write-Host "`n[1/5] Removing auto-generated schema.resolvers.go..."
if (Test-Path "internal\graph\schema.resolvers.go") {
    Remove-Item "internal\graph\schema.resolvers.go" -Force
    Write-Host "      Deleted!" -ForegroundColor Green
} else {
    Write-Host "      Already removed." -ForegroundColor Yellow
}

# Step 2: Create base resolvers file
Write-Host "`n[2/5] Creating resolvers_base.go..."
@"
package graph

// This file contains the base resolver types that connect to generated code

// Query returns QueryResolver implementation
func (r *Resolver) Query() QueryResolver { 
    return &queryResolver{r} 
}

// Mutation returns MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver { 
    return &mutationResolver{r} 
}

// Subscription returns SubscriptionResolver implementation
func (r *Resolver) Subscription() SubscriptionResolver { 
    return &subscriptionResolver{r} 
}

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
"@ | Out-File -FilePath "internal\graph\resolvers_base.go" -Encoding UTF8
Write-Host "      Created!" -ForegroundColor Green

# Step 3: Update gqlgen.yml to prevent auto-generation
Write-Host "`n[3/5] Updating gqlgen.yml..."
@"
schema:
  - internal/graph/schema/*.graphql

exec:
  filename: internal/graph/generated.go
  package: graph

model:
  filename: internal/graph/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: internal/graph
  package: graph
  filename_template: "{name}.resolvers_skip.go"
  omit_template_comment: true

models:
  Time:
    model: time.Time

skip_mod_tidy: true
"@ | Out-File -FilePath "gqlgen.yml" -Encoding UTF8
Write-Host "      Updated!" -ForegroundColor Green

# Step 4: Clean and regenerate
Write-Host "`n[4/5] Cleaning old generated files..."
Remove-Item "internal\graph\generated.go" -ErrorAction SilentlyContinue
Remove-Item "internal\graph\model\models_gen.go" -ErrorAction SilentlyContinue
Write-Host "      Cleaned!" -ForegroundColor Green

Write-Host "`n[5/5] Regenerating GraphQL code..."
go run github.com/99designs/gqlgen generate

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== SUCCESS! ===" -ForegroundColor Green
    Write-Host "Now trying to build..."
    go build ./internal/graph/...
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✓ Everything compiled successfully!" -ForegroundColor Green
    } else {
        Write-Host "`n✗ Build failed. Check errors above." -ForegroundColor Red
    }
} else {
    Write-Host "`n✗ Generation failed. Check errors above." -ForegroundColor Red
}