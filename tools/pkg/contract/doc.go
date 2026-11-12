// Package contract holds the types, shared errors and validation rules of the
// Memorandum-tex template contract v1 (docs/contrato-plantillas.md).
//
// It never uses the network or runs processes. It only reads files when a
// Context asks for the checks that need the template folder (entry, preview,
// components, license texts), so Creative-Ink can import it to validate the
// manifests it downloads.
//
// Validation returns a Report of Findings instead of an error: an invalid
// manifest is a result, and every Finding carries the stable rule code of
// §12 (V01…V52), its level, the JSON field and the line in the source.
package contract
