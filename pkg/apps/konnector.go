package apps

import (
	"encoding/json"
	"io"
	"time"

	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/permissions"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

// KonnManifest contains all the informations associated with an installed
// konnector.
type KonnManifest struct {
	DocRev string `json:"_rev,omitempty"`

	Name       string `json:"name"`
	NamePrefix string `json:"name_prefix,omitempty"`
	Editor     string `json:"editor"`
	Icon       string `json:"icon"`

	Type        string           `json:"type,omitempty"`
	License     string           `json:"license,omitempty"`
	Language    string           `json:"language,omitempty"`
	VendorLink  string           `json:"vendor_link"`
	Locales     *json.RawMessage `json:"locales,omitempty"`
	Langs       *json.RawMessage `json:"langs,omitempty"`
	Platforms   *json.RawMessage `json:"platforms,omitempty"`
	Categories  *json.RawMessage `json:"categories,omitempty"`
	Developer   *json.RawMessage `json:"developer,omitempty"`
	Screenshots *json.RawMessage `json:"screenshots,omitempty"`
	Tags        *json.RawMessage `json:"tags,omitempty"`

	Frequency    string           `json:"frequency"`
	DataTypes    *json.RawMessage `json:"data_types"`
	Doctypes     *json.RawMessage `json:"doctypes"`
	Fields       *json.RawMessage `json:"fields"`
	Messages     *json.RawMessage `json:"messages"`
	OAuth        *json.RawMessage `json:"oauth"`
	TimeInterval *json.RawMessage `json:"time_interval"`

	Parameters    *json.RawMessage `json:"parameters,omitempty"`
	Notifications Notifications    `json:"notifications"`

	// OnDeleteAccount can be used to specify a file path which will be executed
	// when an account associated with the konnector is deleted.
	OnDeleteAccount string `json:"on_delete_account,omitempty"`

	DocSlug        string          `json:"slug"`
	DocState       State           `json:"state"`
	DocSource      string          `json:"source"`
	DocVersion     string          `json:"version"`
	DocPermissions permissions.Set `json:"permissions"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Err string `json:"error,omitempty"`
	err error
}

// ID is part of the Manifest interface
func (m *KonnManifest) ID() string { return m.DocType() + "/" + m.DocSlug }

// Rev is part of the Manifest interface
func (m *KonnManifest) Rev() string { return m.DocRev }

// DocType is part of the Manifest interface
func (m *KonnManifest) DocType() string { return consts.Konnectors }

// Clone is part of the Manifest interface
func (m *KonnManifest) Clone() couchdb.Doc {
	cloned := *m

	cloned.DocPermissions = make(permissions.Set, len(m.DocPermissions))
	copy(cloned.DocPermissions, m.DocPermissions)

	cloned.Locales = cloneRawMessage(m.Locales)
	cloned.Langs = cloneRawMessage(m.Langs)
	cloned.Platforms = cloneRawMessage(m.Platforms)
	cloned.Categories = cloneRawMessage(m.Categories)
	cloned.Developer = cloneRawMessage(m.Developer)
	cloned.Screenshots = cloneRawMessage(m.Screenshots)
	cloned.Tags = cloneRawMessage(m.Tags)
	cloned.Parameters = cloneRawMessage(m.Parameters)

	cloned.DataTypes = cloneRawMessage(m.DataTypes)
	cloned.Doctypes = cloneRawMessage(m.Doctypes)
	cloned.Fields = cloneRawMessage(m.Fields)
	cloned.Messages = cloneRawMessage(m.Messages)
	cloned.OAuth = cloneRawMessage(m.OAuth)
	cloned.TimeInterval = cloneRawMessage(m.TimeInterval)
	return &cloned
}

// SetID is part of the Manifest interface
func (m *KonnManifest) SetID(id string) {}

// SetRev is part of the Manifest interface
func (m *KonnManifest) SetRev(rev string) { m.DocRev = rev }

// Source is part of the Manifest interface
func (m *KonnManifest) Source() string { return m.DocSource }

// Version is part of the Manifest interface
func (m *KonnManifest) Version() string { return m.DocVersion }

// Slug is part of the Manifest interface
func (m *KonnManifest) Slug() string { return m.DocSlug }

// State is part of the Manifest interface
func (m *KonnManifest) State() State { return m.DocState }

// LastUpdate is part of the Manifest interface
func (m *KonnManifest) LastUpdate() time.Time { return m.UpdatedAt }

// SetState is part of the Manifest interface
func (m *KonnManifest) SetState(state State) { m.DocState = state }

// SetVersion is part of the Manifest interface
func (m *KonnManifest) SetVersion(version string) { m.DocVersion = version }

// AppType is part of the Manifest interface
func (m *KonnManifest) AppType() AppType { return Konnector }

// Permissions is part of the Manifest interface
func (m *KonnManifest) Permissions() permissions.Set {
	return m.DocPermissions
}

// SetError is part of the Manifest interface
func (m *KonnManifest) SetError(err error) {
	m.SetState(Errored)
	m.Err = err.Error()
	m.err = err
}

// Error is part of the Manifest interface
func (m *KonnManifest) Error() error { return m.err }

// Match is part of the Manifest interface
func (m *KonnManifest) Match(field, value string) bool {
	switch field {
	case "slug":
		return m.DocSlug == value
	case "state":
		return m.DocState == State(value)
	}
	return false
}

// ReadManifest is part of the Manifest interface
func (m *KonnManifest) ReadManifest(r io.Reader, slug, sourceURL string) error {
	var newManifest KonnManifest
	if err := json.NewDecoder(r).Decode(&newManifest); err != nil {
		return ErrBadManifest
	}

	newManifest.SetID(m.ID())
	newManifest.SetRev(m.Rev())
	newManifest.SetState(m.State())
	newManifest.CreatedAt = m.CreatedAt
	newManifest.DocSlug = slug
	newManifest.DocSource = sourceURL
	if newManifest.Parameters == nil {
		newManifest.Parameters = m.Parameters
	}

	*m = newManifest
	return nil
}

// Create is part of the Manifest interface
func (m *KonnManifest) Create(db prefixer.Prefixer) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	if err := couchdb.CreateNamedDocWithDB(db, m); err != nil {
		return err
	}
	_, err := permissions.CreateKonnectorSet(db, m.Slug(), m.Permissions())
	return err
}

// Update is part of the Manifest interface
func (m *KonnManifest) Update(db prefixer.Prefixer) error {
	m.UpdatedAt = time.Now()
	err := couchdb.UpdateDoc(db, m)
	if err != nil {
		return err
	}
	_, err = permissions.UpdateKonnectorSet(db, m.Slug(), m.Permissions())
	return err
}

// Delete is part of the Manifest interface
func (m *KonnManifest) Delete(db prefixer.Prefixer) error {
	err := permissions.DestroyKonnector(db, m.Slug())
	if err != nil && !couchdb.IsNotFoundError(err) {
		return err
	}
	return couchdb.DeleteDoc(db, m)
}

// GetKonnectorBySlug fetch the manifest of a konnector from the database given
// a slug.
func GetKonnectorBySlug(db prefixer.Prefixer, slug string) (*KonnManifest, error) {
	if slug == "" || !slugReg.MatchString(slug) {
		return nil, ErrInvalidSlugName
	}
	man := &KonnManifest{}
	err := couchdb.GetDoc(db, consts.Konnectors, consts.Konnectors+"/"+slug, man)
	if couchdb.IsNotFoundError(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return man, nil
}

// ListKonnectors returns the list of installed konnectors applications.
//
// TODO: pagination
func ListKonnectors(db prefixer.Prefixer) ([]Manifest, error) {
	var docs []*KonnManifest
	req := &couchdb.AllDocsRequest{Limit: 100}
	err := couchdb.GetAllDocs(db, consts.Konnectors, req, &docs)
	if err != nil {
		return nil, err
	}
	mans := make([]Manifest, len(docs))
	for i, m := range docs {
		mans[i] = m
	}
	return mans, nil
}

var _ Manifest = &KonnManifest{}
