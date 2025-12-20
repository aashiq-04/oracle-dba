package graph

import (
	"github.com/aashiq-04/oracle-dba/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver that holds all services
type Resolver struct {
	authService   *service.AuthService
	rbacService   *service.RBACService
	oracleService *service.OracleService
}

// NewResolver creates a new resolver with injected services
func NewResolver(
	authService *service.AuthService,
	rbacService *service.RBACService,
	oracleService *service.OracleService,
) *Resolver {
	return &Resolver{
		authService:   authService,
		rbacService:   rbacService,
		oracleService: oracleService,
	}
}