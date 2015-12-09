package httpwaymid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/corneldamian/httpway"
)

//session
type Session struct {
	id           string
	data         map[string]interface{}
	dataSync     *sync.RWMutex
	creationTime time.Time
	accessTime   time.Time
}

func NewSession(id string) *Session {
	return &Session{
		id:           id,
		data:         make(map[string]interface{}),
		dataSync:     &sync.RWMutex{},
		creationTime: time.Now(),
		accessTime:   time.Now(),
	}
}

func (s *Session) Id() string {
	return s.id
}

func (s *Session) IsAuth() bool {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok := s.data["_isAuth"]
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

	v, ok := s.data["_username"]
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

	v, ok := s.data[key]
	if !ok {
		return 0
	}
	return v.(int)
}

func (s *Session) GetString(key string) string {
	s.dataSync.RLock()
	defer s.dataSync.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return ""
	}
	return v.(string)
}

//sessions manager
type SessionManager struct {
	sessions     map[string]*Session
	sessionsSync *sync.RWMutex
}

// create a new session manager
func NewSessionManager(timeout, expiration int, log httpway.Logger) *SessionManager {
	sm := &SessionManager{
		sessions:     make(map[string]*Session),
		sessionsSync: &sync.RWMutex{},
	}

	go sm.gc(time.Duration(timeout)*time.Second, time.Duration(expiration)*time.Second, log)

	return sm
}

func (sm *SessionManager) gc(timeout, expiration time.Duration, log httpway.Logger) {
	for {
		time.Sleep(60 * time.Second)

		deleteList := make([]string, 0, 100)
		sm.sessionsSync.RLock()
		for k, v := range sm.sessions {
			if expiration > 0 {
				if time.Since(v.creationTime) > expiration {
					deleteList = append(deleteList, k)
					continue
				}
			}

			if timeout > 0 {
				if time.Since(v.accessTime) > timeout {
					deleteList = append(deleteList, k)
				}
			}
		}
		sm.sessionsSync.RUnlock()

		if len(deleteList) > 0 {
			log.Debug("Going to delete %d sessions from memory", len(deleteList))

			sm.sessionsSync.Lock()
			for _, k := range deleteList {
				delete(sm.sessions, k)
			}
			sm.sessionsSync.Unlock()
		}
	}
}

//will get the session or create it if not found, cookie will be set with the new session
func (sm *SessionManager) Get(w http.ResponseWriter, r *http.Request, log httpway.Logger) httpway.Session {
	sessionId := ""

	cook, err := r.Cookie("_s")
	if err == nil {
		sessionId = cook.Value
	}

	return sm.GetById(sessionId, w, r, log)
}

func (sm *SessionManager) GetById(sessionId string, w http.ResponseWriter, r *http.Request, log httpway.Logger) httpway.Session {
	sm.sessionsSync.RLock()
	s, found := sm.sessions[sessionId]
	sm.sessionsSync.RUnlock()

	if found {
		s.accessTime = time.Now()

		log.Debug("Session %s found, created on: %s last access on: %s", s.id, s.creationTime, s.accessTime)
		return s
	}

	s = sm.newSession(w)

	log.Debug("New session %s created", s.id)

	return s
}

func (sm *SessionManager) Has(sessionId string) bool {
	sm.sessionsSync.RLock()
	defer sm.sessionsSync.RUnlock()

	_, ok := sm.sessions[sessionId]

	return ok
}

func (sm *SessionManager) Set(w http.ResponseWriter, r *http.Request, session httpway.Session, log httpway.Logger) {
	sm.sessionsSync.Lock()
	sm.sessions[session.Id()] = session.(*Session)
	sm.sessionsSync.Unlock()

	log.Debug("Session %s set", session.Id())
}

func (sm *SessionManager) newSession(w http.ResponseWriter) *Session {
	b := make([]byte, 32)
	n, err := rand.Read(b)
	id := ""

	if n != len(b) || err != nil {
		t := make([]byte, 32)
		binary.LittleEndian.PutUint64(t, uint64(time.Now().UnixNano()))
		hasher := md5.New()
		hasher.Write(t)
		id = hex.EncodeToString(hasher.Sum(nil))
	} else {
		id = hex.EncodeToString(b)
	}

	s := NewSession(id)

	sm.sessionsSync.Lock()
	sm.sessions[s.id] = s
	sm.sessionsSync.Unlock()

	newCook := &http.Cookie{Name: "_s", Value: s.Id(), Path: "/"}
	http.SetCookie(w, newCook)

	return s
}
