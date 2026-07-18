// Package command contains write-side use-cases for the Invoice context.
//
// Rules:
//   - Handlers change state only through domain aggregates.
//   - Handlers are concrete types, injected directly into delivery (no reflection bus).
//   - Integration events are written to the transactional outbox in the same TX.
package command
