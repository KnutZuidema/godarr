// Package bencoding provides functionality for de- and encoding values from and to a bencoding representation as
// described here: http://bittorrent.org/beps/bep_0003.html.
//
// The usage of this package is similar to the json package.
//
// For customizing how the attributes of a struct are represented struct tags may be used just like for json
// representation.
//
// Example:
//   type Test struct {
//   	Attr0 []byte
//   	Attr1 int             `bencode:"int_attr"`
//   	Attr2 string          `bencode:"str_attr,omitempty"`
//   	Attr3 map[string]bool `bencode:"-"`
//   }
// The becoding representation of the zero value of this object would look like this:
// d5:Attr00:8:int_attri0ee
//
// If no name is given via the tag the attribute name is used.
//
// The omitempty option skips the element if the element is equal the zero value of that type.
//
// To use the omitempty option without assigning a custom name the leading comma still has to be specified, otherwise
// the bencoding name of the attribute is recognized as "omitempty".
//
// If the tag value is simply "-" the attribute is skipped regardless of its value.
//
// Embedded structs are squashed if no custom name is specified, which means its attributes will be used as if they were
// attributes of the original struct.
//
// The previously mentioned options also apply to embedded structs.
package bencoding
