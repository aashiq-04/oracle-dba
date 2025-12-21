package graph

import "github.com/aashiq-04/oracle-dba/internal/service"

type Resolver struct {
    authService   *service.AuthService
    rbacService   *service.RBACService
    oracleService *service.OracleService
}

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
