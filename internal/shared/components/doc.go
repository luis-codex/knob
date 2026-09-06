// Package components agrupa las piezas de interfaz con estado propio. No
// conocen el dominio: reciben datos ya formateados.
//
// No son modelos de Bubble Tea, sino structs con métodos imperativos y View().
// El único tea.Model es internal/app, lo que mantiene el enrutado en un sitio.
package components
