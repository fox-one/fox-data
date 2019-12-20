package session

import (
	"bytes"
	"context"
	"io"

	gSession "github.com/fox-one/gin-contrib/session"
	"github.com/spf13/viper"
)

// Session session
type Session struct {
	*gSession.Session

	// shared configuration
	v *viper.Viper
}

// New new session with data
func New(data []byte) (*Session, error) {
	return NewWithReader(bytes.NewReader(data))
}

// NewWithReader new session with reader
func NewWithReader(r io.Reader) (*Session, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}

	return NewWithViper(v)
}

// NewWithViper new session with viper
func NewWithViper(v *viper.Viper) (*Session, error) {
	s := &Session{
		Session: gSession.NewWithViper(v),
		v:       v,
	}
	return s, nil
}

// Copy copy
func (s *Session) Copy() *Session {
	return &Session{
		Session: s.Session.Copy(),
		v:       s.v,
	}
}

// MysqlBegin mysql begin
func (s *Session) MysqlBegin() *Session {
	s = s.Copy()
	s.Session = s.Session.MysqlBegin()
	return s
}

// MysqlReadOnWrite mysql read on write
func (s *Session) MysqlReadOnWrite() *Session {
	s = s.Copy()
	s.Session = s.Session.MysqlReadOnWrite()
	return s
}

// WithContext with context
func (s *Session) WithContext(ctx context.Context) *Session {
	if ctx == nil {
		panic("nil context")
	}

	cp := s.Copy()
	cp.Session = cp.Session.WithContext(ctx)
	return cp
}

// CMCKey return a cmc api key
func (s *Session) CMCKey() string {
	return s.v.GetString("cmc.key")
}
