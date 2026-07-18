// Package query contains read-side use-cases for the Invoice context.
//
// Rules:
//   - Handlers read DTO/read-models directly from PostgreSQL.
//   - No aggregate loading, no write-repository usage on the query path.
//   - Handlers are concrete types, injected directly into delivery (no reflection bus).
package query
