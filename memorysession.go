package httpwaymid

import (
	"sync"
	"net/http"

	"github.com/corneldamian/httpway"
	"crypto/rand"
	"encoding/hex"
	"crypto/md5"
	"time"
	"encoding/binary"
)

//session
type Session struct {
	id string
	data map[string]interface{}
	dataSync *sync.RWMutex
}

func newSession() *Session {
	id := ""

	b := make([]byte, 32)
	n, err := rand.Read(b)

	if n != len(b) || err != nil {
		t := make([]byte, 32)
		binary.LittleEndian.PutUint64(t, uint64(time.Now().UnixNano()))
		hasher := md5.New()
		hasher.Write(t)
		id = hex.EncodeToString(hasher.Sum(nil))
	} else {
		id = hex.EncodeToString(b)
	}

	return &Session{
		id: id,
		data: make(map[string]interface{}),
		dataSync: &sync.RWMutex{},
	}
}

func (s *Session) Id() string {
	return s.id
}

func (s *Session) IsAuth() bool {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok :=s.data["_isAuth"]
	if !ok {
		return false
	}
	return v.(bool)
}

func (s *Session) SetAuth(isAuth bool) {
	s.dataSync.Lock()
	defer s.dataSync.Unlock()

	s.data["_isAuth"] = isAuth
}

func (s *Session) Username() string {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok :=s.data["_username"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (s *Session) SetUsername(username string) {
	s.dataSync.Lock()
	defer s.dataSync.Unlock()

	s.data["_username"] = username
}

func (s *Session) Set(key string, val interface{}) {
	s.dataSync.Lock()
	defer s.dataSync.Unlock()

	s.data[key] = val
}

func (s *Session) Get(key string) interface{} {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	return s.data[key]
}

func (s *Session) GetInt(key string) int {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok :=s.data[key]
	if !ok {
		return 0
	}
	return v.(int)
}

func (s *Session) GetString(key string) string {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok :=s.data[key]
	if !ok {
		return ""
	}
	return v.(string)
}

//sessions manager
type SessionManager struct {
	sessions map[string] *Session
	sessionsSync *sync.RWMutex
}

// create a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		sessionsSync: &sync.RWMutex{},
	}
}

//will get the session or create it if not found, cookie will be set with the new session
func (sm *SessionManager) Get(w http.ResponseWriter, r *http.Request) httpway.Session {
	sessionId:= ""

	cook, err := r.Cookie("_s")
	if err == nil {
		sessionId = cook.Value
	}

	sm.sessionsSync.RLock()
	s, found := sm.sessions[sessionId]
	sm.sessionsSync.RUnlock()

	if found {
		return s
	}

	return sm.newSession(w)
}

func (sm *SessionManager) Set(w http.ResponseWriter, r *http.Request, session httpway.Session) {
	sm.sessionsSync.Lock()
	sm.sessions[session.Id()] = session.(*Session)
	sm.sessionsSync.Unlock()
}

func (sm *SessionManager) newSession(w http.ResponseWriter) *Session {
	s := newSession()

	sm.sessionsSync.Lock()
	sm.sessions[s.id] = s
	sm.sessionsSync.Unlock()


	newCook := &http.Cookie{Name: "_s", Value: s.Id(), Path: "/"}
	http.SetCookie(w, newCook)

	return s
}
