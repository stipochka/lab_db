package fyne

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Resource represents a single binary resource, such as an image or font.
// A resource has an identifying name and byte array content.
// The serialised path of a resource can be obtained which may result in a
// blocking filesystem write operation.
type Resource interface {
	Name() string
	Content() []byte
}

// ThemedResource is a version of a resource that can be updated to match a certain theme color.
// The [ThemeColorName] will be used to look up the color for the current theme and colorize the resource.
//
// Since: 2.5
type ThemedResource interface {
	Resource
	ThemeColorName() ThemeColorName
}

// StaticResource is a bundled resource compiled into the application.
// These resources are normally generated by the fyne_bundle command included in
// the Fyne toolkit.
type StaticResource struct {
	StaticName    string
	StaticContent []byte
}

// Name returns the unique name of this resource, usually matching the file it
// was generated from.
func (r *StaticResource) Name() string {
	return r.StaticName
}

// Content returns the bytes of the bundled resource, no compression is applied
// but any compression on the resource is retained.
func (r *StaticResource) Content() []byte {
	return r.StaticContent
}

// NewStaticResource returns a new static resource object with the specified
// name and content. Creating a new static resource in memory results in
// sharable binary data that may be serialised to the system cache location.
func NewStaticResource(name string, content []byte) *StaticResource {
	return &StaticResource{
		StaticName:    name,
		StaticContent: content,
	}
}

// LoadResourceFromPath creates a new [StaticResource] in memory using the contents of the specified file.
func LoadResourceFromPath(path string) (Resource, error) {
	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	return NewStaticResource(name, bytes), nil
}

// LoadResourceFromURLString creates a new [StaticResource] in memory using the body of the specified URL.
func LoadResourceFromURLString(urlStr string) (Resource, error) {
	res, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(urlStr)
	return NewStaticResource(name, bytes), nil
}