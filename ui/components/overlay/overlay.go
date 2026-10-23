// Package overlay collects modal layers (pickers, dialogs) so the app
// root can lay them out above every page.
package overlay

import "github.com/mehran9675/giode"

var layers []giode.Element

// Add registers a layer; call it once, when the component is created.
// The layer lays out nothing while closed.
func Add(el giode.Element) { layers = append(layers, el) }

// All returns every registered layer in registration order.
func All() []giode.Element { return layers }
