// Package components groups the interface pieces with their own state. They do
// not know the domain: they receive data already formatted.
//
// They are not Bubble Tea models, but structs with imperative methods and
// View(). The only tea.Model is internal/app, which keeps routing in one
// place.
package components
